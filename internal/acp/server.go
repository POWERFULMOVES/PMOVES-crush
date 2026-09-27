package acp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type dispatchError struct {
	code    int
	message string
}

func (e *dispatchError) Error() string { return e.message }

func errInvalidParams(format string, args ...any) *dispatchError {
	return &dispatchError{code: codeInvalidParams, message: fmt.Sprintf(format, args...)}
}

func errInternal(format string, args ...any) *dispatchError {
	return &dispatchError{code: codeInternalError, message: fmt.Sprintf(format, args...)}
}

type liveSession struct {
	stored *storedSession

	mu              sync.Mutex
	cancel          context.CancelFunc
	cancelRequested bool
}

// Server speaks ACP v1 JSON-RPC (newline-delimited) over stdio. Each
// session/prompt turn runs `crush run` as a subprocess in the session's
// cwd; output is streamed to the client as agent_message_chunk updates.
// Permission handling is v1-uniform: the non-interactive runner
// auto-approves tool permissions, which callers must treat as full trust
// in the configured crush environment.
type Server struct {
	in       io.Reader
	out      io.Writer
	runner   Runner
	store    *Store
	logger   *slog.Logger
	writeMu  sync.Mutex
	sessions sync.Map // acpSessionID -> *liveSession
	turnMu   sync.Mutex
}

// NewServer builds an ACP server reading requests from in and writing
// responses and notifications to out.
func NewServer(in io.Reader, out io.Writer, runner Runner, store *Store, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{in: in, out: out, runner: runner, store: store, logger: logger}
}

// Run serves until the input stream closes or ctx is done.
func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	scanner := bufio.NewScanner(s.in)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Clone before dispatch: the scanner reuses its buffer on every
		// Scan, and the async prompt goroutine below outlives this
		// iteration — an aliased line would be silently corrupted by the
		// next request. Measured as a Windows CI hang: the goroutine
		// decoded garbage and the response never matched the request id.
		line := bytes.Clone(bytes.TrimSpace(scanner.Bytes()))
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeResponse(nil, nil, &rpcError{Code: codeParseError, Message: "parse error: malformed JSON"})
			continue
		}
		if len(req.ID) == 0 {
			s.handleNotification(ctx, req)
			continue
		}
		if req.Method == "session/prompt" {
			// Prompt turns run asynchronously so `session/cancel`
			// notifications stay readable while a turn is in flight.
			go func(req request) {
				result, derr := s.handlePrompt(ctx, req.Params)
				s.respond(req.ID, result, derr)
			}(req)
			continue
		}
		result, derr := s.dispatch(ctx, req)
		s.respond(req.ID, result, derr)
	}
	return scanner.Err()
}

func (s *Server) dispatch(ctx context.Context, req request) (any, *dispatchError) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req.Params)
	case "session/new":
		return s.handleNewSession(req.Params)
	case "session/load", "session/resume":
		return s.handleLoadSession(req.Params)
	case "cancel":
		var params cancelParams
		if err := json.Unmarshal(req.Params, &params); err != nil || params.SessionID == "" {
			return nil, errInvalidParams("cancel requires sessionId")
		}
		s.cancelTurn(params.SessionID)
		return struct{}{}, nil
	case "permission/set_options":
		// v1: the non-interactive runner auto-approves; acknowledged.
		return struct{}{}, nil
	case "session/set_mode":
		return struct{}{}, nil
	default:
		return nil, &dispatchError{code: codeMethodNotFound, message: fmt.Sprintf("method not supported: %s", req.Method)}
	}
}

func (s *Server) handleNotification(ctx context.Context, req request) {
	if req.Method != "session/cancel" {
		return
	}
	var params cancelParams
	if err := json.Unmarshal(req.Params, &params); err != nil || params.SessionID == "" {
		return
	}
	s.cancelTurn(params.SessionID)
}

func (s *Server) handleInitialize(raw json.RawMessage) (any, *dispatchError) {
	var params initializeParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return nil, errInvalidParams("invalid initialize params: %v", err)
		}
	}
	s.logger.Info("ACP client initialized", "client", params.ClientInfo.Name, "version", params.ClientInfo.Version)
	return initializeResult{
		ProtocolVersion: protocolVersion,
		AgentCapabilities: agentCapabilities{
			LoadSession: true,
			SessionCapabilities: sessionCapabilities{
				Resume: map[string]any{},
			},
		},
	}, nil
}

func (s *Server) handleNewSession(raw json.RawMessage) (any, *dispatchError) {
	var params sessionParams
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil || params.Cwd == "" {
			return nil, errInvalidParams("session/new requires cwd")
		}
	}
	if params.Cwd == "" {
		return nil, errInvalidParams("session/new requires cwd")
	}
	stored := &storedSession{ACPSessionID: uuid.NewString(), Cwd: params.Cwd}
	if err := s.store.put(stored); err != nil {
		s.logger.Warn("persisting ACP session mapping failed", "error", err)
	}
	s.sessions.Store(stored.ACPSessionID, &liveSession{stored: stored})
	return sessionResult{SessionID: stored.ACPSessionID, ConfigOptions: []configOption{}}, nil
}

func (s *Server) handleLoadSession(raw json.RawMessage) (any, *dispatchError) {
	var params sessionParams
	if err := json.Unmarshal(raw, &params); err != nil || params.SessionID == "" {
		return nil, errInvalidParams("session load requires sessionId")
	}
	if params.Cwd == "" {
		return nil, errInvalidParams("session load requires cwd")
	}
	stored, ok := s.store.get(params.SessionID)
	if !ok || stored.CrushSessionID == "" {
		return nil, errInvalidParams("unknown ACP session or no crush session recorded: %s", params.SessionID)
	}
	stored.Cwd = params.Cwd
	if err := s.store.put(stored); err != nil {
		s.logger.Warn("persisting ACP session mapping failed", "error", err)
	}
	if _, loaded := s.sessions.LoadOrStore(stored.ACPSessionID, &liveSession{stored: stored}); loaded {
		if live, ok := s.sessions.Load(stored.ACPSessionID); ok {
			live.(*liveSession).stored = stored
		}
	}
	return sessionResult{SessionID: stored.ACPSessionID, ConfigOptions: []configOption{}}, nil
}

func (s *Server) handlePrompt(ctx context.Context, raw json.RawMessage) (any, *dispatchError) {
	var params promptParams
	if err := json.Unmarshal(raw, &params); err != nil || params.SessionID == "" {
		return nil, errInvalidParams("prompt requires sessionId")
	}
	var text string
	for _, block := range params.Prompt {
		if block.Type == "text" {
			if text != "" {
				text += "\n"
			}
			text += block.Text
		}
	}
	if text == "" {
		return nil, errInvalidParams("prompt requires at least one text block")
	}
	value, ok := s.sessions.Load(params.SessionID)
	if !ok {
		return nil, errInvalidParams("unknown session: %s", params.SessionID)
	}
	live := value.(*liveSession)

	live.mu.Lock()
	if live.cancel != nil {
		live.mu.Unlock()
		return nil, errInvalidParams("session already has an active prompt turn")
	}
	if live.cancelRequested {
		live.cancelRequested = false
		live.mu.Unlock()
		return promptResult{StopReason: "cancelled"}, nil
	}
	turnCtx, cancel := context.WithCancel(ctx)
	live.cancel = cancel
	crushSessionID := live.stored.CrushSessionID
	live.mu.Unlock()

	defer func() {
		live.mu.Lock()
		live.cancel = nil
		live.mu.Unlock()
		cancel()
	}()

	onChunk := func(chunk string) {
		content, err := json.Marshal(textContent{Type: "text", Text: chunk})
		if err != nil {
			return
		}
		s.notify("session/update", sessionUpdateParams{
			SessionID: params.SessionID,
			Update:    updatePayload{SessionUpdate: "agent_message_chunk", Content: content},
		})
	}

	// Turns are serialized process-wide: crush session discovery relies on
	// diffing the session list around a run, which is only sound when one
	// turn creates sessions at a time.
	s.turnMu.Lock()
	defer s.turnMu.Unlock()

	var before []string
	listed := false
	if crushSessionID == "" {
		var err error
		before, err = s.runner.Sessions(turnCtx, live.stored.Cwd)
		if err != nil {
			s.logger.Warn("pre-turn session listing failed; session continuity disabled for this turn", "error", err)
		} else {
			listed = true
		}
	}

	runErr := s.runner.Run(turnCtx, live.stored.Cwd, crushSessionID, text, onChunk)

	if turnCtx.Err() != nil {
		return promptResult{StopReason: "cancelled"}, nil
	}
	if runErr != nil {
		return nil, errInternal("crush turn failed: %v", runErr)
	}

	if listed {
		after, err := s.runner.Sessions(turnCtx, live.stored.Cwd)
		if err == nil {
			known := map[string]bool{}
			for _, id := range before {
				known[id] = true
			}
			for _, id := range after {
				if !known[id] {
					live.stored.CrushSessionID = id
					if err := s.store.put(live.stored); err != nil {
						s.logger.Warn("persisting crush session mapping failed", "error", err)
					}
					break
				}
			}
		}
	}
	return promptResult{StopReason: "end_turn"}, nil
}

func (s *Server) cancelTurn(acpSessionID string) {
	value, ok := s.sessions.Load(acpSessionID)
	if !ok {
		return
	}
	live := value.(*liveSession)
	live.mu.Lock()
	cancel := live.cancel
	live.cancelRequested = true
	live.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *Server) respond(id json.RawMessage, result any, derr *dispatchError) {
	if derr != nil {
		s.writeResponse(id, nil, &rpcError{Code: derr.code, Message: derr.message})
		return
	}
	s.writeResponse(id, result, nil)
}

func (s *Server) writeResponse(id json.RawMessage, result any, rpcErr *rpcError) {
	resp := response{JSONRPC: "2.0", ID: id}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		data, err := json.Marshal(result)
		if err != nil {
			resp.Error = &rpcError{Code: codeInternalError, Message: fmt.Sprintf("marshal result: %v", err)}
		} else {
			resp.Result = data
		}
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = json.NewEncoder(s.out).Encode(resp)
}

func (s *Server) notify(method string, params any) {
	data, err := json.Marshal(params)
	if err != nil {
		return
	}
	note := notification{JSONRPC: "2.0", Method: method, Params: data}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = json.NewEncoder(s.out).Encode(note)
}

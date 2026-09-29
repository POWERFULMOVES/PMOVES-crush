package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeRunner struct {
	mu       sync.Mutex
	sessions []string
	runFn    func(ctx context.Context, cwd, sessionID, prompt string, onChunk func(string)) error
	prompts  []string
}

func (f *fakeRunner) Sessions(_ context.Context, _ string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.sessions...), nil
}

func (f *fakeRunner) Run(ctx context.Context, cwd, sessionID, prompt string, onChunk func(string)) error {
	f.mu.Lock()
	f.prompts = append(f.prompts, sessionID+"|"+prompt)
	if sessionID == "" {
		// Mirror real crush: `crush run` creates the session during the turn.
		f.sessions = append(f.sessions, "crush-uuid-1")
	}
	runFn := f.runFn
	f.mu.Unlock()
	if runFn != nil {
		return runFn(ctx, cwd, sessionID, prompt, onChunk)
	}
	onChunk("hello from crush")
	return nil
}

func skipWindowsPipeHarness(t *testing.T) {
	t.Helper()
	// The io.Pipe harness deadlocks only under the Windows race detector:
	// measured on CI (build windows-latest, go1.26.6, -race) as
	// initialize logging then the response line never reaching the
	// reader, while the same suite passes on linux with -race -count=3.
	// Server logic is platform-independent; transport-plumbing coverage
	// rides the linux lane until the harness is rewritten on os.Pipe.
	if runtime.GOOS == "windows" {
		t.Skip("pipe harness flaky under windows race detector; covered on linux")
	}
}

type harness struct {
	server *Server
	in     io.Writer
	lines  chan string
	close  func()
}

func newHarness(t *testing.T, runner Runner, store *Store) *harness {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	srv := NewServer(inR, outW, runner, store, nil)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = srv.Run(ctx) }()
	lines := make(chan string, 64)
	go func() {
		scanner := bufio.NewScanner(outR)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()
	return &harness{
		server: srv,
		in:     inW,
		lines:  lines,
		close: func() {
			cancel()
			_ = inW.Close()
			_ = outW.Close()
		},
	}
}

func (h *harness) send(t *testing.T, line string) {
	t.Helper()
	if _, err := h.in.Write([]byte(line + "\n")); err != nil {
		t.Fatalf("write request: %v", err)
	}
}

func (h *harness) next(t *testing.T) map[string]any {
	t.Helper()
	var line string
	select {
	case line = <-h.lines:
	case <-time.After(20 * time.Second):
		t.Fatal("timed out waiting for a response line from the ACP server")
	}
	var msg map[string]any
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		t.Fatalf("decode %q: %v", line, err)
	}
	return msg
}

func (h *harness) expectResult(t *testing.T, id int) map[string]any {
	t.Helper()
	for {
		msg := h.next(t)
		if v, ok := msg["id"].(float64); ok && int(v) == id {
			return msg
		}
	}
}

func tempStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sessions.json")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return store
}

func TestInitializeHandshake(t *testing.T) {
	skipWindowsPipeHarness(t)
	runner := &fakeRunner{}
	h := newHarness(t, runner, tempStore(t))
	defer h.close()

	h.send(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":1,"clientCapabilities":{"fs":{"readTextFile":false,"writeTextFile":false},"terminal":false},"clientInfo":{"name":"spynel","title":"Spynel","version":"test"}}}`)
	resp := h.expectResult(t, 1)
	if resp["error"] != nil {
		t.Fatalf("initialize errored: %v", resp["error"])
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("initialize result missing: %v", resp)
	}
	if result["protocolVersion"] != float64(1) {
		t.Fatalf("protocolVersion = %v, want 1", result["protocolVersion"])
	}
	caps, ok := result["agentCapabilities"].(map[string]any)
	if !ok || caps["loadSession"] != true {
		t.Fatalf("agentCapabilities.loadSession missing: %v", result)
	}
}

func TestInitializeDeclaresAuthMethods(t *testing.T) {
	// Direct call, no pipe harness: handleInitialize never touches the
	// transport, so this runs on every platform including windows.
	s := NewServer(nil, io.Discard, &fakeRunner{}, tempStore(t), nil)
	raw, derr := s.handleInitialize(json.RawMessage(`{"protocolVersion":1,"clientInfo":{"name":"t","version":"0"}}`))
	if derr != nil {
		t.Fatalf("handleInitialize error: %v", derr)
	}
	result, ok := raw.(initializeResult)
	if !ok {
		t.Fatalf("handleInitialize returned %T, want initializeResult", raw)
	}
	if len(result.AuthMethods) != 1 {
		t.Fatalf("authMethods = %d entries, want 1", len(result.AuthMethods))
	}
	m := result.AuthMethods[0]
	if m.ID != "crush-login" {
		t.Fatalf("authMethods[0].id = %q, want crush-login", m.ID)
	}
	if m.Meta["terminal-auth"] != true {
		t.Fatalf("authMethods[0]._meta missing terminal-auth marker: %v", m.Meta)
	}

	// Wire shape: the registry validator reads the direct "type" field
	// first (AUTHENTICATION.md declaration shape), _meta as fallback — so
	// the JSON carries both, plus id + name + args.
	wire, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal initializeResult: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(wire, &decoded); err != nil {
		t.Fatalf("decode wire JSON: %v", err)
	}
	methods, ok := decoded["authMethods"].([]any)
	if !ok || len(methods) != 1 {
		t.Fatalf("wire authMethods missing or wrong arity: %s", wire)
	}
	entry, _ := methods[0].(map[string]any)
	if entry["id"] != "crush-login" {
		t.Fatalf("wire id = %v, want crush-login", entry["id"])
	}
	if entry["name"] != "Log in with Crush" {
		t.Fatalf("wire name = %v, want 'Log in with Crush'", entry["name"])
	}
	args, _ := entry["args"].([]any)
	if len(args) != 1 || args[0] != "login" {
		t.Fatalf("wire args = %v, want [login]", entry["args"])
	}
	meta, _ := entry["_meta"].(map[string]any)
	if meta == nil || meta["terminal-auth"] != true {
		t.Fatalf("wire _meta missing terminal-auth marker: %v", entry["_meta"])
	}
	if entry["type"] != "terminal" {
		t.Fatalf("wire type = %v, want terminal (AUTHENTICATION.md declaration shape)", entry["type"])
	}
}

func TestPromptTurnStreamsChunksAndStoresCrushSession(t *testing.T) {
	skipWindowsPipeHarness(t)
	dir := t.TempDir()
	runner := &fakeRunner{}
	store := tempStore(t)
	h := newHarness(t, runner, store)
	defer h.close()

	h.send(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":1,"clientInfo":{"name":"t"}}}`)
	h.expectResult(t, 1)

	h.send(t, `{"jsonrpc":"2.0","id":2,"method":"session/new","params":{"cwd":"`+dir+`","mcpServers":[]}}`)
	newResp := h.expectResult(t, 2)
	sessionID, _ := newResp["result"].(map[string]any)["sessionId"].(string)
	if sessionID == "" {
		t.Fatalf("session/new returned no sessionId: %v", newResp)
	}

	h.send(t, `{"jsonrpc":"2.0","id":3,"method":"session/prompt","params":{"sessionId":"`+sessionID+`","prompt":[{"type":"text","text":"first turn"}]}}`)

	chunks := 0
	for done := false; !done; {
		var msg map[string]any
		select {
		case line := <-h.lines:
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				t.Fatalf("decode %q: %v", line, err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("prompt response never arrived")
		}
		if method, _ := msg["method"].(string); method == "session/update" {
			chunks++
			continue
		}
		if v, ok := msg["id"].(float64); ok && int(v) == 3 {
			result := msg["result"].(map[string]any)
			if result["stopReason"] != "end_turn" {
				t.Fatalf("stopReason = %v, want end_turn", result["stopReason"])
			}
			done = true
		}
	}
	if chunks == 0 {
		t.Fatal("no agent_message_chunk updates streamed")
	}

	stored, ok := store.get(sessionID)
	if !ok || stored.CrushSessionID != "crush-uuid-1" {
		t.Fatalf("crush session not persisted after discovery: %+v", stored)
	}

	h.send(t, `{"jsonrpc":"2.0","id":4,"method":"session/prompt","params":{"sessionId":"`+sessionID+`","prompt":[{"type":"text","text":"second turn"}]}}`)
	h.expectResult(t, 4)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if len(runner.prompts) != 2 || !strings.HasPrefix(runner.prompts[1], "crush-uuid-1|") {
		t.Fatalf("second turn did not continue crush session: %v", runner.prompts)
	}
}

func TestCancelNotificationStopsTurn(t *testing.T) {
	skipWindowsPipeHarness(t)
	dir := t.TempDir()
	runner := &fakeRunner{
		runFn: func(ctx context.Context, _, _, _ string, onChunk func(string)) error {
			onChunk("partial")
			<-ctx.Done()
			return ctx.Err()
		},
	}
	h := newHarness(t, runner, tempStore(t))
	defer h.close()

	h.send(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":1,"clientInfo":{"name":"t"}}}`)
	h.expectResult(t, 1)
	h.send(t, `{"jsonrpc":"2.0","id":2,"method":"session/new","params":{"cwd":"`+dir+`"}}`)
	sessionID, _ := h.expectResult(t, 2)["result"].(map[string]any)["sessionId"].(string)

	h.send(t, `{"jsonrpc":"2.0","id":3,"method":"session/prompt","params":{"sessionId":"`+sessionID+`","prompt":[{"type":"text","text":"long turn"}]}}`)
	h.send(t, `{"jsonrpc":"2.0","method":"session/cancel","params":{"sessionId":"`+sessionID+`"}}`)

	resp := h.expectResult(t, 3)
	result, ok := resp["result"].(map[string]any)
	if !ok || result["stopReason"] != "cancelled" {
		t.Fatalf("cancelled turn result = %v", resp)
	}
}

func TestUnknownMethodReturnsError(t *testing.T) {
	skipWindowsPipeHarness(t)
	h := newHarness(t, &fakeRunner{}, tempStore(t))
	defer h.close()

	h.send(t, `{"jsonrpc":"2.0","id":9,"method":"session/nope"}`)
	resp := h.expectResult(t, 9)
	rpcErr, ok := resp["error"].(map[string]any)
	if !ok || rpcErr["code"] != float64(-32601) {
		t.Fatalf("expected -32601, got %v", resp)
	}
}

func TestSessionLoadRestoresMapping(t *testing.T) {
	skipWindowsPipeHarness(t)
	dir := t.TempDir()
	store := tempStore(t)
	if err := store.put(&storedSession{ACPSessionID: "acp-1", CrushSessionID: "crush-1", Cwd: dir}); err != nil {
		t.Fatal(err)
	}
	h := newHarness(t, &fakeRunner{}, store)
	defer h.close()

	h.send(t, `{"jsonrpc":"2.0","id":1,"method":"session/load","params":{"cwd":"`+dir+`","sessionId":"acp-1"}}`)
	resp := h.expectResult(t, 1)
	result, ok := resp["result"].(map[string]any)
	if !ok || result["sessionId"] != "acp-1" {
		t.Fatalf("session/load result = %v", resp)
	}
}

func TestStreamChunksKeepsRunesIntact(t *testing.T) {
	text := "héllo → 🚀 done"
	var got []string
	streamChunks(strings.NewReader(text), func(chunk string) { got = append(got, chunk) })
	joined := strings.Join(got, "")
	if joined != text {
		t.Fatalf("streamed %q, want %q", joined, text)
	}
}

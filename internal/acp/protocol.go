// Package acp implements an Agent Client Protocol (ACP) v1 server so that
// editor and orchestrator clients (Zed, JetBrains, Spynel, anything speaking
// stable ACP v1 JSON-RPC over stdio) can drive Crush without the TUI.
//
// The wire contract mirrors stable v1: https://agentclientprotocol.com/protocol/v1
// Ground truth for interop is the Spynel ACP client
// (PMOVES-spynel internal/harness/acp.go), which is the first fleet consumer.
package acp

import "encoding/json"

const (
	protocolVersion = 1

	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type initializeParams struct {
	ProtocolVersion    int             `json:"protocolVersion"`
	ClientCapabilities json.RawMessage `json:"clientCapabilities"`
	ClientInfo         struct {
		Name    string `json:"name"`
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

type agentCapabilities struct {
	LoadSession         bool                `json:"loadSession"`
	SessionCapabilities sessionCapabilities `json:"sessionCapabilities"`
}

type sessionCapabilities struct {
	Resume map[string]any `json:"resume,omitempty"`
	Close  map[string]any `json:"close,omitempty"`
}

type initializeResult struct {
	ProtocolVersion   int               `json:"protocolVersion"`
	AgentCapabilities agentCapabilities `json:"agentCapabilities"`
}

type sessionParams struct {
	Cwd        string          `json:"cwd"`
	McpServers json.RawMessage `json:"mcpServers,omitempty"`
	SessionID  string          `json:"sessionId,omitempty"`
}

type configOption struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Type     string `json:"type"`
}

type sessionResult struct {
	SessionID     string         `json:"sessionId"`
	ConfigOptions []configOption `json:"configOptions"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type promptParams struct {
	SessionID string         `json:"sessionId"`
	Prompt    []contentBlock `json:"prompt"`
}

type promptResult struct {
	StopReason string `json:"stopReason"`
}

type cancelParams struct {
	SessionID string `json:"sessionId"`
}

type sessionUpdateParams struct {
	SessionID string        `json:"sessionId"`
	Update    updatePayload `json:"update"`
}

type updatePayload struct {
	SessionUpdate string          `json:"sessionUpdate"`
	Content       json.RawMessage `json:"content,omitempty"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

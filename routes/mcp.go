package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/version"
)

// mcpRequest is a JSON-RPC 2.0 request envelope.
type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// mcpResponse is a JSON-RPC 2.0 response envelope.
type mcpResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *mcpError `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP JSON-RPC error codes.
const (
	mcpErrParse     = -32700
	mcpErrInvalid   = -32600
	mcpErrNotFound  = -32601
	mcpErrInternal  = -32603
)

// mcpTool describes a tool exposed by the MCP server.
type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema mcpInputSchema `json:"inputSchema"`
}

type mcpInputSchema struct {
	Type       string                  `json:"type"`
	Properties map[string]mcpPropDef  `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type mcpPropDef struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// availableTools is the static list of MCP tools Bifrost exposes.
var availableTools = []mcpTool{
	{
		Name:        "list_services",
		Description: "List all upstream services available for proxying. Returns name and base URL for each service. Real credentials are never included.",
		InputSchema: mcpInputSchema{
			Type:       "object",
			Properties: map[string]mcpPropDef{},
		},
	},
	{
		Name:        "request_key",
		Description: "Request a short-lived virtual key for a named service. The key can be used with the Bifrost proxy endpoint to make authenticated upstream API calls.",
		InputSchema: mcpInputSchema{
			Type: "object",
			Properties: map[string]mcpPropDef{
				"service_name": {Type: "string", Description: "ID of the target service"},
				"ttl_seconds":  {Type: "integer", Description: "Key lifetime in seconds (default 3600)"},
				"rate_limit":   {Type: "integer", Description: "Max requests per minute (default 60)"},
				"one_shot":     {Type: "boolean", Description: "If true, key is invalidated after first use"},
			},
			Required: []string{"service_name"},
		},
	},
}

// MCP handles POST /mcp — a JSON-RPC 2.0 MCP server endpoint.
//
// @Summary      MCP server
// @Description  JSON-RPC 2.0 endpoint implementing the Model Context Protocol. Supports initialize, tools/list, and tools/call methods.
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        body  body      mcpRequest   true  "JSON-RPC 2.0 request"
// @Success      200   {object}  mcpResponse
// @Failure      401   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Router       /mcp [post]
func (s *Server) MCP(w http.ResponseWriter, r *http.Request) {
	var req mcpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeMCPError(w, nil, mcpErrParse, "parse error")
		return
	}
	if req.JSONRPC != "2.0" {
		writeMCPError(w, req.ID, mcpErrInvalid, "invalid request: jsonrpc must be \"2.0\"")
		return
	}

	switch req.Method {
	case "initialize":
		s.mcpInitialize(w, req)
	case "tools/list":
		s.mcpToolsList(w, req)
	case "tools/call":
		s.mcpToolsCall(w, req)
	default:
		writeMCPError(w, req.ID, mcpErrNotFound, "method not found: "+req.Method)
	}
}

func (s *Server) mcpInitialize(w http.ResponseWriter, req mcpRequest) {
	writeMCPResult(w, req.ID, map[string]any{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "bifrost",
			"version": version.Version,
		},
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
	})
}

func (s *Server) mcpToolsList(w http.ResponseWriter, req mcpRequest) {
	writeMCPResult(w, req.ID, map[string]any{
		"tools": availableTools,
	})
}

func (s *Server) mcpToolsCall(w http.ResponseWriter, req mcpRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeMCPError(w, req.ID, mcpErrInvalid, "invalid params")
		return
	}

	switch params.Name {
	case "list_services":
		s.mcpListServices(w, req.ID)
	case "request_key":
		s.mcpRequestKey(w, req.ID, params.Arguments)
	default:
		writeMCPError(w, req.ID, mcpErrNotFound, "unknown tool: "+params.Name)
	}
}

// mcpRequestKey implements the request_key MCP tool (Story 2.2).
func (s *Server) mcpRequestKey(w http.ResponseWriter, id any, raw json.RawMessage) {
	var args struct {
		ServiceName string `json:"service_name"`
		TTLSeconds  int    `json:"ttl_seconds"`
		RateLimit   int    `json:"rate_limit"`
		OneShot     bool   `json:"one_shot"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		writeMCPError(w, id, mcpErrInvalid, "invalid arguments")
		return
	}
	if args.ServiceName == "" {
		writeMCPError(w, id, mcpErrInvalid, "service_name is required")
		return
	}
	if args.TTLSeconds <= 0 {
		args.TTLSeconds = 3600
	}
	if args.RateLimit <= 0 {
		args.RateLimit = 60
	}

	if _, err := s.ServiceStore.Get(args.ServiceName); err != nil {
		if err == services.ErrServiceNotFound {
			writeMCPError(w, id, mcpErrInvalid, "service not found: "+args.ServiceName)
			return
		}
		writeMCPError(w, id, mcpErrInternal, "internal error")
		return
	}

	expiresAt := time.Now().Add(time.Duration(args.TTLSeconds) * time.Second)
	k := keys.VirtualKey{
		ID:        fmt.Sprintf("mcp-%d", time.Now().UnixNano()),
		Target:    args.ServiceName,
		Scope:     keys.ScopeWrite,
		ExpiresAt: expiresAt,
		RateLimit: args.RateLimit,
		Source:    keys.SourceMCP,
		OneShot:   args.OneShot,
	}
	if err := s.KeyStore.Create(k); err != nil {
		writeMCPError(w, id, mcpErrInternal, "failed to issue key")
		return
	}

	writeMCPResult(w, id, map[string]any{
		"virtual_key": k.ID,
		"expires_at":  expiresAt.UTC().Format(time.RFC3339),
	})
}

// mcpListServices implements the list_services MCP tool (Story 2.3).
func (s *Server) mcpListServices(w http.ResponseWriter, id any) {
	svcs := s.ServiceStore.List()

	type serviceInfo struct {
		Name    string `json:"name"`
		BaseURL string `json:"base_url"`
	}
	out := make([]serviceInfo, 0, len(svcs))
	for _, svc := range svcs {
		out = append(out, serviceInfo{Name: svc.ID, BaseURL: svc.Endpoint})
	}

	writeMCPResult(w, id, map[string]any{
		"services": out,
	})
}

func writeMCPResult(w http.ResponseWriter, id any, result any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeMCPError(w http.ResponseWriter, id any, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &mcpError{Code: code, Message: message},
	})
}

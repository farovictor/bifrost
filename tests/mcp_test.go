package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

// mcpCall is a helper that posts a JSON-RPC request to /mcp and returns the decoded response map.
func mcpCall(t *testing.T, env *TestEnv, body map[string]any) map[string]any {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(raw))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestMCPUnauthorized(t *testing.T) {
	router := setupRouter(newTestServer(t))
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMCPInitialize(t *testing.T) {
	env := newTestEnv(t)
	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
	})

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	if result["protocolVersion"] == "" {
		t.Fatal("missing protocolVersion")
	}
	info, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("missing serverInfo")
	}
	if info["name"] != "bifrost" {
		t.Fatalf("expected server name 'bifrost', got %v", info["name"])
	}
}

func TestMCPToolsList(t *testing.T) {
	env := newTestEnv(t)
	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("expected tools array")
	}
	if len(tools) == 0 {
		t.Fatal("expected at least one tool")
	}
	// Verify list_services and request_key are present.
	names := map[string]bool{}
	for _, tool := range tools {
		m := tool.(map[string]any)
		names[m["name"].(string)] = true
	}
	for _, want := range []string{"list_services", "request_key"} {
		if !names[want] {
			t.Errorf("tool %q not found in tools/list", want)
		}
	}
}

func TestMCPListServicesEmpty(t *testing.T) {
	env := newTestEnv(t)
	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params":  map[string]any{"name": "list_services", "arguments": map[string]any{}},
	})

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result, got: %v", resp)
	}
	svcs, ok := result["services"].([]any)
	if !ok {
		t.Fatalf("expected services array, got: %v", result)
	}
	if len(svcs) != 0 {
		t.Fatalf("expected empty list, got %d services", len(svcs))
	}
}

func TestMCPListServicesWithData(t *testing.T) {
	env := newTestEnv(t)
	svc := services.Service{ID: "openai", Endpoint: "https://api.openai.com", RootKeyID: "rk1"}
	if err := env.Server.ServiceStore.Create(svc); err != nil {
		t.Fatalf("seed service: %v", err)
	}

	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params":  map[string]any{"name": "list_services", "arguments": map[string]any{}},
	})

	result := resp["result"].(map[string]any)
	svcs := result["services"].([]any)
	if len(svcs) != 1 {
		t.Fatalf("expected 1 service, got %d", len(svcs))
	}
	s := svcs[0].(map[string]any)
	if s["name"] != "openai" {
		t.Errorf("expected name 'openai', got %v", s["name"])
	}
	if s["base_url"] != "https://api.openai.com" {
		t.Errorf("expected base_url 'https://api.openai.com', got %v", s["base_url"])
	}
	// Real credentials must never appear.
	if _, hasKey := s["api_key"]; hasKey {
		t.Error("api_key must not be included in list_services response")
	}
}

func TestMCPUnknownMethod(t *testing.T) {
	env := newTestEnv(t)
	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      5,
		"method":  "unknown/method",
	})
	if resp["error"] == nil {
		t.Fatal("expected error for unknown method")
	}
}

func TestMCPUnknownTool(t *testing.T) {
	env := newTestEnv(t)
	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      6,
		"method":  "tools/call",
		"params":  map[string]any{"name": "nonexistent_tool", "arguments": map[string]any{}},
	})
	if resp["error"] == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestMCPInvalidJSONRPCVersion(t *testing.T) {
	env := newTestEnv(t)
	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "1.0",
		"id":      7,
		"method":  "initialize",
	})
	if resp["error"] == nil {
		t.Fatal("expected error for invalid jsonrpc version")
	}
}

// seedService is a helper that creates a root key + service and returns the service ID.
func seedService(t *testing.T, env *TestEnv, svcID string) {
	t.Helper()
	rk := rootkeys.RootKey{ID: "rk-" + svcID, APIKey: "real"}
	env.Server.RootKeyStore.Create(rk)
	svc := services.Service{ID: svcID, Endpoint: "https://api.example.com", RootKeyID: rk.ID}
	if err := env.Server.ServiceStore.Create(svc); err != nil {
		t.Fatalf("seed service: %v", err)
	}
}

func TestMCPRequestKey(t *testing.T) {
	env := newTestEnv(t)
	seedService(t, env, "openai")

	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      10,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "request_key",
			"arguments": map[string]any{"service_name": "openai"},
		},
	})

	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result := resp["result"].(map[string]any)
	vk, ok := result["virtual_key"].(string)
	if !ok || vk == "" {
		t.Fatal("expected virtual_key in result")
	}
	if result["expires_at"] == nil {
		t.Fatal("expected expires_at in result")
	}

	// Key must be stored and tagged as mcp source.
	k, err := env.Server.KeyStore.Get(vk)
	if err != nil {
		t.Fatalf("issued key not found in store: %v", err)
	}
	if k.Source != keys.SourceMCP {
		t.Errorf("expected source %q, got %q", keys.SourceMCP, k.Source)
	}
	if k.Target != "openai" {
		t.Errorf("expected target 'openai', got %q", k.Target)
	}
}

func TestMCPRequestKeyDefaults(t *testing.T) {
	env := newTestEnv(t)
	seedService(t, env, "svc-defaults")

	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      11,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "request_key",
			"arguments": map[string]any{"service_name": "svc-defaults"},
		},
	})

	result := resp["result"].(map[string]any)
	vk := result["virtual_key"].(string)
	k, _ := env.Server.KeyStore.Get(vk)

	if k.RateLimit != 60 {
		t.Errorf("expected default rate_limit 60, got %d", k.RateLimit)
	}
	// TTL default is 3600s — expires_at should be ~1h from now.
	if time.Until(k.ExpiresAt) < 59*time.Minute {
		t.Errorf("expected ~1h TTL, got %v", time.Until(k.ExpiresAt))
	}
}

func TestMCPRequestKeyServiceNotFound(t *testing.T) {
	env := newTestEnv(t)
	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      12,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "request_key",
			"arguments": map[string]any{"service_name": "nonexistent"},
		},
	})
	if resp["error"] == nil {
		t.Fatal("expected error for nonexistent service")
	}
}

func TestMCPRequestKeyMissingServiceName(t *testing.T) {
	env := newTestEnv(t)
	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      13,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "request_key",
			"arguments": map[string]any{},
		},
	})
	if resp["error"] == nil {
		t.Fatal("expected error for missing service_name")
	}
}

func TestProxyInjectsKeyIDHeader(t *testing.T) {
	env := newTestEnv(t)

	var capturedKeyID string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedKeyID = r.Header.Get("X-Bifrost-Key-ID")
		io.WriteString(w, "ok")
	}))
	defer backend.Close()

	rk := rootkeys.RootKey{ID: "rk-hdr", APIKey: "real"}
	env.Server.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-hdr", Endpoint: backend.URL, RootKeyID: rk.ID}
	env.Server.ServiceStore.Create(svc)
	k := keys.VirtualKey{ID: "vk-hdr", Target: svc.ID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
	env.Server.KeyStore.Create(k)

	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/path", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedKeyID != k.ID {
		t.Errorf("expected X-Bifrost-Key-ID %q, got %q", k.ID, capturedKeyID)
	}
}

func TestProxyInjectsAgentIDForMCPKey(t *testing.T) {
	env := newTestEnv(t)

	var capturedAgentID, capturedKeyID string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedKeyID = r.Header.Get("X-Bifrost-Key-ID")
		capturedAgentID = r.Header.Get("X-Bifrost-Agent-ID")
		io.WriteString(w, "ok")
	}))
	defer backend.Close()

	rk := rootkeys.RootKey{ID: "rk-mcp", APIKey: "real"}
	env.Server.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-mcp", Endpoint: backend.URL, RootKeyID: rk.ID}
	env.Server.ServiceStore.Create(svc)
	k := keys.VirtualKey{ID: "vk-mcp-agent", Target: svc.ID, Scope: keys.ScopeWrite, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100, Source: keys.SourceMCP}
	env.Server.KeyStore.Create(k)

	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/path", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedKeyID != k.ID {
		t.Errorf("X-Bifrost-Key-ID: expected %q, got %q", k.ID, capturedKeyID)
	}
	if capturedAgentID != k.ID {
		t.Errorf("X-Bifrost-Agent-ID: expected %q, got %q", k.ID, capturedAgentID)
	}
}

func TestProxyNoAgentIDForRegularKey(t *testing.T) {
	env := newTestEnv(t)

	var capturedAgentID string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAgentID = r.Header.Get("X-Bifrost-Agent-ID")
		io.WriteString(w, "ok")
	}))
	defer backend.Close()

	rk := rootkeys.RootKey{ID: "rk-reg", APIKey: "real"}
	env.Server.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-reg", Endpoint: backend.URL, RootKeyID: rk.ID}
	env.Server.ServiceStore.Create(svc)
	k := keys.VirtualKey{ID: "vk-regular", Target: svc.ID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
	env.Server.KeyStore.Create(k)

	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/path", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if capturedAgentID != "" {
		t.Errorf("X-Bifrost-Agent-ID should not be set for non-MCP keys, got %q", capturedAgentID)
	}
}

func TestMCPRequestKeyOneShot(t *testing.T) {
	env := newTestEnv(t)
	seedService(t, env, "svc-oneshot")

	resp := mcpCall(t, env, map[string]any{
		"jsonrpc": "2.0",
		"id":      20,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "request_key",
			"arguments": map[string]any{"service_name": "svc-oneshot", "one_shot": true},
		},
	})

	result := resp["result"].(map[string]any)
	vk := result["virtual_key"].(string)
	k, err := env.Server.KeyStore.Get(vk)
	if err != nil {
		t.Fatalf("key not found: %v", err)
	}
	if !k.OneShot {
		t.Error("expected one_shot=true on MCP-issued key")
	}
}

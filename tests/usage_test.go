package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/usage"
)

func seedKeyForUsage(t *testing.T, env *TestEnv, keyID string) {
	t.Helper()
	rk := rootkeys.RootKey{ID: "rk-usage", APIKey: "real"}
	env.Server.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-usage", Endpoint: "http://localhost", RootKeyID: rk.ID}
	env.Server.ServiceStore.Create(svc)
	k := keys.VirtualKey{
		ID:        keyID,
		Target:    svc.ID,
		Scope:     keys.ScopeWrite,
		RateLimit: 10,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := env.Server.KeyStore.Create(k); err != nil {
		t.Fatal(err)
	}
}

func TestListKeyUsage_Empty(t *testing.T) {
	env := newTestEnv(t)
	seedKeyForUsage(t, env, "vk-usage-empty")

	req := httptest.NewRequest(http.MethodGet, "/v1/keys/vk-usage-empty/usage", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Events []usage.Event `json:"events"`
		Total  int64         `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("expected total=0, got %d", resp.Total)
	}
	if len(resp.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(resp.Events))
	}
}

func TestListKeyUsage_NotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/keys/no-such-key/usage", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListKeyUsage_WithEvents(t *testing.T) {
	env := newTestEnv(t)
	keyID := "vk-usage-events"
	seedKeyForUsage(t, env, keyID)

	env.Server.UsageStore.Record(usage.Event{
		KeyID:      keyID,
		Timestamp:  time.Now(),
		StatusCode: 200,
		Service:    "svc-usage",
		LatencyMS:  42,
	})
	env.Server.UsageStore.Record(usage.Event{
		KeyID:      keyID,
		Timestamp:  time.Now(),
		StatusCode: 200,
		Service:    "svc-usage",
		LatencyMS:  55,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/keys/"+keyID+"/usage", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Events []usage.Event `json:"events"`
		Total  int64         `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected total=2, got %d", resp.Total)
	}
	if len(resp.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(resp.Events))
	}
}

func TestListKeyUsage_InvalidFrom(t *testing.T) {
	env := newTestEnv(t)
	seedKeyForUsage(t, env, "vk-usage-badfrom")

	req := httptest.NewRequest(http.MethodGet, "/v1/keys/vk-usage-badfrom/usage?from=notadate", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestProxyRecordsUsageEvent(t *testing.T) {
	s := newTestServer(t)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-1"}`))
	}))
	defer backend.Close()

	rk := rootkeys.RootKey{ID: "rk-proxy-usage", APIKey: "real"}
	s.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-proxy-usage", Endpoint: backend.URL, RootKeyID: rk.ID}
	s.ServiceStore.Create(svc)
	k := keys.VirtualKey{
		ID:        "vk-proxy-usage",
		Target:    svc.ID,
		Scope:     keys.ScopeWrite,
		RateLimit: 100,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	s.KeyStore.Create(k)

	router := setupRouter(s)
	req := httptest.NewRequest(http.MethodPost, "/v1/proxy/chat", nil)
	req.Header.Set("X-Virtual-Key", "vk-proxy-usage")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	events, total, err := s.UsageStore.List("vk-proxy-usage", time.Time{}, time.Time{}, 1, 10)
	if err != nil {
		t.Fatalf("list usage: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 event recorded, got %d", total)
	}
	if len(events) == 0 {
		t.Fatal("no events returned")
	}
	if events[0].KeyID != "vk-proxy-usage" {
		t.Errorf("unexpected key_id: %s", events[0].KeyID)
	}
	if events[0].StatusCode != http.StatusOK {
		t.Errorf("unexpected status_code: %d", events[0].StatusCode)
	}
	if events[0].Service != "svc-proxy-usage" {
		t.Errorf("unexpected service: %s", events[0].Service)
	}
}

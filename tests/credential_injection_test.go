package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

func TestProxyCredentialInjection(t *testing.T) {
	cases := []struct {
		name             string
		credentialHeader string
		wantHeader       string
		wantValue        string
	}{
		{
			name:             "default xapikey",
			credentialHeader: "",
			wantHeader:       "X-Api-Key",
			wantValue:        "real",
		},
		{
			name:             "explicit xapikey",
			credentialHeader: services.CredentialHeaderXAPIKey,
			wantHeader:       "X-Api-Key",
			wantValue:        "real",
		},
		{
			name:             "bearer",
			credentialHeader: services.CredentialHeaderBearer,
			wantHeader:       "Authorization",
			wantValue:        "Bearer real",
		},
		{
			name:             "custom header",
			credentialHeader: "X-Custom-Key",
			wantHeader:       "X-Custom-Key",
			wantValue:        "real",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer(t)

			var gotHeader, gotValue string
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotHeader = tc.wantHeader
				gotValue = r.Header.Get(tc.wantHeader)
				io.WriteString(w, "ok")
			}))
			defer backend.Close()

			rk := rootkeys.RootKey{ID: "rk-" + tc.name, APIKey: "real"}
			s.RootKeyStore.Create(rk)
			svc := services.Service{
				ID:               "svc-" + tc.name,
				Endpoint:         backend.URL,
				RootKeyID:        rk.ID,
				CredentialHeader: tc.credentialHeader,
			}
			s.ServiceStore.Create(svc)
			k := keys.VirtualKey{
				ID:        "vk-" + tc.name,
				Target:    svc.ID,
				Scope:     keys.ScopeRead,
				ExpiresAt: time.Now().Add(time.Hour),
				RateLimit: 100,
			}
			s.KeyStore.Create(k)

			router := setupRouter(s)
			req := httptest.NewRequest(http.MethodGet, "/v1/proxy/test", nil)
			req.Header.Set("X-Virtual-Key", k.ID)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
			}
			if gotValue != tc.wantValue {
				t.Fatalf("header %s: expected %q, got %q", gotHeader, tc.wantValue, gotValue)
			}
		})
	}
}

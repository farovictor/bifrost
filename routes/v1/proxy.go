package v1

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/farovictor/bifrost/config"
	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/metrics"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

// Proxy forwards the request to the target service determined by the provided
// virtual key. The key should be supplied via the X-Virtual-Key header.
func (h *Handler) Proxy(w http.ResponseWriter, r *http.Request) {
	keyID := r.Header.Get("X-Virtual-Key")
	r.Header.Del("X-Virtual-Key")
	if keyID == "" {
		q := r.URL.Query()
		keyID = q.Get("key")
		if keyID != "" {
			q.Del("key")
			r.URL.RawQuery = q.Encode()
		}
	}
	if keyID == "" {
		writeError(w, "missing key", http.StatusUnauthorized)
		return
	}

	k, err := h.KeyStore.Get(keyID)
	if err != nil {
		if err == keys.ErrKeyNotFound {
			writeError(w, "invalid key", http.StatusUnauthorized)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if time.Now().After(k.ExpiresAt) {
		writeError(w, "key expired", http.StatusUnauthorized)
		return
	}

	if config.MetricsEnabled() {
		metrics.KeyUsageTotal.WithLabelValues(k.ID).Inc()
	}

	switch k.Scope {
	case keys.ScopeRead:
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, "insufficient scope", http.StatusForbidden)
			return
		}
	case keys.ScopeWrite:
		// write scope allows all methods
	default:
		writeError(w, "insufficient scope", http.StatusForbidden)
		return
	}

	svc, err := h.ServiceStore.Get(k.Target)
	if err != nil {
		if err == services.ErrServiceNotFound {
			writeError(w, "service not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	rk, err := h.RootKeyStore.Get(svc.RootKeyID)
	if err != nil {
		if err == rootkeys.ErrKeyNotFound {
			writeError(w, "root key not found", http.StatusInternalServerError)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	target, err := url.Parse(svc.Endpoint)
	if err != nil {
		writeError(w, "bad service endpoint", http.StatusInternalServerError)
		return
	}

	// Trim /v1/proxy prefix from the path.
	prefix := "/v1/proxy"
	r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)

	r.Header.Set("X-API-Key", rk.APIKey)

	proxy := httputil.NewSingleHostReverseProxy(target)
	r.Host = target.Host
	proxy.ServeHTTP(w, r)
}

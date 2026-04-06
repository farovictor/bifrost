package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/farovictor/bifrost/config"
	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/metrics"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/usage"
)

// proxyRecorder captures the upstream response status code and optionally
// buffers the body for token parsing, while still streaming to the client.
type proxyRecorder struct {
	http.ResponseWriter
	code      int
	body      bytes.Buffer
	trackBody bool
}

func (r *proxyRecorder) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *proxyRecorder) Write(b []byte) (int, error) {
	if r.trackBody {
		r.body.Write(b)
	}
	return r.ResponseWriter.Write(b)
}

// upstreamUsage is the subset of an OpenAI-compatible response body we care about.
type upstreamUsage struct {
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

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

	if k.OneShot && k.Used {
		writeError(w, "key already used", http.StatusUnauthorized)
		return
	}

	if k.TokenBudget > 0 && h.UsageStore != nil {
		if h.UsageStore.TotalTokens(k.ID) >= k.TokenBudget {
			writeError(w, "token budget exceeded", http.StatusTooManyRequests)
			return
		}
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

	injectCredential(r, svc.CredentialHeader, rk.APIKey)

	// Inject identity headers — server-side only, consumers cannot spoof these.
	r.Header.Set("X-Bifrost-Key-ID", k.ID)
	if k.Source == keys.SourceMCP {
		r.Header.Set("X-Bifrost-Agent-ID", k.ID)
	}

	// Mark one-shot keys as used before forwarding — prevents replay even if
	// the upstream returns an error.
	if k.OneShot {
		k.Used = true
		h.KeyStore.Update(k.ID, k)
	}

	trackTokens := config.TrackTokens()
	rec := &proxyRecorder{ResponseWriter: w, code: http.StatusOK, trackBody: trackTokens}

	start := time.Now()
	proxy := httputil.NewSingleHostReverseProxy(target)
	r.Host = target.Host
	proxy.ServeHTTP(rec, r)
	latency := time.Since(start).Milliseconds()

	if h.UsageStore != nil {
		// Snapshot total before this request so we can detect threshold crossings.
		var prevTotal int
		alerting := k.AlertThreshold > 0 && k.AlertWebhook != ""
		if alerting {
			prevTotal = h.UsageStore.TotalTokens(k.ID)
		}

		ev := usage.Event{
			KeyID:      k.ID,
			Timestamp:  time.Now(),
			StatusCode: rec.code,
			Service:    k.Target,
			LatencyMS:  latency,
		}
		if trackTokens && rec.code == http.StatusOK {
			var body upstreamUsage
			if err := json.Unmarshal(rec.body.Bytes(), &body); err == nil {
				ev.PromptTokens = body.Usage.PromptTokens
				ev.CompletionTokens = body.Usage.CompletionTokens
				ev.TotalTokens = body.Usage.TotalTokens
			}
		}
		h.UsageStore.Record(ev) //nolint:errcheck

		// Fire webhook when the alert threshold is crossed for the first time.
		if alerting {
			newTotal := prevTotal + ev.TotalTokens
			if prevTotal < k.AlertThreshold && newTotal >= k.AlertThreshold {
				go fireAlertWebhook(k.AlertWebhook, k.ID, k.Target, k.AlertThreshold, newTotal)
			}
		}
	}
}

// alertPayload is the JSON body sent to the alert webhook.
type alertPayload struct {
	Event        string    `json:"event"`
	KeyID        string    `json:"key_id"`
	Service      string    `json:"service"`
	Threshold    int       `json:"threshold"`
	TotalTokens  int       `json:"total_tokens"`
	Timestamp    time.Time `json:"timestamp"`
}

// fireAlertWebhook posts an alert to the configured URL. Called in a goroutine —
// failures are logged but do not affect the proxied response.
func fireAlertWebhook(webhookURL, keyID, service string, threshold, totalTokens int) {
	payload, err := json.Marshal(alertPayload{
		Event:       "token_threshold_crossed",
		KeyID:       keyID,
		Service:     service,
		Threshold:   threshold,
		TotalTokens: totalTokens,
		Timestamp:   time.Now().UTC(),
	})
	if err != nil {
		return
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(payload)) //nolint:noctx
	if err != nil {
		logging.Logger.Warn().Err(err).Str("key_id", keyID).Msg("alert webhook failed")
		return
	}
	resp.Body.Close()
}

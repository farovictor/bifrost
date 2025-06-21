package routes

import "net/http"

// Healthz responds with a simple string to indicate the service is running.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

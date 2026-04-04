package routes

import "net/http"

// Healthz responds with a simple string to indicate the service is running.
//
// @Summary      Health check
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "ok"
// @Router       /healthz [get]
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

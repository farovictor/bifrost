package routes

import (
	"encoding/json"
	"net/http"

	"github.com/FokusInternal/bifrost/pkg/version"
)

// Version responds with the application's version.
func Version(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": version.Version})
}

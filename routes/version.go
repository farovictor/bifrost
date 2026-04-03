package routes

import (
	"encoding/json"
	"net/http"

	"github.com/farovictor/bifrost/pkg/version"
)

// Version responds with the application's version.
//
// @Summary      Get version
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /version [get]
func Version(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": version.Version})
}

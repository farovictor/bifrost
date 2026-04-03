package routes

import (
	"net/http"

	"github.com/swaggo/swag"
)

// OpenAPISpec serves the generated OpenAPI 3.0 spec as JSON.
//
// @Summary      OpenAPI spec (JSON)
// @Tags         health
// @Produce      json
// @Success      200  {object}  object
// @Router       /docs/openapi.json [get]
func OpenAPISpec(w http.ResponseWriter, r *http.Request) {
	doc, err := swag.ReadDoc()
	if err != nil {
		http.Error(w, "spec unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(doc))
}

// OpenAPISpecYAML serves the generated OpenAPI spec as YAML.
//
// @Summary      OpenAPI spec (YAML)
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string
// @Router       /docs/openapi.yaml [get]
func OpenAPISpecYAML(w http.ResponseWriter, r *http.Request) {
	// swag only registers JSON; serve the embedded YAML file
	http.ServeFile(w, r, "docs/swagger/swagger.yaml")
}

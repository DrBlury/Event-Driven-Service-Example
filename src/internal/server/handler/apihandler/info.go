package apihandler

import (
	server "drblury/event-driven-service/internal/server/generated"
	_ "embed"
	"html/template"
	"net/http"
)

// embed the openapi JSON and HTML file into the binary
// so we can serve them without reading from the filesystem

//go:embed embedded/stoplight.html
var openapiHTMLStoplight []byte

// Get status
// (GET /info/status)
func (ah *APIHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ah.RespondWithJSON(w, r, http.StatusOK, map[string]string{"status": "HEALTHY"})
}

// Get version
// (GET /info/version)
func (ah *APIHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	ah.RespondWithJSON(w, r, http.StatusOK, ah.Info)
}

// Get openapi JSON
// (GET /info/openapi.json)
func (ah *APIHandler) GetOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	swagger, err := server.GetSwagger()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, err := swagger.MarshalJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get openapi HTML
// (GET /info/openapi.html)
func (ah *APIHandler) GetOpenAPIHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	t, err := template.New("openapi").Parse(string(openapiHTMLStoplight))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// replace the base URL in the HTML file
	// with the actual base URL of the server
	// and render to the response writer
	err = t.Execute(w, map[string]string{
		"BaseURL": ah.BaseURL,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

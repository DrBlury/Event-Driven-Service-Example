package info

import (
	"html/template"
	"net/http"
)

// GetStatus returns a simple health payload that can be used for lightweight diagnostics.
func (ih *InfoHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ih.respondProbe(w, r, http.StatusOK, "HEALTHY")
}

// GetHealthz implements the liveness probe recommended for Kubernetes.
func (ih *InfoHandler) GetHealthz(w http.ResponseWriter, r *http.Request) {
	if err := ih.runChecks(r.Context(), ih.livenessChecks); err != nil {
		ih.HandleAPIError(w, r, http.StatusServiceUnavailable, err, "liveness probe failed")
		return
	}
	ih.respondProbe(w, r, http.StatusOK, "ok")
}

// GetReadyz implements the readiness probe recommended for Kubernetes.
func (ih *InfoHandler) GetReadyz(w http.ResponseWriter, r *http.Request) {
	if err := ih.runChecks(r.Context(), ih.readinessChecks); err != nil {
		ih.HandleAPIError(w, r, http.StatusServiceUnavailable, err, "readiness probe failed")
		return
	}
	ih.respondProbe(w, r, http.StatusOK, "ready")
}

// GetVersion returns the structure provided by the configured InfoProvider.
func (ih *InfoHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	payload := ih.infoProvider()
	if payload == nil {
		payload = map[string]string{}
	}
	ih.RespondWithJSON(w, r, http.StatusOK, payload)
}

// GetOpenAPIJSON streams the configured OpenAPI JSON document to the caller.
func (ih *InfoHandler) GetOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bytes, err := ih.swaggerProvider()
	if err != nil {
		ih.HandleInternalServerError(w, r, err, "failed to load swagger spec")
		return
	}

	if _, err = w.Write(bytes); err != nil {
		ih.HandleInternalServerError(w, r, err, "failed to write swagger response")
		return
	}
}

// GetOpenAPIHTML renders an embedded Stoplight viewer that fetches the OpenAPI document from the JSON endpoint.
func (ih *InfoHandler) GetOpenAPIHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.New("openapi").Parse(string(openapiHTMLStoplight))
	if err != nil {
		ih.HandleInternalServerError(w, r, err, "failed to parse openapi template")
		return
	}

	if err := tmpl.Execute(w, map[string]string{
		"BaseURL": ih.baseURL,
	}); err != nil {
		ih.HandleInternalServerError(w, r, err, "failed to render openapi template")
		return
	}
}

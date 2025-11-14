package api

import (
	_ "embed"
	"errors"
	"html/template"
	"net/http"
)

//go:embed embedded/stoplight.html
var openapiHTMLStoplight []byte

type InfoProvider func() any

type SwaggerProvider func() ([]byte, error)

type InfoOption func(*InfoHandler)

type InfoHandler struct {
	*Responder
	baseURL         string
	infoProvider    InfoProvider
	swaggerProvider SwaggerProvider
}

func NewInfoHandler(opts ...InfoOption) *InfoHandler {
	ih := &InfoHandler{
		Responder: NewResponder(),
		infoProvider: func() any {
			return map[string]string{}
		},
		swaggerProvider: func() ([]byte, error) {
			return nil, errors.New("api swagger provider not configured")
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(ih)
		}
	}
	return ih
}

func WithInfoResponder(responder *Responder) InfoOption {
	return func(ih *InfoHandler) {
		if responder != nil {
			ih.Responder = responder
		}
	}
}

func WithBaseURL(baseURL string) InfoOption {
	return func(ih *InfoHandler) {
		ih.baseURL = baseURL
	}
}

func WithInfoProvider(provider InfoProvider) InfoOption {
	return func(ih *InfoHandler) {
		if provider != nil {
			ih.infoProvider = provider
		}
	}
}

func WithSwaggerProvider(provider SwaggerProvider) InfoOption {
	return func(ih *InfoHandler) {
		if provider != nil {
			ih.swaggerProvider = provider
		}
	}
}

func (ih *InfoHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ih.RespondWithJSON(w, r, http.StatusOK, map[string]string{"status": "HEALTHY"})
}

func (ih *InfoHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	payload := ih.infoProvider()
	if payload == nil {
		payload = map[string]string{}
	}
	ih.RespondWithJSON(w, r, http.StatusOK, payload)
}

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

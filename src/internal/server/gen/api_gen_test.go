package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestErrorTypes tests the error type implementations
func TestUnescapedCookieParamError(t *testing.T) {
	t.Parallel()

	underlyingErr := errors.New("underlying error")
	err := &UnescapedCookieParamError{
		ParamName: "session_id",
		Err:       underlyingErr,
	}

	expectedMsg := "error unescaping cookie parameter 'session_id'"
	if err.Error() != expectedMsg {
		t.Errorf("expected %q, got %q", expectedMsg, err.Error())
	}

	if err.Unwrap() != underlyingErr {
		t.Error("Unwrap should return the underlying error")
	}
}

func TestUnmarshalingParamError(t *testing.T) {
	t.Parallel()

	underlyingErr := errors.New("json syntax error")
	err := &UnmarshalingParamError{
		ParamName: "data",
		Err:       underlyingErr,
	}

	if !strings.Contains(err.Error(), "data") {
		t.Error("error message should contain param name")
	}
	if !strings.Contains(err.Error(), "json syntax error") {
		t.Error("error message should contain underlying error")
	}

	if err.Unwrap() != underlyingErr {
		t.Error("Unwrap should return the underlying error")
	}
}

func TestRequiredParamError(t *testing.T) {
	t.Parallel()

	err := &RequiredParamError{
		ParamName: "id",
	}

	if !strings.Contains(err.Error(), "id") {
		t.Error("error message should contain param name")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Error("error message should indicate required")
	}
}

func TestRequiredHeaderError(t *testing.T) {
	t.Parallel()

	underlyingErr := errors.New("header missing")
	err := &RequiredHeaderError{
		ParamName: "Authorization",
		Err:       underlyingErr,
	}

	if !strings.Contains(err.Error(), "Authorization") {
		t.Error("error message should contain param name")
	}

	if err.Unwrap() != underlyingErr {
		t.Error("Unwrap should return the underlying error")
	}
}

func TestInvalidParamFormatError(t *testing.T) {
	t.Parallel()

	underlyingErr := errors.New("expected int")
	err := &InvalidParamFormatError{
		ParamName: "count",
		Err:       underlyingErr,
	}

	if !strings.Contains(err.Error(), "count") {
		t.Error("error message should contain param name")
	}
	if !strings.Contains(err.Error(), "expected int") {
		t.Error("error message should contain underlying error")
	}

	if err.Unwrap() != underlyingErr {
		t.Error("Unwrap should return the underlying error")
	}
}

func TestTooManyValuesForParamError(t *testing.T) {
	t.Parallel()

	err := &TooManyValuesForParamError{
		ParamName: "filter",
		Count:     5,
	}

	if !strings.Contains(err.Error(), "filter") {
		t.Error("error message should contain param name")
	}
	if !strings.Contains(err.Error(), "5") {
		t.Error("error message should contain count")
	}
}

// TestResponseTypes tests the response type Visit implementations
func TestCreateExampleRecord201JSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	msg := "created"
	status := "success"
	response := CreateExampleRecord201JSONResponse{
		Message: &msg,
		Status:  &status,
	}

	err := response.VisitCreateExampleRecordResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 201 {
		t.Errorf("expected status 201, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json content type")
	}
}

func TestCreateExampleRecord400JSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := CreateExampleRecord400JSONResponse{
		Title:  "Bad Request",
		Status: 400,
	}

	err := response.VisitCreateExampleRecordResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateExampleRecorddefaultJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := CreateExampleRecorddefaultJSONResponse{
		Body: ProblemDetails{
			Title:  "Internal Error",
			Status: 500,
		},
		StatusCode: 500,
	}

	err := response.VisitCreateExampleRecordResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGetHealthz200JSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetHealthz200JSONResponse{
		Status: "healthy",
	}

	err := response.VisitGetHealthzResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetHealthz400ApplicationProblemPlusJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetHealthz400ApplicationProblemPlusJSONResponse{
		Title:  "Bad Request",
		Status: 400,
	}

	err := response.VisitGetHealthzResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/problem+json" {
		t.Errorf("expected application/problem+json content type")
	}
}

func TestGetHealthz503ApplicationProblemPlusJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetHealthz503ApplicationProblemPlusJSONResponse{
		Title:  "Service Unavailable",
		Status: 503,
	}

	err := response.VisitGetHealthzResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 503 {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestGetOpenAPIHTML200TexthtmlResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	htmlContent := "<html><body>API Docs</body></html>"
	response := GetOpenAPIHTML200TexthtmlResponse{
		Body:          strings.NewReader(htmlContent),
		ContentLength: int64(len(htmlContent)),
	}

	err := response.VisitGetOpenAPIHTMLResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/html" {
		t.Errorf("expected text/html content type")
	}
}

func TestGetOpenAPIHTML200TexthtmlResponseWithCloser(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	htmlContent := "<html><body>API Docs</body></html>"
	response := GetOpenAPIHTML200TexthtmlResponse{
		Body: io.NopCloser(strings.NewReader(htmlContent)),
	}

	err := response.VisitGetOpenAPIHTMLResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetOpenAPIHTML400ApplicationProblemPlusJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetOpenAPIHTML400ApplicationProblemPlusJSONResponse{
		Title:  "Bad Request",
		Status: 400,
	}

	err := response.VisitGetOpenAPIHTMLResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetOpenAPIJSON200Responses(t *testing.T) {
	t.Parallel()

	t.Run("ApplicationSchemaPlusJSON", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		response := GetOpenAPIJSON200ApplicationSchemaPlusJSONResponse{
			"openapi": "3.0.0",
		}

		err := response.VisitGetOpenAPIJSONResponse(w)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Header().Get("Content-Type") != "application/schema+json" {
			t.Errorf("expected application/schema+json content type")
		}
	})

	t.Run("ApplicationVndOaiOpenapi", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		response := GetOpenAPIJSON200ApplicationVndOaiOpenapiPlusJSONVersion31Response{
			"openapi": "3.1.0",
		}

		err := response.VisitGetOpenAPIJSONResponse(w)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestGetOpenAPIJSON400ApplicationProblemPlusJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetOpenAPIJSON400ApplicationProblemPlusJSONResponse{
		Title:  "Bad Request",
		Status: 400,
	}

	err := response.VisitGetOpenAPIJSONResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetStatus200JSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetStatus200JSONResponse{
		Status: "ok",
	}

	err := response.VisitGetStatusResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetStatus400ApplicationProblemPlusJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetStatus400ApplicationProblemPlusJSONResponse{
		Title:  "Bad Request",
		Status: 400,
	}

	err := response.VisitGetStatusResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetVersion200JSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetVersion200JSONResponse{
		Version:    "1.0.0",
		CommitHash: "abc123",
		CommitDate: "2024-01-01",
		BuildDate:  "2024-01-02",
		Details:    "Test build",
	}

	err := response.VisitGetVersionResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetVersion400ApplicationProblemPlusJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetVersion400ApplicationProblemPlusJSONResponse{
		Title:  "Bad Request",
		Status: 400,
	}

	err := response.VisitGetVersionResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetReadyz200JSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetReadyz200JSONResponse{
		Status: "ready",
	}

	err := response.VisitGetReadyzResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetReadyz400ApplicationProblemPlusJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetReadyz400ApplicationProblemPlusJSONResponse{
		Title:  "Bad Request",
		Status: 400,
	}

	err := response.VisitGetReadyzResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetReadyz503ApplicationProblemPlusJSONResponse(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := GetReadyz503ApplicationProblemPlusJSONResponse{
		Title:  "Service Unavailable",
		Status: 503,
	}

	err := response.VisitGetReadyzResponse(w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 503 {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

// TestGetSwagger tests the swagger specification retrieval
func TestGetSwagger(t *testing.T) {
	t.Parallel()

	swagger, err := GetSwagger()
	if err != nil {
		t.Fatalf("unexpected error getting swagger: %v", err)
	}
	if swagger == nil {
		t.Error("swagger should not be nil")
	}
}

func TestPathToRawSpec(t *testing.T) {
	t.Parallel()

	t.Run("with path", func(t *testing.T) {
		t.Parallel()
		result := PathToRawSpec("/api/v1/spec.json")
		if len(result) != 1 {
			t.Errorf("expected 1 entry, got %d", len(result))
		}
	})

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()
		result := PathToRawSpec("")
		if len(result) != 0 {
			t.Errorf("expected 0 entries, got %d", len(result))
		}
	})
}

// Mock implementation of ServerInterface for testing
type mockServerImpl struct{}

func (m *mockServerImpl) CreateExampleRecord(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (m *mockServerImpl) GetAsyncAPIHTML(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockServerImpl) GetAsyncAPIJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockServerImpl) GetHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockServerImpl) GetOpenAPIHTML(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockServerImpl) GetOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockServerImpl) GetStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockServerImpl) GetVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockServerImpl) GetReadyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestHandler(t *testing.T) {
	t.Parallel()

	mock := &mockServerImpl{}
	handler := Handler(mock)

	if handler == nil {
		t.Error("handler should not be nil")
	}
}

func TestHandlerFromMux(t *testing.T) {
	t.Parallel()

	mock := &mockServerImpl{}
	mux := http.NewServeMux()
	handler := HandlerFromMux(mock, mux)

	if handler == nil {
		t.Error("handler should not be nil")
	}
}

func TestHandlerFromMuxWithBaseURL(t *testing.T) {
	t.Parallel()

	mock := &mockServerImpl{}
	mux := http.NewServeMux()
	handler := HandlerFromMuxWithBaseURL(mock, mux, "/api/v1")

	if handler == nil {
		t.Error("handler should not be nil")
	}
}

func TestHandlerWithOptions(t *testing.T) {
	t.Parallel()

	mock := &mockServerImpl{}
	options := StdHTTPServerOptions{
		BaseURL: "/api",
		Middlewares: []MiddlewareFunc{
			func(next http.Handler) http.Handler {
				return next
			},
		},
	}
	handler := HandlerWithOptions(mock, options)

	if handler == nil {
		t.Error("handler should not be nil")
	}
}

func TestServerInterfaceWrapper(t *testing.T) {
	t.Parallel()

	mock := &mockServerImpl{}
	wrapper := &ServerInterfaceWrapper{
		Handler: mock,
	}

	tests := []struct {
		name    string
		method  string
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"GetHealthz", "GET", "/healthz", wrapper.GetHealthz},
		{"GetReadyz", "GET", "/readyz", wrapper.GetReadyz},
		{"GetStatus", "GET", "/info/status", wrapper.GetStatus},
		{"GetVersion", "GET", "/info/version", wrapper.GetVersion},
		{"GetOpenAPIHTML", "GET", "/info/openapi.html", wrapper.GetOpenAPIHTML},
		{"GetOpenAPIJSON", "GET", "/info/openapi.json", wrapper.GetOpenAPIJSON},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			tc.handler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestServerInterfaceWrapperCreateExampleRecord(t *testing.T) {
	t.Parallel()

	mock := &mockServerImpl{}
	wrapper := &ServerInterfaceWrapper{
		Handler: mock,
	}

	req := httptest.NewRequest("POST", "/examples", nil)
	w := httptest.NewRecorder()
	wrapper.CreateExampleRecord(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestServerInterfaceWrapperWithMiddleware(t *testing.T) {
	t.Parallel()

	mock := &mockServerImpl{}
	middlewareCalled := false
	wrapper := &ServerInterfaceWrapper{
		Handler: mock,
		HandlerMiddlewares: []MiddlewareFunc{
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					middlewareCalled = true
					next.ServeHTTP(w, r)
				})
			},
		},
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	wrapper.GetHealthz(w, req)

	if !middlewareCalled {
		t.Error("middleware should have been called")
	}
}

// Mock implementation of StrictServerInterface for testing
type mockStrictServerImpl struct{}

func (m *mockStrictServerImpl) CreateExampleRecord(ctx context.Context, request CreateExampleRecordRequestObject) (CreateExampleRecordResponseObject, error) {
	msg := "created"
	return CreateExampleRecord201JSONResponse{Message: &msg}, nil
}

func (m *mockStrictServerImpl) GetAsyncAPIHTML(ctx context.Context, request GetAsyncAPIHTMLRequestObject) (GetAsyncAPIHTMLResponseObject, error) {
	return GetAsyncAPIHTML200TexthtmlResponse{Body: strings.NewReader("<html></html>")}, nil
}

func (m *mockStrictServerImpl) GetAsyncAPIJSON(ctx context.Context, request GetAsyncAPIJSONRequestObject) (GetAsyncAPIJSONResponseObject, error) {
	return GetAsyncAPIJSON200JSONResponse{"asyncapi": "3.0.0"}, nil
}

func (m *mockStrictServerImpl) GetHealthz(ctx context.Context, request GetHealthzRequestObject) (GetHealthzResponseObject, error) {
	return GetHealthz200JSONResponse{Status: "healthy"}, nil
}

func (m *mockStrictServerImpl) GetOpenAPIHTML(ctx context.Context, request GetOpenAPIHTMLRequestObject) (GetOpenAPIHTMLResponseObject, error) {
	return GetOpenAPIHTML200TexthtmlResponse{Body: strings.NewReader("<html></html>")}, nil
}

func (m *mockStrictServerImpl) GetOpenAPIJSON(ctx context.Context, request GetOpenAPIJSONRequestObject) (GetOpenAPIJSONResponseObject, error) {
	return GetOpenAPIJSON200ApplicationSchemaPlusJSONResponse{"openapi": "3.0.0"}, nil
}

func (m *mockStrictServerImpl) GetStatus(ctx context.Context, request GetStatusRequestObject) (GetStatusResponseObject, error) {
	return GetStatus200JSONResponse{Status: "ok"}, nil
}

func (m *mockStrictServerImpl) GetVersion(ctx context.Context, request GetVersionRequestObject) (GetVersionResponseObject, error) {
	return GetVersion200JSONResponse{Version: "1.0.0"}, nil
}

func (m *mockStrictServerImpl) GetReadyz(ctx context.Context, request GetReadyzRequestObject) (GetReadyzResponseObject, error) {
	return GetReadyz200JSONResponse{Status: "ready"}, nil
}

func TestNewStrictHandler(t *testing.T) {
	t.Parallel()

	mock := &mockStrictServerImpl{}
	handler := NewStrictHandler(mock, nil)

	if handler == nil {
		t.Error("handler should not be nil")
	}
}

func TestNewStrictHandlerWithOptions(t *testing.T) {
	t.Parallel()

	mock := &mockStrictServerImpl{}
	options := StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}
	handler := NewStrictHandlerWithOptions(mock, nil, options)

	if handler == nil {
		t.Error("handler should not be nil")
	}
}

func TestStrictHandlerEndpoints(t *testing.T) {
	t.Parallel()

	mock := &mockStrictServerImpl{}
	handler := NewStrictHandler(mock, nil)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		handler        func(http.ResponseWriter, *http.Request)
		expectedStatus int
	}{
		{"GetHealthz", "GET", "/healthz", "", handler.GetHealthz, http.StatusOK},
		{"GetReadyz", "GET", "/readyz", "", handler.GetReadyz, http.StatusOK},
		{"GetStatus", "GET", "/info/status", "", handler.GetStatus, http.StatusOK},
		{"GetVersion", "GET", "/info/version", "", handler.GetVersion, http.StatusOK},
		{"GetOpenAPIHTML", "GET", "/info/openapi.html", "", handler.GetOpenAPIHTML, http.StatusOK},
		{"GetOpenAPIJSON", "GET", "/info/openapi.json", "", handler.GetOpenAPIJSON, http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			req := httptest.NewRequest(tc.method, tc.path, body)
			w := httptest.NewRecorder()
			tc.handler(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestStrictHandlerCreateExampleRecord(t *testing.T) {
	t.Parallel()

	mock := &mockStrictServerImpl{}
	handler := NewStrictHandler(mock, nil)

	body := `{"recordId": "123", "title": "Test", "meta": {"requestedBy": "test", "priority": 1, "desiredStartDate": {"year": 2024, "month": 1, "day": 1}}}`
	req := httptest.NewRequest("POST", "/examples", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateExampleRecord(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestStrictHandlerCreateExampleRecordInvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockStrictServerImpl{}
	handler := NewStrictHandler(mock, nil)

	body := `invalid json`
	req := httptest.NewRequest("POST", "/examples", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateExampleRecord(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// Test strict handler with error response
type mockStrictServerImplWithError struct{}

func (m *mockStrictServerImplWithError) CreateExampleRecord(ctx context.Context, request CreateExampleRecordRequestObject) (CreateExampleRecordResponseObject, error) {
	return nil, errors.New("internal error")
}

func (m *mockStrictServerImplWithError) GetAsyncAPIHTML(ctx context.Context, request GetAsyncAPIHTMLRequestObject) (GetAsyncAPIHTMLResponseObject, error) {
	return nil, errors.New("internal error")
}

func (m *mockStrictServerImplWithError) GetAsyncAPIJSON(ctx context.Context, request GetAsyncAPIJSONRequestObject) (GetAsyncAPIJSONResponseObject, error) {
	return nil, errors.New("internal error")
}

func (m *mockStrictServerImplWithError) GetHealthz(ctx context.Context, request GetHealthzRequestObject) (GetHealthzResponseObject, error) {
	return nil, errors.New("internal error")
}

func (m *mockStrictServerImplWithError) GetOpenAPIHTML(ctx context.Context, request GetOpenAPIHTMLRequestObject) (GetOpenAPIHTMLResponseObject, error) {
	return nil, errors.New("internal error")
}

func (m *mockStrictServerImplWithError) GetOpenAPIJSON(ctx context.Context, request GetOpenAPIJSONRequestObject) (GetOpenAPIJSONResponseObject, error) {
	return nil, errors.New("internal error")
}

func (m *mockStrictServerImplWithError) GetStatus(ctx context.Context, request GetStatusRequestObject) (GetStatusResponseObject, error) {
	return nil, errors.New("internal error")
}

func (m *mockStrictServerImplWithError) GetVersion(ctx context.Context, request GetVersionRequestObject) (GetVersionResponseObject, error) {
	return nil, errors.New("internal error")
}

func (m *mockStrictServerImplWithError) GetReadyz(ctx context.Context, request GetReadyzRequestObject) (GetReadyzResponseObject, error) {
	return nil, errors.New("internal error")
}

func TestStrictHandlerWithErrors(t *testing.T) {
	t.Parallel()

	mock := &mockStrictServerImplWithError{}
	handler := NewStrictHandler(mock, nil)

	tests := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"GetHealthz", handler.GetHealthz},
		{"GetReadyz", handler.GetReadyz},
		{"GetStatus", handler.GetStatus},
		{"GetVersion", handler.GetVersion},
		{"GetOpenAPIHTML", handler.GetOpenAPIHTML},
		{"GetOpenAPIJSON", handler.GetOpenAPIJSON},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			tc.handler(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("expected status 500, got %d", w.Code)
			}
		})
	}
}

func TestStrictHandlerWithMiddleware(t *testing.T) {
	t.Parallel()

	mock := &mockStrictServerImpl{}
	middlewareCalled := false

	middleware := func(f StrictHandlerFunc, operationID string) StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			middlewareCalled = true
			return f(ctx, w, r, request)
		}
	}

	handler := NewStrictHandler(mock, []StrictMiddlewareFunc{middleware})

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	handler.GetHealthz(w, req)

	if !middlewareCalled {
		t.Error("middleware should have been called")
	}
}

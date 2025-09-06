package echo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	svc "github.com/hesusruiz/isbetmf/tmfserver/service"
	"github.com/labstack/echo/v4"
)

// ServiceInterface defines the methods used by the handler
type ServiceInterface interface {
	CreateGenericObject(req *svc.Request) *svc.Response
	GetGenericObject(req *svc.Request) *svc.Response
	UpdateGenericObject(req *svc.Request) *svc.Response
	DeleteGenericObject(req *svc.Request) *svc.Response
	ListGenericObjects(req *svc.Request) *svc.Response
}

// testHandler wraps the original handler to use an interface
type testHandler struct {
	service ServiceInterface
}

func (h *testHandler) ListGenericObjects(c echo.Context) error {
	jwtToken := svc.ExtractJWTToken(c.Request().Header.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Request().Method,
		Action:       "LIST",
		ResourceName: c.Param("resourceName"),
		QueryParams:  c.QueryParams(),
		AccessToken:  jwtToken,
	}

	resp := h.service.ListGenericObjects(req)
	return SendResponse(c, resp)
}

// mockService is a mock service for testing
type mockService struct{}

func (m *mockService) CreateGenericObject(req *svc.Request) *svc.Response {
	return &svc.Response{StatusCode: http.StatusNotImplemented}
}

func (m *mockService) GetGenericObject(req *svc.Request) *svc.Response {
	return &svc.Response{StatusCode: http.StatusNotImplemented}
}

func (m *mockService) UpdateGenericObject(req *svc.Request) *svc.Response {
	return &svc.Response{StatusCode: http.StatusNotImplemented}
}

func (m *mockService) DeleteGenericObject(req *svc.Request) *svc.Response {
	return &svc.Response{StatusCode: http.StatusNotImplemented}
}

func (m *mockService) ListGenericObjects(req *svc.Request) *svc.Response {
	// Return an empty list response
	headers := make(map[string]string)
	headers["X-Total-Count"] = "0"

	return &svc.Response{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       []map[string]any{}, // Empty array
	}
}

// TestListGenericObjectsEmptyList tests that the handler returns an empty JSON array and proper X-Total-Count header
func TestListGenericObjectsEmptyList(t *testing.T) {
	// Create handler with mock service
	mockSvc := &mockService{}
	handler := &testHandler{service: mockSvc}

	// Create echo instance
	e := echo.New()

	// Setup route
	e.GET("/:resourceName", handler.ListGenericObjects)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/NonExistentResource", nil)
	rec := httptest.NewRecorder()

	// Execute request
	c := e.NewContext(req, rec)
	c.SetPath("/:resourceName")
	c.SetParamNames("resourceName")
	c.SetParamValues("NonExistentResource")

	// Call handler
	err := handler.ListGenericObjects(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Check X-Total-Count header
	totalCount := rec.Header().Get("X-Total-Count")
	if totalCount != "0" {
		t.Fatalf("expected X-Total-Count=0, got %s", totalCount)
	}

	// Check content type
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %s", contentType)
	}

	// Check body is an empty JSON array
	var body []map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if len(body) != 0 {
		t.Fatalf("expected empty array, got %d items", len(body))
	}
}

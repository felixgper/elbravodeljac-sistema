package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestArchitectureNeverClaimsBusinessOrDatabaseReadiness(t *testing.T) {
	handler := New()
	for path, status := range map[string]int{"/healthz": 200, "/readyz": 503} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != status {
			t.Fatalf("%s: got %d, want %d", path, w.Code, status)
		}
	}
	for _, route := range Routes {
		path := route.Path
		// Sustituye los parámetros de ruta para probar el registro real del mux.
		for strings.Contains(path, "{") {
			start := strings.Index(path, "{")
			end := strings.Index(path[start:], "}") + start
			path = path[:start] + "1" + path[end+1:]
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(route.Method, path, nil))
		if w.Code != http.StatusNotImplemented || !strings.Contains(w.Body.String(), "NOT_IMPLEMENTED") {
			t.Fatalf("%s %s: got %d %s", route.Method, path, w.Code, w.Body.String())
		}
	}
}

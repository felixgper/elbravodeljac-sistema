package httpapi

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontendSameOriginAndPublicFilesOnly(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets", "js"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><head></head><body>aprobado</body></html>"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "js", "app.js"), []byte("export const publicAsset=true;"), 0600); err != nil {
		t.Fatal(err)
	}
	h, err := New(&stubStore{}, Options{FrontendDir: dir, Origin: "http://localhost:8080", Key: []byte(strings.Repeat("k", 32))})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path     string
		status   int
		contains string
	}{
		{"/", 200, `content="real"`}, {"/index.html", 200, `content="real"`},
		{"/demo", 404, ""}, {"/demo/", 404, ""}, {"/demo/index.html", 404, ""},
		{"/demo/?mode=real", 404, ""}, {"/demo/assets/", 404, ""}, {"/demo/assets/js/app.js", 404, ""},
		{"/checks/fixtures/legacy-demo-loader.mjs", 404, ""},
		{"/assets/js/app.js", 200, "publicAsset"},
		{"/backend/.env", 404, ""}, {"/assets/js/", 404, ""}, {"/database/postgresql/esquema_entrega.sql", 404, ""},
		{"/api/v1/clientes", 401, "UNAUTHORIZED"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			h.ServeHTTP(res, httptest.NewRequest("GET", tc.path, nil))
			if res.Code != tc.status || !strings.Contains(res.Body.String(), tc.contains) {
				t.Fatalf("%d %s", res.Code, res.Body.String())
			}
		})
	}
}

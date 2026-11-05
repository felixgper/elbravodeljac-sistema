package httpapi

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Only the approved dist entry point and public assets are exposed, never the
// repository root, SQL, environment files or directory listings.
func mountFrontend(mux *http.ServeMux, dir string) error {
	index, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		return err
	}
	assets := http.FileServer(http.Dir(filepath.Join(dir, "assets")))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "/index.html", "/demo/", "/demo/index.html":
			mode := "real"
			if strings.HasPrefix(r.URL.Path, "/demo/") {
				mode = "demo"
			}
			page := bytes.Replace(index, []byte("<head>"), []byte("<head>\n<meta name=\"bravo-mode\" content=\""+mode+"\">"), 1)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(page)
		default:
			prefix := "/assets/"
			if strings.HasPrefix(r.URL.Path, "/demo/assets/") {
				prefix = "/demo/assets/"
			}
			if !strings.HasPrefix(r.URL.Path, prefix) {
				http.NotFound(w, r)
				return
			}
			name := strings.TrimPrefix(r.URL.Path, prefix)
			if strings.Contains(name, "\\") || strings.Contains(name, ":") {
				http.NotFound(w, r)
				return
			}
			for _, segment := range strings.Split(name, "/") {
				if segment == "" || strings.HasPrefix(segment, ".") {
					http.NotFound(w, r)
					return
				}
			}
			switch strings.ToLower(filepath.Ext(name)) {
			case ".js", ".css", ".svg", ".png", ".jpg", ".jpeg", ".webp", ".ico", ".woff", ".woff2":
			default:
				http.NotFound(w, r)
				return
			}
			info, err := os.Stat(filepath.Join(dir, "assets", filepath.FromSlash(name)))
			if err != nil || !info.Mode().IsRegular() {
				http.NotFound(w, r)
				return
			}
			http.StripPrefix(strings.TrimSuffix(prefix, "/"), assets).ServeHTTP(w, r)
		}
	})
	return nil
}

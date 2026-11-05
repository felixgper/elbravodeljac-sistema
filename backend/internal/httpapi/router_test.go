package httpapi

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/application"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubStore struct {
	application.Store
	user domain.User
}

func (s *stubStore) Ping(context.Context) error                             { return nil }
func (s *stubStore) User(context.Context, string) (domain.User, error)      { return s.user, nil }
func (s *stubStore) FindLogin(context.Context, string) (domain.User, error) { return s.user, nil }
func (s *stubStore) Clients(context.Context, domain.Page) (domain.List[domain.Client], error) {
	return domain.List[domain.Client]{Data: []domain.Client{}}, nil
}
func TestHTTPAuthenticationCSRFAndLivePermissions(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("unit-test-password"), bcrypt.MinCost)
	store := &stubStore{user: domain.User{ID: "1", Active: true, PasswordHash: string(hash), Permissions: []string{"clientes.ver"}}}
	h, err := New(store, Options{Origin: "https://example.test", Key: []byte(strings.Repeat("k", 32)), SecureCookies: true})
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path, body, origin, csrf string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		r.Header.Set("X-CSRF-Token", csrf)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := call("GET", "/api/v1/clientes", "", "", "", nil); w.Code != 401 {
		t.Fatal(w.Code)
	}
	if w := call("POST", "/api/v1/auth/login", `{"username":"u","password":"unit-test-password"}`, "https://evil.test", "", nil); w.Code != 403 {
		t.Fatal(w.Code)
	}
	w := call("POST", "/api/v1/auth/login", `{"username":"u","password":"unit-test-password"}`, "https://example.test", "", nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("insecure cookie")
	}
	if strings.Contains(w.Body.String(), "PasswordHash") || strings.Contains(w.Body.String(), "$2") || strings.Contains(w.Body.String(), cookie.Value) {
		t.Fatal("credential leaked")
	}
	var response struct {
		CSRF string `json:"csrfToken"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	if w = call("GET", "/api/v1/clientes", "", "", "", cookie); w.Code != 200 {
		t.Fatal(w.Code)
	}
	store.user.Permissions = nil
	if w = call("GET", "/api/v1/clientes", "", "", "", cookie); w.Code != 403 {
		t.Fatal("stale permissions", w.Code)
	}
	if w = call("POST", "/api/v1/auth/logout", "{}", "https://example.test", "", cookie); w.Code != 403 {
		t.Fatal("missing CSRF", w.Code)
	}
	if w = call("POST", "/api/v1/auth/logout", "{}", "https://evil.test", response.CSRF, cookie); w.Code != 403 {
		t.Fatal("bad origin", w.Code)
	}
	if w = call("POST", "/api/v1/auth/logout", "{}", "https://example.test", response.CSRF, cookie); w.Code != 204 {
		t.Fatal(w.Code)
	}
	if w = call("GET", "/api/v1/auth/me", "", "", "", cookie); w.Code != 401 {
		t.Fatal("logout not revoked", w.Code)
	}
}
func TestLoginLimiterAndDecoder(t *testing.T) {
	l := loginLimiter{entries: map[string]loginWindow{}}
	for i := 0; i < 5; i++ {
		if !l.allow("ip") {
			t.Fatal("early limit")
		}
	}
	if l.allow("ip") {
		t.Fatal("unbounded login")
	}
	for _, body := range []string{`{"unknown":1}`, `{} {}`, strings.Repeat("x", 65537)} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		var v struct {
			Name string `json:"name"`
		}
		if decode(w, r, &v) {
			t.Fatal("accepted invalid JSON")
		}
	}
}

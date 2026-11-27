package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Synthetic stores: no SQL/Mongo credentials and no official database access.
type surveyTestSQL struct {
	stubStore
	delivered bool
}

func (s *surveyTestSQL) SurveyDelivery(ctx context.Context, id string) (domain.SurveyDelivery, error) {
	if !s.delivered {
		return domain.SurveyDelivery{}, domain.ErrNotFound
	}
	return domain.SurveyDelivery{ID: "7", OrderID: id}, nil
}

type surveyMemory struct {
	mu     sync.Mutex
	v      domain.SurveyInvitation
	exists bool
}

func (m *surveyMemory) IssueSurvey(ctx context.Context, v domain.SurveyInvitation, replace bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.exists && (!replace || m.v.Response != nil || v.CreatedAt.Sub(m.v.CreatedAt) < time.Minute) {
		return domain.ErrConflict
	}
	m.v = v
	m.exists = true
	return nil
}
func (m *surveyMemory) SurveyByOrder(ctx context.Context, id string) (domain.SurveyInvitation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.exists || m.v.OrderID != id {
		return domain.SurveyInvitation{}, domain.ErrNotFound
	}
	return m.v, nil
}
func (m *surveyMemory) SurveyByHash(ctx context.Context, hash string) (domain.SurveyInvitation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.exists || m.v.TokenHash != hash {
		return domain.SurveyInvitation{}, domain.ErrNotFound
	}
	return m.v, nil
}
func (m *surveyMemory) AnswerSurvey(ctx context.Context, hash string, v domain.SurveyResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.v.TokenHash != hash || m.v.Response != nil || !m.v.ExpiresAt.After(time.Now()) {
		return domain.ErrConflict
	}
	m.v.Response = &v
	return nil
}
func (m *surveyMemory) SurveyStatistics(context.Context, time.Time, time.Time) (domain.SurveyStats, error) {
	return domain.SurveyStats{Total: 1, Average: 5, Aspects: [5]float64{1, 2, 3, 4, 5}, Distribution: [5]int{0, 0, 0, 0, 1}}, nil
}
func (m *surveyMemory) SurveyTracking(context.Context, int) ([]domain.SurveyInvitation, error) {
	return []domain.SurveyInvitation{m.v}, nil
}

func TestSurveyHTTPWorkflow(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("synthetic-test-password"), bcrypt.MinCost)
	sql := &surveyTestSQL{stubStore: stubStore{user: domain.User{ID: "1", Active: true, PasswordHash: string(hash), Permissions: []string{"pagos.ver", "pagos.entregar", "reportes.ver"}}}}
	memory := &surveyMemory{}
	dir := t.TempDir()
	for _, file := range []string{"index.html", "encuesta.html"} {
		b, err := os.ReadFile(filepath.Join("../../../dist", file))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, file), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	h, err := New(sql, Options{Origin: "https://workshop.test", Key: []byte(strings.Repeat("k", 32)), Surveys: memory, FrontendDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	var cookie *http.Cookie
	csrf := ""
	counter := 0
	call := func(method, path, body string, admin bool) *httptest.ResponseRecorder {
		counter++
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", counter)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", "https://workshop.test")
		if admin && cookie != nil {
			r.AddCookie(cookie)
			r.Header.Set("X-CSRF-Token", csrf)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := call("GET", "/api/v1/encuestas/estadisticas", "", false); w.Code != 401 {
		t.Fatal(w.Code)
	}
	w := call("POST", "/api/v1/auth/login", `{"username":"test","password":"synthetic-test-password"}`, false)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	cookie = w.Result().Cookies()[0]
	var login struct {
		CSRF string `json:"csrfToken"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &login)
	csrf = login.CSRF
	if w = call("POST", "/api/v1/ordenes/1/encuesta", `{}`, true); w.Code != 404 {
		t.Fatal("pending delivery", w.Code)
	}
	sql.delivered = true
	w = call("POST", "/api/v1/ordenes/1/encuesta", `{}`, true)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var issued struct {
		URL string `json:"url"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &issued)
	parts := strings.Split(issued.URL, "#t=")
	if len(parts) != 2 {
		t.Fatal("missing fragment")
	}
	token := parts[1]
	// Reopen using a fresh HTTP server instance and the same private key/store.
	reopened, e := New(sql, Options{Origin: "https://workshop.test", Key: []byte(strings.Repeat("k", 32)), Surveys: memory})
	if e != nil {
		t.Fatal(e)
	}
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"username":"test","password":"synthetic-test-password"}`))
	loginReq.Header.Set("Origin", "https://workshop.test")
	loginReq.Header.Set("Content-Type", "application/json")
	loginResult := httptest.NewRecorder()
	reopened.ServeHTTP(loginResult, loginReq)
	if loginResult.Code != 200 {
		t.Fatal("relogin after API restart")
	}
	req := httptest.NewRequest("GET", "/api/v1/ordenes/1/encuesta", nil)
	req.AddCookie(loginResult.Result().Cookies()[0])
	rw := httptest.NewRecorder()
	reopened.ServeHTTP(rw, req)
	var recovered struct {
		URL string `json:"url"`
	}
	_ = json.Unmarshal(rw.Body.Bytes(), &recovered)
	if rw.Code != 200 || recovered.URL != issued.URL {
		t.Fatal("link not recovered after reopening")
	}
	sql.user.Permissions = []string{"pagos.ver"}
	rw = httptest.NewRecorder()
	reopened.ServeHTTP(rw, req)
	if strings.Contains(rw.Body.String(), "/encuesta#t=") || strings.Contains(rw.Body.String(), "tokenCipher") {
		t.Fatal("read-only account received reusable bearer")
	}
	sql.user.Permissions = []string{"pagos.ver", "pagos.entregar", "reportes.ver"}
	if strings.Contains(w.Body.String(), memory.v.TokenHash) {
		t.Fatal("hash leaked")
	}
	if w = call("POST", "/api/v1/ordenes/1/encuesta", `{}`, true); w.Code != 200 {
		t.Fatal("idempotent invitation")
	}
	if w = call("POST", "/api/v1/ordenes/1/encuesta", `{"regenerate":true}`, true); w.Code != 409 {
		t.Fatal("regeneration cooldown")
	}
	public := func(path, payload string) *httptest.ResponseRecorder {
		return call("POST", "/api/v1/encuestas/publica/"+path, payload, false)
	}
	w = public("validar", fmt.Sprintf(`{"token":%q}`, token))
	if w.Code != 200 || strings.Contains(w.Body.String(), "deliveryId") || strings.Contains(w.Body.String(), "orderId") {
		t.Fatal("public disclosure", w.Body.String())
	}
	for _, ratings := range []string{`[1,2,3,4]`, `[1,2,3,4,5,1]`, `[1,2,3,4,6]`, `[1,2,3,4,0]`, `[1,2,3,4,4.5]`, `["1",2,3,4,5]`} {
		w = public("responder", fmt.Sprintf(`{"token":%q,"ratings":%s}`, token, ratings))
		if w.Code < 400 {
			t.Fatal("bad ratings", ratings)
		}
	}
	w = public("responder", fmt.Sprintf(`{"token":%q,"ratings":[1,2,3,4,5],"comment":%q}`, token, strings.Repeat("a", 1201)))
	if w.Code != 422 {
		t.Fatal("long comment", w.Code)
	}
	memory.v.ExpiresAt = time.Now().Add(-time.Second)
	w = public("validar", fmt.Sprintf(`{"token":%q}`, token))
	if w.Code != 410 {
		t.Fatal("expiry")
	}
	memory.v.ExpiresAt = time.Now().Add(time.Hour)
	w = public("validar", `{"token":"invalid"}`)
	if w.Code != 404 {
		t.Fatal("invalid token")
	}
	// Competing HTTP submissions, with an atomic in-memory persistence double.
	// Real MongoDB atomicity is tested separately in the opt-in QA integration.
	var wg sync.WaitGroup
	statuses := make(chan int, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := httptest.NewRequest("POST", "/api/v1/encuestas/publica/responder", strings.NewReader(fmt.Sprintf(`{"token":%q,"ratings":[1,2,3,4,5],"comment":"Gracias"}`, token)))
			r.RemoteAddr = fmt.Sprintf("198.51.100.%d:1234", i)
			r.Header.Set("Origin", "https://workshop.test")
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			statuses <- w.Code
		}(i)
	}
	wg.Wait()
	close(statuses)
	wins := 0
	for code := range statuses {
		if code == 201 {
			wins++
		} else if code != 409 {
			t.Fatal("unexpected concurrent response", code)
		}
	}
	if wins != 1 {
		t.Fatal("duplicate accepted response", wins)
	}
	w = public("responder", fmt.Sprintf(`{"token":%q,"ratings":[1,2,3,4,5]}`, token))
	if w.Code != 409 {
		t.Fatal("second response")
	}
	w = call("GET", "/api/v1/encuestas", "", true)
	if w.Code != 200 || strings.Contains(w.Body.String(), "Gracias") || strings.Contains(w.Body.String(), "ratings") {
		t.Fatal("tracking leak", w.Body.String())
	}
	w = call("GET", "/api/v1/encuestas/estadisticas", "", true)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"average":5`) {
		t.Fatal("stats", w.Body.String())
	}
	for _, perms := range [][]string{nil, {"ordenes.ver"}, {"pagos.ver"}} {
		sql.user.Permissions = perms
		for _, path := range []string{"/api/v1/encuestas/estadisticas"} {
			if w = call("GET", path, "", true); w.Code != 403 {
				t.Fatal("stats authorization", perms, w.Code)
			}
		}
		if w = call("POST", "/api/v1/ordenes/1/encuesta", `{}`, true); w.Code != 403 {
			t.Fatal("issue authorization", perms, w.Code)
		}
	}
	sql.user.Permissions = []string{"reportes.ver"}
	if w = call("GET", "/api/v1/encuestas/estadisticas", "", true); w.Code != 200 {
		t.Fatal("effective reporting permission", w.Code)
	}
	if w = call("GET", "/api/v1/encuestas", "", true); w.Code != 403 {
		t.Fatal("tracking permission", w.Code)
	}
	sql.user.Permissions = []string{"pagos.entregar"}
	rNoCSRF := httptest.NewRequest("POST", "/api/v1/ordenes/1/encuesta", strings.NewReader(`{}`))
	rNoCSRF.AddCookie(cookie)
	rNoCSRF.Header.Set("Origin", "https://workshop.test")
	rNoCSRF.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rNoCSRF)
	if w.Code != 403 {
		t.Fatal("administrative CSRF")
	}
	for i := 0; i < 31; i++ {
		r := httptest.NewRequest("POST", "/api/v1/encuestas/publica/validar", strings.NewReader(`{"token":"invalid"}`))
		r.RemoteAddr = "203.0.113.100:2345"
		r.Header.Set("Origin", "https://workshop.test")
		r.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if i == 30 && w.Code != 429 {
			t.Fatal("public rate limit", w.Code)
		}
	}
	for _, path := range []string{"/demo/", "/demo/index.html", "/demo/assets/"} {
		if w = call("GET", path, "", false); w.Code != 404 {
			t.Fatal("demo enabled")
		}
	}
	w = call("GET", "/encuesta", "", false)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Tu opinión nos importa") || !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Fatal("public page")
	}
	r := httptest.NewRequest("POST", "/api/v1/encuestas/publica/validar", strings.NewReader(`{}`))
	r.Header.Set("Origin", "https://evil.test")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross origin")
	}
}

// A real local reverse proxy, with no external connections or database access.
func TestSurveyReverseProxyBoundary(t *testing.T) {
	sql := &surveyTestSQL{stubStore: stubStore{user: domain.User{ID: "1", Active: true}}}
	h, err := New(sql, Options{Origin: "https://workshop.test", Key: []byte(strings.Repeat("k", 32)), Surveys: &surveyMemory{}})
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(h)
	defer upstream.Close()
	target, _ := url.Parse(upstream.URL)
	proxy := httptest.NewServer(httputil.NewSingleHostReverseProxy(target))
	defer proxy.Close()
	send := func(path, origin string, i int) *http.Response {
		req, err := http.NewRequest("POST", proxy.URL+path, strings.NewReader(`{"token":"invalid"}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", origin)
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", i))
		req.Header.Set("X-Real-IP", fmt.Sprintf("192.0.2.%d", i))
		res, err := proxy.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res
	}
	if res := send("/api/v1/encuestas/publica/validar", "https://evil.test", 0); res.StatusCode != 403 {
		t.Fatal("proxy lost origin protection", res.StatusCode)
	}
	for i := 0; i < 31; i++ {
		res := send("/api/v1/encuestas/publica/validar", "https://workshop.test", i)
		expected := 404
		if i == 30 {
			expected = 429
		}
		if res.StatusCode != expected {
			t.Fatalf("request=%d status=%d expected=%d", i, res.StatusCode, expected)
		}
		if i == 30 && res.Header.Get("Retry-After") != "60" {
			t.Fatal("retry guidance missing")
		}
	}
	// Both public operations share their budget; changing spoofed IP headers cannot
	// bypass it. The login rate limiter is independent of the survey budget.
	if res := send("/api/v1/encuestas/publica/responder", "https://workshop.test", 99); res.StatusCode != 429 {
		t.Fatal("submission bypassed proxy limit")
	}
	if res := send("/api/v1/auth/login", "https://workshop.test", 99); res.StatusCode == 429 {
		t.Fatal("survey abuse consumed login budget")
	}
	t.Log("Proxy clients share the connection-IP budget: 30/min total, not 30/min per forwarded IP; no trusted-proxy configuration added.")
}

func TestSurveyClientIPTrust(t *testing.T) {
	s := &server{options: Options{SurveyTrustedProxyIPs: []netip.Addr{netip.MustParseAddr("172.19.0.3"), netip.MustParseAddr("::1")}}}
	for _, tc := range []struct {
		peer      string
		forwarded []string
		want      string
	}{
		{"172.19.0.3:1234", []string{"198.51.100.1"}, "198.51.100.1"},
		{"172.19.0.4:1234", []string{"198.51.100.1"}, "172.19.0.4"},
		{"[::1]:1234", []string{"2001:db8::1"}, "2001:db8::1"},
		{"172.19.0.3:1234", []string{"::ffff:198.51.100.1"}, "198.51.100.1"},
		{"172.19.0.3:1234", []string{"198.51.100.1, 198.51.100.2"}, "172.19.0.3"},
		{"172.19.0.3:1234", []string{"198.51.100.1", "198.51.100.2"}, "172.19.0.3"},
		{"172.19.0.3:1234", []string{"unknown"}, "172.19.0.3"},
		{"172.19.0.3:1234", []string{"fe80::1%eth0"}, "172.19.0.3"},
		{"172.19.0.3:1234", []string{"0.0.0.0"}, "172.19.0.3"},
		{"172.19.0.3:1234", []string{"ff02::1"}, "172.19.0.3"},
		{"172.19.0.3:1234", nil, "172.19.0.3"},
	} {
		r := httptest.NewRequest("POST", "/", nil)
		r.RemoteAddr = tc.peer
		for _, value := range tc.forwarded {
			r.Header.Add("X-Forwarded-For", value)
		}
		r.Header.Set("X-Real-IP", "192.0.2.99")
		if got := s.surveyClientIP(r); got != tc.want {
			t.Fatalf("peer=%s forwarded=%v got=%s want=%s", tc.peer, tc.forwarded, got, tc.want)
		}
	}
}

func TestSurveyTrustedProxyBudgets(t *testing.T) {
	sql := &surveyTestSQL{stubStore: stubStore{user: domain.User{ID: "1", Active: true}}}
	h, err := New(sql, Options{Origin: "https://workshop.test", Key: []byte(strings.Repeat("k", 32)), Surveys: &surveyMemory{}, SurveyTrustedProxyIPs: []netip.Addr{netip.MustParseAddr("172.19.0.3")}})
	if err != nil {
		t.Fatal(err)
	}
	send := func(peer, client string) int {
		r := httptest.NewRequest("POST", "/api/v1/encuestas/publica/validar", strings.NewReader(`{"token":"invalid"}`))
		r.RemoteAddr = peer
		r.Header.Set("Origin", "https://workshop.test")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Forwarded-For", client)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	// Two legitimate clients behind the same verified peer have independent budgets.
	for i := 0; i < 30; i++ {
		for _, client := range []string{"198.51.100.1", "198.51.100.2"} {
			if got := send("172.19.0.3:1234", client); got != 404 {
				t.Fatalf("client %s request %d: %d", client, i, got)
			}
		}
	}
	if got := send("172.19.0.3:1234", "198.51.100.1"); got != 429 {
		t.Fatal("first client budget not enforced", got)
	}
	if got := send("172.19.0.3:1234", "198.51.100.3"); got != 404 {
		t.Fatal("unrelated client blocked", got)
	}
	// Direct requests from an untrusted peer cannot obtain new budgets by spoofing headers.
	for i := 0; i < 31; i++ {
		want := 404
		if i == 30 {
			want = 429
		}
		if got := send("192.0.2.1:1234", fmt.Sprintf("198.51.100.%d", i+1)); got != want {
			t.Fatalf("spoof request %d: got %d want %d", i, got, want)
		}
	}
}

func TestSurveySanitizingProxyRejectsSpoofedClients(t *testing.T) {
	sql := &surveyTestSQL{stubStore: stubStore{user: domain.User{ID: "1", Active: true}}}
	h, err := New(sql, Options{Origin: "https://workshop.test", Key: []byte(strings.Repeat("k", 32)), Surveys: &surveyMemory{}, SurveyTrustedProxyIPs: []netip.Addr{netip.MustParseAddr("127.0.0.1")}})
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(h)
	defer upstream.Close()
	target, _ := url.Parse(upstream.URL)
	// Exercise the edge's verified policy with real local HTTP sockets, rather
	// than accepting an arbitrary client-supplied forwarded identity.
	proxy := httptest.NewServer(&httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) { r.SetURL(target); r.SetXForwarded() }})
	defer proxy.Close()
	for i := 0; i < 31; i++ {
		r, err := http.NewRequest("POST", proxy.URL+"/api/v1/encuestas/publica/validar", strings.NewReader(`{"token":"invalid"}`))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Origin", "https://workshop.test")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", i+1))
		r.Header.Set("X-Real-IP", fmt.Sprintf("192.0.2.%d", i+1))
		res, err := proxy.Client().Do(r)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		want := 404
		if i == 30 {
			want = 429
		}
		if res.StatusCode != want {
			t.Fatalf("spoofed client request %d: %d want %d", i, res.StatusCode, want)
		}
	}
}

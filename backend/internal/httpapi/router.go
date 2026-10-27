package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/application"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/auth"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Options struct {
	Technical     application.Technical
	Origin        string
	Key           []byte
	SecureCookies bool
}
type server struct {
	store     application.Store
	options   Options
	tokens    *auth.Manager
	dummyHash []byte
	limiter   loginLimiter
}
type endpoint struct {
	method, path, permission string
	handler                  func(http.ResponseWriter, *http.Request, domain.User)
}
type loginWindow struct {
	count   int
	expires time.Time
}
type loginLimiter struct {
	mu      sync.Mutex
	entries map[string]loginWindow
}

func (l *loginLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for k, w := range l.entries {
		if !w.expires.After(now) {
			delete(l.entries, k)
		}
	}
	w := l.entries[ip]
	if w.expires.IsZero() {
		if len(l.entries) >= 1024 {
			return false
		}
		w.expires = now.Add(time.Minute)
	}
	w.count++
	l.entries[ip] = w
	return w.count <= 5
}
func New(store application.Store, opts Options) (http.Handler, error) {
	if store == nil || opts.Origin == "" {
		return nil, errors.New("store and origin are required")
	}
	tokens, err := auth.New(opts.Key)
	if err != nil {
		return nil, err
	}
	dummy, err := bcrypt.GenerateFromPassword([]byte("dummy-credential-never-an-account"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	s := &server{store: store, options: opts, tokens: tokens, dummyHash: dummy, limiter: loginLimiter{entries: map[string]loginWindow{}}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(r.Context()); err != nil {
			writeError(w, 503, "NOT_READY", "PostgreSQL no está disponible.")
			return
		}
		if opts.Technical != nil {
			if err := opts.Technical.Ping(r.Context()); err != nil {
				writeError(w, 503, "NOT_READY", "MongoDB técnico no está disponible o inicializado.")
				return
			}
		}
		respond(w, 200, map[string]any{"status": "ready", "technicalConfigured": opts.Technical != nil})
	})
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	implemented := map[string]bool{"POST /api/v1/auth/login": true}
	for _, e := range append(append(append(s.endpoints(), s.workshopEndpoints()...), s.relationalEndpoints()...), s.technicalEndpoints()...) {
		implemented[e.method+" "+e.path] = true
		mux.HandleFunc(e.method+" "+e.path, s.protect(e.permission, e.handler))
	}
	for _, route := range Routes {
		if implemented[route.Method+" "+route.Path] {
			continue
		}
		mux.HandleFunc(route.Method+" "+route.Path, s.protect("", func(w http.ResponseWriter, r *http.Request, u domain.User) {
			writeError(w, 501, "NOT_IMPLEMENTED", "Módulo pendiente de la siguiente fase.")
		}))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Frame-Options", "DENY")
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		mux.ServeHTTP(w, r.WithContext(ctx))
	}), nil
}
func (s *server) originOK(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Origin") != s.options.Origin || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		writeError(w, 403, "CSRF", "Origen no permitido.")
		return false
	}
	return true
}
func (s *server) protect(permission string, next func(http.ResponseWriter, *http.Request, domain.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("bravo_access")
		if err != nil {
			writeError(w, 401, "UNAUTHORIZED", "Inicia sesión.")
			return
		}
		claims, err := s.tokens.Parse(cookie.Value)
		if err != nil {
			writeError(w, 401, "UNAUTHORIZED", "Sesión inválida o vencida.")
			return
		}
		if _, err = domain.ID(claims.Subject); err != nil {
			writeError(w, 401, "UNAUTHORIZED", "Sesión inválida.")
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			if !s.originOK(w, r) {
				return
			}
			if !auth.ValidCSRF(claims.CSRF, r.Header.Get("X-CSRF-Token")) {
				writeError(w, 403, "CSRF", "Token CSRF inválido.")
				return
			}
		}
		u, err := s.store.User(r.Context(), claims.Subject)
		if errors.Is(err, domain.ErrNotFound) || (err == nil && !u.Active) {
			s.tokens.Revoke(claims.ID)
			writeError(w, 401, "UNAUTHORIZED", "Sesión inválida.")
			return
		}
		if err != nil {
			s.failure(w, err)
			return
		}
		if !auth.ValidCSRF(auth.CredentialVersion(u.PasswordHash), claims.CredentialVersion) {
			s.tokens.Revoke(claims.ID)
			writeError(w, 401, "UNAUTHORIZED", "Las credenciales cambiaron; inicia sesión nuevamente.")
			return
		}
		if permission != "" && !u.Can(permission) {
			writeError(w, 403, "FORBIDDEN", "Tu usuario no tiene este permiso.")
			return
		}
		if r.URL.Path == "/api/v1/auth/me" {
			respond(w, 200, map[string]any{"data": u, "csrfToken": claims.CSRF})
			return
		}
		if r.URL.Path == "/api/v1/auth/logout" {
			s.tokens.Revoke(claims.ID)
			s.cookie(w, "", -1)
			w.WriteHeader(204)
			return
		}
		next(w, r, u)
	}
}
func (s *server) cookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: "bravo_access", Value: value, Path: "/", HttpOnly: true, Secure: s.options.SecureCookies, SameSite: http.SameSiteStrictMode, MaxAge: maxAge})
}
func (s *server) login(w http.ResponseWriter, r *http.Request) {
	if !s.originOK(w, r) {
		return
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip == "" {
		ip = r.RemoteAddr
	}
	if !s.limiter.allow(ip) {
		w.Header().Set("Retry-After", "60")
		writeError(w, 429, "RATE_LIMITED", "Espera antes de volver a intentar.")
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Remember bool   `json:"remember"`
	}
	if !decode(w, r, &input) {
		return
	}
	if input.Username == "" || len(input.Username) > 600 || len(input.Password) == 0 || len(input.Password) > 72 {
		writeError(w, 401, "INVALID_CREDENTIALS", "Credenciales inválidas.")
		return
	}
	u, err := s.store.FindLogin(r.Context(), strings.TrimSpace(input.Username))
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		s.failure(w, err)
		return
	}
	hash := []byte(u.PasswordHash)
	if err != nil || !u.Active {
		hash = s.dummyHash
	}
	valid := bcrypt.CompareHashAndPassword(hash, []byte(input.Password)) == nil
	if !valid || err != nil || !u.Active {
		writeError(w, 401, "INVALID_CREDENTIALS", "Credenciales inválidas.")
		return
	}
	token, claims, err := s.tokens.Issue(u.ID, u.PasswordHash)
	if err != nil {
		s.failure(w, err)
		return
	}
	s.cookie(w, token, int(auth.TTL.Seconds()))
	respond(w, 200, map[string]any{"data": u, "csrfToken": claims.CSRF})
}
func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	respond(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func (s *server) failure(w http.ResponseWriter, err error) {
	var validation *domain.Validation
	switch {
	case errors.As(err, &validation):
		respond(w, 422, map[string]any{"error": map[string]any{"code": "VALIDATION", "message": validation.Message, "fields": map[string]string{validation.Field: validation.Message}}})
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, 404, "NOT_FOUND", "No se encontró el registro.")
	case errors.Is(err, domain.ErrConflict):
		writeError(w, 409, "CONFLICT", "Registro duplicado, referencias existentes o conflicto concurrente.")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, 403, "FORBIDDEN", "No tienes permiso sobre esta orden o asignación.")
	default:
		log.Print("API: fallo de persistencia/servicio (detalle omitido para proteger credenciales)")
		writeError(w, 503, "UNAVAILABLE", "Servicio temporalmente no disponible.")
	}
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	return decodeLimit(w, r, v, 64<<10)
}
func decodeLimit(w http.ResponseWriter, r *http.Request, v any, limit int64) bool {
	if strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]) != "application/json" {
		writeError(w, 415, "CONTENT_TYPE", "Usa application/json.")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		writeError(w, 400, "INVALID_JSON", "JSON inválido, demasiado grande o con campos desconocidos.")
		return false
	}
	if err := d.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "INVALID_JSON", "Envía un único objeto JSON.")
		return false
	}
	return true
}
func pathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if _, err := domain.ID(id); err != nil {
		writeError(w, 400, "INVALID_ID", "ID inválido.")
		return "", false
	}
	return id, true
}
func page(r *http.Request) (domain.Page, error) {
	p := domain.Page{Number: 1, Size: 20, Query: r.URL.Query().Get("q"), ClientID: r.URL.Query().Get("clientId")}
	for key, dst := range map[string]*int{"page": &p.Number, "pageSize": &p.Size} {
		if value := r.URL.Query().Get(key); value != "" {
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 || n > 1000000 {
				return p, domain.Invalid(key, "Paginación inválida.")
			}
			*dst = n
		}
	}
	if p.Size > 100 || len(p.Query) > 600 {
		return p, domain.Invalid("query", "Consulta o tamaño de página fuera del límite.")
	}
	if p.ClientID != "" {
		if _, err := domain.ID(p.ClientID); err != nil {
			return p, domain.Invalid("clientId", "Cliente inválido.")
		}
	}
	return p, nil
}

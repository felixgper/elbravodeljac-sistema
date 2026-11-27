package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/application"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/auth"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

func (s *server) mountSurveys(mux *http.ServeMux) {
	for _, e := range []endpoint{
		{"POST", "/api/v1/ordenes/{id}/encuesta", "pagos.entregar", s.issueSurvey},
		{"GET", "/api/v1/ordenes/{id}/encuesta", "pagos.ver", s.surveyStatus},
		{"GET", "/api/v1/encuestas", "pagos.ver", s.surveyTracking},
		{"GET", "/api/v1/encuestas/estadisticas", "reportes.ver", s.surveyStats},
	} {
		mux.HandleFunc(e.method+" "+e.path, s.protect(e.permission, e.handler))
	}
	mux.HandleFunc("POST /api/v1/encuestas/publica/validar", s.publicSurvey)
	mux.HandleFunc("POST /api/v1/encuestas/publica/responder", s.publicSurvey)
}
func (s *server) surveysAvailable(w http.ResponseWriter) bool {
	if s.options.Surveys == nil {
		writeError(w, 503, "UNAVAILABLE", "Encuestas no disponibles.")
		return false
	}
	return true
}
func (s *server) issueSurvey(w http.ResponseWriter, r *http.Request, u domain.User) {
	if !s.surveysAvailable(w) {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var input struct {
		Regenerate bool `json:"regenerate"`
	}
	if !decode(w, r, &input) {
		return
	}
	reader, ok := s.store.(application.SurveyDeliveryReader)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Entrega no verificable.")
		return
	}
	delivery, err := reader.SurveyDelivery(r.Context(), id)
	if err != nil {
		s.failure(w, err)
		return
	}
	if !input.Regenerate {
		old, readErr := s.options.Surveys.SurveyByOrder(r.Context(), id)
		if readErr == nil && old.Response == nil && old.ExpiresAt.After(time.Now()) && old.TokenCipher != "" {
			s.surveyLinkResult(w, 200, old)
			return
		}
		if readErr != nil && !errors.Is(readErr, domain.ErrNotFound) {
			s.failure(w, readErr)
			return
		}
	}
	token, hash, err := domain.NewSurveyToken()
	if err != nil {
		s.failure(w, err)
		return
	}
	idBytes := make([]byte, 16)
	if _, err = rand.Read(idBytes); err != nil {
		s.failure(w, err)
		return
	}
	now := time.Now().UTC()
	v := domain.SurveyInvitation{ID: hex.EncodeToString(idBytes), OrderID: delivery.OrderID, DeliveryID: delivery.ID, TokenHash: hash, CreatedAt: now, ExpiresAt: now.Add(domain.SurveyLifetime), Version: domain.SurveyVersion}
	v.TokenCipher, err = auth.SealSurveyLink(s.options.Key, token, surveyLinkBinding(v))
	if err != nil {
		s.failure(w, err)
		return
	}
	if err = s.options.Surveys.IssueSurvey(r.Context(), v, input.Regenerate); err != nil {
		if !input.Regenerate && errors.Is(err, domain.ErrConflict) {
			old, e := s.options.Surveys.SurveyByOrder(r.Context(), id)
			if e == nil && old.Response == nil && old.ExpiresAt.After(time.Now()) && old.TokenCipher != "" {
				s.surveyLinkResult(w, 200, old)
				return
			}
		}

		s.failure(w, err)
		return
	}
	// Read back the immutable ID and authenticated ciphertext; never log secrets.
	v, err = s.options.Surveys.SurveyByOrder(r.Context(), id)
	if err != nil {
		s.failure(w, err)
		return
	}
	if v.TokenHash != hash {
		s.failure(w, domain.ErrConflict)
		return
	}
	s.surveyLinkResult(w, 201, v)
}
func surveyLinkBinding(v domain.SurveyInvitation) string {
	return "v1:" + v.OrderID + ":" + v.DeliveryID + ":" + v.TokenHash
}
func (s *server) surveyLinkResult(w http.ResponseWriter, status int, v domain.SurveyInvitation) {
	token, err := auth.OpenSurveyLink(s.options.Key, v.TokenCipher, surveyLinkBinding(v))
	if err != nil || domain.SurveyTokenHash(token) != v.TokenHash {
		s.failure(w, errors.New("survey link cannot be recovered"))
		return
	}
	respond(w, status, map[string]any{"data": v.PublicStatus(time.Now()), "url": strings.TrimRight(s.options.Origin, "/") + "/encuesta#t=" + token})
}
func (s *server) surveyStatus(w http.ResponseWriter, r *http.Request, u domain.User) {
	if !s.surveysAvailable(w) {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := s.options.Surveys.SurveyByOrder(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		respond(w, 200, map[string]any{"data": nil})
		return
	}
	if err != nil {
		s.failure(w, err)
		return
	}
	if u.Can("pagos.entregar") && v.Response == nil && v.ExpiresAt.After(time.Now()) && v.TokenCipher != "" {
		s.surveyLinkResult(w, 200, v)
		return
	}
	respond(w, 200, map[string]any{"data": v.PublicStatus(time.Now())})
}
func (s *server) surveyTracking(w http.ResponseWriter, r *http.Request, u domain.User) {
	if !s.surveysAvailable(w) {
		return
	}
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 100000 {
			writeError(w, 400, "INVALID_PAGE", "Página inválida.")
			return
		}
		page = v
	}
	rows, err := s.options.Surveys.SurveyTracking(r.Context(), page)
	if err != nil {
		s.failure(w, err)
		return
	}
	for i := range rows {
		rows[i] = rows[i].PublicStatus(time.Now())
	}
	respond(w, 200, map[string]any{"data": rows, "page": page, "pageSize": 50})
}
func (s *server) surveyStats(w http.ResponseWriter, r *http.Request, u domain.User) {
	if !s.surveysAvailable(w) {
		return
	}
	start, end, err := statsRange(r, true)
	if err != nil {
		s.failure(w, err)
		return
	}
	v, err := s.options.Surveys.SurveyStatistics(r.Context(), start, end)
	if err != nil {
		s.failure(w, err)
		return
	}
	respond(w, 200, map[string]any{"data": v})
}
func (s *server) publicSurvey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	if !s.originOK(w, r) || !s.surveysAvailable(w) {
		return
	}
	// Forwarded IPs are accepted only from explicitly configured proxy peers.
	// This budget remains separate from the existing login limiter.
	if !s.surveyLimiter.allowN(s.surveyClientIP(r), 30) {
		w.Header().Set("Retry-After", "60")
		writeError(w, 429, "RATE_LIMITED", "Espera un minuto antes de volver a intentar.")
		return
	}
	var input struct {
		Token   string `json:"token"`
		Ratings []int  `json:"ratings,omitempty"`
		Comment string `json:"comment,omitempty"`
	}
	if !decodeLimit(w, r, &input, 8<<10) {
		return
	}
	hash := domain.SurveyTokenHash(input.Token)
	if hash == "" {
		writeError(w, 404, "INVALID_LINK", "Enlace de encuesta no válido.")
		return
	}
	v, err := s.options.Surveys.SurveyByHash(r.Context(), hash)
	if err != nil {
		s.failure(w, err)
		return
	}
	if v.Response != nil {
		writeError(w, 409, "SURVEY_USED", "Esta encuesta ya fue respondida. ¡Gracias!")
		return
	}
	if !v.ExpiresAt.After(time.Now()) {
		writeError(w, 410, "SURVEY_EXPIRED", "El enlace ha vencido. Solicita uno nuevo al taller.")
		return
	}
	if strings.HasSuffix(r.URL.Path, "/validar") {
		respond(w, 200, map[string]any{"data": map[string]any{"version": v.Version, "expiresAt": v.ExpiresAt}})
		return
	}
	if len(input.Ratings) != 5 {
		writeError(w, 422, "VALIDATION", "Responde exactamente cinco preguntas.")
		return
	}
	answer := domain.SurveyResponse{Comment: strings.TrimSpace(input.Comment), At: time.Now().UTC()}
	copy(answer.Ratings[:], input.Ratings)
	if err = answer.Validate(); err != nil {
		s.failure(w, err)
		return
	}
	if err = s.options.Surveys.AnswerSurvey(r.Context(), hash, answer); err != nil {
		s.failure(w, err)
		return
	}
	respond(w, 201, map[string]any{"data": map[string]bool{"accepted": true}})
}

// The current Caddy edge replaces incoming X-Forwarded-For with one client IP.
// Chains, duplicate headers and malformed addresses fail closed to the peer;
// accepting arbitrary forwarded lists would allow client-controlled identities.
func (s *server) surveyClientIP(r *http.Request) string {
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	ip, err := netip.ParseAddr(peer)
	if err != nil {
		return peer
	}
	ip = ip.Unmap()
	for _, trusted := range s.options.SurveyTrustedProxyIPs {
		if trusted.Unmap() != ip {
			continue
		}
		values := r.Header.Values("X-Forwarded-For")
		if len(values) != 1 {
			break
		}
		client, err := netip.ParseAddr(strings.TrimSpace(values[0]))
		if err == nil && client.Zone() == "" && !client.IsUnspecified() && !client.IsMulticast() {
			return client.Unmap().String()
		}
		break
	}
	return ip.String()
}

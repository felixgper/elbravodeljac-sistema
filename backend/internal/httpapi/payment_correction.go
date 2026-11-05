package httpapi

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
)

type correctingStore interface {
	ReversePayment(context.Context, string, string, string, string, *domain.PaymentReplacement, domain.User) (any, error)
}

func (s *server) reversePayment(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	payment := r.PathValue("paymentId")
	if _, err := domain.ID(payment); err != nil {
		s.failure(w, err)
		return
	}
	var v struct {
		RequestID   string                     `json:"requestId"`
		Reason      string                     `json:"reason"`
		Replacement *domain.PaymentReplacement `json:"replacement"`
	}
	if !decode(w, r, &v) {
		return
	}
	store, ok := s.store.(correctingStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Rectificación no disponible.")
		return
	}
	out, err := store.ReversePayment(r.Context(), id, payment, v.RequestID, v.Reason, v.Replacement, u)
	s.result(w, 200, out, err)
}

type attentionHistoryStore interface {
	AttentionHistory(context.Context, string, string, domain.User) (any, error)
}

func (s *server) clientHistory(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.Client(r.Context(), id); err != nil {
		s.failure(w, err)
		return
	}
	store, ok := s.store.(attentionHistoryStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Historial no disponible.")
		return
	}
	out, err := store.AttentionHistory(r.Context(), id, "", u)
	s.result(w, 200, out, err)
}

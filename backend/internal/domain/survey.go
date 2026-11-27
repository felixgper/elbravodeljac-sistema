package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const SurveyVersion = 1
const SurveyLifetime = 14 * 24 * time.Hour
const SurveyCommentLimit = 1200

var ErrSurveyExpired = errors.New("survey expired")

type SurveyResponse struct {
	Ratings [5]int    `json:"ratings" bson:"ratings"`
	Comment string    `json:"comment" bson:"comment"`
	At      time.Time `json:"at" bson:"at"`
}

func (v SurveyResponse) Validate() error {
	for _, n := range v.Ratings {
		if n < 1 || n > 5 {
			return Invalid("ratings", "Responde las cinco preguntas con estrellas enteras entre 1 y 5.")
		}
	}
	if !utf8.ValidString(v.Comment) || utf8.RuneCountInString(v.Comment) > SurveyCommentLimit {
		return Invalid("comment", "El comentario admite hasta 1200 caracteres.")
	}
	return nil
}

// Response is embedded so consumption and persistence are a single-document
// atomic write, including on standalone MongoDB (no distributed transaction).
type SurveyInvitation struct {
	ID          string          `json:"id" bson:"_id"`
	DeliveryID  string          `json:"deliveryId" bson:"deliveryId"`
	OrderID     string          `json:"orderId" bson:"orderId"`
	TokenCipher string          `json:"-" bson:"tokenCipher,omitempty"`
	TokenHash   string          `json:"-" bson:"tokenHash"`
	CreatedAt   time.Time       `json:"createdAt" bson:"createdAt"`
	ExpiresAt   time.Time       `json:"expiresAt" bson:"expiresAt"`
	Version     int             `json:"version" bson:"version"`
	Response    *SurveyResponse `json:"-" bson:"response,omitempty"`
	AnsweredAt  *time.Time      `json:"answeredAt,omitempty" bson:"-"`
	Status      string          `json:"status" bson:"-"`
}

func (v SurveyInvitation) PublicStatus(now time.Time) SurveyInvitation {
	v.Status = "generada"
	if v.Response != nil {
		v.Status = "respondida"
		at := v.Response.At
		v.AnsweredAt = &at
	} else if !v.ExpiresAt.After(now) {
		v.Status = "vencida"
	}
	v.Response = nil
	return v
}

type SurveyDelivery struct{ ID, OrderID string }
type SurveyStats struct {
	Total        int        `json:"total"`
	Average      float64    `json:"average"`
	Aspects      [5]float64 `json:"aspects"`
	Distribution [5]int     `json:"distribution"`
}

func NewSurveyToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	return token, SurveyTokenHash(token), nil
}
func SurveyTokenHash(token string) string {
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(b) != 32 || strings.TrimSpace(token) != token || base64.RawURLEncoding.EncodeToString(b) != token {
		return ""
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"sync"
	"time"
)

const TTL = 15 * time.Minute
const Issuer = "elbravodeljac-api"
const Audience = "elbravodeljac-admin"

type Claims struct {
	CSRF string `json:"csrf"`
	jwt.RegisteredClaims
}

// Allowlist in memory: logout revokes immediately; restart invalidates sessions.
// A multi-instance deployment needs shared session storage.
type Manager struct {
	key      []byte
	mu       sync.Mutex
	sessions map[string]time.Time
}

func New(key []byte) (*Manager, error) {
	if len(key) < 32 {
		return nil, errors.New("JWT signing key requires 32 random bytes")
	}
	return &Manager{key: append([]byte(nil), key...), sessions: map[string]time.Time{}}, nil
}
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func (m *Manager) Issue(subject string) (string, *Claims, error) {
	jti, err := randomToken()
	if err != nil {
		return "", nil, err
	}
	csrf, err := randomToken()
	if err != nil {
		return "", nil, err
	}
	now := time.Now()
	c := &Claims{CSRF: csrf, RegisteredClaims: jwt.RegisteredClaims{Issuer: Issuer, Audience: jwt.ClaimStrings{Audience}, Subject: subject, ID: jti, IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(TTL))}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.key)
	if err != nil {
		return "", nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, at := range m.sessions {
		if !at.After(now) {
			delete(m.sessions, id)
		}
	}
	if len(m.sessions) >= 10000 {
		return "", nil, errors.New("session capacity reached")
	}
	m.sessions[jti] = c.ExpiresAt.Time
	return token, c, nil
}
func (m *Manager) Parse(raw string) (*Claims, error) {
	c := &Claims{}
	token, err := jwt.ParseWithClaims(raw, c, func(t *jwt.Token) (any, error) { return m.key, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(Issuer), jwt.WithAudience(Audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid || c.Subject == "" || c.ID == "" || c.CSRF == "" || c.IssuedAt == nil || c.NotBefore == nil {
		return nil, errors.New("invalid session")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	expiry, ok := m.sessions[c.ID]
	if !ok || !expiry.After(time.Now()) {
		delete(m.sessions, c.ID)
		return nil, errors.New("expired or revoked session")
	}
	return c, nil
}
func (m *Manager) Revoke(id string) { m.mu.Lock(); delete(m.sessions, id); m.mu.Unlock() }
func ValidCSRF(expected, actual string) bool {
	return actual != "" && subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

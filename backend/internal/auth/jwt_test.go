package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"testing"
	"time"
)

func TestJWTValidationAndRevocation(t *testing.T) {
	key := []byte(strings.Repeat("k", 32))
	m, _ := New(key)
	raw, c, err := m.Issue("1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Parse(raw); err != nil {
		t.Fatal(err)
	}
	if !ValidCSRF(c.CSRF, c.CSRF) || ValidCSRF(c.CSRF, "") || ValidCSRF(c.CSRF, "wrong") {
		t.Fatal("CSRF binding")
	}
	other, _ := New([]byte(strings.Repeat("z", 32)))
	if _, err = other.Parse(raw); err == nil {
		t.Fatal("wrong key")
	}
	for _, mutation := range []string{"expired", "issuer", "audience", "algorithm", "unsigned", "missing-exp"} {
		copy := *c
		method := jwt.SigningMethodHS256
		switch mutation {
		case "expired":
			copy.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
		case "issuer":
			copy.Issuer = "evil"
		case "audience":
			copy.Audience = jwt.ClaimStrings{"evil"}
		case "missing-exp":
			copy.ExpiresAt = nil
		}
		var bad string
		if mutation == "unsigned" {
			bad, _ = jwt.NewWithClaims(jwt.SigningMethodNone, &copy).SignedString(jwt.UnsafeAllowNoneSignatureType)
		} else if mutation == "algorithm" {
			bad, _ = jwt.NewWithClaims(jwt.SigningMethodHS384, &copy).SignedString(key)
		} else {
			bad, _ = jwt.NewWithClaims(method, &copy).SignedString(key)
		}
		if _, err = m.Parse(bad); err == nil {
			t.Fatal("accepted", mutation)
		}
	}
	m.Revoke(c.ID)
	if _, err = m.Parse(raw); err == nil {
		t.Fatal("revocation failed")
	}
	restarted, _ := New(key)
	if _, err = restarted.Parse(raw); err == nil {
		t.Fatal("restart must invalidate sessions")
	}
}

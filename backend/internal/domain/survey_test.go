package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSurveyValidationAndSecrets(t *testing.T) {
	valid := SurveyResponse{Ratings: [5]int{1, 2, 3, 4, 5}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		for _, n := range []int{-1, 0, 6} {
			v := valid
			v.Ratings[i] = n
			if v.Validate() == nil {
				t.Fatalf("accepted %d in question %d", n, i)
			}
		}
	}
	valid.Comment = strings.Repeat("á", 1200)
	if valid.Validate() != nil {
		t.Fatal("unicode limit")
	}
	valid.Comment += "a"
	if valid.Validate() == nil {
		t.Fatal("comment overflow")
	}
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		token, hash, err := NewSurveyToken()
		if err != nil || len(token) != 43 || len(hash) != 64 || seen[token] || SurveyTokenHash(token) != hash {
			t.Fatal("unsafe token")
		}
		seen[token] = true
	}
	for _, token := range []string{"", strings.Repeat("a", 42), strings.Repeat("!", 43), strings.Repeat("a", 43) + "="} {
		if SurveyTokenHash(token) != "" {
			t.Fatal("invalid token accepted")
		}
	}
	now := time.Now()
	v := SurveyInvitation{TokenHash: "secret-hash", ExpiresAt: now.Add(time.Hour), Response: &SurveyResponse{Comment: "private-comment", At: now}}
	out := v.PublicStatus(now)
	b, _ := json.Marshal(out)
	if out.Status != "respondida" || out.AnsweredAt == nil || strings.Contains(string(b), "secret") || strings.Contains(string(b), "private") || strings.Contains(string(b), "ratings") {
		t.Fatal(string(b))
	}
	v.Response = nil
	v.ExpiresAt = now
	if v.PublicStatus(now).Status != "vencida" {
		t.Fatal("expiry boundary")
	}
	v.ExpiresAt = now.Add(time.Second)
	if v.PublicStatus(now).Status != "generada" {
		t.Fatal("status")
	}
}

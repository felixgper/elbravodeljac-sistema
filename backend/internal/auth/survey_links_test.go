package auth

import (
	"strings"
	"testing"
)

func TestSurveyLinkEncryption(t *testing.T) {
	key := []byte(strings.Repeat("k", 32))
	binding := "v1:order:delivery:hash"
	token := strings.Repeat("a", 43)
	sealed, err := SealSurveyLink(key, token, binding)
	if err != nil || strings.Contains(sealed, token) {
		t.Fatal("token encryption failed")
	}
	opened, err := OpenSurveyLink(append([]byte(nil), key...), sealed, binding)
	if err != nil || opened != token {
		t.Fatal("restored key cannot recover link")
	}
	for _, tc := range []struct {
		key            []byte
		value, binding string
	}{{[]byte(strings.Repeat("x", 32)), sealed, binding}, {key, sealed, binding + "tampered"}, {key, sealed[:20], binding}, {key, "!", binding}, {nil, sealed, binding}} {
		if _, err := OpenSurveyLink(tc.key, tc.value, tc.binding); err == nil {
			t.Fatal("unauthenticated recovery accepted")
		}
	}
	other, _ := SealSurveyLink(key, token, binding)
	if other == sealed {
		t.Fatal("nonce reused")
	}
}

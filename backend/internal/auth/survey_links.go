package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// Purpose-separated encryption key; never use a JWT directly as a survey link.
// Preserve the application's private key when backing up/restoring invitations.
func surveyCipher(key []byte) (cipher.AEAD, error) {
	if len(key) < 32 {
		return nil, errors.New("survey encryption key unavailable")
	}
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte("elbravodeljac/survey-link/aes-gcm/v1"))
	b, err := aes.NewCipher(h.Sum(nil))
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(b)
}
func SealSurveyLink(key []byte, token, binding string) (string, error) {
	c, err := surveyCipher(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, c.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(c.Seal(nonce, nonce, []byte(token), []byte(binding))), nil
}
func OpenSurveyLink(key []byte, sealed, binding string) (string, error) {
	c, err := surveyCipher(key)
	if err != nil {
		return "", err
	}
	if len(sealed) > 256 {
		return "", errors.New("invalid survey ciphertext")
	}
	b, err := base64.RawURLEncoding.DecodeString(sealed)
	if err != nil || len(b) < c.NonceSize()+c.Overhead() {
		return "", errors.New("invalid survey ciphertext")
	}
	plain, err := c.Open(nil, b[:c.NonceSize()], b[c.NonceSize():], []byte(binding))
	if err != nil {
		return "", errors.New("survey ciphertext authentication failed")
	}
	return string(plain), nil
}

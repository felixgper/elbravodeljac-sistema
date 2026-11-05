package config

import (
	"encoding/base64"
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Addr, DSN, Origin string
	Key               []byte
	SecureCookies     bool
}

func ValidateDSN(dsn string) error {
	if dsn == "" {
		return errors.New("POSTGRES_DSN es obligatorio")
	}
	c, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return errors.New("POSTGRES_DSN inválido")
	}
	// Phase A intentionally refuses the real DB and all ambiguous database names.
	name := strings.ToLower(c.ConnConfig.Database)
	if name != "elbravodeljac_test" && name != "elbravodeljac_dev" {
		return errors.New("Fase A solo permite bases elbravodeljac_test o elbravodeljac_dev")
	}
	return nil
}
func Load() (Config, error) {
	c := Config{Addr: os.Getenv("HTTP_ADDR"), DSN: os.Getenv("POSTGRES_DSN"), Origin: os.Getenv("APP_ORIGIN"), SecureCookies: true}
	if c.Addr == "" {
		c.Addr = "127.0.0.1:8080"
	}
	if err := ValidateDSN(c.DSN); err != nil {
		return c, err
	}
	u, err := url.Parse(c.Origin)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return c, errors.New("APP_ORIGIN debe ser un origen exacto sin ruta")
	}
	c.Key, err = base64.StdEncoding.DecodeString(os.Getenv("JWT_SECRET_BASE64"))
	if err != nil || len(c.Key) < 32 {
		return c, errors.New("JWT_SECRET_BASE64 requiere al menos 32 bytes aleatorios codificados en base64")
	}
	if os.Getenv("COOKIE_SECURE") == "false" {
		if u.Scheme != "http" || (u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1") {
			return c, errors.New("Cookies sin Secure solo se permiten en HTTP localhost")
		}
		c.SecureCookies = false
	}
	if c.SecureCookies && u.Scheme != "https" {
		return c, errors.New("Cookies Secure requieren APP_ORIGIN HTTPS; para pruebas locales configura COOKIE_SECURE=false")
	}
	return c, nil
}

package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrForbidden = errors.New("forbidden")

type Validation struct{ Field, Message string }

func (e *Validation) Error() string       { return e.Field + ": " + e.Message }
func Invalid(field, message string) error { return &Validation{field, message} }

type User struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	Active       bool     `json:"active"`
	Roles        []string `json:"roles"`
	Permissions  []string `json:"permissions"`
	PasswordHash string   `json:"-"`
}

func (u User) Can(code string) bool {
	for _, permission := range u.Permissions {
		if permission == code {
			return true
		}
	}
	return false
}

type ClientInput struct {
	Name           string `json:"name"`
	DocumentType   string `json:"documentType"`
	DocumentNumber string `json:"documentNumber"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	Email          string `json:"email"`
}
type Client struct {
	ID string `json:"id"`
	ClientInput
}
type ContactInput struct {
	Name           string `json:"name"`
	DocumentType   string `json:"documentType"`
	DocumentNumber string `json:"documentNumber"`
	Phone          string `json:"phone"`
	Relation       string `json:"relation"`
}
type Contact struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
	Active   bool   `json:"active"`
	ContactInput
}
type VehicleInput struct {
	ClientID string `json:"clientId"`
	Plate    string `json:"plate"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
	Mileage  string `json:"mileage"`
	Notes    string `json:"notes"`
}

// PATCH deliberately excludes plate and clientId; transfer has a separate command.
type VehiclePatch struct {
	Brand   *string `json:"brand"`
	Model   *string `json:"model"`
	Year    *int    `json:"year"`
	Mileage *string `json:"mileage"`
	Notes   *string `json:"notes"`
}
type Vehicle struct {
	ID string `json:"id"`
	VehicleInput
}
type Transfer struct {
	NewClientID string `json:"newClientId"`
	Verified    bool   `json:"transferVerified"`
	Observation string `json:"observation"`
}
type Page struct {
	Number, Size    int
	Query, ClientID string
}
type List[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}
type Meta struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

func ID(value string) (int64, error) {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 || strconv.FormatInt(n, 10) != value {
		return 0, Invalid("id", "Debe ser un ID decimal positivo.")
	}
	return n, nil
}

var decimal = regexp.MustCompile(`^(0|[1-9][0-9]{0,7})(\.[0-9]{1,2})?$`)

// Kilometraje NUMERIC(10,2), represented as integer hundredths without floats.
func Mileage(value string) (int64, error) {
	if !decimal.MatchString(value) {
		return 0, Invalid("mileage", "Usa un decimal entre 0.00 y 99999999.99, con hasta dos decimales.")
	}
	p := strings.SplitN(value, ".", 2)
	whole, _ := strconv.ParseInt(p[0], 10, 64)
	fraction := "00"
	if len(p) == 2 {
		fraction = (p[1] + "0")[:2]
	}
	part, _ := strconv.ParseInt(fraction, 10, 64)
	return whole*100 + part, nil
}
func FormatMileage(n int64) string { return fmt.Sprintf("%d.%02d", n/100, n%100) }
func text(field, value string, max int, required bool) error {
	if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		return Invalid(field, "Texto inválido.")
	}
	if required && strings.TrimSpace(value) == "" {
		return Invalid(field, "Campo obligatorio.")
	}
	if utf8.RuneCountInString(value) > max {
		return Invalid(field, fmt.Sprintf("Máximo %d caracteres; no se truncará el valor.", max))
	}
	return nil
}
func documentAndPhone(kind, number, phone string) error {
	if kind != "DNI" && kind != "RUC" && kind != "CE" && kind != "Pasaporte" {
		return Invalid("documentType", "Tipo de documento inválido.")
	}
	if !regexp.MustCompile(`^[A-Z0-9]{5,20}$`).MatchString(number) || (kind == "DNI" && !regexp.MustCompile(`^[0-9]{8}$`).MatchString(number)) || (kind == "RUC" && !regexp.MustCompile(`^[0-9]{11}$`).MatchString(number)) {
		return Invalid("documentNumber", "Número de documento inválido para este tipo.")
	}
	if !regexp.MustCompile(`^\+?[0-9]{7,15}$`).MatchString(phone) {
		return Invalid("phone", "Usa un teléfono de 7 a 15 dígitos, con + opcional.")
	}
	return nil
}
func normalizePhone(value string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '(' || r == ')' || r == '-' {
			return -1
		}
		return r
	}, strings.TrimSpace(value))
}
func (v *ClientInput) Validate() error {
	v.DocumentNumber = strings.ToUpper(strings.TrimSpace(v.DocumentNumber))
	v.Phone = normalizePhone(v.Phone)
	v.Email = strings.ToLower(strings.TrimSpace(v.Email))
	for _, e := range []error{text("name", v.Name, 150, true), text("documentType", v.DocumentType, 20, true), text("documentNumber", v.DocumentNumber, 20, true), text("phone", v.Phone, 20, true), text("email", v.Email, 150, false), text("address", v.Address, 10000, true)} {
		if e != nil {
			return e
		}
	}
	if err := documentAndPhone(v.DocumentType, v.DocumentNumber, v.Phone); err != nil {
		return err
	}
	if v.Email != "" && !regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`).MatchString(v.Email) {
		return Invalid("email", "Correo inválido.")
	}
	return nil
}
func (v *ContactInput) Validate() error {
	v.DocumentNumber = strings.ToUpper(strings.TrimSpace(v.DocumentNumber))
	v.Phone = normalizePhone(v.Phone)
	for _, e := range []error{text("name", v.Name, 150, true), text("documentType", v.DocumentType, 20, true), text("documentNumber", v.DocumentNumber, 20, true), text("phone", v.Phone, 20, true), text("relation", v.Relation, 50, true)} {
		if e != nil {
			return e
		}
	}
	return documentAndPhone(v.DocumentType, v.DocumentNumber, v.Phone)
}
func (v *VehicleInput) NormalizeValidate() error {
	v.Plate = strings.ToUpper(strings.Join(strings.Fields(v.Plate), ""))
	if _, err := ID(v.ClientID); err != nil {
		return Invalid("clientId", "Cliente inválido.")
	}
	for _, e := range []error{text("plate", v.Plate, 12, true), text("brand", v.Brand, 40, true), text("model", v.Model, 60, true), text("notes", v.Notes, 10000, false)} {
		if e != nil {
			return e
		}
	}
	if len(v.Plate) < 5 || !regexp.MustCompile(`^[A-Z0-9]+(-[A-Z0-9]+)*$`).MatchString(v.Plate) {
		return Invalid("plate", "Placa inválida.")
	}
	if v.Year < 1950 || v.Year > time.Now().Year()+1 || v.Year > 2100 {
		return Invalid("year", "Año fuera del rango permitido.")
	}
	n, err := Mileage(v.Mileage)
	if err != nil {
		return err
	}
	v.Mileage = FormatMileage(n)
	return nil
}
func (v VehiclePatch) Apply(old VehicleInput) VehicleInput {
	if v.Brand != nil {
		old.Brand = *v.Brand
	}
	if v.Model != nil {
		old.Model = *v.Model
	}
	if v.Year != nil {
		old.Year = *v.Year
	}
	if v.Mileage != nil {
		old.Mileage = *v.Mileage
	}
	if v.Notes != nil {
		old.Notes = *v.Notes
	}
	return old
}
func (v Transfer) Validate() error {
	if _, err := ID(v.NewClientID); err != nil {
		return Invalid("newClientId", "Cliente inválido.")
	}
	if !v.Verified {
		return Invalid("transferVerified", "La transferencia debe estar verificada.")
	}
	return text("observation", v.Observation, 4000, false)
}

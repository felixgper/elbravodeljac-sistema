package technicalsheet

import (
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

//go:embed assets/*
var assets embed.FS

type Item struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}
type Group struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Items   []Item `json:"items"`
	Options []Item `json:"options"`
}
type Catalogue struct {
	Groups        []Group `json:"groups"`
	Maintenance   []Group `json:"maintenance"`
	OtherServices []Item  `json:"otherServices"`
}

var Catalog Catalogue

func init() {
	b, _ := assets.ReadFile("assets/catalogo.json")
	if err := json.Unmarshal(b, &Catalog); err != nil {
		panic(err)
	}
}

type Execution struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}
type Reconciliation struct {
	By     string `json:"by"`
	At     string `json:"at"`
	Hash   string `json:"hash"`
	Reason string `json:"reason"`
}
type Control struct {
	Execution        map[string]Execution `json:"execution"`
	Reconciliation   *Reconciliation      `json:"reconciliation,omitempty"`
	AdjustmentReason string               `json:"adjustmentReason,omitempty"`

	Version        int64               `json:"version"`
	Inspections    map[string]string   `json:"inspections"`
	Maintenance    map[string][]string `json:"maintenance"`
	OtherServices  []string            `json:"otherServices"`
	Tires          map[string]string   `json:"tires"`
	OtherFluids    string              `json:"otherFluids"`
	Coils          string              `json:"coils"`
	ClientNotes    string              `json:"clientNotes"`
	TechnicalNotes string              `json:"technicalNotes"`
	NextMileage    string              `json:"nextMileage"`
	NextDate       string              `json:"nextDate"`
	PerformedItems []string            `json:"performedItems"`
}

func (c Control) Validate() error {
	if len(c.Execution) > 100 {
		return fmt.Errorf("Máximo 100 trabajos por orden")
	}
	for _, entry := range c.Execution {
		if entry.Status != "Pendiente" && entry.Status != "Realizado" && entry.Status != "No realizado" {
			return fmt.Errorf("Estado de trabajo inválido")
		}
		if utf8.RuneCountInString(entry.Notes) > 1200 {
			return fmt.Errorf("Observaciones: máximo 1200 caracteres")
		}
		if entry.Status == "No realizado" && strings.TrimSpace(entry.Notes) == "" {
			return fmt.Errorf("Explica por qué no se realizó el trabajo")
		}
	}
	if c.Version < 0 {
		return fmt.Errorf("Versión inválida")
	}
	valid := map[string]bool{}
	for _, g := range Catalog.Groups {
		for _, i := range g.Items {
			valid[i.ID] = true
		}
	}
	for k, v := range c.Inspections {
		if !valid[k] || (v != "Bien" && v != "Atención" && v != "Urgente" && v != "Sin evaluar") {
			return fmt.Errorf("Inspección o estado inválido")
		}
	}
	choices := map[string]map[string]bool{}
	for _, g := range Catalog.Maintenance {
		choices[g.ID] = map[string]bool{}
		for _, i := range g.Options {
			choices[g.ID][i.ID] = true
		}
	}
	for k, values := range c.Maintenance {
		if choices[k] == nil {
			return fmt.Errorf("Categoría de mantenimiento inválida")
		}
		seen := map[string]bool{}
		for _, v := range values {
			if !choices[k][v] || seen[v] {
				return fmt.Errorf("Opción de mantenimiento inválida o duplicada")
			}
			seen[v] = true
		}
	}
	valid = map[string]bool{}
	for _, i := range Catalog.OtherServices {
		valid[i.ID] = true
	}
	seen := map[string]bool{}
	for _, k := range c.OtherServices {
		if !valid[k] || seen[k] {
			return fmt.Errorf("Servicio inválido o duplicado")
		}
		seen[k] = true
	}
	for k, v := range c.Tires {
		if k != "frontLeft" && k != "frontRight" && k != "rearLeft" && k != "rearRight" {
			return fmt.Errorf("Neumático inválido")
		}
		if v != "" {
			n, err := strconv.ParseFloat(v, 64)
			if err != nil || n < 0 || n > 150 || strings.ContainsAny(v, "eENaInf") {
				return fmt.Errorf("PSI debe ser un número entre 0 y 150")
			}
		}
	}
	for _, v := range []string{c.OtherFluids, c.Coils, c.ClientNotes, c.TechnicalNotes} {
		if utf8.RuneCountInString(v) > 1200 || strings.ContainsRune(v, 0) {
			return fmt.Errorf("Observaciones: máximo 1200 caracteres")
		}
	}
	if c.NextMileage != "" {
		n, err := strconv.ParseUint(c.NextMileage, 10, 32)
		if err != nil || n > 999999999 {
			return fmt.Errorf("Próximo kilometraje inválido")
		}
	}
	if c.NextDate != "" {
		if _, err := time.Parse("2006-01-02", c.NextDate); err != nil {
			return fmt.Errorf("Fecha recomendada inválida")
		}
	}
	if len(c.PerformedItems) > 100 {
		return fmt.Errorf("Demasiados trabajos realizados")
	}
	return nil
}

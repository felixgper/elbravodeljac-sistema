package technicalsheet

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Quotation is a read-only snapshot. Monetary values come from PostgreSQL;
// rendering never recomputes discounts, subtotals, or the authorized amount.
type Quotation struct {
	Workshop Workshop `json:"workshop"`
	Client   string   `json:"client"`
	Document string   `json:"document"`
	Phone    string   `json:"phone"`
	Email    string   `json:"email"`
	Address  string   `json:"address"`
	Plate    string   `json:"plate"`
	Vehicle  string   `json:"vehicle"`
	Quote    struct {
		ID            string `json:"id"`
		Version       int    `json:"version"`
		ParentQuoteID string `json:"parentQuoteId"`
		ReceptionID   string `json:"receptionId"`
		CreatedAt     string `json:"createdAt"`
		Status        string `json:"status"`
		Total         string `json:"total"`
		Approval      *struct {
			Name    string `json:"name"`
			Method  string `json:"method"`
			At      string `json:"at"`
			Version int    `json:"version"`
		} `json:"approval"`
		Items []struct {
			Line
			Type     string `json:"type"`
			Origin   string `json:"origin"`
			PartKind string `json:"partKind"`
		} `json:"items"`
	} `json:"quote"`
}

// RenderQuotation paginates descriptions (including a single very long item),
// repeats the table header, and keeps the stored totals and approval version.
func RenderQuotation(raw []byte) ([]byte, error) {
	var s Quotation
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	q := s.Quote
	var pages [][]byte
	var c *canvas
	y := 0.0
	finish := func() error {
		if c.err != nil {
			return fmt.Errorf("La cotización contiene caracteres no disponibles en la tipografía PDF; no se ha omitido ni cambiado el texto")
		}
		c.text(34, 809, 8, fmt.Sprintf("Cotización %s · versión %d | Página %d", q.ID, q.Version, len(pages)+1), false)
		pages = append(pages, append([]byte(nil), c.Bytes()...))
		return nil
	}
	newPage := func() {
		c = &canvas{}
		fmt.Fprint(c, "0.38 0.05 0.12 rg 0 755 595.5 87 re f\n")
		c.bold = true
		c.text(34, 36, 18, or(s.Workshop.Name, "Taller · datos no registrados"), true)
		c.text(34, 62, 12, fmt.Sprintf("COTIZACIÓN N.º %s · VERSIÓN %d", q.ID, q.Version), true)
		c.bold = false
		y = 111
	}
	// Text blocks always continue on another page rather than shrink or truncate.
	block := func(text string, size float64, bold bool) error {
		for _, line := range quoteWrap(text, size, 525) {
			if y > 765 {
				if err := finish(); err != nil {
					return err
				}
				newPage()
			}
			c.bold = bold
			c.text(34, y, size, line, false)
			y += size + 5
		}
		c.bold = false
		return nil
	}
	newPage()
	for _, text := range []string{
		"RUC: " + or(s.Workshop.RUC, "No registrado"),
		"Dirección del taller: " + or(s.Workshop.Address, "No registrada"),
		"Fecha de emisión: " + date(q.CreatedAt) + " | Estado: " + q.Status,
		"Cliente: " + s.Client + " | Documento: " + s.Document,
		"Teléfono: " + or(s.Phone, "No registrado") + " | Correo: " + or(s.Email, "No registrado"),
		"Dirección del cliente: " + or(s.Address, "No registrada"),
		"Vehículo: " + s.Vehicle + " | Placa: " + s.Plate + " | Recepción: " + q.ReceptionID,
	} {
		if err := block(text, 10, false); err != nil {
			return nil, err
		}
	}
	if q.ParentQuoteID != "" {
		if err := block("Cotización adicional a la N.º "+q.ParentQuoteID, 10, false); err != nil {
			return nil, err
		}
	}
	y += 10
	header := func() {
		fmt.Fprintf(c, "0.38 0.05 0.12 rg 34 %.2f 527 24 re f\n", height-y-7)
		c.bold = true
		for _, h := range []struct {
			x     float64
			label string
		}{{40, "Descripción / tipo"}, {307, "Cant."}, {353, "Precio S/"}, {424, "Dto./u. S/"}, {498, "Subtotal S/"}} {
			c.text(h.x, y, 9, h.label, true)
		}
		c.bold = false
		y += 27
	}
	if y > 690 {
		if err := finish(); err != nil {
			return nil, err
		}
		newPage()
	}
	header()
	for n, item := range q.Items {
		label := fmt.Sprintf("%d. %s", n+1, item.Concept)
		lines := quoteWrap(label, 10, 252)
		meta := item.Type
		if item.Type == "Repuesto" {
			if item.PartKind != "" {
				meta += " · " + item.PartKind
			}
			if item.Origin != "" {
				meta += " · " + item.Origin
			}
		}
		lines = append(lines, quoteWrap(meta, 10, 252)...)
		for j, line := range lines {
			if y > 755 {
				if err := finish(); err != nil {
					return nil, err
				}
				newPage()
				header()
			}
			c.text(40, y, 10, line, false)
			if j == 0 {
				for _, v := range []struct {
					x float64
					s string
				}{{307, item.Quantity}, {353, item.UnitPrice}, {424, item.Discount}, {498, item.Subtotal}} {
					c.text(v.x, y, 8, v.s, false)
				}
			}
			y += 15
		}
		fmt.Fprintf(c, "0.88 0.83 0.84 RG 34 %.2f m 561 %.2f l S\n", height-y+5, height-y+5)
		y += 8
	}
	y += 12
	if err := block("TOTAL COTIZADO: S/ "+q.Total, 15, true); err != nil {
		return nil, err
	}
	if err := block("Descuentos expresados por unidad. Importes en soles (PEN).", 9, false); err != nil {
		return nil, err
	}
	y += 10
	if q.Status == "Aprobada" && q.Approval != nil {
		version := "no registrada"
		if q.Approval.Version > 0 {
			version = fmt.Sprint(q.Approval.Version)
		}
		for _, text := range []string{"Aprobación registrada · versión " + version, "Cliente / persona registrada: " + or(q.Approval.Name, "No registrada"), "Medio: " + or(q.Approval.Method, "No registrado") + " | Fecha: " + or(date(q.Approval.At), "No registrada")} {
			if err := block(text, 10, false); err != nil {
				return nil, err
			}
		}
	} else {
		note := "Pendiente de aprobación del cliente. Este documento no autoriza la ejecución de trabajos."
		if q.Status == "Rechazada" {
			note = "Cotización rechazada. Este documento no autoriza la ejecución de trabajos."
		}
		if q.Status == "Aprobada" {
			note = "Estado aprobado histórico; no se registraron los datos de la aprobación."
		}
		if err := block(note, 10, false); err != nil {
			return nil, err
		}
	}
	if err := finish(); err != nil {
		return nil, err
	}
	return buildDocument(false, pages...)
}

// Conservative Helvetica/Bold glyph widths: even an unbroken sequence of W's
// fits its column. Keep every rune; newlines and long words remain printable.
func quoteGlyph(r rune) float64 {
	switch {
	case strings.ContainsRune(" ilI.,:;!'|", r):
		return .36
	case r > 127 || strings.ContainsRune("MW@%mw", r):
		return 1.02
	default:
		return .75
	}
}
func quoteWrap(text string, size, max float64) []string {
	var out []string
	for _, paragraph := range strings.Split(text, "\n") {
		line := ""
		used := 0.0
		for _, word := range strings.Fields(paragraph) {
			w := 0.0
			for _, r := range word {
				w += quoteGlyph(r) * size
			}
			if line != "" && used+size*.28+w > max {
				out = append(out, line)
				line = ""
				used = 0
			}
			if line != "" {
				line += " "
				used += size * .28
			}
			for _, r := range word {
				n := quoteGlyph(r) * size
				if used+n > max && line != "" {
					out = append(out, line)
					line = ""
					used = 0
				}
				line += string(r)
				used += n
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

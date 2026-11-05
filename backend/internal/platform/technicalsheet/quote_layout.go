package technicalsheet

import (
	"fmt"
	"strings"
)

const quoteQuantity = 376.0
const quoteAmount = 444.0
const quoteBottom = 751.0

// Values returned by PostgreSQL are printed verbatim. In particular, Importe is
// the stored subtotal, not the unit price and not an extra grouped charge.
func renderQuotationLayout(s Quotation) ([]byte, error) {
	q := s.Quote
	var pages [][]byte
	var c *canvas
	y := 0.0
	status := q.Status
	if status == "Pendiente" {
		status = "Pendiente de aprobación"
	}
	finish := func() error {
		if c.err != nil {
			return fmt.Errorf("La cotización contiene caracteres no disponibles en la tipografía PDF; no se ha omitido ni cambiado el texto")
		}
		pages = append(pages, append([]byte(nil), c.Bytes()...))
		return nil
	}
	newPage := func(first bool) {
		c = &canvas{}
		documentHeader(c, "COTIZACIÓN", "N.º "+q.ID, fmt.Sprintf("VERSIÓN %d", q.Version), s.Workshop.Name)
		y = 135
		if !first {
			for _, text := range []string{"Cliente: " + s.Client + " · Placa: " + s.Plate, "Estado: " + status + " · Emisión: " + date(q.CreatedAt)} {
				for _, line := range quoteWrap(text, 9, 527) {
					c.text(34, y, 9, line, false)
					y += 14
				}
			}
			y += 10
		}
	}
	newPage(true)
	// Workshop legal/contact data remain sourced from the same snapshot.
	legal := []string{}
	if s.Workshop.RUC != "" {
		legal = append(legal, "RUC: "+s.Workshop.RUC)
	}
	if s.Workshop.Address != "" {
		legal = append(legal, s.Workshop.Address)
	}
	if s.Workshop.Phone != "" {
		legal = append(legal, "Teléfono: "+s.Workshop.Phone)
	}
	for _, line := range quoteWrap(strings.Join(legal, " · "), 9, 527) {
		if line != "" {
			c.text(34, y, 9, line, false)
			y += 13
		}
	}
	y += 16
	left := []string{"Cliente: " + s.Client, "Documento: " + s.Document}
	for _, entry := range []struct{ label, value string }{{"Teléfono", s.Phone}, {"Correo", s.Email}, {"Dirección", s.Address}} {
		if entry.value != "" {
			left = append(left, entry.label+": "+entry.value)
		}
	}
	right := []string{"Vehículo: " + s.Vehicle, "Placa: " + s.Plate, "Emisión: " + or(date(q.CreatedAt), "No registrada")}
	if q.ReceptionID != "" {
		right = append(right, "Recepción: "+q.ReceptionID)
	}
	if q.ParentQuoteID != "" {
		right = append(right, "Adicional a cotización: "+q.ParentQuoteID)
	}
	columns := []struct {
		x, w   float64
		title  string
		values []string
	}{{34, 255, "CLIENTE", left}, {306, 255, "VEHÍCULO Y ATENCIÓN", right}}
	maxY := y
	for _, col := range columns {
		height := 30.0
		for _, value := range col.values {
			height += float64(len(quoteWrap(value, 9.5, col.w-20))) * 12.5
		}
		documentPanel(c, col.x, y-12, col.w, height)
		col.x += 10
		col.w -= 20
		c.bold = true
		c.text(col.x, y, 9, col.title, false)
		c.bold = false
		cy := y + 20
		for _, text := range col.values {
			for _, line := range quoteWrap(text, 9.5, col.w) {
				c.text(col.x, cy, 9.5, line, false)
				cy += 12.5
			}
		}
		if cy > maxY {
			maxY = cy
		}
	}
	y = maxY + 6
	// Metadata is bounded by existing field lengths, but long addresses/names
	// continue independently rather than overlap the table or footer.
	if y > quoteBottom-90 {
		if err := finish(); err != nil {
			return nil, err
		}
		newPage(false)
	}
	c.bold = true
	for _, line := range quoteWrap("ESTADO: "+status, 10, 527) {
		c.text(34, y, 10, line, false)
		y += 15
	}
	c.bold = false
	y += 10
	tableHeader := func() {
		quoteRect(c, 34, y, 527, 28, brandBurgundy, false)
		quoteGrid(c, y, 28)
		c.bold = true
		c.text(44, y+18, 9, "DESCRIPCIÓN / CONCEPTO", true)
		quoteCenter(c, 376, 444, y+18, 9, "CANTIDAD", true)
		quoteCenter(c, 444, 561, y+18, 9, "IMPORTE", true)
		c.bold = false
		y += 28
	}
	section := func(name string) {
		quoteRect(c, 34, y, 527, 23, brandSurface, false)
		c.bold = true
		c.text(44, y+16, 9.5, name, false)
		c.bold = false
		y += 23
	}
	tableHeader()
	for _, group := range []struct{ name, kind string }{{"SERVICIOS", "Servicio"}, {"REPUESTOS", "Repuesto"}, {"OTROS CONCEPTOS", ""}} {
		items := []QuotationItem{}
		for _, item := range q.Items {
			if item.Type == group.kind || group.kind == "" && item.Type != "Servicio" && item.Type != "Repuesto" {
				items = append(items, item)
			}
		}
		if len(items) == 0 {
			continue
		}
		firstLines, firstDetail := quoteRowContent(items[0])
		firstHeight := float64(len(firstLines))*13 + float64(len(firstDetail))*11 + 18
		if firstHeight > quoteBottom-235 {
			firstHeight = 60
		}
		if y+23+firstHeight > quoteBottom {
			if err := finish(); err != nil {
				return nil, err
			}
			newPage(false)
			tableHeader()
		}
		section(group.name)
		for itemIndex, item := range items {
			lines, detail := quoteRowContent(item)
			rowHeight := float64(len(lines))*13 + float64(len(detail))*11 + 18
			// Normal rows are indivisible. A row taller than a full page needs
			// explicitly labelled continuation; amount/quantity print only once.
			if y+rowHeight > quoteBottom && rowHeight <= quoteBottom-235 {
				if err := finish(); err != nil {
					return nil, err
				}
				newPage(false)
				tableHeader()
				section(group.name + " · CONTINUACIÓN")
			}
			all := []struct {
				text          string
				size, leading float64
			}{}
			for _, line := range lines {
				all = append(all, struct {
					text          string
					size, leading float64
				}{line, 10.5, 13})
			}
			for _, line := range detail {
				all = append(all, struct {
					text          string
					size, leading float64
				}{line, 9, 11})
			}
			start := y
			if itemIndex%2 == 1 && y+rowHeight <= quoteBottom {
				quoteRect(c, 34, y, 527, rowHeight, brandSurface, false)
			}
			cy := y + 14
			printed := false
			for _, line := range all {
				if cy+line.leading > quoteBottom {
					quoteGrid(c, start, cy-start)
					if err := finish(); err != nil {
						return nil, err
					}
					newPage(false)
					tableHeader()
					section(group.name + " · CONTINUACIÓN")
					start = y
					cy = y + 14
					c.text(44, cy, 8, "Concepto anterior · continuación sin cargo adicional", false)
					cy += 15
				}
				c.text(44, cy, line.size, line.text, false)
				if !printed {
					quoteCenter(c, 376, 444, cy, 10.5, item.Quantity, false)
					quoteRightText(c, 551, cy, 10.5, "S/ "+item.Subtotal, false)
					printed = true
				}
				cy += line.leading
			}
			y = cy + 4
			quoteGrid(c, start, y-start)
		}
	}
	if len(q.Items) == 0 {
		c.text(44, y+20, 10, "Sin conceptos registrados", false)
		quoteGrid(c, y, 32)
		y += 32
	}
	notes := []string{"Importes en soles (PEN). Subtotales con descuentos incluidos."}
	if q.Status == "Aprobada" && q.Approval != nil {
		version := "no registrada"
		if q.Approval.Version > 0 {
			version = fmt.Sprint(q.Approval.Version)
		}
		notes = append(notes, "Aprobación registrada · versión "+version, "Persona registrada: "+or(q.Approval.Name, "No registrada"), "Medio: "+or(q.Approval.Method, "No registrado")+" · Fecha: "+or(date(q.Approval.At), "No registrada"))
	} else if q.Status == "Aprobada" {
		notes = append(notes, "Estado aprobado histórico; no se registraron los datos de la aprobación.")
	} else {
		notes = append(notes, "Presupuesto propuesto. Este documento no constituye una autorización de trabajo.")
	}
	needed := 58.0
	for _, note := range notes {
		needed += float64(len(quoteWrap(note, 9, 527)))*13 + 4
	}
	if y+needed > quoteBottom {
		if err := finish(); err != nil {
			return nil, err
		}
		newPage(false)
	}
	y += 8
	quoteRect(c, 336, y, 225, 40, brandBurgundy, false)
	c.text(34, y+18, 9, "Moneda: soles (PEN)", false)
	c.bold = true
	c.text(346, y+15, 9, "TOTAL COTIZADO", true)
	quoteRightText(c, 551, y+32, 16, "S/ "+q.Total, true)
	c.bold = false
	y += 54
	for _, note := range notes {
		for _, line := range quoteWrap(note, 9, 527) {
			if y > quoteBottom {
				if err := finish(); err != nil {
					return nil, err
				}
				newPage(false)
			}
			c.text(34, y, 9, line, false)
			y += 13
		}
		y += 4
	}
	if err := finish(); err != nil {
		return nil, err
	}
	return finishDocument(pages, fmt.Sprintf("Cotización %s · versión %d", q.ID, q.Version))
}

func quoteRect(c *canvas, x, y, w, h float64, color string, border bool) {
	mode := "f"
	if border {
		mode = "B"
	}
	fmt.Fprintf(c, "q %s rg 0.65 0.61 0.62 RG 0.6 w %.2f %.2f %.2f %.2f re %s Q\n", color, x, height-y-h, w, h, mode)
}
func quoteGrid(c *canvas, y, h float64) {
	fmt.Fprintf(c, "q 0.65 0.61 0.62 RG 0.6 w 34 %.2f 527 %.2f re S\n", height-y-h, h)
	for _, x := range []float64{quoteQuantity, quoteAmount} {
		fmt.Fprintf(c, "%.2f %.2f m %.2f %.2f l S\n", x, height-y, x, height-y-h)
	}
	fmt.Fprint(c, "Q\n")
}

// Helvetica widths for monetary alignment (digits are tabular at 556 units).
func quoteTextWidth(text string, size float64, bold bool) float64 {
	upper := []float64{667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778, 667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611}
	if bold {
		upper = []float64{722, 722, 722, 722, 667, 611, 778, 722, 278, 556, 722, 611, 833, 722, 778, 667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611}
	}
	w := 0.0
	for _, r := range text {
		switch {
		case r >= '0' && r <= '9':
			w += 556
		case r >= 'A' && r <= 'Z':
			w += upper[r-'A']
		case strings.ContainsRune(" ./", r):
			w += 278
		case r == '-':
			w += 333
		default:
			w += 556
		}
	}
	return w * size / 1000
}
func quoteRightText(c *canvas, right, y, size float64, text string, white bool) {
	c.text(right-quoteTextWidth(text, size, c.bold), y, size, text, white)
}
func quoteCenter(c *canvas, left, right, y, size float64, text string, white bool) {
	c.text((left+right-quoteTextWidth(text, size, c.bold))/2, y, size, text, white)
}
func quoteIsZero(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r != '0' && r != '.' {
			return false
		}
	}
	return true
}

func quoteRowContent(item QuotationItem) ([]string, []string) {
	lines := quoteWrap(item.Concept, 10.5, 322)
	meta := "P. unit.: S/ " + item.UnitPrice + " · Dto./u.: S/ " + item.Discount
	if item.Type == "Repuesto" {
		if item.PartKind != "" {
			meta += " · " + item.PartKind
		}
		if item.Origin != "" {
			meta += " · Origen: " + item.Origin
		}
	}
	if quoteIsZero(item.Subtotal) {
		meta += " · Sin cargo independiente"
	}
	detail := quoteWrap(meta, 9, 322)
	return lines, detail
}

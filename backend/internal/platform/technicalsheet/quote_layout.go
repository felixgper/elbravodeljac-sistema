package technicalsheet

import (
	"fmt"
	"strings"
)

const quoteQuantity = 376.0
const quoteAmount = 444.0
const quoteBottom = 741.0

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
		quoteRect(c, 34, 774, 527, 32, "0.96 0.94 0.95", false)
		c.bold = true
		quoteCenter(c, 34, 561, 794, 11, "¡GRACIAS POR CONFIAR EN NOSOTROS!", false)
		c.bold = false
		c.text(34, 823, 8, fmt.Sprintf("Cotización %s · versión %d", q.ID, q.Version), false)
		quoteRightText(c, 561, 823, 8, fmt.Sprintf("Página %d", len(pages)+1), false)
		if c.err != nil {
			return fmt.Errorf("La cotización contiene caracteres no disponibles en la tipografía PDF; no se ha omitido ni cambiado el texto")
		}
		pages = append(pages, append([]byte(nil), c.Bytes()...))
		return nil
	}
	newPage := func(first bool) {
		c = &canvas{}
		quoteRect(c, 34, 30, 527, 82, "0.98 0.97 0.97", false)
		quoteRect(c, 34, 109, 527, 3, "0.38 0.05 0.12", false)
		quoteRect(c, 47, 36, 70, 70, "0.38 0.05 0.12", false)
		drawQuoteLogo(c, 53, 42, 58)
		c.bold = true
		name := quoteWrap(or(s.Workshop.Name, "EL BRAVO DEL JAC"), 16, 285)
		for i, line := range name {
			if i < 2 {
				c.text(132, 58+float64(i)*19, 16, line, false)
			}
		}
		c.bold = false
		c.text(132, 96, 9, "TALLER AUTOMOTRIZ", false)
		c.bold = true
		c.text(434, 50, 9, "COTIZACIÓN", false)
		for i, line := range quoteWrap("N.º "+q.ID, 10, 117) {
			c.text(434, 69+float64(i)*13, 10, line, false)
		}
		c.bold = false
		c.text(434, 96, 9, fmt.Sprintf("VERSIÓN %d", q.Version), false)
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
	// Workshop and quote data follow the two-column structure of the reference.
	left := []string{or(s.Workshop.Name, "Nombre comercial no registrado"), "RUC: " + or(s.Workshop.RUC, "No registrado"), "Dirección: " + or(s.Workshop.Address, "No registrada")}
	if s.Workshop.Phone != "" {
		left = append(left, "Teléfono: "+s.Workshop.Phone)
	}
	emission := "Emisión: " + or(date(q.CreatedAt), "No registrada")
	if q.ReceptionID != "" {
		emission += " · Recepción: " + q.ReceptionID
	}
	right := []string{emission, "Cliente: " + s.Client, "Documento: " + s.Document, "Placa: " + s.Plate, "Vehículo: " + s.Vehicle}
	if q.ParentQuoteID != "" {
		right = append(right, "Adicional a cotización: "+q.ParentQuoteID)
	}
	columns := []struct {
		x, w   float64
		title  string
		values []string
	}{{34, 243, "DATOS DEL TALLER", left}, {306, 255, "DATOS DE LA COTIZACIÓN", right}}
	maxY := y
	for _, col := range columns {
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
		quoteRect(c, 34, y, 527, 28, "0.94 0.93 0.94", true)
		quoteGrid(c, y, 28)
		c.bold = true
		c.text(44, y+18, 9, "DESCRIPCIÓN DE SERVICIOS", false)
		quoteCenter(c, 376, 444, y+18, 9, "CANTIDAD", false)
		quoteCenter(c, 444, 561, y+18, 9, "IMPORTE", false)
		c.bold = false
		y += 28
	}
	section := func(name string) {
		quoteRect(c, 34, y, 527, 23, "0.38 0.05 0.12", true)
		c.bold = true
		c.text(44, y+16, 9.5, name, true)
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
		for _, item := range items {
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
	needed := 48.0
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
	quoteRect(c, 34, y, 527, 35, "0.38 0.05 0.12", false)
	c.bold = true
	c.text(44, y+23, 12, "TOTAL COTIZADO:", true)
	quoteRightText(c, 551, y+23, 14, "S/ "+q.Total, true)
	c.bold = false
	y += 47
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
	return buildDocument(false, pages...)
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

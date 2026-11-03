package technicalsheet

import (
	"fmt"
	"strings"
)

type sheetLayout struct {
	snapshot Snapshot
	pages    [][]byte
	c        *canvas
	y        float64
}

func (p *sheetLayout) finish() error {
	if p.c.err != nil {
		return p.c.err
	}
	p.pages = append(p.pages, append([]byte(nil), p.c.Bytes()...))
	return nil
}
func (p *sheetLayout) next() error {
	if p.c != nil {
		if err := p.finish(); err != nil {
			return err
		}
	}
	p.c = &canvas{}
	documentHeader(p.c, "FICHA TÉCNICA", p.snapshot.Code, p.snapshot.Plate, p.snapshot.Workshop.Name)
	p.y = 124
	return nil
}
func (p *sheetLayout) room(h float64) {
	if p.y+h > documentBottom {
		_ = p.next()
	}
}
func (p *sheetLayout) heading(title string) {
	p.room(36)
	documentTitle(p.c, 34, p.y+12, 11, title)
	quoteRect(p.c, 34, p.y+19, 527, .6, "0.79 0.79 0.80", false)
	p.y += 25
}
func (p *sheetLayout) paragraph(label, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	p.room(28)
	if label != "" {
		documentTitle(p.c, 34, p.y, 9, label)
		p.y += 12
	}
	for _, line := range quoteWrap(value, 9.5, 527) {
		if p.y+13 > documentBottom {
			_ = p.next()
			if label != "" {
				documentTitle(p.c, 34, p.y, 9, label+" · CONTINUACIÓN")
				p.y += 14
			}
		}
		p.c.text(34, p.y, 9.5, line, false)
		p.y += 12
	}
	p.y += 5
}
func (p *sheetLayout) note(label, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	p.room(26)
	for _, line := range quoteWrap(label+": "+value, 9.5, 527) {
		if p.y+13 > documentBottom {
			_ = p.next()
			documentTitle(p.c, 34, p.y, 9, label+" · CONTINUACIÓN")
			p.y += 14
		}
		p.c.text(34, p.y, 9.5, line, false)
		p.y += 12
	}
	p.y += 5
}

type sheetPair struct{ label, value string }

func (p *sheetLayout) pairs(values []sheetPair) {
	for i := 0; i < len(values); i += 2 {
		left := quoteWrap(values[i].value, 9, 243)
		var right []string
		if i+1 < len(values) {
			right = quoteWrap(values[i+1].value, 9, 243)
		}
		rows := len(left)
		if len(right) > rows {
			rows = len(right)
		}
		p.room(30)
		p.c.bold = true
		p.c.text(34, p.y, 8, values[i].label, false)
		if i+1 < len(values) {
			p.c.text(306, p.y, 8, values[i+1].label, false)
		}
		p.c.bold = false
		p.y += 14
		for n := 0; n < rows; n++ {
			p.room(13)
			if n < len(left) {
				p.c.text(34, p.y, 9, left[n], false)
			}
			if n < len(right) {
				p.c.text(306, p.y, 9, right[n], false)
			}
			p.y += 12
		}
		p.y += 11
	}
}
func (p *sheetLayout) table(lines []Line, title string) {
	if len(lines) == 0 {
		return
	}
	header := func() {
		p.room(70)
		p.heading(title)
		quoteRect(p.c, 34, p.y, 527, 23, brandBurgundy, false)
		p.c.bold = true
		p.c.text(44, p.y+15, 8.5, "DESCRIPCIÓN / TRABAJO REALIZADO", true)
		quoteCenter(p.c, 376, 444, p.y+15, 8.5, "CANT.", true)
		quoteRightText(p.c, 551, p.y+15, 8.5, "IMPORTE", true)
		p.c.bold = false
		p.y += 23
	}
	section := func(kind string) {
		name := map[string]string{"Servicio": "SERVICIOS REALIZADOS", "Repuesto": "REPUESTOS UTILIZADOS"}[kind]
		if name == "" {
			name = "TRABAJOS REGISTRADOS"
		}
		quoteRect(p.c, 34, p.y, 527, 18, brandSurface, false)
		documentTitle(p.c, 44, p.y+12, 8.5, name)
		p.y += 18
	}
	// Keep the initial title, column labels, category and a real first row together.
	first := lines[0]
	firstHeight := float64(len(quoteWrap(first.Concept, 10, 322)))*13 + float64(len(quoteWrap("P. unit.: S/ "+first.UnitPrice+" · Dto./u.: S/ "+first.Discount+orNote(first.Notes), 8.5, 322)))*11 + 14
	if firstHeight > 100 {
		firstHeight = 100
	}
	p.room(25 + 23 + 18 + firstHeight)
	header()
	prior := "__first__"
	for index, line := range lines {
		content := quoteWrap(line.Concept, 10, 322)
		details := quoteWrap("P. unit.: S/ "+line.UnitPrice+" · Dto./u.: S/ "+line.Discount+orNote(line.Notes), 8.5, 322)
		h := float64(len(content))*13 + float64(len(details))*11 + 14
		extra := 0.0
		if prior != line.Type {
			extra = 18
		}
		if p.y+h+extra > documentBottom && h < 500 {
			_ = p.next()
			header()
			prior = "__first__"
		}
		if prior != line.Type {
			if p.y+50 > documentBottom {
				_ = p.next()
				header()
			}
			section(line.Type)
			prior = line.Type
		}
		start := p.y
		if index%2 == 1 && p.y+h <= documentBottom {
			quoteRect(p.c, 34, p.y, 527, h, brandSurface, false)
		}
		first := true
		for _, part := range []struct {
			lines         []string
			size, leading float64
		}{{content, 10, 13}, {details, 8.5, 11}} {
			for _, text := range part.lines {
				if p.y+part.leading+7 > documentBottom {
					quoteGrid(p.c, start, p.y-start)
					_ = p.next()
					header()
					section(line.Type)
					p.c.text(44, p.y+11, 8, "Concepto anterior · continuación sin cargo adicional", false)
					p.y += 22
					start = p.y
				}
				p.c.text(44, p.y+13, part.size, text, false)
				if first {
					quoteCenter(p.c, 376, 444, p.y+13, 10, line.Quantity, false)
					quoteRightText(p.c, 551, p.y+13, 10, "S/ "+line.Subtotal, false)
					first = false
				}
				p.y += part.leading
			}
		}
		p.y += 9
		quoteGrid(p.c, start, p.y-start)
	}
	p.y += 10
}
func renderSheetLayout(s Snapshot) ([]byte, error) {
	p := &sheetLayout{snapshot: s}
	_ = p.next()
	legal := []string{}
	if s.Workshop.RUC != "" {
		legal = append(legal, "RUC: "+s.Workshop.RUC)
	}
	if s.Workshop.Address != "" {
		legal = append(legal, s.Workshop.Address)
	}
	p.paragraph("", strings.Join(legal, " · "))
	p.note("Ingreso / salida", or(date(s.In), "No registrado")+" · "+or(date(s.Out), "No registrada"))
	p.heading("CLIENTE Y VEHÍCULO")
	left := []string{s.Client, "Documento: " + or(s.Document, "No registrado")}
	for _, entry := range []sheetPair{{"Teléfono", s.Phone}, {"Correo", s.Email}, {"Dirección", s.Address}} {
		if entry.value != "" {
			left = append(left, entry.label+": "+entry.value)
		}
	}
	right := []string{s.Vehicle, "Placa: " + s.Plate, "Kilometraje: " + s.Mileage + " km", "Mecánico: " + or(s.Mechanic, "No registrado")}
	var columns [2][]string
	for col, values := range [][]string{left, right} {
		for _, value := range values {
			columns[col] = append(columns[col], quoteWrap(value, 9.5, 233)...)
		}
	}
	rows := len(columns[0])
	if len(columns[1]) > rows {
		rows = len(columns[1])
	}
	documentPanel(p.c, 34, p.y-9, 527, float64(rows)*13+5)
	for row := 0; row < rows; row++ {
		p.room(14)
		for col := 0; col < 2; col++ {
			if row < len(columns[col]) {
				p.c.text(44+float64(col)*272, p.y, 9.5, columns[col][row], false)
			}
		}
		p.y += 13
	}
	p.y += 8
	p.note("Motivo de ingreso", s.Symptoms)
	diagnosis := s.Diagnosis
	if s.Evaluation != "" && s.Evaluation != diagnosis {
		if diagnosis != "" {
			diagnosis += "\nEvaluación técnica: "
		}
		diagnosis += s.Evaluation
	}
	p.note("Diagnóstico técnico", diagnosis)
	p.note("Recomendaciones", s.Recommendations)
	p.note("Observaciones del diagnóstico", s.DiagnosisNotes)
	var ordered []Line
	for _, kind := range []string{"Servicio", "Repuesto", ""} {
		for _, line := range s.Lines {
			if line.Type == kind || (kind == "" && line.Type != "Servicio" && line.Type != "Repuesto") {
				ordered = append(ordered, line)
			}
		}
	}
	p.table(ordered, "TRABAJOS REALIZADOS")
	if len(s.Lines) == 0 {
		p.paragraph("TRABAJOS REALIZADOS", "Sin trabajos confirmados en el control técnico")
	}
	p.room(48)
	quoteRect(p.c, 336, p.y, 225, 40, brandBurgundy, false)
	p.c.text(34, p.y+18, 9, "Importes en soles (PEN)", false)
	p.c.bold = true
	p.c.text(346, p.y+15, 9, "TOTAL DE LA ATENCIÓN", true)
	quoteRightText(p.c, 551, p.y+32, 16, "S/ "+s.Total, true)
	p.c.bold = false
	p.y += 44
	if len(s.Advances) > 0 || len(s.Tests) > 0 {
		p.heading("AVANCES Y PRUEBAS REGISTRADOS")
	}
	for _, advance := range s.Advances {
		p.note("Avance", advance)
	}
	for _, test := range s.Tests {
		p.note(test.Kind, test.Result+orNote(test.Notes))
	}
	// The technical control starts on its own page. Longer controls continue
	// without compressing the approved work summary or the inspection cards.
	if p.y > 124 {
		_ = p.next()
	}
	renderSheetControl(p, s)
	if err := p.finish(); err != nil {
		return nil, err
	}
	return finishDocument(p.pages, "Ficha técnica · "+s.Code+" · "+s.Plate)
}

// Plan with the same drawing functions, so keep-together decisions cannot drift
// from the actual typography. The plan is in memory and never renders a PDF.
func (p *sheetLayout) keepBlock(draw func(*sheetLayout)) {
	plan := &sheetLayout{snapshot: p.snapshot}
	_ = plan.next()
	draw(plan)
	if len(plan.pages) == 0 && plan.c.err == nil {
		p.room(plan.y - 124)
	}
	draw(p)
}
func renderSheetControl(p *sheetLayout, s Snapshot) {
	// Keep the three framed categories together, without shrinking their text.
	maxInspection := 0.0
	for _, g := range Catalog.Groups {
		if h := sheetInspectionHeight(g, s.Control.Tires); h > maxInspection {
			maxInspection = h
		}
	}
	p.room(maxInspection + 48)
	p.heading("CONTROL TÉCNICO REGISTRADO")
	p.y += 4
	p.c.text(34, p.y, 8.5, "B: Bien · A: Atención · U: Urgente. Sin marcar: componente sin evaluar.", false)
	p.y += 10
	for index := 0; index < len(Catalog.Groups); index += 3 {
		p.room(maxInspection + 8)
		for col := 0; col < 3 && index+col < len(Catalog.Groups); col++ {
			g := Catalog.Groups[index+col]
			x := 34 + float64(col)*179
			sheetCard(p.c, x, p.y, 169, maxInspection)
			sheetInspectionGroup(p.c, g, s.Control.Inspections, x, p.y, s.Control.Tires)
		}
		p.y += maxInspection + 8
	}
	if s.Control.OtherFluids != "" && s.Control.Coils != "" {
		rows := []sheetPair{{"Otros líquidos", s.Control.OtherFluids}, {"Bobinas", s.Control.Coils}}
		var wrapped [2][]string
		n := 0
		for i, value := range rows {
			wrapped[i] = sheetWrap(value.label+": "+value.value, 9.5, 255, false)
			if len(wrapped[i]) > n {
				n = len(wrapped[i])
			}
		}
		p.room(float64(n) * 13)
		for i := range rows {
			for row, text := range wrapped[i] {
				p.c.text(34+float64(i)*272, p.y+float64(row)*13, 9.5, text, false)
			}
		}
		p.y += float64(n) * 13

	} else {
		p.paragraph("OTROS LÍQUIDOS", s.Control.OtherFluids)
		p.paragraph("BOBINAS", s.Control.Coils)
	}
	p.heading("MANTENIMIENTO PREVENTIVO Y CORRECTIVO")
	legend := "Sin marcar: sin selección registrada."
	p.c.text(561-sheetTextWidth(legend, 8.5, false), p.y-13, 8.5, legend, false)
	p.y += 10
	p.maintenanceChecks(s.Control)
	// Keep the last section and its closing together when possible, so a longer
	// control does not leave only the signatures on an otherwise empty page.
	p.keepBlock(func(target *sheetLayout) {
		target.heading("OTROS SERVICIOS Y REPARACIONES REGISTRADOS")
		target.y += 11
		target.otherServiceChecks(s.Control.OtherServices)
		target.keepBlock(func(last *sheetLayout) { renderSheetClosing(last, s) })
	})

}
func renderSheetClosing(p *sheetLayout, s Snapshot) {
	p.note("Observaciones del cliente", s.Control.ClientNotes)
	p.note("Observaciones técnicas", s.Control.TechnicalNotes)
	if r := s.Control.Reconciliation; r != nil {
		p.paragraph("CONCILIACIÓN AUTORIZADA", fmt.Sprintf("Usuario %s: %s. Presupuesto S/ %s; importe definitivo S/ %s.", r.By, r.Reason, s.AuthorizedTotal, s.Total))
	}
	// The closing fields are one measured block: they must not orphan signatures.
	p.room(48)
	documentTitle(p.c, 34, p.y, 8.5, "PRÓXIMO MANTENIMIENTO (KM)")
	documentTitle(p.c, 306, p.y, 8.5, "FECHA RECOMENDADA")
	p.c.text(34, p.y+13, 9.5, or(s.Control.NextMileage, "Sin recomendación registrada"), false)
	p.c.text(306, p.y+13, 9.5, nextDate(s.Control.NextDate), false)
	p.c.text(34, p.y+39, 8.5, "Conformidad del cliente: ____________________", false)
	p.c.text(306, p.y+39, 8.5, "Conformidad del operario: ____________________", false)
	p.y += 48

}

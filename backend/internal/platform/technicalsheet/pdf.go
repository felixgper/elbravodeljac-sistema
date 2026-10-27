package technicalsheet

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"image/jpeg"
	"strings"
	"time"
	"unicode/utf8"
)

type Line struct {
	ID        string `json:"id"`
	Concept   string `json:"concept"`
	Quantity  string `json:"quantity"`
	UnitPrice string `json:"unitPrice"`
	Discount  string `json:"discount"`
	Subtotal  string `json:"subtotal"`
}
type Workshop struct {
	Name    string `json:"name"`
	RUC     string `json:"ruc"`
	Address string `json:"address"`
}
type Snapshot struct {
	Workshop Workshop `json:"workshop"`
	Code     string   `json:"code"`
	Client   string   `json:"client"`
	Document string   `json:"document"`
	Phone    string   `json:"phone"`
	Address  string   `json:"address"`
	Email    string   `json:"email"`
	Plate    string   `json:"plate"`
	Vehicle  string   `json:"vehicle"`
	Mileage  string   `json:"mileage"`
	Symptoms string   `json:"symptoms"`
	In       string   `json:"in"`
	Out      string   `json:"out"`
	Mechanic string   `json:"mechanic"`
	Total    string   `json:"total"`
	Lines    []Line   `json:"lines"`
	Advances []string `json:"advances"`
	Control  Control  `json:"control"`
}

const width = 595.5
const height = 842.25

func date(s string) string {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return s
	}
	loc, _ := time.LoadLocation("America/Lima")
	return t.In(loc).Format("02/01/2006")
}

type canvas struct {
	bytes.Buffer
	err  error
	bold bool
}

func (c *canvas) text(x, y, size float64, s string, white bool) {
	if s == "" {
		return
	}
	b, err := charmap.Windows1252.NewEncoder().Bytes([]byte(s))
	if err != nil {
		c.err = fmt.Errorf("La ficha contiene caracteres fuera de la tipografía de la plantilla")
		return
	}
	color := "0.20 0.04 0.09"
	if white {
		color = "1 1 1"
	}
	escaped := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)", "\r", " ", "\n", " ").Replace(string(b))
	font := "/F1"
	if c.bold {
		font = "/F2"
	}
	fmt.Fprintf(c, "BT %s %.2f Tf %s rg 1 0 0 1 %.2f %.2f Tm (%s) Tj ET\n", font, size, color, x, height-y, escaped)
}
func wrap(s string, size, w float64) []string {
	limit := int(w / (size * .55))
	if limit < 1 {
		limit = 1
	}
	var out []string
	for _, paragraph := range strings.Split(s, "\n") {
		line := ""
		for _, word := range strings.Fields(paragraph) {
			for utf8.RuneCountInString(word) > limit {
				if line != "" {
					out = append(out, line)
					line = ""
				}
				r := []rune(word)
				out = append(out, string(r[:limit]))
				word = string(r[limit:])
			}
			if utf8.RuneCountInString(line+" "+word) > limit {
				out = append(out, line)
				line = word
			} else {
				if line != "" {
					line += " "
				}
				line += word
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
func (c *canvas) box(x, y, w, h, max float64, s string, white bool) {
	for size := max; size >= 5.5; size -= .5 {
		lines := wrap(s, size, w)
		if float64(len(lines))*size*1.15 <= h {
			for i, l := range lines {
				c.text(x, y+size+float64(i)*size*1.15, size, l, white)
			}
			return
		}
	}
	c.err = fmt.Errorf("El contenido supera el espacio imprimible de la ficha original de dos páginas; no se ha recortado información")
}
func (c *canvas) boxBold(x, y, w, h, max float64, s string, white bool) {
	c.bold = true
	c.box(x, y, w, h, max, s, white)
	c.bold = false
}
func (c *canvas) mark(x, y float64, chosen bool) {
	fmt.Fprintf(c, "0.45 G 0.6 w %.2f %.2f 8 8 re S\n", x-4, height-y-4)
	if chosen {
		c.text(x-2.7, y+2.7, 8, "X", false)
	}
}
func (c *canvas) workMark(x, y float64) {
	const r = 8.0
	const k = 4.42
	fmt.Fprintf(c, "0.11 0.48 0.25 rg %.2f %.2f m %.2f %.2f %.2f %.2f %.2f %.2f c %.2f %.2f %.2f %.2f %.2f %.2f c %.2f %.2f %.2f %.2f %.2f %.2f c %.2f %.2f %.2f %.2f %.2f %.2f c f\n", x+r, height-y, x+r, height-y+k, x+k, height-y+r, x, height-y+r, x-k, height-y+r, x-r, height-y+k, x-r, height-y, x-r, height-y-k, x-k, height-y-r, x, height-y-r, x+k, height-y-r, x+r, height-y-k, x+r, height-y)
	fmt.Fprintf(c, "1 1 1 RG 1.5 w %.2f %.2f m %.2f %.2f l %.2f %.2f l S\n", x-4, height-y, x-1, height-y-3, x+4, height-y+4)
}

func Render(raw []byte) ([]byte, error) {
	var s Snapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	p1, p2 := &canvas{}, &canvas{}
	for _, p := range []*canvas{p1, p2} {
		fmt.Fprintf(p, "q %.2f 0 0 %.2f 0 0 cm /Bg Do Q\n", width, height)
	}
	p1.text(490, 67, 8, "RUC "+s.Workshop.RUC, true)
	p1.box(455, 73, 106, 10, 7.5, s.Workshop.Address, true)
	p2.box(45, 802, 305, 14, 9, s.Workshop.Address+" - RUC "+s.Workshop.RUC, true)
	p1.boxBold(185, 141, 187, 43, 16, s.Symptoms, false)
	p1.text(399, 171, 9, date(s.In), false)
	p1.text(490, 171, 9, date(s.Out), false)
	p1.boxBold(43.5, 241, 242, 23, 15, s.Client, false)
	p1.box(43.5, 280, 115, 12, 9, s.Document, false)
	p1.box(166, 280, 120, 12, 9, s.Phone, false)
	p1.box(43.5, 309, 115, 15, 8, s.Address, false)
	p1.box(166, 309, 120, 15, 8, s.Email, false)
	p1.boxBold(480, 224, 67, 18, 12, s.Plate, false)
	p1.boxBold(317, 253, 238, 22, 14, s.Vehicle, false)
	p1.text(317, 301, 9, s.Mileage+" km", false)
	p1.box(439, 292, 116, 20, 9, s.Mechanic, false)
	performed := map[string]bool{}
	for _, id := range s.Control.PerformedItems {
		performed[id] = true
	}
	var work []string
	for _, l := range s.Lines {
		if performed[l.ID] {
			work = append(work, l.Concept)
		}
	}

	if len(work) == 0 {
		work = []string{"Sin trabajos confirmados en el control técnico"}
	}
	chunks := make([][]string, 4)
	for i, l := range work {
		chunks[i%4] = append(chunks[i%4], l)
	}
	for i, lines := range chunks {
		if len(lines) > 0 {
			x := 66.6 + float64(i%2)*271.3
			y := 376.0 + float64(i/2)*46
			p1.box(x, y, 214, 27, 9, strings.Join(lines, "; "), false)
			if len(performed) > 0 {
				p1.workMark(x-17, y+11)
			}
		}
	}
	n := len(s.Lines)
	if n == 0 {
		return nil, fmt.Errorf("Cotización sin conceptos")
	}
	rows := n
	if rows < 4 {
		rows = 4
	}
	row := 114.0 / float64(rows)
	if row < 7 {
		return nil, fmt.Errorf("Demasiados conceptos para imprimir sin recortes en las dos páginas originales")
	}
	// Hide old row separators only within the detail body; preserve its original bounds/header.
	fmt.Fprintf(p1, "1 1 1 rg 30 %.2f 535 114 re f\n", height-626)
	for j := 0; j < rows; j++ {
		y := 512 + float64(j)*row
		if j%2 == 1 {
			fmt.Fprintf(p1, "0.97 0.975 0.98 rg 30 %.2f 535 %.2f re f\n", height-y-row, row)
		}
		fmt.Fprintf(p1, "0.90 G 0.4 w 30 %.2f m 565 %.2f l S\n", height-y-row, height-y-row)
	}
	for i, l := range s.Lines {
		y := 512 + float64(i)*row
		p1.text(42, y+9, 7, fmt.Sprint(i+1), false)
		p1.box(72, y, 427, row-1, 9, l.Concept+" | "+l.Quantity+" x S/ "+l.UnitPrice+" - dto/u S/ "+l.Discount, false)
		p1.box(505, y, 51, row-1, 9, "S/ "+l.Subtotal, false)
	}
	p1.boxBold(458, 642, 95, 27, 21, "S/ "+s.Total, true)
	p1.box(124, 772, 150, 10, 7, "Operario: "+s.Mechanic, false)
	p2.box(509, 15, 52, 14, 10, s.Plate, true)
	p2.box(426, 32, 135, 11, 8, s.Vehicle+" - "+s.Mileage+" km", true)
	p2.text(211, 82, 6, "Sin marca = Sin evaluar", false)
	for _, g := range Catalog.Groups {
		for _, i := range g.Items {
			v := s.Control.Inspections[i.ID]
			offset := map[string]float64{"Bien": 0, "Atención": 13.5, "Urgente": 27}[v]
			if v != "" && v != "Sin evaluar" {
				p2.text(i.X+offset-2, i.Y+3, 8, "X", false)
			}
		}
	}
	for _, g := range Catalog.Maintenance {
		selected := map[string]bool{}
		for _, v := range s.Control.Maintenance[g.ID] {
			selected[v] = true
		}
		for _, i := range g.Options {
			if selected[i.ID] || (g.ID == "aceite" && i.ID == "0") || (g.ID == "marca_aceite" && i.ID == "0") || (g.ID == "filtro_aceite" && i.ID == "1") {
				p2.mark(i.X, i.Y, selected[i.ID])
			}
		}
	}
	selected := map[string]bool{}
	for _, v := range s.Control.OtherServices {
		selected[v] = true
	}
	for _, i := range Catalog.OtherServices {
		if selected[i.ID] {
			p2.text(i.X-2, i.Y+3, 8, "X", false)
		}
	}
	p2.box(246, 219, 118, 13, 7, s.Control.OtherFluids, false)
	p2.box(374, 338, 178, 13, 8, s.Control.Coils, false)
	for i, k := range []string{"frontLeft", "frontRight", "rearLeft", "rearRight"} {
		if value := s.Control.Tires[k]; value != "" {
			x := 438.0
			if i%2 == 1 {
				x = 525
			}
			y := 240.5
			if i >= 2 {
				y = 253
			}
			fmt.Fprintf(p2, "1 1 1 rg %.2f %.2f 13 10 re f\n", x-1, height-y-2)
			p2.text(x, y, 6.5, value, false)
		}
	}
	p2.box(41, 661, 237, 70, 9, s.Control.ClientNotes, false)
	p2.box(315, 661, 237, 70, 9, s.Control.TechnicalNotes, false)
	p2.box(350, 759, 205, 14, 8, "a los "+or(s.Control.NextMileage, "Sin evaluar")+" km - fecha "+nextDate(s.Control.NextDate), false)
	if p1.err != nil {
		return nil, p1.err
	}
	if p2.err != nil {
		return nil, p2.err
	}
	return buildPDF(p1.Bytes(), p2.Bytes())
}
func nextDate(v string) string {
	if v == "" {
		return "Sin evaluar"
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return v
	}
	return t.Format("02/01/2006")
}
func or(v, otherwise string) string {
	if v == "" {
		return otherwise
	}
	return v
}
func buildPDF(a, b []byte) ([]byte, error) {
	objects := [][]byte{[]byte("<< /Type /Catalog /Pages 2 0 R >>"), []byte("<< /Type /Pages /Kids [4 0 R 7 0 R] /Count 2 >>"), []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")}
	for i, content := range [][]byte{a, b} {
		base := 4 + i*3
		objects = append(objects, []byte(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /Font << /F1 3 0 R /F2 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >> >> /XObject << /Bg %d 0 R >> >> /Contents %d 0 R >>", width, height, base+2, base+1)))
		var z bytes.Buffer
		w := zlib.NewWriter(&z)
		w.Write(content)
		w.Close()
		objects = append(objects, stream(fmt.Sprintf("/Filter /FlateDecode /Length %d", z.Len()), z.Bytes()))
		jpg, err := assets.ReadFile(fmt.Sprintf("assets/pagina-%d.jpg", i+1))
		if err != nil {
			return nil, err
		}
		cfg, err := jpeg.DecodeConfig(bytes.NewReader(jpg))
		if err != nil {
			return nil, err
		}
		objects = append(objects, stream(fmt.Sprintf("/Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d", cfg.Width, cfg.Height, len(jpg)), jpg))
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n", i+1)
		out.Write(obj)
		out.WriteString("\nendobj\n")
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, v := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", v)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return out.Bytes(), nil
}
func stream(dict string, data []byte) []byte {
	return append(append([]byte("<< "+dict+" >>\nstream\n"), data...), []byte("\nendstream")...)
}

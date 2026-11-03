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
	Type      string `json:"type"`
	Notes     string `json:"notes"`
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
type TechnicalTest struct {
	Kind   string `json:"kind"`
	Result string `json:"result"`
	Notes  string `json:"notes"`
}
type Snapshot struct {
	Diagnosis       string          `json:"diagnosis"`
	Evaluation      string          `json:"evaluation"`
	Recommendations string          `json:"recommendations"`
	DiagnosisNotes  string          `json:"diagnosisNotes"`
	Tests           []TechnicalTest `json:"tests"`
	AuthorizedTotal string          `json:"authorizedTotal"`
	Workshop        Workshop        `json:"workshop"`
	Code            string          `json:"code"`
	Client          string          `json:"client"`
	Document        string          `json:"document"`
	Phone           string          `json:"phone"`
	Address         string          `json:"address"`
	Email           string          `json:"email"`
	Plate           string          `json:"plate"`
	Vehicle         string          `json:"vehicle"`
	Mileage         string          `json:"mileage"`
	Symptoms        string          `json:"symptoms"`
	In              string          `json:"in"`
	Out             string          `json:"out"`
	Mechanic        string          `json:"mechanic"`
	Total           string          `json:"total"`
	Lines           []Line          `json:"lines"`
	Advances        []string        `json:"advances"`
	Control         Control         `json:"control"`
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
	color string
	bytes.Buffer
	err    error
	bold   bool
	extras []string
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
	color := or(c.color, "0.12 0.12 0.13")
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
	c.extras = append(c.extras, s)
	c.text(x, y+6, 5.5, "Ver información completa en anexo.", white)
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

// Render uses the current visual layout from an immutable attention snapshot.
// Updating presentation never rewrites historical archived PDF bytes.
func Render(raw []byte) ([]byte, error) {
	var snapshot Snapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return renderSheetLayout(snapshot)
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
func buildPDF(pages ...[]byte) ([]byte, error) {
	return buildDocument(true, pages...)
}

// Current documents embed the official raster emblem; legacy backgrounds remain available only for compatibility.
func buildDocument(template bool, pages ...[]byte) ([]byte, error) {
	objects := [][]byte{[]byte("<< /Type /Catalog /Pages 2 0 R >>"), []byte("PAGES_PLACEHOLDER"), []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")}
	kids := ""
	for i := range pages {
		kids += fmt.Sprintf("%d 0 R ", 4+i*3)
	}
	objects[1] = []byte(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", kids, len(pages)))
	for i, content := range pages {
		base := 4 + i*3
		objects = append(objects, []byte(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /Font << /F1 3 0 R /F2 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >> >> /XObject << /Bg %d 0 R >> >> /Contents %d 0 R >>", width, height, base+2, base+1)))
		var z bytes.Buffer
		w := zlib.NewWriter(&z)
		w.Write(content)
		w.Close()
		objects = append(objects, stream(fmt.Sprintf("/Filter /FlateDecode /Length %d", z.Len()), z.Bytes()))
		if !template {
			logo, err := quoteLogoObject()
			if err != nil {
				return nil, err
			}
			objects = append(objects, logo)
			continue
		}
		if i >= 2 {
			objects = append(objects, []byte("<< >>"))
			continue
		}
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

func orNote(notes string) string {
	if notes == "" {
		return ""
	}
	return " (Obs.: " + notes + ")"
}

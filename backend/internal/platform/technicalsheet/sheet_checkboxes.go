package technicalsheet

import (
	"fmt"
	"strings"
)

// Helvetica AFM widths, matching the PDF base fonts. Accurate metrics keep
// labels readable without the false line breaks of the generic quote estimator.
var sheetHelvetica = [...]float64{278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278, 556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556, 1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778, 667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556, 333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556, 556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584}
var sheetHelveticaBold = [...]float64{278, 333, 474, 556, 556, 889, 722, 238, 333, 333, 389, 584, 278, 333, 278, 278, 556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 333, 333, 584, 584, 584, 611, 975, 722, 722, 722, 722, 667, 611, 778, 722, 278, 556, 722, 611, 833, 722, 778, 667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 333, 278, 333, 584, 556, 333, 556, 611, 556, 611, 556, 333, 611, 611, 278, 278, 556, 278, 889, 611, 611, 611, 611, 389, 556, 333, 611, 556, 778, 556, 556, 500, 389, 280, 389, 584}

func sheetTextWidth(text string, size float64, bold bool) float64 {
	widths := sheetHelvetica
	if bold {
		widths = sheetHelveticaBold
	}
	w := 0.0
	for _, r := range text {
		for _, group := range []string{"aàáâãäå", "eèéêë", "iìíîï", "oòóôõöø", "uùúûü", "nñ", "cç", "yýÿ", "AÀÁÂÃÄÅ", "EÈÉÊË", "IÌÍÎÏ", "OÒÓÔÕÖØ", "UÙÚÛÜ", "NÑ", "CÇ", "YÝŸ"} {
			if strings.ContainsRune(group, r) {
				r = []rune(group)[0]
				break
			}
		}
		if r >= 32 && r <= 126 {
			w += widths[r-32]
		} else if r == '—' {
			w += 1000
		} else if r == '·' {
			w += 278
		} else {
			w += 1000
		}
	}
	return w * size / 1000 * 1.02
}
func sheetWrap(text string, size, width float64, bold bool) []string {
	var result []string
	for _, paragraph := range strings.Split(text, "\n") {
		runes := []rune(paragraph)
		if len(runes) == 0 {
			result = append(result, "")
			continue
		}
		for start := 0; start < len(runes); {
			end, lastSpace, used := start, -1, 0.0
			for end < len(runes) && used+sheetTextWidth(string(runes[end]), size, bold) <= width {
				used += sheetTextWidth(string(runes[end]), size, bold)
				if runes[end] == ' ' {
					lastSpace = end
				}
				end++
			}
			if end < len(runes) && lastSpace > start {
				end = lastSpace + 1
			}
			if end == start {
				end++
			}
			result = append(result, string(runes[start:end]))
			start = end
		}
	}
	return result
}

const sheetBoxSize = 8.5

// Printed marks are drawn only from the recorded control, never quote items.
func sheetCheckbox(c *canvas, x, y float64, checked bool, color string) {
	fmt.Fprintf(c, "q %s RG 0.65 w %.2f %.2f %.2f %.2f re S Q\n", color, x, height-y-sheetBoxSize, sheetBoxSize, sheetBoxSize)
	if checked {
		c.text(x+1.2, y+7.1, 8, "X", false)
	}
}

// A bordered card keeps each category distinct. Measurements and drawing share
// the same font and line widths; the card height is common across the three columns.
func sheetInspectionHeight(g Group, tires map[string]string) float64 {
	h := 34.0
	for _, item := range g.Items {
		h += float64(len(sheetWrap(item.Label, 9.2, 104, false)))*12.5 + 1
	}
	if g.ID == "neumaticos" {
		h += 29
		for _, id := range []string{"frontLeft", "frontRight", "rearLeft", "rearRight"} {
			if n := len(sheetWrap(or(tires[id], "—"), 9.2, 27, false)); n > 1 {
				h += float64(n-1) * 12
			}
		}
	}
	return h + 8
}
func sheetCard(c *canvas, x, y, w, h float64) {
	r := 3.0
	k := r * .55228475
	left, right, bottom, top := x, x+w, height-y-h, height-y
	fmt.Fprintf(c, "q 0.84 0.84 0.86 RG 0.65 w %.2f %.2f m %.2f %.2f l %.2f %.2f %.2f %.2f %.2f %.2f c %.2f %.2f l %.2f %.2f %.2f %.2f %.2f %.2f c %.2f %.2f l %.2f %.2f %.2f %.2f %.2f %.2f c %.2f %.2f l %.2f %.2f %.2f %.2f %.2f %.2f c S Q\n",
		left+r, bottom, right-r, bottom, right-r+k, bottom, right, bottom+r-k, right, bottom+r,
		right, top-r, right, top-r+k, right-r+k, top, right-r, top,
		left+r, top, left+r-k, top, left, top-r+k, left, top-r,
		left, bottom+r, left, bottom+r-k, left+r-k, bottom, left+r, bottom)
}
func sheetInspectionGroup(c *canvas, g Group, values map[string]string, x, y float64, tires ...map[string]string) float64 {
	documentTitle(c, x+8, y+14, 9.2, g.Label)
	quoteRect(c, x+8, y+21, 153, .4, "0.87 0.87 0.89", false)
	for index, label := range []string{"B", "A", "U"} {
		c.text(x+120+float64(index)*13, y+16, 8.5, label, false)
	}
	y += 34
	for _, item := range g.Items {
		for index, state := range []string{"Bien", "Atención", "Urgente"} {
			sheetCheckbox(c, x+119+float64(index)*13, y-7, values[item.ID] == state, []string{"0.13 0.42 0.26", "0.60 0.42 0.10", brandBurgundy}[index])
		}
		for _, line := range sheetWrap(item.Label, 9.2, 104, false) {
			c.text(x+8, y, 9.2, line, false)
			y += 12.5
		}
		y++
	}
	if g.ID == "neumaticos" && len(tires) > 0 {
		c.text(x+8, y+3, 8.5, "Presión PSI (— = Sin evaluar)", false)
		y += 17
		for row := 0; row < 2; row++ {
			maxLines := 1
			for col := 0; col < 2; col++ {
				index := row*2 + col
				id := []string{"frontLeft", "frontRight", "rearLeft", "rearRight"}[index]
				label := []string{"Del. piloto", "Del. copil.", "Tras. piloto", "Tras. copil."}[index]
				xx := x + 8 + float64(col)*79
				lines := sheetWrap(or(tires[0][id], "—"), 9.2, 27, false)
				if len(lines) > maxLines {
					maxLines = len(lines)
				}
				c.text(xx, y, 8.2, label, false)
				for i, text := range lines {
					c.text(xx+51, y+float64(i)*12, 9.2, text, false)
				}
			}
			y += float64(maxLines) * 12
		}
	}
	return y + 8
}

// The same flow is used for measurement and drawing, preserving long catalogue
// labels and every stored selection without inventing or shortening options.
func sheetOptionFlow(c *canvas, options []Item, selected []string, x, y, w float64) float64 {
	chosen := map[string]bool{}
	for _, id := range selected {
		chosen[id] = true
	}
	cursor, top := x, y
	known := map[string]bool{}
	for _, option := range options {
		known[option.ID] = true
		tw := sheetTextWidth(option.Label, 9.2, false) + 18.5
		if cursor > x && cursor+tw > x+w {
			cursor = x
			top += 13
		}
		if tw > w {
			if cursor > x {
				cursor = x
				top += 13
			}
			if c != nil {
				sheetCheckbox(c, cursor, top-7, chosen[option.ID], "0.35 0.35 0.38")
			}
			for _, line := range sheetWrap(option.Label, 9.2, w-15, false) {
				if c != nil {
					c.text(cursor+15, top, 9.2, line, false)
				}
				top += 13
			}
			cursor = x
			continue
		}
		if c != nil {
			sheetCheckbox(c, cursor, top-7, chosen[option.ID], "0.35 0.35 0.38")
			c.text(cursor+15, top, 9.2, option.Label, false)
		}
		cursor += tw
	}
	top += 13
	for _, id := range selected {
		if !known[id] {
			for _, line := range sheetWrap("Referencia registrada: "+id, 9.2, w, false) {
				if c != nil {
					c.text(x, top, 9.2, line, false)
				}
				top += 13
			}
		}
	}
	return top
}
func sheetMaintenanceHeight(g Group, selected []string) float64 {
	label := float64(len(sheetWrap(g.Label, 9.2, 60, true))) * 13
	options := sheetOptionFlow(nil, g.Options, selected, 0, 0, 171)
	if options > label {
		return options + 2
	}
	return label + 2
}
func (p *sheetLayout) maintenanceChecks(control Control) {
	var heights []float64
	total := 0.0
	for index := 0; index < len(Catalog.Maintenance); index += 2 {
		h := 0.0
		for col := 0; col < 2 && index+col < len(Catalog.Maintenance); col++ {
			g := Catalog.Maintenance[index+col]
			if v := sheetMaintenanceHeight(g, control.Maintenance[g.ID]); v > h {
				h = v
			}
		}
		heights = append(heights, h)
		total += h
	}
	framed := total+8 <= documentBottom-124
	if framed {
		p.room(total + 8)
		sheetCard(p.c, 34, p.y-10, 527, total)
		quoteRect(p.c, 297.5, p.y-3, .4, total-14, "0.87 0.87 0.89", false)
	}
	for row, h := range heights {
		index := row * 2
		if !framed {
			p.room(h + 8)
		}
		for col := 0; col < 2 && index+col < len(Catalog.Maintenance); col++ {
			g := Catalog.Maintenance[index+col]
			x := 34 + float64(col)*272
			if !framed {
				sheetCard(p.c, x, p.y-10, 255, h)
			}
			if row < len(heights)-1 {
				quoteRect(p.c, x+8, p.y+h-10, 239, .3, "0.91 0.91 0.92", false)
			}
			quoteRect(p.c, x+72, p.y-5, .4, h-10, "0.89 0.89 0.90", false)
			for line, label := range sheetWrap(g.Label, 9.2, 60, true) {
				documentTitle(p.c, x+8, p.y+float64(line)*13, 9.2, label)
			}
			sheetOptionFlow(p.c, g.Options, control.Maintenance[g.ID], x+79, p.y, 171)
		}
		p.y += h
	}
}
func (p *sheetLayout) otherServiceChecks(selected []string) {
	chosen := map[string]bool{}
	known := map[string]bool{}
	for _, id := range selected {
		chosen[id] = true
	}
	total := 0.0
	for index := 0; index < len(Catalog.OtherServices); index += 3 {
		h := 0.0
		for col := 0; col < 3 && index+col < len(Catalog.OtherServices); col++ {
			v := float64(len(sheetWrap(Catalog.OtherServices[index+col].Label, 9.2, 138, false)))*13 + 4
			if v > h {
				h = v
			}
		}
		total += h
	}
	p.room(total + 8)
	sheetCard(p.c, 34, p.y-11, 527, total)
	for index := 0; index < len(Catalog.OtherServices); index += 3 {
		h := 0.0
		for col := 0; col < 3 && index+col < len(Catalog.OtherServices); col++ {
			item := Catalog.OtherServices[index+col]
			v := float64(len(sheetWrap(item.Label, 9.2, 138, false)))*13 + 4
			if v > h {
				h = v
			}
		}
		p.room(h + 8)
		for col := 0; col < 3 && index+col < len(Catalog.OtherServices); col++ {
			item := Catalog.OtherServices[index+col]
			known[item.ID] = true
			x := 34 + float64(col)*179
			sheetCheckbox(p.c, x+7, p.y-7, chosen[item.ID], "0.35 0.35 0.38")
			for row, label := range sheetWrap(item.Label, 9.2, 138, false) {
				p.c.text(x+23, p.y+float64(row)*13, 9.2, label, false)
			}
		}
		p.y += h
	}
	for _, id := range selected {
		if !known[id] {
			p.note("Referencia registrada", id)
		}
	}
}

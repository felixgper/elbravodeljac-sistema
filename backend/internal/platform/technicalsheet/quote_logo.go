package technicalsheet

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
)

// The first-party SVG is embedded for deployment, then drawn as PDF vectors;
// there is no screenshot, foreign branding, raster logo or network dependency.
func drawQuoteLogo(c *canvas, x, y, w float64) {
	raw, err := assets.ReadFile("assets/bravo-borgona.svg")
	if err != nil {
		c.err = err
		return
	}
	var svg struct {
		Paths []struct {
			D      string `xml:"d,attr"`
			Fill   string `xml:"fill,attr"`
			Stroke string `xml:"stroke,attr"`
			Width  string `xml:"stroke-width,attr"`
		} `xml:"path"`
	}
	if err = xml.Unmarshal(raw, &svg); err != nil {
		c.err = err
		return
	}
	fmt.Fprintf(c, "q %.5f 0 0 %.5f %.2f %.2f cm\n", w/120, -w/120, x, height-y)
	for _, p := range svg.Paths {
		if p.Fill != "" {
			quoteSVGColor(c, p.Fill, "rg")
		}
		if p.Stroke != "" {
			quoteSVGColor(c, p.Stroke, "RG")
			fmt.Fprintf(c, "%s w 1 J 1 j\n", p.Width)
		}
		if err = quoteSVGPath(c, p.D); err != nil {
			c.err = err
			return
		}
		mode := "f"
		if p.Stroke != "" {
			mode = "S"
			if p.Fill != "" {
				mode = "B"
			}
		}
		fmt.Fprintln(c, mode)
	}
	fmt.Fprint(c, "Q\n")
}
func quoteSVGColor(c *canvas, hex, operator string) {
	var r, g, b uint64
	if len(hex) == 7 {
		r, _ = strconv.ParseUint(hex[1:3], 16, 8)
		g, _ = strconv.ParseUint(hex[3:5], 16, 8)
		b, _ = strconv.ParseUint(hex[5:7], 16, 8)
	} else if len(hex) == 4 {
		r, _ = strconv.ParseUint(hex[1:2]+hex[1:2], 16, 8)
		g, _ = strconv.ParseUint(hex[2:3]+hex[2:3], 16, 8)
		b, _ = strconv.ParseUint(hex[3:4]+hex[3:4], 16, 8)
	}
	fmt.Fprintf(c, "%.4f %.4f %.4f %s\n", float64(r)/255, float64(g)/255, float64(b)/255, operator)
}

var quoteSVGToken = regexp.MustCompile(`[A-Za-z]|[-+]?(?:\d*\.)?\d+`)

func quoteSVGPath(c *canvas, path string) error {
	tokens := quoteSVGToken.FindAllString(path, -1)
	i := 0
	cmd := byte(0)
	x, y, sx, sy := 0.0, 0.0, 0.0, 0.0
	for i < len(tokens) {
		if len(tokens[i]) == 1 && (tokens[i][0] >= 'A' && tokens[i][0] <= 'Z' || tokens[i][0] >= 'a' && tokens[i][0] <= 'z') {
			cmd = tokens[i][0]
			i++
		}
		relative := cmd >= 'a' && cmd <= 'z'
		upper := cmd
		if relative {
			upper -= 32
		}
		if upper == 'Z' {
			fmt.Fprintln(c, "h")
			x, y = sx, sy
			cmd = 0
			continue
		}
		n := 0
		switch upper {
		case 'M', 'L':
			n = 2
		case 'H', 'V':
			n = 1
		case 'C':
			n = 6
		default:
			return fmt.Errorf("Comando SVG institucional no admitido: %c", cmd)
		}
		if i+n > len(tokens) {
			return fmt.Errorf("Logo institucional incompleto")
		}
		v := make([]float64, n)
		for k := range v {
			parsed, err := strconv.ParseFloat(tokens[i+k], 64)
			if err != nil {
				return err
			}
			v[k] = parsed
		}
		i += n
		if relative {
			switch upper {
			case 'H':
				v[0] += x
			case 'V':
				v[0] += y
			default:
				for k := 0; k < n; k += 2 {
					v[k] += x
					v[k+1] += y
				}
			}
		}
		switch upper {
		case 'M':
			x, y = v[0], v[1]
			sx, sy = x, y
			fmt.Fprintf(c, "%.3f %.3f m\n", x, y)
			cmd = 'L'
			if relative {
				cmd = 'l'
			}
		case 'L':
			x, y = v[0], v[1]
			fmt.Fprintf(c, "%.3f %.3f l\n", x, y)
		case 'H':
			x = v[0]
			fmt.Fprintf(c, "%.3f %.3f l\n", x, y)
		case 'V':
			y = v[0]
			fmt.Fprintf(c, "%.3f %.3f l\n", x, y)
		case 'C':
			fmt.Fprintf(c, "%.3f %.3f %.3f %.3f %.3f %.3f c\n", v[0], v[1], v[2], v[3], v[4], v[5])
			x, y = v[4], v[5]
		}
	}
	return nil
}

package technicalsheet

import "fmt"

const brandBurgundy = "0.25 0.035 0.085"
const brandSurface = "0.965 0.965 0.97"
const documentMargin = 34.0
const documentWidth = 527.0
const documentBottom = 756.0

// Official shield, unchanged raster; the small charcoal plate is intentional.
// References are graphic inspiration only: no commercial conditions are added.
func documentHeader(c *canvas, title, reference, revision, workshop string) {
	quoteRect(c, 34, 28, 527, 77, brandBurgundy, false)
	quoteRect(c, 550, 28, 11, 77, "0.64 0.06 0.17", false)
	drawQuoteLogo(c, 44, 35, 56)
	c.bold = true
	for i, line := range quoteWrap(or(workshop, "EL BRAVO DEL JAC"), 16, 260) {
		c.text(112, 53+float64(i)*18, 16, line, true)
	}
	c.bold = false
	c.text(112, 86, 8, "TALLER AUTOMOTRIZ", true)
	c.bold = true
	quoteRightText(c, 538, 49, 12, title, true)
	c.bold = false
	quoteRightText(c, 538, 71, 10, reference, true)
	quoteRightText(c, 538, 91, 8, revision, true)
}
func documentFooter(c *canvas, reference string, page, total int) {
	c.text(34, 780, 7.5, reference, false)
	quoteRect(c, 34, 792, 527, 24, brandBurgundy, false)
	c.bold = true
	c.text(44, 807, 8, "GRACIAS POR CONFIAR EN NOSOTROS", true)
	c.bold = false
	quoteRightText(c, 551, 807, 8, fmt.Sprintf("EL BRAVO DEL JAC · PÁGINA %d DE %d", page, total), true)
}
func finishDocument(pages [][]byte, reference string) ([]byte, error) {
	for i, content := range pages {
		c := &canvas{}
		c.Write(content)
		documentFooter(c, reference, i+1, len(pages))
		if c.err != nil {
			return nil, c.err
		}
		pages[i] = c.Bytes()
	}
	return buildDocument(false, pages...)
}
func documentTitle(c *canvas, x, y, size float64, text string) {
	priorColor, priorBold := c.color, c.bold
	c.color = brandBurgundy
	c.bold = true
	c.text(x, y, size, text, false)
	c.color = priorColor
	c.bold = priorBold
}
func documentPanel(c *canvas, x, y, w, h float64) { quoteRect(c, x, y, w, h, brandSurface, true) }

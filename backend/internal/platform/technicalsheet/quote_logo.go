package technicalsheet

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image/png"
)

// The official raster is kept unchanged. Only its shield viewport is rendered;
// no transparency, vector reconstruction or replacement artwork is invented.
func drawQuoteLogo(c *canvas, x, y, w float64) {
	h := w * 430 / 410
	fmt.Fprintf(c, "q %.3f 0 0 %.3f %.3f %.3f cm /Bg Do Q\n", w, h, x, height-y-h)
}
func quoteLogoObject() ([]byte, error) {
	raw, err := assets.ReadFile("assets/diseno19-original.png")
	if err != nil {
		return nil, err
	}
	im, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	var pixels bytes.Buffer
	for y := 170; y < 600; y++ {
		for x := 130; x < 540; x++ {
			r, g, b, _ := im.At(x, y).RGBA()
			pixels.Write([]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8)})
		}
	}
	var compressed bytes.Buffer
	z := zlib.NewWriter(&compressed)
	if _, err = z.Write(pixels.Bytes()); err != nil {
		return nil, err
	}
	if err = z.Close(); err != nil {
		return nil, err
	}
	return stream(fmt.Sprintf("/Type /XObject /Subtype /Image /Width 410 /Height 430 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode /Length %d", compressed.Len()), compressed.Bytes()), nil
}

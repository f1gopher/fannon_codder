package terrain

import (
	"image"
	"image/color"
)

// SilhouettePad is the empty border PropSilhouette adds on every side so
// the blur is not cut off. Callers shift the anchor by this many pixels.
const SilhouettePad = 14

// A hut or crate shadow is the sprite's own alpha, softened. Light comes
// from the north-west, so the copy sits a little south-east of the sprite.
// Shifts are world pixels.
const (
	PropShiftX = 1.75
	PropShiftY = 2.75
)

// PropSilhouette is a soft black copy of src's alpha. The shape is the
// sprite. Soldiers stay on the oval in render.
func PropSilhouette(src *image.RGBA) *image.RGBA {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 1 || h < 1 {
		return nil
	}
	pad := SilhouettePad
	pw, ph := w+2*pad, h+2*pad
	a := make([]float32, pw*ph)
	for y := 0; y < h; y++ {
		row := src.Pix[y*src.Stride:]
		for x := 0; x < w; x++ {
			a[(y+pad)*pw+(x+pad)] = float32(row[x*4+3])
		}
	}
	a = boxBlur(boxBlur(a, pw, ph, 7), pw, ph, 7)
	out := image.NewRGBA(image.Rect(0, 0, pw, ph))
	for y := 0; y < ph; y++ {
		for x := 0; x < pw; x++ {
			v := a[y*pw+x] * 0.5
			if v < 1 {
				continue
			}
			if v > 255 {
				v = 255
			}
			out.SetRGBA(x, y, color.RGBA{A: uint8(v)})
		}
	}
	return out
}

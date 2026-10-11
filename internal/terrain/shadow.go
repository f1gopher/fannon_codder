package terrain

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// moundW is the tree contact oval in world pixels. It matches the battle's
// old dummy image of 14*ArtScale pixels (ArtScale is 8), read back as 14.
const moundW = 14.0

var moundOnce struct {
	img *ebiten.Image
}

func moundOval() *ebiten.Image {
	if moundOnce.img != nil {
		return moundOnce.img
	}
	const w, h = 64, 32
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	cx := (w - 1) / 2.0
	cy := (h - 1) / 2.0
	rx := w / 2.0
	ry := h / 2.0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			d := dx*dx + dy*dy
			if d >= 1 {
				continue
			}
			fade := (1 - d) * (1 - d)
			rgba.SetRGBA(x, y, color.RGBA{A: uint8(fade * 150)})
		}
	}
	moundOnce.img = ebiten.NewImageFromImage(rgba)
	return moundOnce.img
}

// DrawShadow sits the tree contact oval on the canopy foot. Placement
// matches render.castShadow for a 14-world-pixel mound. Unit shadows stay
// in render.
func (c Canopy) DrawShadow(dst *ebiten.Image, originX, originY, scale float64) {
	if dst == nil {
		return
	}
	if scale <= 0 {
		scale = 1
	}
	rw := moundW
	rh := rw * 0.35
	oval := moundOval()
	b := oval.Bounds()
	x := originX + c.FootX
	y := originY + c.FootY
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(rw/float64(b.Dx())*scale, rh/float64(b.Dy())*scale)
	op.GeoM.Translate((x-rw/2)*scale, (y-rh*0.85)*scale)
	dst.DrawImage(oval, op)
}

package render

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	ground "fannon-codder/internal/terrain"
)

// shadowOval is a soft black ellipse drawn under a painted body. It is not
// part of the sheet. Rectangle placeholders do not cast one, so a missing
// sheet still looks like the chunk 32 field.
var shadowOnce struct {
	img *ebiten.Image
}

func shadowOval() *ebiten.Image {
	if shadowOnce.img != nil {
		return shadowOnce.img
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
	shadowOnce.img = ebiten.NewImageFromImage(rgba)
	return shadowOnce.img
}

// castShadow sits an oval on the anchor. Most of it is above that point,
// under the body.
func castShadow(dst *ebiten.Image, x, y float64, src *ebiten.Image) {
	if dst == nil || src == nil {
		return
	}
	rw := float64(src.Bounds().Dx()) / ArtScale
	if rw < 1 {
		rw = 1
	}
	drawShadowOval(dst, x, y, rw, rw*0.35, 0, -rw*0.35*0.85)
}

func drawShadowOval(dst *ebiten.Image, x, y, rw, rh, ox, oy float64) {
	if dst == nil || rw < 1 || rh <= 0 {
		return
	}
	oval := shadowOval()
	b := oval.Bounds()
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(rw/float64(b.Dx())*ps, rh/float64(b.Dy())*ps)
	op.GeoM.Translate((x+ox-rw/2)*ps, (y+oy)*ps)
	dst.DrawImage(oval, op)
}

var formCache struct {
	mu sync.Mutex
	m  map[*image.RGBA]*ebiten.Image
}

func formImage(src *image.RGBA) *ebiten.Image {
	formCache.mu.Lock()
	defer formCache.mu.Unlock()
	if formCache.m == nil {
		formCache.m = map[*image.RGBA]*ebiten.Image{}
	}
	if img := formCache.m[src]; img != nil {
		return img
	}
	sil := ground.PropSilhouette(src)
	if sil == nil {
		return nil
	}
	img := ebiten.NewImageFromImage(sil)
	formCache.m[src] = img
	return img
}

// castForm draws the soft silhouette on the sheet anchor, shifted
// south-east. x and y are that anchor in logical pixels.
func castForm(dst *ebiten.Image, src *image.RGBA, ax, ay int, x, y float64) {
	if dst == nil || src == nil {
		return
	}
	img := formImage(src)
	if img == nil {
		return
	}
	sc := SpriteScale()
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(-float64(ax+ground.SilhouettePad), -float64(ay+ground.SilhouettePad))
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate((x+ground.PropShiftX)*ps, (y+ground.PropShiftY)*ps)
	dst.DrawImage(img, op)
}

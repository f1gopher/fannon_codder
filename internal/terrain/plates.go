package terrain

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

//go:embed plates/*.png
var plateFS embed.FS

// plateScale is painted pixels per world pixel. One keeps a campaign map
// inside a single copy of each plate, so the picture is not a tiled stamp.
const plateScale = 1.0

// painted is one overhead fill from the approved meadow picture.
type painted struct {
	w, h int
	pix  []uint8
}

var (
	grassPlate   *painted
	canopyPlate  *painted
	shallowPlate *painted
	deepPlate    *painted
	sandPlate    *painted
	snowPlate    *painted
	rockPlate    *painted
	icePlate     *painted
	dirtPlate    *painted
)

func init() {
	grassPlate = mustPlate("plates/grass.png")
	canopyPlate = mustPlate("plates/canopy.png")
	shallowPlate = mustPlate("plates/shallow.png")
	deepPlate = mustPlate("plates/deep.png")
	sandPlate = mustPlate("plates/sand.png")
	snowPlate = mustPlate("plates/snow.png")
	rockPlate = mustPlate("plates/rock.png")
	icePlate = mustPlate("plates/ice.png")
	dirtPlate = mustPlate("plates/dirt.png")
}

func mustPlate(name string) *painted {
	b, err := plateFS.ReadFile(name)
	if err != nil {
		panic(err)
	}
	src, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	r := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(dst, dst.Bounds(), src, r.Min, draw.Src)
	return &painted{w: dst.Rect.Dx(), h: dst.Rect.Dy(), pix: dst.Pix}
}

// at samples the painting in world space. Past the picture the sample mirrors
// so a larger map does not show a cut.
func (p *painted) at(wx, wy float64) color.RGBA {
	return p.atScale(wx, wy, plateScale)
}

// atScale uses painted pixels per world pixel. Canopy and rock are denser
// than the ground plates so one tree reads as crowns, not a colour blob.
func (p *painted) atScale(wx, wy, scale float64) color.RGBA {
	x := mirror(wx*scale, p.w)
	y := mirror(wy*scale, p.h)
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	if x0 >= p.w-1 {
		x0 = p.w - 2
	}
	if y0 >= p.h-1 {
		y0 = p.h - 2
	}
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	fx := x - float64(x0)
	fy := y - float64(y0)
	c00 := p.px(x0, y0)
	c10 := p.px(x0+1, y0)
	c01 := p.px(x0, y0+1)
	c11 := p.px(x0+1, y0+1)
	return color.RGBA{
		R: lerpByte(lerpByte(c00.R, c10.R, fx), lerpByte(c01.R, c11.R, fx), fy),
		G: lerpByte(lerpByte(c00.G, c10.G, fx), lerpByte(c01.G, c11.G, fx), fy),
		B: lerpByte(lerpByte(c00.B, c10.B, fx), lerpByte(c01.B, c11.B, fx), fy),
		A: 255,
	}
}

func (p *painted) px(x, y int) color.RGBA {
	i := (y*p.w + x) * 4
	return color.RGBA{R: p.pix[i], G: p.pix[i+1], B: p.pix[i+2], A: 255}
}

func mirror(v float64, n int) float64 {
	if n <= 1 {
		return 0
	}
	span := float64(n - 1)
	p := math.Mod(v, span*2)
	if p < 0 {
		p += span * 2
	}
	if p > span {
		p = span*2 - p
	}
	return p
}

func lerpByte(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + (float64(b)-float64(a))*t)
}

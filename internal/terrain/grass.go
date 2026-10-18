package terrain

import (
	"image/color"
	"math"
)

// Meadow colours follow the jungle ground plate's body. That plate averages
// about 114, 120, 56. GrassLo and GrassHi are a brighter green than the
// plate, so the field does not use them.
var (
	grassShade = color.RGBA{R: 46, G: 62, B: 30, A: 255}
	grassSun   = color.RGBA{R: 178, G: 184, B: 82, A: 255}
	grassDry   = color.RGBA{R: 168, G: 176, B: 96, A: 255}
	grassDark  = color.RGBA{R: 28, G: 44, B: 20, A: 255}
)

// grassPix is one jungle paint, one colour per world pixel. A per-source-pixel
// noise does not fit the one-second paint budget, and the draw's linear
// filter covers the step.
var (
	grassPix []color.RGBA
	grassW   int
)

// prepareGrass builds the jungle field for a map of w by h world pixels.
// Other kinds clear it. The next paint replaces it.
func prepareGrass(w, h int, kind Kind) {
	grassPix = nil
	grassW = 0
	if kind != Jungle || w < 1 || h < 1 {
		return
	}
	pix := make([]color.RGBA, w*h)
	for y := 0; y < h; y++ {
		wy := float64(y) + 0.5
		row := y * w
		for x := 0; x < w; x++ {
			pix[row+x] = grassColor(float64(x)+0.5, wy)
		}
	}
	grassPix = pix
	grassW = w
}

// grassSample is the jungle colour at a world position. During a jungle
// paint it reads the prepared field.
func grassSample(wx, wy float64) color.RGBA {
	if grassPix == nil || grassW < 1 {
		return grassColor(wx, wy)
	}
	x := int(math.Floor(wx))
	y := int(math.Floor(wy))
	h := len(grassPix) / grassW
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= grassW {
		x = grassW - 1
	}
	if y >= h {
		y = h - 1
	}
	return grassPix[y*grassW+x]
}

// grassColor is jungle open ground at one world pixel. The colour depends
// only on that position, so the field does not repeat a plate. Coordinates
// snap to the world-pixel centre.
func grassColor(wx, wy float64) color.RGBA {
	wx = math.Floor(wx) + 0.5
	wy = math.Floor(wy) + 0.5
	// Wind runs south-east, the same way canopy shadow falls.
	// (0.82, 0.42) normalised. The cross axis is the short side of a blade.
	const (
		windX  = 0.890
		windY  = 0.456
		crossX = -0.456
		crossY = 0.890
	)
	field := valueNoise(wx, wy, 88, 11)
	drift := valueNoise(wx*0.58, wy*1.08, 38, 12)
	t := clamp01(0.06 + 0.88*field + 0.12*(drift-0.5))
	c := mixRGBA(grassShade, grassSun, smooth01(t))
	c = mixRGBA(c, grassDry, smooth01((drift-0.55)/0.40)*0.22)
	// Shade the low field again so the sun and the shade stay apart.
	c = mixRGBA(c, grassDark, smooth01(0.42-field)*0.45)

	// Drift bends a short south-east grain so it is not one ruled diagonal.
	bend := (drift - 0.5) * 10
	along := wx*windX + wy*windY + bend
	across := wx*crossX + wy*crossY
	grain := valueNoise(along, across*1.8, 3.4, 19)
	return shiftBlade(c, (grain-0.5)*0.12)
}

func shiftBlade(c color.RGBA, d float64) color.RGBA {
	r := float64(c.R) + d*80
	g := float64(c.G) + d*110
	b := float64(c.B) + d*40
	if r > g {
		r = g
	}
	if b > g {
		b = g
	}
	return color.RGBA{
		R: clampByte(int(r + 0.5)),
		G: clampByte(int(g + 0.5)),
		B: clampByte(int(b + 0.5)),
		A: 255,
	}
}

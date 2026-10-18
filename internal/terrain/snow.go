package terrain

import (
	"image/color"
	"math"
)

// Drift colours follow the snow ground plate's body. That plate averages
// about 187, 193, 214. SnowLo and SnowHi are a flatter white than the
// plate, so the field does not use them.
var (
	snowShade  = color.RGBA{R: 132, G: 152, B: 190, A: 255}
	snowSun    = color.RGBA{R: 236, G: 238, B: 244, A: 255}
	snowHollow = color.RGBA{R: 108, G: 128, B: 178, A: 255}
	snowCrest  = color.RGBA{R: 246, G: 248, B: 252, A: 255}
)

// snowPix is one snow paint, one colour per world pixel. A per-source-pixel
// noise does not fit the one-second paint budget, and the draw's linear
// filter covers the step.
var (
	snowPix []color.RGBA
	snowW   int
)

// prepareSnow builds the snow field for a map of w by h world pixels.
// Other kinds clear it. The next paint replaces it.
func prepareSnow(w, h int, kind Kind) {
	snowPix = nil
	snowW = 0
	if kind != Snow || w < 1 || h < 1 {
		return
	}
	pix := make([]color.RGBA, w*h)
	for y := 0; y < h; y++ {
		wy := float64(y) + 0.5
		row := y * w
		for x := 0; x < w; x++ {
			pix[row+x] = snowColor(float64(x)+0.5, wy)
		}
	}
	snowPix = pix
	snowW = w
}

// snowSample is the snow colour at a world position. During a snow paint
// it reads the prepared field.
func snowSample(wx, wy float64) color.RGBA {
	if snowPix == nil || snowW < 1 {
		return snowColor(wx, wy)
	}
	x := int(math.Floor(wx))
	y := int(math.Floor(wy))
	h := len(snowPix) / snowW
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= snowW {
		x = snowW - 1
	}
	if y >= h {
		y = h - 1
	}
	return snowPix[y*snowW+x]
}

// snowColor is snow open ground at one world pixel. The colour depends
// only on that position, so the field does not repeat a plate. Coordinates
// snap to the world-pixel centre. Broad drifts carry the colour. A short
// south-east grain keeps the surface from reading as a flat fill.
func snowColor(wx, wy float64) color.RGBA {
	wx = math.Floor(wx) + 0.5
	wy = math.Floor(wy) + 0.5
	// Light falls from the north-west, the same way canopy shadow falls.
	// (0.82, 0.42) normalised. The cross axis is the short side of a drift.
	const (
		windX  = 0.890
		windY  = 0.456
		crossX = -0.456
		crossY = 0.890
	)
	// A broad field sets the banks. A smaller octave, pulled a little along
	// the wind, breaks each bank so the field is not one soft gradient and
	// not a set of ruled stripes.
	along := wx*windX + wy*windY
	across := wx*crossX + wy*crossY
	broad := valueNoise(wx, wy, 86, 21)
	mid := valueNoise(wx+along*0.35, wy+across*0.2, 30, 23)
	field := clamp01(broad*0.74 + mid*0.26)
	t := clamp01(0.06 + 0.90*field)
	c := mixRGBA(snowShade, snowSun, smooth01(t))
	c = mixRGBA(c, snowCrest, smooth01((broad-0.60)/0.28)*0.38)
	c = mixRGBA(c, snowHollow, smooth01(0.42-broad)*0.55)

	grain := valueNoise(along, across*1.35, 2.6, 27)
	return shiftDrift(c, (grain-0.5)*0.14)
}

func shiftDrift(c color.RGBA, d float64) color.RGBA {
	r := float64(c.R) + d*46
	g := float64(c.G) + d*40
	b := float64(c.B) + d*28
	if g > b {
		g = b
	}
	if r > g {
		r = g
	}
	return color.RGBA{
		R: clampByte(int(r + 0.5)),
		G: clampByte(int(g + 0.5)),
		B: clampByte(int(b + 0.5)),
		A: 255,
	}
}

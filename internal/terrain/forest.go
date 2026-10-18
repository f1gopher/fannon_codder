package terrain

import (
	"image/color"
	"math"
)

// Moss colours follow the forest ground plate's body. That plate averages
// about 67, 89, 48. The jungle meadow is a brighter yellow-green, so this
// field stays darker and does not reuse those colours.
var (
	mossShade = color.RGBA{R: 36, G: 52, B: 26, A: 255}
	mossSun   = color.RGBA{R: 96, G: 124, B: 66, A: 255}
	mossDark  = color.RGBA{R: 22, G: 34, B: 16, A: 255}
)

// forestPix is one forest paint, one colour per world pixel. A per-source-pixel
// noise does not fit the one-second paint budget, and the draw's linear
// filter covers the step.
var (
	forestPix []color.RGBA
	forestW   int
)

// prepareForest builds the forest field for a map of w by h world pixels.
// Other kinds clear it. The next paint replaces it.
func prepareForest(w, h int, kind Kind) {
	forestPix = nil
	forestW = 0
	if kind != Forest || w < 1 || h < 1 {
		return
	}
	pix := make([]color.RGBA, w*h)
	for y := 0; y < h; y++ {
		wy := float64(y) + 0.5
		row := y * w
		for x := 0; x < w; x++ {
			pix[row+x] = forestColor(float64(x)+0.5, wy)
		}
	}
	forestPix = pix
	forestW = w
}

// forestSample is the forest colour at a world position. During a forest
// paint it reads the prepared field.
func forestSample(wx, wy float64) color.RGBA {
	if forestPix == nil || forestW < 1 {
		return forestColor(wx, wy)
	}
	x := int(math.Floor(wx))
	y := int(math.Floor(wy))
	h := len(forestPix) / forestW
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= forestW {
		x = forestW - 1
	}
	if y >= h {
		y = h - 1
	}
	return forestPix[y*forestW+x]
}

// forestColor is forest open ground at one world pixel. The colour depends
// only on that position, so the field does not repeat a plate. Coordinates
// snap to the world-pixel centre. Shade pools carry the colour. A short
// south-east grain keeps the surface from reading as a flat fill.
func forestColor(wx, wy float64) color.RGBA {
	wx = math.Floor(wx) + 0.5
	wy = math.Floor(wy) + 0.5
	// Light falls from the north-west, the same way canopy shadow falls.
	// (0.82, 0.42) normalised. The cross axis is the short side of the grain.
	const (
		windX  = 0.890
		windY  = 0.456
		crossX = -0.456
		crossY = 0.890
	)
	field := valueNoise(wx, wy, 96, 31)
	drift := valueNoise(wx*0.62, wy*1.05, 34, 33)
	t := clamp01(0.10 + 0.78*field + 0.10*(drift-0.5))
	c := mixRGBA(mossShade, mossSun, smooth01(t))
	// The low field goes darker again, the way shade sits under a canopy.
	c = mixRGBA(c, mossDark, smooth01(0.46-field)*0.55)

	bend := (drift - 0.5) * 8
	along := wx*windX + wy*windY + bend
	across := wx*crossX + wy*crossY
	grain := valueNoise(along, across*2.1, 2.2, 37)
	clump := valueNoise(along*0.7, across*1.3, 6.5, 39)
	return shiftMoss(c, (grain-0.5)*0.26+(clump-0.5)*0.16)
}

func shiftMoss(c color.RGBA, d float64) color.RGBA {
	r := float64(c.R) + d*34
	g := float64(c.G) + d*52
	b := float64(c.B) + d*14
	if r > g {
		r = g
	}
	if b > r {
		b = r
	}
	return color.RGBA{
		R: clampByte(int(r + 0.5)),
		G: clampByte(int(g + 0.5)),
		B: clampByte(int(b + 0.5)),
		A: 255,
	}
}

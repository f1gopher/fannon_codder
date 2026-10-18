package terrain

import (
	"image/color"
	"math"
)

// Dune colours follow the desert ground plate's body. That plate averages
// about 214, 190, 138. SandLo and SandHi are a hotter orange than the
// plate, and the beach plate is hotter still, so the field uses neither.
var (
	sandShade  = color.RGBA{R: 176, G: 148, B: 98, A: 255}
	sandSun    = color.RGBA{R: 244, G: 222, B: 170, A: 255}
	sandHollow = color.RGBA{R: 142, G: 114, B: 70, A: 255}
	sandCrest  = color.RGBA{R: 252, G: 238, B: 204, A: 255}
)

// sandPix is one desert paint, one colour per world pixel. A per-source-pixel
// noise does not fit the one-second paint budget, and the draw's linear
// filter covers the step.
var (
	sandPix []color.RGBA
	sandW   int
)

// prepareSand builds the desert field for a map of w by h world pixels.
// Other kinds clear it. The next paint replaces it. Beach sand and
// quicksand stay on the shared sand plate.
func prepareSand(w, h int, kind Kind) {
	sandPix = nil
	sandW = 0
	if kind != Desert || w < 1 || h < 1 {
		return
	}
	pix := make([]color.RGBA, w*h)
	for y := 0; y < h; y++ {
		wy := float64(y) + 0.5
		row := y * w
		for x := 0; x < w; x++ {
			pix[row+x] = sandColor(float64(x)+0.5, wy)
		}
	}
	sandPix = pix
	sandW = w
}

// sandSample is the desert colour at a world position. During a desert
// paint it reads the prepared field.
func sandSample(wx, wy float64) color.RGBA {
	if sandPix == nil || sandW < 1 {
		return sandColor(wx, wy)
	}
	x := int(math.Floor(wx))
	y := int(math.Floor(wy))
	h := len(sandPix) / sandW
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= sandW {
		x = sandW - 1
	}
	if y >= h {
		y = h - 1
	}
	return sandPix[y*sandW+x]
}

// sandColor is desert open ground at one world pixel. The colour depends
// only on that position, so the field does not repeat a plate. Coordinates
// snap to the world-pixel centre. Broad banks carry the colour. A short
// south-east grain keeps the surface from reading as a flat fill.
func sandColor(wx, wy float64) color.RGBA {
	wx = math.Floor(wx) + 0.5
	wy = math.Floor(wy) + 0.5
	// Light falls from the north-west, the same way canopy shadow falls.
	// (0.82, 0.42) normalised. The cross axis is the short side of a ripple.
	const (
		windX  = 0.890
		windY  = 0.456
		crossX = -0.456
		crossY = 0.890
	)
	// A wide ripple runs across the wind. The lee side is the darker half,
	// so the mark is a dune face and not a thin line. A shorter ripple, a
	// bend, and a slow change in strength keep the faces from repeating
	// as one ruled grating.
	along := wx*windX + wy*windY
	across := wx*crossX + wy*crossY
	bend := (valueNoise(wx, wy, 84, 41) - 0.5) * 46
	amp := 0.42 + 0.58*valueNoise(wx, wy, 130, 59)
	wave := math.Sin((across + bend) * (2 * math.Pi / 88))
	wave2 := math.Sin((across*0.55 + along*0.35 + bend*0.3) * (2 * math.Pi / 37))
	w := (wave*0.86 + wave2*0.14) * amp
	lee := clamp01(0.48 - w*0.62)
	c := mixRGBA(sandSun, sandShade, smooth01(lee))
	c = mixRGBA(c, sandHollow, smooth01(lee-0.74)*0.40)
	c = mixRGBA(c, sandCrest, smooth01(0.15+w)*0.14)
	grain := valueNoise(along, across*1.15, 2.4, 47)
	return shiftRipple(c, (grain-0.5)*0.22)
}

func shiftRipple(c color.RGBA, d float64) color.RGBA {
	r := float64(c.R) + d*36
	g := float64(c.G) + d*32
	b := float64(c.B) + d*22
	if g > r {
		g = r
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

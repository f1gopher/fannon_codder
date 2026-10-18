package terrain

import (
	"image/color"
	"math"

	"fannon-codder/internal/sim"
)

// Water is a colour along the distance to land, not a sample of the shallow
// or deep plates. The line itself moves with a smooth world-locked wave so a
// straight run of water cells does not paint a ruled edge. Each kind keeps
// its own pale and deep colours. Jungle and forest meet a warm beach. Desert
// meets its own sand. Snow meets a pale frozen lip, not that beach.
const (
	// shoreWave is the most the line may leave the cells, in world pixels.
	// The edge pass paints two cells inland, which covers this wave.
	shoreWave = 18.0
	// waterReach is how far offshore the pale colour has given way to deep.
	waterReach = 42.0
)

var (
	shoreToWater []float32
	shoreToLand  []float32
	shoreTW      int
	shoreTH      int
	shoreFlowG   *noiseGrid
	shoreDriftG  *noiseGrid
)

func prepareShoreTiles(surf []surface, tw, th int) {
	w := tw * sim.TileSize
	h := th * sim.TileSize
	n := w * h
	toWater := make([]float32, n)
	toLand := make([]float32, n)
	const inf float32 = 1e6
	any := false
	span := sim.TileSize
	for y := 0; y < h; y++ {
		ty := y / span
		row := y * w
		for x := 0; x < w; x++ {
			s := surf[ty*tw+x/span]
			i := row + x
			if isWetSurf(s) {
				any = true
				toWater[i] = 0
				toLand[i] = inf
			} else {
				toWater[i] = inf
				toLand[i] = 0
			}
		}
	}
	if !any {
		bindShore(nil, nil, 0, 0)
		return
	}
	chamfer(toWater, w, h)
	chamfer(toLand, w, h)
	bindShore(toWater, toLand, w, h)
}

// smoothShore turns the cell distance into one signed field and blurs it,
// so a staircase of tiles becomes a slope instead of a row of square notches.
func smoothShore(toWater, toLand []float32, w, h int) []float32 {
	sd := make([]float32, w*h)
	for i := range sd {
		dw := toWater[i]
		dl := toLand[i]
		if dw > 1e5 {
			dw = 1000
		}
		if dl > 1e5 {
			dl = 1000
		}
		if dw < dl {
			sd[i] = -dl
		} else {
			sd[i] = dw
		}
	}
	// Two wide box passes turn a one-tile step into a slope. A straight
	// coast is a ramp, and a ramp survives the blur, so the wave still
	// draws that line.
	return boxBlur(boxBlur(sd, w, h, 22), w, h, 22)
}

func boxBlur(src []float32, w, h, rad int) []float32 {
	tmp := make([]float32, len(src))
	dst := make([]float32, len(src))
	pref := make([]float32, max(w, h)+1)
	for y := 0; y < h; y++ {
		row := y * w
		pref[0] = 0
		for x := 0; x < w; x++ {
			pref[x+1] = pref[x] + src[row+x]
		}
		for x := 0; x < w; x++ {
			x0 := x - rad
			x1 := x + rad
			if x0 < 0 {
				x0 = 0
			}
			if x1 >= w {
				x1 = w - 1
			}
			tmp[row+x] = (pref[x1+1] - pref[x0]) / float32(x1-x0+1)
		}
	}
	for x := 0; x < w; x++ {
		pref[0] = 0
		for y := 0; y < h; y++ {
			pref[y+1] = pref[y] + tmp[y*w+x]
		}
		for y := 0; y < h; y++ {
			y0 := y - rad
			y1 := y + rad
			if y0 < 0 {
				y0 = 0
			}
			if y1 >= h {
				y1 = h - 1
			}
			dst[y*w+x] = (pref[y1+1] - pref[y0]) / float32(y1-y0+1)
		}
	}
	return dst
}

func isWetSurf(s surface) bool {
	return s == surfShallow || s == surfDeep || s == surfBridge
}

func bindShore(toWater, toLand []float32, w, h int) {
	shoreTW, shoreTH = w, h
	if toWater == nil || w < 2 || h < 2 {
		shoreToWater = nil
		shoreToLand = nil
		shoreFlowG = nil
		shoreDriftG = nil
		return
	}
	shoreToWater = smoothShore(toWater, toLand, w, h)
	shoreToLand = nil
	// Coarse grids, one sample per eight world pixels. The paint then only
	// lerps, which keeps a full phase inside the one-second budget.
	shoreFlowG = buildShoreGrid(w, h, 16, 7, 3.5, 96, 40, 18, 71, 83, 89)
	shoreDriftG = buildShoreGrid(w, h, 12, 5, 0, 52, 24, 1, 91, 97, 1)
}

func buildShoreGrid(worldW, worldH int, ampA, ampB, ampC, scaleA, scaleB, scaleC float64, seedA, seedB, seedC int) *noiseGrid {
	const step = 8.0
	w := int(float64(worldW)/step) + 3
	h := int(float64(worldH)/step) + 3
	v := make([]float32, w*h)
	for y := 0; y < h; y++ {
		wy := (float64(y) - 1) * step
		for x := 0; x < w; x++ {
			wx := (float64(x) - 1) * step
			a := valueNoise(wx, wy, scaleA, seedA)
			b := valueNoise(wx, wy, scaleB, seedB)
			c := 0.5
			if ampC != 0 {
				c = valueNoise(wx, wy, scaleC, seedC)
			}
			v[y*w+x] = float32((a-0.5)*ampA + (b-0.5)*ampB + (c-0.5)*ampC)
		}
	}
	return &noiseGrid{w: w, h: h, step: step, v: v}
}

// shoreField is the chamfer distance to the cells. Positive is land,
// negative is water. shoreSigned adds the wave.
func shoreField(wx, wy float64) (float64, bool) {
	if shoreToWater == nil || shoreTW < 2 || shoreTH < 2 {
		return 0, false
	}
	return bilerpField(shoreToWater, shoreTW, shoreTH, wx, wy), true
}

func shoreSigned(wx, wy float64) (float64, bool) {
	sd, ok := shoreField(wx, wy)
	if !ok {
		return 0, false
	}
	return sd + shoreFlow(wx, wy), true
}

func shoreFlow(wx, wy float64) float64 {
	if shoreFlowG == nil {
		return 0
	}
	f := sampleNoise(shoreFlowG, wx, wy)
	if f > shoreWave {
		return shoreWave
	}
	if f < -shoreWave {
		return -shoreWave
	}
	return f
}

func shoreDrift(wx, wy float64) float64 {
	if shoreDriftG == nil {
		return 0
	}
	return sampleNoise(shoreDriftG, wx, wy)
}

func bilerpField(f []float32, w, h int, wx, wy float64) float64 {
	x := wx - 0.5
	y := wy - 0.5
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > float64(w-1) {
		x = float64(w - 1)
	}
	if y > float64(h-1) {
		y = float64(h - 1)
	}
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1
	if x1 >= w {
		x1 = w - 1
	}
	if y1 >= h {
		y1 = h - 1
	}
	fx := x - float64(x0)
	fy := y - float64(y0)
	sx := fx * fx * (3 - 2*fx)
	sy := fy * fy * (3 - 2*fy)
	v00 := float64(f[y0*w+x0])
	v10 := float64(f[y0*w+x1])
	v01 := float64(f[y1*w+x0])
	v11 := float64(f[y1*w+x1])
	a := v00 + (v10-v00)*sx
	b := v01 + (v11-v01)*sx
	return a + (b-a)*sy
}

func shoreRGBA(lk look, wx, wy float64) color.RGBA {
	switch lk.kind {
	case Snow:
		return color.RGBA{R: 214, G: 224, B: 230, A: 255}
	case Desert:
		return sandSample(wx, wy)
	default:
		return beachSandColor(wx, wy)
	}
}

// beachSandColor is the warm strip where jungle or forest meets water.
// It is not plates/sand.png, so the beach does not repeat that picture.
func beachSandColor(wx, wy float64) color.RGBA {
	n := 0.0
	if shoreDriftG != nil {
		n = shoreDrift(wx, wy) / 10
	}
	return color.RGBA{
		R: clampByte(int(204 + n*16)),
		G: clampByte(int(168 + n*12)),
		B: clampByte(int(104 + n*8)),
		A: 255,
	}
}

func waterEnds(k Kind) (pale, deep color.RGBA) {
	switch k {
	case Forest:
		return color.RGBA{R: 78, G: 138, B: 104, A: 255}, color.RGBA{R: 10, G: 46, B: 38, A: 255}
	case Desert:
		return color.RGBA{R: 176, G: 198, B: 164, A: 255}, color.RGBA{R: 32, G: 104, B: 128, A: 255}
	case Snow:
		return color.RGBA{R: 164, G: 196, B: 230, A: 255}, color.RGBA{R: 14, G: 36, B: 74, A: 255}
	default:
		return color.RGBA{R: 118, G: 196, B: 206, A: 255}, color.RGBA{R: 12, G: 82, B: 110, A: 255}
	}
}

// waterColor fades the shore material into pale water, then into the deep.
// into is the distance offshore at this pixel. fade is that distance pulled
// toward a second shore, so a narrow channel stays pale.
func waterColor(lk look, wx, wy, into, fade float64) color.RGBA {
	sand := shoreRGBA(lk, wx, wy)
	pale, deep := waterEnds(lk.kind)
	wet := smooth01((into + 0.8) / 11)
	col := mixRGBA(sand, pale, wet)
	d := fade + shoreDrift(wx, wy)
	if d < 0 {
		d = 0
	}
	return mixRGBA(col, deep, smooth01(d/waterReach))
}

// waterMask keeps the sand lip still. The shimmer starts once the colour
// has left the sand and become water.
func waterMask(into float64) uint8 {
	if into > 4.5 {
		return maskWater
	}
	return maskNone
}

// wetSand is the beach where the wave has crossed onto a water cell.
// The meadow stays on land, so this fringe is sand rather than grass.
func wetSand(lk look, wx, wy, sd float64) color.RGBA {
	c := shoreRGBA(lk, wx, wy)
	if sd < 2.6 {
		wet := waterColor(lk, wx, wy, -sd, -sd)
		c = mixRGBA(c, wet, smooth01((2.6-sd)/2.6)*0.92)
	}
	return c
}

// featherLand blends ground into the shore, then a little water, so neither
// side of the line is a hard cut.
func featherLand(lk look, ground color.RGBA, wx, wy, sd, reach float64) color.RGBA {
	sand := shoreRGBA(lk, wx, wy)
	const feather = 3.2
	t := 1.0
	if sd > reach-feather {
		t = (reach - sd) / feather
		if t < 0 {
			t = 0
		}
	}
	c := mixRGBA(ground, sand, smooth01(t))
	if sd < 2.6 {
		wet := waterColor(lk, wx, wy, -sd, -sd)
		c = mixRGBA(c, wet, smooth01((2.6-sd)/2.6)*0.92)
	}
	return c
}

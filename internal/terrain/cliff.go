package terrain

import (
	"image/color"
	"math"

	"fannon-codder/internal/sim"
)

// A snow cliff is the region, shifted by a smooth world-locked wave, the
// same idea as the shore. The face is the cliff palette: a thin snow lip,
// cooler rock toward the foot, and a short shadow on the ground to the
// south. It does not sample the rock plate. Other terrains still do.
//
// cliffWave is the most the line may leave the region, in world pixels.
// The shore moves more. This is the smaller wander.
const cliffWave = 8

// cliffFoot is the contact shadow just south of the face.
var cliffFoot = color.RGBA{R: 28, G: 34, B: 42, A: 255}

// cliffPaint is the baked snow-cliff field for one map. cover is 1 on the
// face and 0 off it, with a short feather on the line. hot marks the band
// the paint has to consult. reveal is the surface under a cliff pixel the
// wave has uncovered.
type cliffPaint struct {
	w, h   int
	cover  []float32
	shade  []float32
	hot    []byte
	col    []color.RGBA
	reveal []surface
	ramp   []float32
	open   surface
}

var snowCliff *cliffPaint

func prepareSnowCliff(grid []surface, w, h int, kind Kind, open surface, regions []ChartRegion) {
	snowCliff = nil
	if kind != Snow || w < 2 || h < 2 {
		return
	}
	any := false
	for _, s := range grid {
		if s == surfCliff {
			any = true
			break
		}
	}
	if !any {
		return
	}
	snowCliff = bakeSnowCliff(grid, w, h, open, regions)
}

// prepareSnowCliffTiles expands a tile map to world pixels and bakes the
// same field. A snow map with no cliff clears the field.
func prepareSnowCliffTiles(surf []surface, tw, th int, kind Kind, open surface) {
	snowCliff = nil
	if kind != Snow || tw < 1 || th < 1 {
		return
	}
	any := false
	for _, s := range surf {
		if s == surfCliff {
			any = true
			break
		}
	}
	if !any {
		return
	}
	w := tw * sim.TileSize
	h := th * sim.TileSize
	grid := make([]surface, w*h)
	span := sim.TileSize
	for y := 0; y < h; y++ {
		ty := y / span
		row := y * w
		srow := ty * tw
		for x := 0; x < w; x++ {
			grid[row+x] = surf[srow+x/span]
		}
	}
	snowCliff = bakeSnowCliff(grid, w, h, open, nil)
}

func bakeSnowCliff(grid []surface, w, h int, open surface, regions []ChartRegion) *cliffPaint {
	const inf float32 = 1e6
	cliffs, ramps := cliffRings(regions)
	useRings := len(cliffs) > 0
	ramp := make([]float32, w*h)
	rampAny := false
	for i, s := range grid {
		if s == surfRamp {
			ramp[i] = 0
			rampAny = true
		} else {
			ramp[i] = inf
		}
	}
	if rampAny && !useRings {
		chamfer(ramp, w, h)
	}
	// The shift is smooth and mostly vertical, so a long face bends in one
	// piece. A binary mask of that shift would stair-step. The distance is
	// sampled at the shifted point instead, which leaves a continuous line.
	warpY := buildShoreGrid(w, h, 13, 6, 0, 150, 62, 1, 41, 67, 1)
	warpX := buildShoreGrid(w, h, 5, 2, 0, 120, 56, 1, 23, 59, 1)

	orig := make([]byte, w*h)
	reveal := make([]surface, w*h)
	toIn := make([]float32, w*h)
	toOut := make([]float32, w*h)
	for i, s := range grid {
		reveal[i] = s
		if s == surfCliff {
			orig[i] = 1
			toIn[i] = 0
			toOut[i] = inf
		} else {
			toIn[i] = inf
			toOut[i] = 0
		}
	}
	var base []float32
	if !useRings {
		chamfer(toIn, w, h)
		chamfer(toOut, w, h)
		base = make([]float32, w*h)
		for i := range base {
			base[i] = toOut[i] - toIn[i]
		}
	}
	spans := cliffSpans(orig, w, h)

	shade := make([]float32, w*h)
	col := make([]color.RGBA, w*h)
	cover := make([]float32, w*h)
	var rampCover []float32
	if len(ramps) > 0 {
		rampCover = make([]float32, w*h)
	}
	hot := make([]byte, w*h)
	for y := 0; y < h; y++ {
		wy := float64(y) + 0.5
		row := y * w
		for x := 0; x < w; x++ {
			i := row + x
			wx := float64(x) + 0.5
			dx := clampAbs(sampleNoise(warpX, wx, wy), cliffWave)
			dy := clampAbs(sampleNoise(warpY, wx, wy), cliffWave)
			damp := 1.0
			if len(ramps) > 0 {
				rd := rampField(wx, wy, ramps)
				if rd < 0 {
					damp = 0
				} else if rd < 18 {
					damp = rd / 18
				}
				// The cut is the ramp polygon, not the raster stair of it.
				rampCover[i] = float32(smooth01((0.9 - rd) / 1.8))
			} else if rampAny && ramp[i] < inf {
				d := float64(ramp[i])
				if d < 18 {
					damp = d / 18
				}
			}
			dx *= damp
			dy *= damp
			sx := wx - dx
			sy := wy - dy
			var dist float64
			if useRings {
				dist = cliffField(sx, sy, cliffs, ramps)
			} else {
				dist = bilerpField(base, w, h, sx, sy)
			}
			cvr := smooth01((dist + 1.2) / 2.4)
			cover[i] = float32(cvr)
			face := 0.55
			if y0, y1, ok := cliffSpanAt(spans, w, sx, sy); ok {
				face = (sy - y0) / (y1 - y0 + 1)
				// The foot's shadow follows the shifted line, so it bends
				// with the face instead of sitting on a pixel step.
				below := sy - (y1 + 0.5)
				if below > 0 && below < 7 && dist < 1 {
					t := 1 - below/7
					shade[i] = float32(t * t * 0.55)
				}
			}
			got := cliffLookup(grid, w, h, sx, sy)
			if got != surfCliff && !cliffHold(got) {
				reveal[i] = got
			} else if orig[i] == 1 {
				reveal[i] = open
			}
			if cvr <= 0 && shade[i] == 0 {
				continue
			}
			col[i] = snowCliffRGB(wx, wy, face)
		}
	}
	for y := 0; y < h; y++ {
		row := y * w
		for x := 0; x < w; x++ {
			i := row + x
			// A cliff pixel the wave slid off has no cover and no shadow, and
			// it still has to be repainted with whatever was underneath.
			rC := float32(0)
			if rampCover != nil {
				rC = rampCover[i]
			}
			if cover[i] <= 0.004 && shade[i] <= 0.004 && orig[i] == 0 && (rC <= 0.004 || rC >= 0.996) {
				continue
			}
			for oy := -1; oy <= 1; oy++ {
				yy := y + oy
				if yy < 0 || yy >= h {
					continue
				}
				for ox := -1; ox <= 1; ox++ {
					xx := x + ox
					if xx < 0 || xx >= w {
						continue
					}
					hot[yy*w+xx] = 1
				}
			}
		}
	}
	return &cliffPaint{
		w: w, h: h,
		cover: cover, shade: shade, hot: hot,
		col: col, reveal: reveal, ramp: rampCover, open: open,
	}
}

func cliffRings(regions []ChartRegion) (cliffs, ramps [][][2]float64) {
	for _, r := range regions {
		if len(r.Points) < 3 {
			continue
		}
		switch r.Tile {
		case sim.TileCliff:
			cliffs = append(cliffs, r.Points)
		case sim.TileRamp:
			ramps = append(ramps, r.Points)
		}
	}
	return cliffs, ramps
}

// ringSDF is the distance to a closed ring. Negative is inside.
func ringSDF(x, y float64, p [][2]float64) float64 {
	best := 1e9
	for i := range p {
		j := i + 1
		if j == len(p) {
			j = 0
		}
		d := distPointSeg(x, y, p[i][0], p[i][1], p[j][0], p[j][1])
		if d < best {
			best = d
		}
	}
	if pointInRing(x, y, p) {
		return -best
	}
	return best
}

func distPointSeg(px, py, ax, ay, bx, by float64) float64 {
	dx := bx - ax
	dy := by - ay
	l2 := dx*dx + dy*dy
	if l2 == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := ((px-ax)*dx + (py-ay)*dy) / l2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}

// cliffField is positive on the rock and negative off it. A ramp ring is a
// hole, so the cut follows that ring instead of the cell stair.
func cliffField(x, y float64, cliffs, ramps [][][2]float64) float64 {
	c := 1e9
	for _, ring := range cliffs {
		d := ringSDF(x, y, ring)
		if d < c {
			c = d
		}
	}
	if c > 1e8 {
		return -1000
	}
	r := 1e9
	for _, ring := range ramps {
		d := ringSDF(x, y, ring)
		if d < r {
			r = d
		}
	}
	if c < 0 && r < 0 {
		return r
	}
	if c < 0 {
		depth := -c
		if r < 1e8 && r < depth {
			return r
		}
		return depth
	}
	return -c
}

func rampField(x, y float64, ramps [][][2]float64) float64 {
	r := 1e9
	for _, ring := range ramps {
		d := ringSDF(x, y, ring)
		if d < r {
			r = d
		}
	}
	return r
}

// cliffSpans lists the vertical runs of cliff in each column. A ramp cuts
// a run, so the gap stays a gap.
func cliffSpans(orig []byte, w, h int) [][][2]int {
	spans := make([][][2]int, w)
	for x := 0; x < w; x++ {
		start := -1
		for y := 0; y < h; y++ {
			on := orig[y*w+x] == 1
			if on && start < 0 {
				start = y
			}
			if !on && start >= 0 {
				spans[x] = append(spans[x], [2]int{start, y - 1})
				start = -1
			}
		}
		if start >= 0 {
			spans[x] = append(spans[x], [2]int{start, h - 1})
		}
	}
	return spans
}

// cliffSpanAt is the run nearest to the shifted point. y0 and y1 are the
// inclusive pixel rows of that run.
func cliffSpanAt(spans [][][2]int, w int, x, y float64) (float64, float64, bool) {
	xi := int(math.Floor(x))
	if xi < 0 {
		xi = 0
	}
	if xi >= w {
		xi = w - 1
	}
	best := 1e9
	y0, y1 := 0.0, 0.0
	ok := false
	for _, sp := range spans[xi] {
		a, b := float64(sp[0]), float64(sp[1])
		d := 0.0
		if y < a {
			d = a - y
		} else if y > b {
			d = y - b
		}
		if !ok || d < best {
			best = d
			y0, y1 = a, b
			ok = true
		}
	}
	return y0, y1, ok
}

// cliffHold is a surface the wave must not paint over. The ramp is the
// gap in the face, so it stays a straight cut.
func cliffHold(s surface) bool {
	switch s {
	case surfRamp, surfBridge, surfTree, surfMine, surfShallow, surfDeep, surfSand:
		return true
	default:
		return false
	}
}

func cliffLookup(grid []surface, w, h int, wx, wy float64) surface {
	x := int(math.Floor(wx))
	y := int(math.Floor(wy))
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= w {
		x = w - 1
	}
	if y >= h {
		y = h - 1
	}
	return grid[y*w+x]
}

func clampAbs(v, lim float64) float64 {
	if v > lim {
		return lim
	}
	if v < -lim {
		return -lim
	}
	return v
}

// snowCliffRGB is one pixel of the face. face is 0 at the north lip and 1
// at the south foot. The lip takes a little snow. The foot goes dark. A
// broad column breaks the gradient so the wall is not one airbrushed ramp.
func snowCliffRGB(wx, wy, face float64) color.RGBA {
	face = clamp01(face)
	pal := snowLook.pal
	c := mixRGBA(pal.CliffHi, pal.CliffLo, smooth01(face))
	cap := clamp01((0.14 - face) / 0.14)
	if cap > 0 {
		c = mixRGBA(c, snowSample(wx, wy), cap*0.75)
	}
	foot := clamp01((face - 0.8) / 0.2)
	if foot > 0 {
		c = mixRGBA(c, cliffFoot, foot*0.7)
	}
	col := valueNoise(math.Floor(wx)+0.5, 0, 40, 19)
	return mixRGBA(c, pal.CliffLo, (col-0.5)*0.2)
}

func (c *cliffPaint) hotAt(wx, wy float64) bool {
	x := int(math.Floor(wx))
	y := int(math.Floor(wy))
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return false
	}
	return c.hot[y*c.w+x] != 0
}

func (c *cliffPaint) atColor(wx, wy float64) color.RGBA {
	x := int(math.Floor(wx))
	y := int(math.Floor(wy))
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= c.w {
		x = c.w - 1
	}
	if y >= c.h {
		y = c.h - 1
	}
	return c.col[y*c.w+x]
}

func (c *cliffPaint) revealAt(wx, wy float64) surface {
	x := int(math.Floor(wx))
	y := int(math.Floor(wy))
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= c.w {
		x = c.w - 1
	}
	if y >= c.h {
		y = c.h - 1
	}
	s := c.reveal[y*c.w+x]
	if s == surfCliff || cliffHold(s) {
		return c.open
	}
	return s
}

// paintSnowCliffPixel feathers the waved face over the colour already
// chosen for this pixel. have is false when c has not been painted yet
// (the region painter). The tile painter passes the colour it already has.
// cliffBlocks is a surface the wave leaves alone. A chart ramp is not in
// that set: its cut is the ring, feathered in paintSnowCliffPixel, so the
// join is not the raster stair.
func cliffBlocks(s surface) bool {
	if s == surfRamp && snowCliff != nil && snowCliff.ramp != nil {
		return false
	}
	return cliffHold(s)
}

func paintSnowCliffPixel(lk look, s surface, wx, wy float64, have bool, c color.RGBA, k uint8) (color.RGBA, uint8) {
	if snowCliff == nil || cliffBlocks(s) || !snowCliff.hotAt(wx, wy) {
		if have {
			return c, k
		}
		return chartPixelPaint(lk, s, wx, wy)
	}
	cover := bilerpField(snowCliff.cover, snowCliff.w, snowCliff.h, wx, wy)
	if cover >= 0.98 && !rampCuts(wx, wy) {
		return snowCliff.atColor(wx, wy), maskNone
	}
	if !have || s == surfCliff || s == surfRamp {
		base := s
		if s == surfCliff || s == surfRamp {
			base = snowCliff.revealAt(wx, wy)
		}
		if have {
			c, k = lk.plateColor(base, wx, wy)
		} else {
			c, k = chartPixelPaint(lk, base, wx, wy)
		}
	}
	if cover > 0 {
		c = mixRGBA(c, snowCliff.atColor(wx, wy), cover)
		if cover > 0.4 {
			k = maskNone
		}
	}
	sh := bilerpField(snowCliff.shade, snowCliff.w, snowCliff.h, wx, wy)
	if sh > 0 && cover < 0.55 {
		c = mixRGBA(c, cliffFoot, sh*(1-cover))
	}
	return mixRamp(lk, wx, wy, have, c, k)
}

func rampCuts(wx, wy float64) bool {
	if snowCliff == nil || snowCliff.ramp == nil {
		return false
	}
	return bilerpField(snowCliff.ramp, snowCliff.w, snowCliff.h, wx, wy) > 0.02
}

func mixRamp(lk look, wx, wy float64, have bool, c color.RGBA, k uint8) (color.RGBA, uint8) {
	if snowCliff == nil || snowCliff.ramp == nil {
		return c, k
	}
	rc := bilerpField(snowCliff.ramp, snowCliff.w, snowCliff.h, wx, wy)
	if rc <= 0.02 {
		return c, k
	}
	var dirt color.RGBA
	var dk uint8
	if have {
		dirt, dk = lk.plateColor(surfRamp, wx, wy)
	} else {
		dirt, dk = chartPixelPaint(lk, surfRamp, wx, wy)
	}
	if rc >= 0.98 {
		return dirt, dk
	}
	return mixRGBA(c, dirt, rc), dk
}

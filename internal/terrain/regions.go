package terrain

import (
	"image"
	"image/color"
	"math"

	"fannon-codder/internal/sim"
)

// Chart is a freeform map. W and H are world pixels. The ground is that
// terrain's open plate. Each region paints over it, and a later region
// wins where they overlap. Trees are a canopy on the open ground.
// Live is the battle's tile grid. A mine paints only while that cell is
// still a mine. The editor leaves Live empty.
type Chart struct {
	W, H    int
	Terrain string
	Regions []ChartRegion
	Live    sim.Map
}

// ChartRegion is one closed ring in world pixels.
type ChartRegion struct {
	Tile   sim.Tile
	Name   string
	Points [][2]float64
}

var chartNow struct {
	hash uint64
	view *View
}

// PrepareChart caches one painted region map.
func PrepareChart(c Chart) *View {
	if c.W < 1 || c.H < 1 {
		return &View{}
	}
	h := hashChart(c)
	if chartNow.view != nil && chartNow.view.ground != nil && chartNow.hash == h {
		return chartNow.view
	}
	painted := paintChart(c)
	view := &View{
		ground: uploadStraight(painted.ground),
		anim:   uploadStraight(painted.anim),
		mask:   uploadStraight(painted.mask),
	}
	for i := range painted.trees {
		can := painted.trees[i]
		for _, run := range can.runs {
			if run.sprite == nil || run.sprite.Bounds().Empty() {
				continue
			}
			view.canopies = append(view.canopies, Canopy{
				img:    uploadStraight(run.sprite),
				WorldX: run.world,
				NorthY: run.north,
				FootX:  run.footX,
				FootY:  run.foot,
			})
		}
	}
	chartNow.hash = h
	chartNow.view = view
	return view
}

// PrepareChartLive paints c, and drops a mine mark whose cell is no longer a mine.
func PrepareChartLive(c Chart, m sim.Map) *View {
	c.Live = m
	return PrepareChart(c)
}

func hashChart(c Chart) uint64 {
	h := uint64(14695981039346656037)
	h = (h ^ uint64(c.W)) * 1099511628211
	h = (h ^ uint64(c.H)) * 1099511628211
	h = (h ^ uint64(KindOf(c.Terrain))) * 1099511628211
	for _, r := range c.Regions {
		h = (h ^ uint64(r.Tile)) * 1099511628211
		for _, p := range r.Points {
			h = (h ^ math.Float64bits(p[0])) * 1099511628211
			h = (h ^ math.Float64bits(p[1])) * 1099511628211
		}
	}
	if c.liveMines() {
		h = (h ^ uint64(c.Live.W)) * 1099511628211
		h = (h ^ uint64(c.Live.H)) * 1099511628211
		n := c.Live.W * c.Live.H
		for _, t := range c.Live.Tiles[:n] {
			h = (h ^ uint64(t)) * 1099511628211
		}
	}
	return h
}

func (c Chart) liveMines() bool {
	m := c.Live
	if m.W < 1 || m.H < 1 || len(m.Tiles) < m.W*m.H {
		return false
	}
	return m.W*sim.TileSize == c.W && m.H*sim.TileSize == c.H
}

func paintChart(c Chart) terrainPaint {
	lk := lookOf(c.Terrain)
	ww, wh := c.W, c.H
	prepareGrass(ww, wh, lk.kind)
	prepareSnow(ww, wh, lk.kind)
	prepareSand(ww, wh, lk.kind)
	prepareForest(ww, wh, lk.kind)
	grid := rasterChart(c, lk.open)
	clearSpentMines(grid, c, lk.open)
	toWater, toLand := chartWaterDist(grid, ww, wh)
	pw := ww * terrainPx
	ph := wh * terrainPx
	ground := image.NewRGBA(image.Rect(0, 0, pw, ph))
	kind := make([]uint8, pw*ph)
	noise := buildNoise(max(1, ww/sim.TileSize), max(1, wh/sim.TileSize), 11)
	for py := 0; py < ph; py++ {
		wy := (float64(py) + 0.5) / float64(terrainPx)
		iy := clampi(int(wy), 0, wh-1)
		row := py * pw
		gy := iy * ww
		for px := 0; px < pw; px++ {
			wx := (float64(px) + 0.5) / float64(terrainPx)
			ix := clampi(int(wx), 0, ww-1)
			gi := gy + ix
			col, k := chartPixel(lk, grid[gi], wx, wy, toWater, toLand, gi, noise)
			setRGBA(ground, px, py, col)
			kind[row+px] = k
		}
	}
	seeds := chartSeeds(grid, ww, wh)
	if len(seeds) > 0 {
		crowns := make([]crown, len(seeds))
		for i, s := range seeds {
			crowns[i] = crownAt(s)
		}
		shadeCrowns(ground, crowns, lk)
	}
	stampChartMines(ground, grid, ww, wh, lk.pal)
	anim, mask := animLayers(ground, kind)
	trees := paintChartCanopies(grid, ww, wh, seeds, lk, noise, ground)
	return terrainPaint{ground: ground, anim: anim, mask: mask, trees: trees}
}

func chartSurface(t sim.Tile, open surface) surface {
	if t == sim.TileMine {
		return surfMine
	}
	return cellSurface(t, open)
}

func rasterChart(c Chart, open surface) []surface {
	g := make([]surface, c.W*c.H)
	for i := range g {
		g[i] = open
	}
	for _, r := range c.Regions {
		if len(r.Points) < 3 {
			continue
		}
		s := chartSurface(r.Tile, open)
		x0, y0, x1, y1 := ringBounds(r.Points, c.W, c.H)
		for y := y0; y < y1; y++ {
			row := y * c.W
			py := float64(y) + 0.5
			for x := x0; x < x1; x++ {
				if pointInRing(float64(x)+0.5, py, r.Points) {
					g[row+x] = s
				}
			}
		}
	}
	return g
}

// Contains reports whether the world point is inside this ring.
func (r ChartRegion) Contains(x, y float64) bool {
	return len(r.Points) >= 3 && pointInRing(x, y, r.Points)
}

// clearSpentMines turns mine pixels back to open ground when that cell
// is no longer a mine. The mark is stamped from those pixels.
func clearSpentMines(grid []surface, c Chart, open surface) {
	if !c.liveMines() {
		return
	}
	ts := sim.TileSize
	for ty := 0; ty < c.Live.H; ty++ {
		for tx := 0; tx < c.Live.W; tx++ {
			if c.Live.At(tx, ty) == sim.TileMine {
				continue
			}
			x0 := tx * ts
			y0 := ty * ts
			for y := y0; y < y0+ts; y++ {
				row := y * c.W
				for x := x0; x < x0+ts; x++ {
					if grid[row+x] == surfMine {
						grid[row+x] = open
					}
				}
			}
		}
	}
}

func ringBounds(pts [][2]float64, w, h int) (x0, y0, x1, y1 int) {
	minX, minY := pts[0][0], pts[0][1]
	maxX, maxY := minX, minY
	for _, p := range pts[1:] {
		if p[0] < minX {
			minX = p[0]
		}
		if p[1] < minY {
			minY = p[1]
		}
		if p[0] > maxX {
			maxX = p[0]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	x0 = clampi(int(math.Floor(minX)), 0, w)
	y0 = clampi(int(math.Floor(minY)), 0, h)
	x1 = clampi(int(math.Ceil(maxX)), 0, w)
	y1 = clampi(int(math.Ceil(maxY)), 0, h)
	return x0, y0, x1, y1
}

func pointInRing(x, y float64, p [][2]float64) bool {
	in := false
	j := len(p) - 1
	for i := range p {
		yi, yj := p[i][1], p[j][1]
		if (yi > y) != (yj > y) {
			xin := (p[j][0]-p[i][0])*(y-yi)/(yj-yi) + p[i][0]
			if x < xin {
				in = !in
			}
		}
		j = i
	}
	return in
}

func chartPixel(lk look, s surface, wx, wy float64, toWater, toLand []float32, gi int, noise *noiseGrid) (color.RGBA, uint8) {
	if s == surfBridge {
		c := lk.pal.PlankGap
		if int(math.Floor(wy/bridgePlank))%2 == 0 {
			c = lk.pal.Plank
		}
		if lk.snowCaps {
			c = mixRGBA(c, color.RGBA{R: 232, G: 242, B: 246, A: 255}, 0.28)
		}
		return c, maskNone
	}
	base := s
	if s == surfTree || s == surfMine {
		base = lk.open
	}
	c, k := lk.plateColor(base, wx, wy)
	if toWater == nil {
		return c, k
	}
	dw := float64(toWater[gi])
	dl := float64(toLand[gi])
	water := s == surfShallow || s == surfDeep
	if lk.beach {
		if !water && s != surfIce && s != surfCliff && s != surfRamp {
			reach := beachSand + (sampleNoise(noise, wx, wy)-0.5)*beachSandWander
			if reach < 2.2 {
				reach = 2.2
			}
			if dw > 0 && dw < reach {
				return sandPlate.at(wx, wy), maskNone
			}
		}
		if s == surfDeep {
			wob := (sampleNoise(noise, wx*0.7, wy*0.7) - 0.5) * beachFadeWander
			d := dl + wob
			span := beachShallowHold + beachWaterFade
			if d < span {
				t := 0.0
				if d > beachShallowHold {
					t = smooth01((d - beachShallowHold) / beachWaterFade)
				}
				return mixRGBA(lk.shallow.at(wx, wy), c, t), maskWater
			}
		}
		return c, k
	}
	if !water && (base == surfGrass || base == surfSnow) && dw > 0 && dw < float64(sim.TileSize) {
		t := 1 - dw/float64(sim.TileSize)
		if lk.shore != nil && lk.shoreMix > 0 {
			c = mixRGBA(c, lk.shore.at(wx, wy), t*lk.shoreMix)
		}
		c = beadFoam(c, lk.pal, wx, wy, dw)
	}
	return c, k
}

// chartWaterDist is a chamfer distance to water and to land, in world pixels.
// Both are nil when the chart has no water.
func chartWaterDist(grid []surface, w, h int) (toWater, toLand []float32) {
	n := w * h
	toWater = make([]float32, n)
	toLand = make([]float32, n)
	const inf float32 = 1e6
	any := false
	for i, s := range grid {
		if s == surfShallow || s == surfDeep {
			any = true
			toWater[i] = 0
			toLand[i] = inf
		} else {
			toWater[i] = inf
			toLand[i] = 0
		}
	}
	if !any {
		return nil, nil
	}
	chamfer(toWater, w, h)
	chamfer(toLand, w, h)
	return toWater, toLand
}

func chamfer(d []float32, w, h int) {
	const straight float32 = 1
	const diag float32 = 1.41421356
	for y := 0; y < h; y++ {
		row := y * w
		for x := 0; x < w; x++ {
			i := row + x
			if x > 0 {
				d[i] = minf32(d[i], d[i-1]+straight)
			}
			if y == 0 {
				continue
			}
			d[i] = minf32(d[i], d[i-w]+straight)
			if x > 0 {
				d[i] = minf32(d[i], d[i-w-1]+diag)
			}
			if x+1 < w {
				d[i] = minf32(d[i], d[i-w+1]+diag)
			}
		}
	}
	for y := h - 1; y >= 0; y-- {
		row := y * w
		for x := w - 1; x >= 0; x-- {
			i := row + x
			if x+1 < w {
				d[i] = minf32(d[i], d[i+1]+straight)
			}
			if y+1 == h {
				continue
			}
			d[i] = minf32(d[i], d[i+w]+straight)
			if x+1 < w {
				d[i] = minf32(d[i], d[i+w+1]+diag)
			}
			if x > 0 {
				d[i] = minf32(d[i], d[i+w-1]+diag)
			}
		}
	}
}

func minf32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

type treeSeed struct {
	x, y   float64
	ox, oy float64
	edge   bool
}

func chartSeeds(grid []surface, w, h int) []treeSeed {
	var out []treeSeed
	for y := sim.TileSize / 2; y < h; y += sim.TileSize {
		for x := sim.TileSize / 2; x < w; x += sim.TileSize {
			if grid[y*w+x] != surfTree {
				continue
			}
			ox, oy, edge := treeOutward(grid, w, h, float64(x), float64(y))
			out = append(out, treeSeed{x: float64(x), y: float64(y), ox: ox, oy: oy, edge: edge})
		}
	}
	return out
}

func treeOutward(grid []surface, w, h int, x, y float64) (float64, float64, bool) {
	const reach = 14.0
	sx, sy := 0.0, 0.0
	for k := 0; k < 8; k++ {
		ang := float64(k) * math.Pi / 4
		xx := int(math.Floor(x + math.Cos(ang)*reach))
		yy := int(math.Floor(y + math.Sin(ang)*reach))
		if xx < 0 || yy < 0 || xx >= w || yy >= h || grid[yy*w+xx] != surfTree {
			sx += math.Cos(ang)
			sy += math.Sin(ang)
		}
	}
	l := math.Hypot(sx, sy)
	if l < 0.5 {
		return 0, 0, false
	}
	return sx / l, sy / l, true
}

func crownAt(s treeSeed) crown {
	tx := int(math.Floor(s.x / float64(sim.TileSize)))
	ty := int(math.Floor(s.y / float64(sim.TileSize)))
	hsh := hashU(tx, ty, 5)
	jx := (float64(hsh&255)/255 - 0.5) * 4
	jy := (float64((hsh>>8)&255)/255-0.5)*3 - 1.2
	cr := crown{
		tx:    tx,
		ty:    ty,
		cx:    s.x + jx,
		cy:    s.y + jy,
		mainR: 14.2 + float64((hsh>>16)&3)*0.25,
		edge:  s.edge,
	}
	for n := 0; n < len(cr.lobes); n++ {
		hi := hashU(tx*4+n, ty*3+n, 9)
		ang := float64(n)*(math.Pi*2/float64(len(cr.lobes))) + (float64(hi&255)/255-0.5)*1.05
		dist := 8.5 + float64((hi>>8)&255)/255*4.5
		cr.lobes[n] = lobe{
			x: cr.cx + math.Cos(ang)*dist,
			y: cr.cy + math.Sin(ang)*dist,
			r: 7.4 + float64((hi>>16)&255)/255*3.6,
		}
	}
	if !s.edge {
		return cr
	}
	cr.lobes[0] = lobe{
		x: cr.cx + s.ox*9,
		y: cr.cy + s.oy*9,
		r: 11 + float64(hsh&3)*0.35,
	}
	for n := 0; n < len(cr.sprigs); n++ {
		hi := hashU(tx*5+n, ty*3+2, 40+n)
		along := 15 + float64(hi%9)
		side := (float64((hi>>8)%100)/50 - 1) * 9
		cr.sprigs[n] = lobe{
			x: s.x + s.ox*along - s.oy*side,
			y: s.y + s.oy*along + s.ox*side,
			r: 4.4 + float64((hi>>16)%4)*0.55,
		}
	}
	return cr
}

func shadeCrowns(dst *image.RGBA, crowns []crown, lk look) {
	pw := dst.Bounds().Dx()
	ph := dst.Bounds().Dy()
	minX, minY := crowns[0].cx, crowns[0].cy
	maxX, maxY := minX, minY
	for _, cr := range crowns[1:] {
		if cr.cx < minX {
			minX = cr.cx
		}
		if cr.cy < minY {
			minY = cr.cy
		}
		if cr.cx > maxX {
			maxX = cr.cx
		}
		if cr.cy > maxY {
			maxY = cr.cy
		}
	}
	const reach = 28.0
	x0 := int(math.Floor((minX - reach) * terrainPx))
	y0 := int(math.Floor((minY - reach) * terrainPx))
	x1 := int(math.Ceil((maxX + reach) * terrainPx))
	y1 := int(math.Ceil((maxY + reach) * terrainPx))
	x0 = clampi(x0, 0, pw)
	y0 = clampi(y0, 0, ph)
	x1 = clampi(x1, 0, pw)
	y1 = clampi(y1, 0, ph)
	strength := make([]uint8, pw*ph)
	for py := y0; py < y1; py += terrainPx {
		wy := (float64(py) + 0.5) / float64(terrainPx)
		for px := x0; px < x1; px += terrainPx {
			wx := (float64(px) + 0.5) / float64(terrainPx)
			d := shadeDist(crowns, wx-3.2, wy-4.0)
			if d > 16 {
				continue
			}
			t := clamp01((d + 6) / 24)
			s := (1 - smooth01(t)) * 0.16
			if s <= 0 {
				continue
			}
			v := uint8(s * 255)
			xEnd := min(px+terrainPx, x1)
			yEnd := min(py+terrainPx, y1)
			for yy := py; yy < yEnd; yy++ {
				row := yy * pw
				for xx := px; xx < xEnd; xx++ {
					if v > strength[row+xx] {
						strength[row+xx] = v
					}
				}
			}
		}
	}
	for i, v := range strength {
		if v == 0 {
			continue
		}
		o := i * 4
		c := color.RGBA{R: dst.Pix[o], G: dst.Pix[o+1], B: dst.Pix[o+2], A: dst.Pix[o+3]}
		c = mixRGBA(c, lk.shadow, float64(v)/255)
		dst.Pix[o] = c.R
		dst.Pix[o+1] = c.G
		dst.Pix[o+2] = c.B
		dst.Pix[o+3] = c.A
	}
}

func stampChartMines(dst *image.RGBA, grid []surface, w, h int, pal Palette) {
	for y := sim.TileSize / 2; y < h; y += sim.TileSize {
		for x := sim.TileSize / 2; x < w; x += sim.TileSize {
			if grid[y*w+x] != surfMine {
				continue
			}
			fillWorldRect(dst, float64(x)-3, float64(y)-3, 6, 6, pal.Mine)
			fillWorldRect(dst, float64(x)-1, float64(y)-1, 2, 2, pal.MineMark)
		}
	}
}

func paintChartCanopies(grid []surface, w, h int, seeds []treeSeed, lk look, noise *noiseGrid, ground *image.RGBA) []treeCanopy {
	if len(seeds) == 0 {
		return nil
	}
	ids := labelTrees(grid, w, h)
	groups := map[int][]treeSeed{}
	var order []int
	for _, s := range seeds {
		id := ids[int(s.y)*w+int(s.x)]
		if id < 0 {
			continue
		}
		if _, ok := groups[id]; !ok {
			order = append(order, id)
		}
		groups[id] = append(groups[id], s)
	}
	var out []treeCanopy
	for _, id := range order {
		out = append(out, paintSeedCanopy(groups[id], lk, noise, ground))
	}
	return out
}

func labelTrees(grid []surface, w, h int) []int {
	ids := make([]int, w*h)
	for i := range ids {
		ids[i] = -1
	}
	next := 0
	stack := make([]int, 0, 64)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			start := y*w + x
			if grid[start] != surfTree || ids[start] >= 0 {
				continue
			}
			ids[start] = next
			stack = append(stack[:0], start)
			for len(stack) > 0 {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				nx, ny := n%w, n/w
				for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					xx, yy := nx+d[0], ny+d[1]
					if xx < 0 || yy < 0 || xx >= w || yy >= h {
						continue
					}
					j := yy*w + xx
					if grid[j] != surfTree || ids[j] >= 0 {
						continue
					}
					ids[j] = next
					stack = append(stack, j)
				}
			}
			next++
		}
	}
	return ids
}

func paintSeedCanopy(seeds []treeSeed, lk look, noise *noiseGrid, ground *image.RGBA) treeCanopy {
	crowns := make([]crown, len(seeds))
	cells := make([]image.Point, len(seeds))
	minX, minY := seeds[0].x, seeds[0].y
	maxX, maxY := minX, minY
	for i, s := range seeds {
		crowns[i] = crownAt(s)
		cells[i] = image.Point{X: crowns[i].tx, Y: crowns[i].ty}
		if s.x < minX {
			minX = s.x
		}
		if s.y < minY {
			minY = s.y
		}
		if s.x > maxX {
			maxX = s.x
		}
		if s.y > maxY {
			maxY = s.y
		}
	}
	const pad = canopyPad + 8
	originX := minX - pad
	originY := minY - pad
	worldW := maxX - minX + pad*2
	worldH := maxY - minY + pad*2
	pw := int(math.Ceil(worldW * terrainPx))
	ph := int(math.Ceil(worldH * terrainPx))
	img := image.NewRGBA(image.Rect(0, 0, pw, ph))
	owner := make([]int16, pw*ph)
	for i := range owner {
		owner[i] = -1
	}
	cover, tips, ccx, ccy := splatCrowns(crowns, originX, originY, pw, ph)
	gw := ground.Bounds().Dx()
	gh := ground.Bounds().Dy()
	for py := 0; py < ph; py++ {
		wy := originY + (float64(py)+0.5)/float64(terrainPx)
		row := py * pw
		for px := 0; px < pw; px++ {
			d := float64(cover[row+px])
			sprig := float64(tips[row+px])
			if d > 12 && sprig < 0.02 {
				continue
			}
			wx := originX + (float64(px)+0.5)/float64(terrainPx)
			a := canopyAlpha(d, wx, wy, noise)
			if sprig > a {
				a = sprig
			}
			if a < 0.012 {
				continue
			}
			c := lk.canopy.atScale(wx, wy, 4)
			c = shadeCrown(c, d, wx-float64(ccx[row+px]), wy-float64(ccy[row+px]), lk, sampleNoise(noise, wx*1.7, wy*1.7))
			c = mergeFringe(c, ground, gw, gh, wx, wy, a, lk)
			c.A = uint8(a * 255)
			if c.A == 0 {
				continue
			}
			setRGBA(img, px, py, c)
			owner[row+px] = int16(nearestTree(cells, wx, wy))
		}
	}
	can := treeCanopy{img: img, originX: originX, originY: originY}
	for i, cell := range cells {
		can.runs = append(can.runs, sliceRun(can, owner, pw, i, cell.X, cell.Y))
	}
	return can
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

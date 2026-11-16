package terrain

import (
	"image"
	"image/color"
	"math"
	"sort"
	"sync"
	"time"

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

// chartPaintMu keeps paintChart off the shared shore and field caches
// while a briefing warm runs beside the game thread.
var chartPaintMu sync.Mutex

// paintCharts counts paintChart calls so a test can see that PrepareChart
// uploaded a warm result instead of painting again.
var paintCharts int

type chartWarmState struct {
	mu    sync.Mutex
	hash  uint64
	paint terrainPaint
	ready bool
	wait  chan struct{}
}

var chartWarm chartWarmState

// WarmChart paints c on a worker. The battle uploads that picture.
// A chart already cached, or already warming, is left alone.
func WarmChart(c Chart) {
	if c.W < 1 || c.H < 1 {
		return
	}
	h := hashChart(c)
	if chartNow.view != nil && chartNow.view.ground != nil && chartNow.hash == h {
		return
	}
	chartWarm.mu.Lock()
	if chartWarm.hash == h && (chartWarm.ready || chartWarm.wait != nil) {
		chartWarm.mu.Unlock()
		return
	}
	done := make(chan struct{})
	chartWarm.hash = h
	chartWarm.paint = terrainPaint{}
	chartWarm.ready = false
	chartWarm.wait = done
	chartWarm.mu.Unlock()
	go func(c Chart, h uint64, done chan struct{}) {
		painted := paintChart(c)
		chartWarm.mu.Lock()
		if chartWarm.wait == done {
			chartWarm.hash = h
			chartWarm.paint = painted
			chartWarm.ready = true
		}
		chartWarm.mu.Unlock()
		close(done)
	}(c, h, done)
}

// takeWarm waits for a warm paint of h and hands it over. The CPU images
// are cleared so they are not kept after the upload.
func takeWarm(h uint64) (terrainPaint, bool) {
	chartWarm.mu.Lock()
	wait := chartWarm.wait
	same := chartWarm.hash == h
	ready := chartWarm.ready && same
	paint := chartWarm.paint
	chartWarm.mu.Unlock()
	if !same {
		return terrainPaint{}, false
	}
	if !ready {
		if wait == nil {
			return terrainPaint{}, false
		}
		<-wait
	}
	chartWarm.mu.Lock()
	defer chartWarm.mu.Unlock()
	if chartWarm.hash != h || !chartWarm.ready || chartWarm.paint.ground == nil {
		return terrainPaint{}, false
	}
	paint = chartWarm.paint
	chartWarm.paint = terrainPaint{}
	chartWarm.ready = false
	chartWarm.wait = nil
	return paint, true
}

// PrepareChart caches one painted region map. A warm paint for this chart
// is uploaded instead of painted again.
func PrepareChart(c Chart) *View {
	if c.W < 1 || c.H < 1 {
		return &View{}
	}
	h := hashChart(c)
	if chartNow.view != nil && chartNow.view.ground != nil && chartNow.hash == h {
		return chartNow.view
	}
	if chartUp.view != nil && chartUp.hash == h {
		uploadMore(true, 0)
		return publishUpload()
	}
	painted, ok := takeWarm(h)
	if !ok {
		painted = paintChart(c)
	}
	startUpload(painted, h)
	uploadMore(true, 0)
	return publishUpload()
}

// chartUpload is a warm paint being moved to the GPU on the game thread.
type chartUpload struct {
	hash  uint64
	paint terrainPaint
	view  *View
	next  int
	total int
}

var chartUp chartUpload

// PumpWarm uploads a finished warm paint on the game thread. budget is how
// long one call may spend on canopy slices after the ground is up. It
// returns true when the chart is cached and the battle will not paint or
// upload it again. Call it from Update.
func PumpWarm(c Chart, budget time.Duration) bool {
	if c.W < 1 || c.H < 1 {
		return false
	}
	h := hashChart(c)
	if chartNow.view != nil && chartNow.view.ground != nil && chartNow.hash == h {
		return true
	}
	if chartUp.view != nil && chartUp.hash != h {
		chartUp = chartUpload{}
	}
	if chartUp.view == nil {
		paint, ok := peekWarm(h)
		if !ok {
			return false
		}
		startUpload(paint, h)
	}
	uploadMore(false, budget)
	if chartUp.next < chartUp.total {
		return false
	}
	publishUpload()
	return true
}

func peekWarm(h uint64) (terrainPaint, bool) {
	chartWarm.mu.Lock()
	defer chartWarm.mu.Unlock()
	if chartWarm.hash != h || !chartWarm.ready || chartWarm.paint.ground == nil {
		return terrainPaint{}, false
	}
	paint := chartWarm.paint
	chartWarm.paint = terrainPaint{}
	chartWarm.ready = false
	chartWarm.wait = nil
	return paint, true
}

func startUpload(paint terrainPaint, h uint64) {
	chartUp = chartUpload{
		hash:  h,
		paint: paint,
		view: &View{
			ground: uploadStraight(paint.ground),
			anim:   uploadStraight(paint.anim),
			mask:   uploadStraight(paint.mask),
		},
		total: runCount(paint.trees),
	}
}

func uploadMore(all bool, budget time.Duration) {
	u := &chartUp
	start := time.Now()
	for u.next < u.total {
		if !all && u.next > 0 && time.Since(start) >= budget {
			return
		}
		run := runAt(u.paint.trees, u.next)
		u.next++
		if run.sprite == nil || run.sprite.Bounds().Empty() {
			continue
		}
		u.view.canopies = append(u.view.canopies, Canopy{
			img:    uploadCanopy(run.sprite),
			padded: true,
			WorldX: run.world,
			NorthY: run.north,
			FootX:  run.footX,
			FootY:  run.foot,
		})
	}
}

func publishUpload() *View {
	chartNow.hash = chartUp.hash
	chartNow.view = chartUp.view
	view := chartUp.view
	chartUp = chartUpload{}
	return view
}

func runCount(trees []treeCanopy) int {
	n := 0
	for i := range trees {
		n += len(trees[i].runs)
	}
	return n
}

func runAt(trees []treeCanopy, n int) treeRun {
	i := 0
	for t := range trees {
		runs := trees[t].runs
		if n < i+len(runs) {
			return runs[n-i]
		}
		i += len(runs)
	}
	return treeRun{}
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
	chartPaintMu.Lock()
	defer chartPaintMu.Unlock()
	paintCharts++
	return paintChartNow(c)
}

func paintChartNow(c Chart) terrainPaint {
	lk := lookOf(c.Terrain)
	ww, wh := c.W, c.H
	prepareGrass(ww, wh, lk.kind)
	prepareSnow(ww, wh, lk.kind)
	prepareSand(ww, wh, lk.kind)
	prepareForest(ww, wh, lk.kind)
	grid := rasterChart(c, lk.open)
	clearSpentMines(grid, c, lk.open)
	prepareSnowCliff(grid, ww, wh, lk.kind, lk.open, c.Regions)
	toWater, toLand := chartWaterDist(grid, ww, wh)
	bindShore(toWater, toLand, ww, wh)
	pw := ww * terrainPx
	ph := wh * terrainPx
	ground := image.NewRGBA(image.Rect(0, 0, pw, ph))
	kind := make([]uint8, pw*ph)
	noise := buildNoise(max(1, ww/sim.TileSize), max(1, wh/sim.TileSize), 11)
	spanMap, span := tilesFromChart(grid, ww, wh)
	chartSpan, chartSpanW, chartSpanH = span, spanMap.W, spanMap.H
	fillChartGround(ground, kind, grid, lk, ww, wh)
	chartSpan = nil
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
	if span != nil {
		stampBridges(ground, anim, mask, spanMap, span, lk)
	}
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
	xs := make([]float64, 0, 8)
	for _, r := range c.Regions {
		if len(r.Points) < 3 {
			continue
		}
		xs = fillRing(g, c.W, xs, r.Points, chartSurface(r.Tile, open))
	}
	return g
}

// fillRing sets samples inside the ring. A sample sits at the cell centre.
// The crossings use the same edge rule as pointInRing.
func fillRing(g []surface, w int, xs []float64, pts [][2]float64, s surface) []float64 {
	x0, y0, x1, y1 := ringBounds(pts, w, len(g)/w)
	for y := y0; y < y1; y++ {
		py := float64(y) + 0.5
		xs = xs[:0]
		j := len(pts) - 1
		for i := range pts {
			yi, yj := pts[i][1], pts[j][1]
			if (yi > py) != (yj > py) {
				xin := (pts[j][0]-pts[i][0])*(py-yi)/(yj-yi) + pts[i][0]
				xs = append(xs, xin)
			}
			j = i
		}
		if len(xs) == 0 {
			continue
		}
		sort.Float64s(xs)
		// From the left of the map, an odd number of crossings still ahead
		// means this centre is inside. Passing a crossing (centre >= xin)
		// drops it, which matches the ray test x < xin.
		inside := len(xs)%2 == 1
		ci := 0
		row := y * w
		for x := x0; x < x1; x++ {
			center := float64(x) + 0.5
			for ci < len(xs) && !(center < xs[ci]) {
				inside = !inside
				ci++
			}
			if inside {
				g[row+x] = s
			}
		}
	}
	return xs
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

// shoreFlat is how far inland a world-pixel centre must be before the
// whole 4×4 source block is past the shore early-out (sd > 36). Samples
// sit at most 0.375 world pixels from that centre.
const shoreFlat = 38.0

// fillChartGround writes the ground colour. Inland procedural pixels are
// one colour per world pixel, so the block is stamped once. Shore, water,
// and plate samples still run per source pixel. Rows do not overlap.
func fillChartGround(ground *image.RGBA, kind []uint8, grid []surface, lk look, ww, wh int) {
	pw := ww * terrainPx
	parallelStripes(wh, func(y int) {
		wy := float64(y) + 0.5
		gy := y * ww
		py := y * terrainPx
		for x := 0; x < ww; x++ {
			wx := float64(x) + 0.5
			s := grid[gy+x]
			if col, k, ok := uniformPixel(lk, s, wx, wy); ok {
				stampBlock(ground, kind, x*terrainPx, py, pw, col, k)
				continue
			}
			for sy := 0; sy < terrainPx; sy++ {
				pyy := py + sy
				wyy := (float64(pyy) + 0.5) / float64(terrainPx)
				row := pyy * pw
				for sx := 0; sx < terrainPx; sx++ {
					pxx := x*terrainPx + sx
					wxx := (float64(pxx) + 0.5) / float64(terrainPx)
					col, k := chartPixel(lk, s, wxx, wyy)
					setRGBA(ground, pxx, pyy, col)
					kind[row+pxx] = k
				}
			}
		}
	})
}

func stampBlock(dst *image.RGBA, kind []uint8, x, y, pw int, c color.RGBA, k uint8) {
	for sy := 0; sy < terrainPx; sy++ {
		o := dst.PixOffset(x, y+sy)
		row := (y+sy)*pw + x
		for sx := 0; sx < terrainPx; sx++ {
			dst.Pix[o] = c.R
			dst.Pix[o+1] = c.G
			dst.Pix[o+2] = c.B
			dst.Pix[o+3] = c.A
			kind[row] = k
			o += 4
			row++
		}
	}
}

// uniformPixel reports the colour of a world pixel when every source sample
// inside it matches chartPixel. wx and wy are the pixel centre.
func uniformPixel(lk look, s surface, wx, wy float64) (color.RGBA, uint8, bool) {
	if snowCliff != nil && !cliffBlocks(s) && snowCliff.hotAt(wx, wy) {
		return color.RGBA{}, 0, false
	}
	switch s {
	case surfBridge, surfShallow, surfDeep, surfIce, surfSand, surfCliff, surfRamp:
		return color.RGBA{}, 0, false
	}
	if shoreToWater != nil {
		sd, ok := shoreField(wx, wy)
		if ok && sd <= shoreFlat {
			return color.RGBA{}, 0, false
		}
	}
	base := s
	if s == surfTree || s == surfMine {
		base = lk.open
	}
	switch lk.kind {
	case Jungle, Desert, Forest:
		if base != surfGrass {
			return color.RGBA{}, 0, false
		}
	case Snow:
		if base != surfSnow && base != surfGrass {
			return color.RGBA{}, 0, false
		}
	default:
		return color.RGBA{}, 0, false
	}
	c, k := lk.plateColor(base, wx, wy)
	return c, k, true
}

func chartPixel(lk look, s surface, wx, wy float64) (color.RGBA, uint8) {
	if snowCliff != nil && !cliffBlocks(s) && snowCliff.hotAt(wx, wy) {
		return paintSnowCliffPixel(lk, s, wx, wy, false, color.RGBA{}, 0)
	}
	return chartPixelPaint(lk, s, wx, wy)
}

// chartSpan is the tile grid for the chart being painted. Bridge pixels
// take the water or bank under the span. stampBridges draws the deck after.
var (
	chartSpan  []surface
	chartSpanW int
	chartSpanH int
)

// tilesFromChart samples each tile at its centre. A bridge cell is one
// whose centre lies in the bridge region, matching how the battle walks.
func tilesFromChart(grid []surface, ww, wh int) (sim.Map, []surface) {
	tw := ww / sim.TileSize
	th := wh / sim.TileSize
	if tw < 1 || th < 1 || len(grid) < ww*wh {
		return sim.Map{}, nil
	}
	surf := make([]surface, tw*th)
	for ty := 0; ty < th; ty++ {
		for tx := 0; tx < tw; tx++ {
			cx := tx*sim.TileSize + sim.TileSize/2
			cy := ty*sim.TileSize + sim.TileSize/2
			surf[ty*tw+tx] = grid[cy*ww+cx]
		}
	}
	return sim.Map{W: tw, H: th}, surf
}

func chartPixelPaint(lk look, s surface, wx, wy float64) (color.RGBA, uint8) {
	if s == surfBridge && chartSpan != nil {
		tx := int(math.Floor(wx / sim.TileSize))
		ty := int(math.Floor(wy / sim.TileSize))
		if tx < 0 {
			tx = 0
		}
		if ty < 0 {
			ty = 0
		}
		if tx >= chartSpanW {
			tx = chartSpanW - 1
		}
		if ty >= chartSpanH {
			ty = chartSpanH - 1
		}
		under := bridgeUnder(chartSpan, chartSpanW, chartSpanH, tx, ty, openGround(chartSpan, chartSpanW, chartSpanH))
		return chartPixelPaint(lk, under, wx, wy)
	}
	// Water replaces the plate colour. plateColor would paint the same
	// water and this function would paint it again.
	if s == surfShallow || s == surfDeep {
		if shoreToWater == nil {
			return lk.plateColor(s, wx, wy)
		}
		sd, ok := shoreField(wx, wy)
		if !ok {
			return lk.plateColor(s, wx, wy)
		}
		sd += shoreFlow(wx, wy)
		if sd > 0 {
			return wetSand(lk, wx, wy, sd), maskNone
		}
		return waterColor(lk, wx, wy, -sd, -sd), waterMask(-sd)
	}
	base := s
	if s == surfTree || s == surfMine {
		base = lk.open
	}
	c, k := lk.plateColor(base, wx, wy)
	if shoreToWater == nil {
		return c, k
	}
	water := s == surfShallow || s == surfDeep
	sd, ok := shoreField(wx, wy)
	if !ok {
		return c, k
	}
	// Far inland the wave cannot reach, so the ground colour stands.
	if !water && sd > 36 {
		return c, k
	}
	sd += shoreFlow(wx, wy)
	if water {
		if sd > 0 {
			return wetSand(lk, wx, wy, sd), maskNone
		}
		return waterColor(lk, wx, wy, -sd, -sd), waterMask(-sd)
	}
	if s == surfIce || s == surfCliff || s == surfRamp || s == surfSand || s == surfBridge {
		return c, k
	}
	if sd < 0 {
		return waterColor(lk, wx, wy, -sd, -sd), waterMask(-sd)
	}
	reach := sandReach(wx, wy)
	if sd >= reach {
		return c, k
	}
	return featherLand(lk, c, wx, wy, sd, reach), maskNone
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
		if isWetSurf(s) {
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
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		chamfer(toWater, w, h)
	}()
	go func() {
		defer wg.Done()
		chamfer(toLand, w, h)
	}()
	wg.Wait()
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
	near := newCrownNear(crowns)
	rows := 0
	if y1 > y0 {
		rows = (y1 - y0 + terrainPx - 1) / terrainPx
	}
	parallelStripes(rows, func(i int) {
		py := y0 + i*terrainPx
		wy := (float64(py) + 0.5) / float64(terrainPx)
		for px := x0; px < x1; px += terrainPx {
			wx := (float64(px) + 0.5) / float64(terrainPx)
			d := near.dist(wx-3.2, wy-4.0)
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
	})
	parallelStripes(ph, func(y int) {
		if y < y0 || y >= y1 {
			return
		}
		row := y * pw
		base := row * 4
		for x := x0; x < x1; x++ {
			v := strength[row+x]
			if v == 0 {
				continue
			}
			o := base + x*4
			c := color.RGBA{R: dst.Pix[o], G: dst.Pix[o+1], B: dst.Pix[o+2], A: dst.Pix[o+3]}
			c = mixRGBA(c, lk.shadow, float64(v)/255)
			dst.Pix[o] = c.R
			dst.Pix[o+1] = c.G
			dst.Pix[o+2] = c.B
			dst.Pix[o+3] = c.A
		}
	})
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
	cover, tips, ccx, ccy, spanLo, spanHi := splatCrowns(crowns, originX, originY, pw, ph)
	gw := ground.Bounds().Dx()
	gh := ground.Bounds().Dy()
	near := newCellNear(cells)
	// A world pixel holds a 4×4 block of source samples. Those samples share
	// a nearest tree when the winner leads by more than the block diagonal.
	const ownerGap = 1.061
	ox4, oy4 := 0, 0
	half := originX == math.Trunc(originX) && originY == math.Trunc(originY)
	if half {
		ox4 = int(originX) * 4
		oy4 = int(originY) * 4
	}
	canopy := lk.canopy
	workers := stripeWorkers(ph)
	localBox := make([][]pxBox, workers)
	for w := 0; w < workers; w++ {
		localBox[w] = make([]pxBox, len(cells))
		for i := range localBox[w] {
			localBox[w][i] = pxBox{minX: pw, minY: ph, maxX: -1, maxY: -1}
		}
	}
	parallelStripesW(ph, func(worker, py int) {
		if spanHi[py] < spanLo[py] {
			return
		}
		wy := originY + (float64(py)+0.5)/float64(terrainPx)
		row := py * pw
		boxes := localBox[worker]
		blockOwn := -2
		for px := spanLo[py]; px <= spanHi[py]; px++ {
			if px&(terrainPx-1) == 0 {
				blockOwn = -2
			}
			d := float64(cover[row+px])
			sprig := float64(tips[row+px])
			if d > 12 && sprig < 0.02 {
				continue
			}
			wx := originX + (float64(px)+0.5)/float64(terrainPx)
			// Foliage noise stays inside ±2.6, so a crown this deep is
			// solid and one this far out is clear. Same result as canopyAlpha.
			var a float64
			if d <= -6.8 {
				a = 1
			} else if d >= 10.1 {
				a = 0
			} else {
				a = canopyAlpha(d, wx, wy, noise)
			}
			if sprig > a {
				a = sprig
			}
			if a < 0.012 {
				continue
			}
			var c color.RGBA
			if half {
				c = canopy.atHalf(ox4+px, oy4+py)
			} else {
				c = canopy.atScale(wx, wy, 4)
			}
			n := 0.0
			if lk.snowCaps {
				n = sampleNoise(noise, wx*1.7, wy*1.7)
			}
			c = shadeCrown(c, d, wx-float64(ccx[row+px]), wy-float64(ccy[row+px]), lk, n)
			if a > 0.5 {
				if a < 0.85 && int(c.R)+int(c.G)+int(c.B) < 96 {
					c = mixRGBA(c, lk.leaf, 0.45)
				}
			} else {
				c = mergeFringe(c, ground, gw, gh, wx, wy, a, lk)
			}
			c.A = uint8(a * 255)
			if c.A == 0 {
				continue
			}
			setRGBA(img, px, py, c)
			if blockOwn == -2 {
				cx := originX + (float64(px&^(terrainPx-1))+2)/float64(terrainPx)
				cy := originY + (float64(py&^(terrainPx-1))+2)/float64(terrainPx)
				idx, gap := near.nearest(cx, cy)
				if gap > ownerGap {
					blockOwn = idx
				} else {
					blockOwn = -1
				}
			}
			own := blockOwn
			if own < 0 {
				own = near.index(wx, wy)
			}
			owner[row+px] = int16(own)
			b := &boxes[own]
			if px < b.minX {
				b.minX = px
			}
			if py < b.minY {
				b.minY = py
			}
			if px > b.maxX {
				b.maxX = px
			}
			if py > b.maxY {
				b.maxY = py
			}
		}
	})
	can := treeCanopy{img: img, originX: originX, originY: originY}
	for i, cell := range cells {
		box := pxBox{minX: pw, minY: ph, maxX: -1, maxY: -1}
		for w := 0; w < workers; w++ {
			b := localBox[w][i]
			if b.maxX < 0 {
				continue
			}
			if b.minX < box.minX {
				box.minX = b.minX
			}
			if b.minY < box.minY {
				box.minY = b.minY
			}
			if b.maxX > box.maxX {
				box.maxX = b.maxX
			}
			if b.maxY > box.maxY {
				box.maxY = b.maxY
			}
		}
		can.runs = append(can.runs, sliceBoxed(can, box.minX, box.minY, box.maxX, box.maxY, cell.X, cell.Y))
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

package mapedit

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"

	"fannon-codder/internal/data"
	"fannon-codder/internal/sim"
	"fannon-codder/internal/terrain"
)

const (
	wheelStep = 48
	tps       = 60
)

// Editor shows one campaign phase. A region map is painted from its
// polygons. A tile map still uses the battle ground. In the regions
// view a left drag adds a tree region for this session. Nothing is written.
type Editor struct {
	phases   []phaseRow
	index    int
	grid     sim.Map
	biome    string
	freeform bool
	chart    terrain.Chart
	err      error
	// start is the player tile. hasStart is false until a phase loads.
	start    [2]int
	hasStart bool
	enemies  []data.Enemy

	scrollX, scrollY float64
	time             float64
	gridView         bool
	stroke           [][2]float64
	added            map[int][]terrain.ChartRegion

	scale      float64
	offW, offH int
	fbW, fbH   int

	drag                     bool
	dragX, dragY             float64
	dragScrollX, dragScrollY float64

	view         *ebiten.Image
	viewW, viewH int
	font         *text.GoTextFaceSource
}

type phaseRow struct {
	file  string
	name  string
	title string
}

// New loads campaign.json. A missing campaign is an error. A bad phase
// stays in the list and shows its error when selected.
func New() (*Editor, error) {
	c, err := data.LoadCampaign()
	if err != nil {
		return nil, err
	}
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return nil, err
	}
	ed := &Editor{font: src, scale: 1}
	for _, file := range c.Phases {
		row := phaseRow{file: file, name: strings.TrimSuffix(file, ".json")}
		if p, err := data.LoadPhase(file); err == nil {
			row.title = p.Title
		}
		ed.phases = append(ed.phases, row)
	}
	ed.load(0)
	return ed, nil
}

func (e *Editor) load(i int) {
	e.drag = false
	e.stroke = nil
	if len(e.phases) == 0 {
		e.err = fmt.Errorf("campaign has no phases")
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(e.phases) {
		i = len(e.phases) - 1
	}
	e.index = i
	e.scrollX, e.scrollY = 0, 0
	e.freeform = false
	e.chart = terrain.Chart{}
	e.start = [2]int{}
	e.hasStart = false
	e.enemies = nil
	p, err := data.LoadPhase(e.phases[i].file)
	if err != nil {
		e.err = err
		e.grid = sim.Map{}
		e.biome = ""
		return
	}
	e.phases[i].title = p.Title
	g, err := p.Grid()
	if err != nil {
		e.err = err
		e.grid = sim.Map{}
		e.biome = p.Terrain
		return
	}
	e.biome = p.Terrain
	e.freeform = p.Map.Freeform()
	e.chart = terrain.Chart{}
	if e.freeform {
		ch, err := chartFrom(p)
		if err != nil {
			e.err = err
			e.grid = sim.Map{}
			return
		}
		e.chart = ch
		e.chart.Regions = append(e.chart.Regions, e.added[i]...)
	}
	e.err = nil
	e.grid = g
	e.start = p.PlayerStart
	e.hasStart = true
	e.enemies = append([]data.Enemy(nil), p.Enemies...)
}

func (e *Editor) label() string {
	if len(e.phases) == 0 {
		return "Fannon Codder"
	}
	row := e.phases[e.index]
	if row.title == "" {
		return "Fannon Codder — " + row.name
	}
	return "Fannon Codder — " + row.name + " " + row.title
}

func (e *Editor) Update() error {
	e.time += 1.0 / tps
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		e.gridView = !e.gridView
		e.stroke = nil
	}
	ebiten.SetWindowTitle(e.label())

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && e.index > 0 {
		e.load(e.index - 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && e.index+1 < len(e.phases) {
		e.load(e.index + 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		e.scrollX -= sim.TileSize
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		e.scrollX += sim.TileSize
	}

	e.pointer()

	wx, wy := ebiten.Wheel()
	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight) {
		wx += wy
		wy = 0
	}
	e.scrollX -= wx * wheelStep
	e.scrollY -= wy * wheelStep
	e.clampScroll()
	return nil
}

// pointer selects a phase on the rail. On the map, a left drag draws a
// tree region in the regions view and scrolls in the other views.
func (e *Editor) pointer() {
	cx, cy := e.cursor()
	rail := RailWidth(e.offW)
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if cx < float64(rail) {
			if cy < InfoTop(e.offH, e.scale) {
				if row, ok := RailHit(cx, cy, rail, RowHeight(e.scale), len(e.phases)); ok {
					e.drag = false
					e.stroke = nil
					e.load(row)
				}
			}
		} else {
			x, y := e.worldCursor(cx, cy)
			if !e.beginStroke(x, y) {
				e.drag = true
				e.dragX, e.dragY = cx, cy
				e.dragScrollX, e.dragScrollY = e.scrollX, e.scrollY
			}
		}
	}
	if len(e.stroke) > 0 {
		x, y := e.worldCursor(cx, cy)
		if down {
			e.extendStroke(x, y)
		} else {
			e.finishStroke(x, y)
		}
	}
	if e.drag {
		if down {
			s := e.scale
			if s <= 0 {
				s = 1
			}
			e.scrollX = e.dragScrollX - (cx-e.dragX)/s
			e.scrollY = e.dragScrollY - (cy-e.dragY)/s
		} else {
			e.drag = false
		}
	}
}

func (e *Editor) clampScroll() {
	mw, mh := e.mapPixels()
	vw, vh := e.viewWorld()
	e.scrollX, e.scrollY = ClampScroll(e.scrollX, e.scrollY, mw, mh, vw, vh)
}

func (e *Editor) mapPixels() (float64, float64) {
	if e.grid.W <= 0 || e.grid.H <= 0 || len(e.grid.Tiles) < e.grid.W*e.grid.H {
		return 0, 0
	}
	return float64(e.grid.W * sim.TileSize), float64(e.grid.H * sim.TileSize)
}

func (e *Editor) viewWorld() (float64, float64) {
	pw, ph := viewportPixels(e.offW, e.offH)
	s := e.scale
	if s <= 0 {
		s = 1
	}
	return float64(pw) / s, float64(ph) / s
}

func viewportPixels(offW, offH int) (int, int) {
	w := offW - RailWidth(offW)
	if w < 1 {
		w = 1
	}
	h := offH
	if h < 1 {
		h = 1
	}
	return w, h
}

func (e *Editor) cursor() (float64, float64) {
	cx, cy := ebiten.CursorPosition()
	return OffscreenPoint(float64(cx), float64(cy), e.fbW, e.fbH, e.offW, e.offH)
}

func (e *Editor) Draw(screen *ebiten.Image) {
	screen.Fill(railBG)
	e.drawRail(screen)
	e.drawMap(screen)
}

func (e *Editor) drawRail(screen *ebiten.Image) {
	rail := RailWidth(e.offW)
	if rail < 1 || e.offH < 1 {
		return
	}
	clip := screen
	if sub, ok := screen.SubImage(image.Rect(0, 0, rail, e.offH)).(*ebiten.Image); ok {
		clip = sub
	}
	rh := RowHeight(e.scale)
	infoTop := InfoTop(e.offH, e.scale)
	for i, row := range e.phases {
		y := float64(i) * rh
		if y+rh > infoTop {
			break
		}
		if i == e.index {
			fillRect(clip, 0, int(y), rail, int(rh)+1, railSel)
		}
		ink := railInk
		if i == e.index {
			ink = railInkSel
		}
		label := row.name
		if row.title != "" {
			label += "  " + row.title
		}
		drawFace(clip, e.font, e.scale, label, 8, y+4, ink)
	}
	e.drawInfo(clip, rail, infoTop, rh)
}

func (e *Editor) infoLines() (terrain, size string) {
	terrain = e.biome
	if terrain == "" {
		terrain = "—"
	}
	if e.err != nil || e.grid.W <= 0 || e.grid.H <= 0 {
		return terrain, "—"
	}
	if e.freeform {
		return terrain, fmt.Sprintf("%d × %d", e.grid.W*sim.TileSize, e.grid.H*sim.TileSize)
	}
	return terrain, fmt.Sprintf("%d × %d tiles", e.grid.W, e.grid.H)
}

func (e *Editor) drawInfo(clip *ebiten.Image, rail int, infoTop, rh float64) {
	if clip == nil || rail < 1 {
		return
	}
	fillRect(clip, 0, int(infoTop), rail, 1, railSel)
	terrain, size := e.infoLines()
	drawFace(clip, e.font, e.scale, "Terrain", 8, infoTop+4, railInk)
	drawFace(clip, e.font, e.scale, terrain, 8, infoTop+4+rh, railInkSel)
	drawFace(clip, e.font, e.scale, size, 8, infoTop+4+rh*2, railInkSel)
	drawFace(clip, e.font, e.scale, "View  V", 8, infoTop+4+rh*3, railInk)
	drawFace(clip, e.font, e.scale, e.viewName(), 8, infoTop+4+rh*4, railInkSel)
}

func (e *Editor) viewName() string {
	if e.gridView {
		if e.freeform {
			return "regions"
		}
		return "tile grid"
	}
	return "terrain"
}

func (e *Editor) drawMap(screen *ebiten.Image) {
	rail := RailWidth(e.offW)
	vw, vh := viewportPixels(e.offW, e.offH)
	if e.view == nil || e.viewW != vw || e.viewH != vh {
		e.view = ebiten.NewImage(vw, vh)
		e.viewW, e.viewH = vw, vh
	}
	e.view.Fill(color.Black)
	if e.err != nil {
		drawFace(e.view, e.font, e.scale, e.err.Error(), 8, 8, railInkSel)
	} else if e.gridView && e.freeform {
		e.drawRegions(e.view)
	} else if e.gridView {
		e.drawTileGrid(e.view)
	} else if e.freeform {
		e.paint(e.view, terrain.PrepareChart(e.chart))
	} else {
		e.paint(e.view, terrain.Prepare(e.grid, e.biome))
	}
	e.drawMarks(e.view)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(rail), 0)
	screen.DrawImage(e.view, op)
}

func (e *Editor) Layout(outsideWidth, outsideHeight int) (int, int) {
	return e.prepare(float64(outsideWidth), float64(outsideHeight))
}

func (e *Editor) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	w, h := e.prepare(outsideWidth, outsideHeight)
	return float64(w), float64(h)
}

func (e *Editor) prepare(outsideWidth, outsideHeight float64) (int, int) {
	s, w, h := Picture(outsideWidth, outsideHeight, monitorScale())
	e.scale = s
	e.offW, e.offH = w, h
	return w, h
}

// DrawFinalScreen blits the offscreen 1:1, nearest, centered. Same contract
// as the game, so the picture scale is not applied twice.
func (e *Editor) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, _ ebiten.GeoM) {
	b := screen.Bounds()
	e.fbW = b.Dx()
	e.fbH = b.Dy()
	screen.Fill(color.Black)
	ob := offscreen.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(float64(b.Dx()-ob.Dx())/2, float64(b.Dy()-ob.Dy())/2)
	screen.DrawImage(offscreen, op)
}

func monitorScale() float64 {
	m := ebiten.Monitor()
	if m == nil {
		return 1
	}
	s := m.DeviceScaleFactor()
	if s <= 0 {
		return 1
	}
	return s
}

var (
	railBG     = color.RGBA{R: 28, G: 36, B: 22, A: 255}
	railSel    = color.RGBA{R: 96, G: 112, B: 58, A: 255}
	railInk    = color.RGBA{R: 186, G: 176, B: 120, A: 255}
	railInkSel = color.RGBA{R: 244, G: 236, B: 200, A: 255}
	gridLine   = color.RGBA{R: 12, G: 12, B: 12, A: 255}
	gridInk    = color.RGBA{R: 244, G: 236, B: 200, A: 255}
	markPlayer = color.RGBA{R: 0xff, G: 0xc4, B: 0x00, A: 255}
	markEnemy  = color.RGBA{R: 0xe0, G: 0x30, B: 0x30, A: 255}
)

func drawFace(dst *ebiten.Image, src *text.GoTextFaceSource, scale float64, str string, x, y float64, c color.Color) {
	if dst == nil || src == nil || str == "" {
		return
	}
	if scale <= 0 {
		scale = 1
	}
	face := &text.GoTextFace{Source: src, Size: 14 * scale / 3}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, str, face, op)
}

var whiteOnce struct {
	img  *ebiten.Image
	once sync.Once
}

func (e *Editor) paint(dst *ebiten.Image, pic *terrain.View) {
	if pic == nil {
		return
	}
	pic.DrawGround(dst, -e.scrollX, -e.scrollY, e.scale, e.time)
	src := pic.Canopies()
	cans := make([]terrain.Canopy, len(src))
	copy(cans, src)
	sort.SliceStable(cans, func(i, j int) bool { return cans[i].FootY < cans[j].FootY })
	for i := range cans {
		cans[i].DrawShadow(dst, -e.scrollX, -e.scrollY, e.scale)
	}
	for i := range cans {
		cans[i].Draw(dst, -e.scrollX, -e.scrollY, e.scale)
	}
}

func chartFrom(p *data.Phase) (terrain.Chart, error) {
	ch := terrain.Chart{
		W:       p.Map.W * sim.TileSize,
		H:       p.Map.H * sim.TileSize,
		Terrain: p.Terrain,
	}
	for i, r := range p.Map.Regions {
		tile, err := data.RegionTile(r.Type)
		if err != nil {
			return terrain.Chart{}, fmt.Errorf("region %d: %w", i, err)
		}
		if len(r.Points) < 3 {
			return terrain.Chart{}, fmt.Errorf("region %d needs at least 3 points", i)
		}
		cr := terrain.ChartRegion{Tile: tile, Name: r.Type}
		for _, pt := range r.Points {
			if len(pt) != 2 {
				return terrain.Chart{}, fmt.Errorf("region %d point", i)
			}
			cr.Points = append(cr.Points, [2]float64{pt[0], pt[1]})
		}
		ch.Regions = append(ch.Regions, cr)
	}
	return ch, nil
}

func (e *Editor) drawRegions(dst *ebiten.Image) {
	if dst == nil || e.chart.W < 1 || e.chart.H < 1 {
		return
	}
	s := e.scale
	if s <= 0 {
		s = 1
	}
	ox := -e.scrollX * s
	oy := -e.scrollY * s
	fillRect(dst, int(math.Floor(ox)), int(math.Floor(oy)), int(math.Ceil(float64(e.chart.W)*s)), int(math.Ceil(float64(e.chart.H)*s)), groundFill(e.biome))
	for _, r := range e.chart.Regions {
		if len(r.Points) < 3 {
			continue
		}
		path := regionPath(r.Points, ox, oy, s)
		fill := &vector.DrawPathOptions{AntiAlias: true}
		fill.ColorScale.ScaleWithColor(tileFill(r.Tile))
		vector.FillPath(dst, path, nil, fill)
		stroke := &vector.DrawPathOptions{AntiAlias: true}
		stroke.ColorScale.ScaleWithColor(gridLine)
		vector.StrokePath(dst, path, &vector.StrokeOptions{Width: float32(math.Max(1, s)), LineJoin: vector.LineJoinRound}, stroke)
		cx, cy := 0.0, 0.0
		for _, p := range r.Points {
			cx += p[0]
			cy += p[1]
		}
		n := float64(len(r.Points))
		drawFace(dst, e.font, s, r.Name, ox+cx/n*s, oy+cy/n*s, gridInk)
	}
	e.drawStroke(dst)
}

func regionPath(pts [][2]float64, ox, oy, scale float64) *vector.Path {
	p := &vector.Path{}
	p.MoveTo(float32(ox+pts[0][0]*scale), float32(oy+pts[0][1]*scale))
	for _, pt := range pts[1:] {
		p.LineTo(float32(ox+pt[0]*scale), float32(oy+pt[1]*scale))
	}
	p.Close()
	return p
}

func groundFill(biome string) color.RGBA {
	switch terrain.KindOf(biome) {
	case terrain.Desert:
		return color.RGBA{R: 196, G: 168, B: 96, A: 255}
	case terrain.Snow:
		return color.RGBA{R: 220, G: 228, B: 232, A: 255}
	case terrain.Forest:
		return color.RGBA{R: 48, G: 92, B: 52, A: 255}
	default:
		return tileFill(sim.TileGrass)
	}
}

// markAt is the screen point of a tile centre in the viewport.
func (e *Editor) markAt(tx, ty int) (float64, float64) {
	s := e.scale
	if s <= 0 {
		s = 1
	}
	c := float64(sim.TileSize) * s
	return -e.scrollX*s + (float64(tx)+0.5)*c, -e.scrollY*s + (float64(ty)+0.5)*c
}

func (e *Editor) drawMarks(dst *ebiten.Image) {
	if dst == nil || e.err != nil || !e.hasStart && len(e.enemies) == 0 {
		return
	}
	s := e.scale
	if s <= 0 {
		s = 1
	}
	for _, en := range e.enemies {
		x, y := e.markAt(en.X, en.Y)
		drawDiamond(dst, x, y, 3.5*s, markEnemy)
	}
	if e.hasStart {
		x, y := e.markAt(e.start[0], e.start[1])
		drawDiamond(dst, x, y, 4.5*s, markPlayer)
	}
}

func drawDiamond(dst *ebiten.Image, x, y, r float64, c color.Color) {
	if r < 1 {
		r = 1
	}
	p := &vector.Path{}
	p.MoveTo(float32(x), float32(y-r))
	p.LineTo(float32(x+r), float32(y))
	p.LineTo(float32(x), float32(y+r))
	p.LineTo(float32(x-r), float32(y))
	p.Close()
	edge := &vector.DrawPathOptions{AntiAlias: true}
	edge.ColorScale.ScaleWithColor(gridLine)
	vector.FillPath(dst, p, nil, edge)
	in := &vector.Path{}
	ri := r * 0.62
	in.MoveTo(float32(x), float32(y-ri))
	in.LineTo(float32(x+ri), float32(y))
	in.LineTo(float32(x), float32(y+ri))
	in.LineTo(float32(x-ri), float32(y))
	in.Close()
	fill := &vector.DrawPathOptions{AntiAlias: true}
	fill.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, in, nil, fill)
}

func (e *Editor) drawTileGrid(dst *ebiten.Image) {
	if dst == nil || e.grid.W <= 0 || e.grid.H <= 0 {
		return
	}
	s := e.scale
	if s <= 0 {
		s = 1
	}
	ts := float64(sim.TileSize) * s
	ox := -e.scrollX * s
	oy := -e.scrollY * s
	for ty := 0; ty < e.grid.H; ty++ {
		for tx := 0; tx < e.grid.W; tx++ {
			tile := e.grid.At(tx, ty)
			x0 := int(math.Floor(ox + float64(tx)*ts))
			y0 := int(math.Floor(oy + float64(ty)*ts))
			x1 := int(math.Floor(ox + float64(tx+1)*ts))
			y1 := int(math.Floor(oy + float64(ty+1)*ts))
			fillRect(dst, x0, y0, x1-x0, y1-y0, tileFill(tile))
			fillRect(dst, x1-1, y0, 1, y1-y0, gridLine)
			fillRect(dst, x0, y1-1, x1-x0, 1, gridLine)
			label := string(data.TileRune(tile))
			drawFace(dst, e.font, s, label, float64(x0)+4, float64(y0)+2, gridInk)
		}
	}
}

func tileFill(t sim.Tile) color.RGBA {
	switch t {
	case sim.TileTree:
		return color.RGBA{R: 24, G: 72, B: 28, A: 255}
	case sim.TileWaterShallow:
		return color.RGBA{R: 64, G: 140, B: 168, A: 255}
	case sim.TileWaterDeep:
		return color.RGBA{R: 16, G: 52, B: 112, A: 255}
	case sim.TileIce:
		return color.RGBA{R: 176, G: 220, B: 232, A: 255}
	case sim.TileQuicksand:
		return color.RGBA{R: 168, G: 140, B: 64, A: 255}
	case sim.TileCliff:
		return color.RGBA{R: 88, G: 88, B: 92, A: 255}
	case sim.TileBridge:
		return color.RGBA{R: 120, G: 76, B: 40, A: 255}
	case sim.TileMine:
		return color.RGBA{R: 160, G: 40, B: 36, A: 255}
	case sim.TileRamp:
		return color.RGBA{R: 140, G: 128, B: 72, A: 255}
	default:
		return color.RGBA{R: 72, G: 124, B: 48, A: 255}
	}
}

func fillRect(dst *ebiten.Image, x, y, w, h int, c color.Color) {
	if dst == nil || w <= 0 || h <= 0 {
		return
	}
	whiteOnce.once.Do(func() {
		whiteOnce.img = ebiten.NewImage(1, 1)
		whiteOnce.img.Fill(color.White)
	})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w), float64(h))
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	dst.DrawImage(whiteOnce.img, op)
}

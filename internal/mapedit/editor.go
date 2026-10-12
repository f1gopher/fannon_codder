package mapedit

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"

	"fannon-codder/internal/data"
	"fannon-codder/internal/sim"
	"fannon-codder/internal/terrain"
)

const (
	wheelStep = 48
	tps       = 60
)

// Editor shows one campaign phase. It reads the phase JSON and paints with
// the battle ground. It does not write files or change cells.
type Editor struct {
	phases []phaseRow
	index  int
	grid   sim.Map
	biome  string
	err    error

	scrollX, scrollY float64
	time             float64

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
	e.err = nil
	e.grid = g
	e.biome = p.Terrain
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

	cx, cy := e.cursor()
	rail := RailWidth(e.offW)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if cy < InfoTop(e.offH, e.scale) {
			if row, ok := RailHit(cx, cy, rail, RowHeight(e.scale), len(e.phases)); ok {
				e.drag = false
				e.load(row)
			}
		} else if cx >= float64(rail) {
			e.drag = true
			e.dragX, e.dragY = cx, cy
			e.dragScrollX, e.dragScrollY = e.scrollX, e.scrollY
		}
	}
	if e.drag {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
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
	} else {
		pic := terrain.Prepare(e.grid, e.biome)
		pic.DrawGround(e.view, -e.scrollX, -e.scrollY, e.scale, e.time)
		src := pic.Canopies()
		cans := make([]terrain.Canopy, len(src))
		copy(cans, src)
		sort.SliceStable(cans, func(i, j int) bool { return cans[i].FootY < cans[j].FootY })
		for i := range cans {
			cans[i].DrawShadow(e.view, -e.scrollX, -e.scrollY, e.scale)
		}
		for i := range cans {
			cans[i].Draw(e.view, -e.scrollX, -e.scrollY, e.scale)
		}
	}
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

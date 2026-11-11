package mapedit

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"fannon-codder/assets/art"
	"fannon-codder/internal/data"
	"fannon-codder/internal/sim"
	"fannon-codder/internal/terrain"
)

// artScale matches the battle sheets: source pixels per world pixel.
const artScale = 8

type propCel struct {
	img      *ebiten.Image
	ax, ay   int
	shadow   *ebiten.Image
	sax, say int
}

var propLib struct {
	once                                                       sync.Once
	door, plain, tent, handover, home, reinforced, factory, computer, grenade, rocket, barrel, civil, hostage, leader *propCel
}

func sheets() (door, plain, grenade, rocket, barrel *propCel) {
	propLib.once.Do(func() {
		propLib.door, _ = loadCel("hut/door")
		propLib.plain, _ = loadCel("hut/plain")
		propLib.tent, _ = loadCel("hut/tent")
		propLib.handover, _ = loadCel("hut/handover")
		propLib.home, _ = loadCel("hut/home")
		propLib.reinforced, _ = loadCel("hut/reinforced")
		propLib.factory, _ = loadCel("hut/factory")
		propLib.computer, _ = loadCel("hut/computer")
		propLib.leader, _ = loadCelRow("leader/idle", 2)
		propLib.grenade, _ = loadCel("crate/grenade")
		propLib.rocket, _ = loadCel("crate/rocket")
		propLib.barrel, _ = loadCel("barrel/drum")
		propLib.civil, _ = loadCelRow("civilian/idle", 2)
		propLib.hostage, _ = loadCelRow("hostage/idle", 2)
	})
	return propLib.door, propLib.plain, propLib.grenade, propLib.rocket, propLib.barrel
}

func civilianCel() *propCel {
	sheets()
	return propLib.civil
}

func hostageCel() *propCel {
	sheets()
	return propLib.hostage
}

func tentCel() *propCel {
	sheets()
	return propLib.tent
}

func handoverCel() *propCel {
	sheets()
	return propLib.handover
}

func homeCel() *propCel {
	sheets()
	return propLib.home
}

func leaderCel() *propCel {
	sheets()
	return propLib.leader
}

func reinforcedCel() *propCel {
	sheets()
	return propLib.reinforced
}

func factoryCel() *propCel {
	sheets()
	return propLib.factory
}

func computerCel() *propCel {
	sheets()
	return propLib.computer
}

type celManifest struct {
	FrameW  int `json:"frameW"`
	FrameH  int `json:"frameH"`
	AnchorX int `json:"anchorX"`
	AnchorY int `json:"anchorY"`
}

func loadCel(key string) (*propCel, error) {
	return loadCelRow(key, 0)
}

// loadCelRow takes one frame row. Soldier sheets stack E, SE, S, so row 2 faces the camera.
func loadCelRow(key string, row int) (*propCel, error) {
	raw, err := art.Files.ReadFile(key + ".json")
	if err != nil {
		return nil, err
	}
	var m celManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	pngBytes, err := art.Files.ReadFile(key + ".png")
	if err != nil {
		return nil, err
	}
	src, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	if m.FrameW < 1 || m.FrameH < 1 {
		b := src.Bounds()
		m.FrameW, m.FrameH = b.Dx(), b.Dy()
	}
	if row < 0 {
		row = 0
	}
	oy := row * m.FrameH
	if oy+m.FrameH > src.Bounds().Dy() {
		oy = 0
	}
	frame := image.NewRGBA(image.Rect(0, 0, m.FrameW, m.FrameH))
	for y := 0; y < m.FrameH; y++ {
		for x := 0; x < m.FrameW; x++ {
			frame.Set(x, y, src.At(src.Bounds().Min.X+x, src.Bounds().Min.Y+oy+y))
		}
	}
	sil := terrain.PropSilhouette(frame)
	cel := &propCel{
		img: ebiten.NewImageFromImage(frame),
		ax:  m.AnchorX,
		ay:  m.AnchorY,
		sax: m.AnchorX + terrain.SilhouettePad,
		say: m.AnchorY + terrain.SilhouettePad,
	}
	if sil != nil {
		cel.shadow = ebiten.NewImageFromImage(sil)
	}
	return cel, nil
}

func (e *Editor) drawProps(dst *ebiten.Image) {
	if dst == nil || e.err != nil {
		return
	}
	s := e.scale
	if s <= 0 {
		s = 1
	}
	for ty := 0; ty < e.grid.H; ty++ {
		for tx := 0; tx < e.grid.W; tx++ {
			if e.grid.At(tx, ty) != sim.TileMine {
				continue
			}
			c := sim.TileCenter(tx, ty)
			bx, by := e.worldToScreen(c.X, c.Y)
			drawBarrel(dst, bx, by, s)
		}
	}
	door, plain, grenade, rocket, barrel := sheets()
	for _, b := range e.buildings {
		tw, th := b.W, b.H
		if tw < 1 {
			tw = 2
		}
		if th < 1 {
			th = 2
		}
		footX := float64(b.X*sim.TileSize) + float64(tw*sim.TileSize)/2
		footY := float64((b.Y + th) * sim.TileSize)
		sx, sy := e.worldToScreen(footX, footY)
		cel := plain
		if b.Kind == "tent" {
			cel = tentCel()
		} else if b.Kind == "handover" {
			cel = handoverCel()
		} else if b.Kind == "home" {
			cel = homeCel()
		} else if b.Kind == "reinforced" {
			cel = reinforcedCel()
		} else if b.Kind == "factory" {
			cel = factoryCel()
		} else if b.Kind == "computer" {
			cel = computerCel()
		} else if b.Door {
			cel = door
		}
		if cel == nil || cel.img == nil {
			drawHutFallback(dst, b, s, e)
			continue
		}
		drawPropShadow(dst, cel, sx, sy, s)
		drawCel(dst, cel, sx, sy, s)
	}
	for _, b := range e.barrels {
		c := sim.TileCenter(b.X, b.Y)
		bx, by := e.worldToScreen(c.X, c.Y)
		if barrel == nil || barrel.img == nil {
			drawBarrel(dst, bx, by, s)
			continue
		}
		drawPropShadow(dst, barrel, bx, by, s)
		drawCel(dst, barrel, bx, by, s)
	}
	civil := civilianCel()
	for _, c := range e.civilians {
		p := sim.TileCenter(c.X, c.Y)
		sx, sy := e.worldToScreen(p.X, p.Y)
		if civil == nil || civil.img == nil {
			fillRect(dst, int(sx-4*s), int(sy-8*s), int(8*s), int(10*s), color.RGBA{R: 0xe6, G: 0xd2, B: 0x3a, A: 0xff})
			continue
		}
		drawPropShadow(dst, civil, sx, sy, s)
		drawCel(dst, civil, sx, sy, s)
	}
	host := hostageCel()
	for _, h := range e.hostages {
		p := sim.TileCenter(h.X, h.Y)
		sx, sy := e.worldToScreen(p.X, p.Y)
		if host == nil || host.img == nil {
			fillRect(dst, int(sx-4*s), int(sy-8*s), int(8*s), int(10*s), color.RGBA{R: 0x3a, G: 0x6e, B: 0xc8, A: 0xff})
			continue
		}
		drawPropShadow(dst, host, sx, sy, s)
		drawCel(dst, host, sx, sy, s)
	}
	for _, p := range e.pickups {
		c := sim.TileCenter(p.X, p.Y)
		sx, sy := e.worldToScreen(c.X, c.Y)
		cel := grenade
		if p.Kind == "rocket" || p.Kind == "rockets" {
			cel = rocket
		}
		if cel == nil || cel.img == nil {
			fillRect(dst, int(sx-6*s), int(sy-6*s), int(12*s), int(12*s), color.RGBA{R: 0x8a, G: 0x8a, B: 0x8a, A: 0xff})
			continue
		}
		drawPropShadow(dst, cel, sx, sy, s)
		drawCel(dst, cel, sx, sy, s)
	}
}

// drawPropShadow is the same south-east silhouette the battle paints under
// a hut or a crate. x and y are the sheet anchor in screen pixels.
func drawPropShadow(dst *ebiten.Image, cel *propCel, x, y, scale float64) {
	if dst == nil || cel == nil || cel.shadow == nil || scale <= 0 {
		return
	}
	sc := scale / artScale
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(-float64(cel.sax), -float64(cel.say))
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x+terrain.PropShiftX*scale, y+terrain.PropShiftY*scale)
	dst.DrawImage(cel.shadow, op)
}

func (e *Editor) worldToScreen(x, y float64) (float64, float64) {
	s := e.scale
	if s <= 0 {
		s = 1
	}
	return (-e.scrollX + x) * s, (-e.scrollY + y) * s
}

func drawCel(dst *ebiten.Image, cel *propCel, x, y, scale float64) bool {
	if cel == nil || cel.img == nil {
		return false
	}
	sc := scale / artScale
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(-float64(cel.ax), -float64(cel.ay))
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x, y)
	dst.DrawImage(cel.img, op)
	return true
}

func drawHutFallback(dst *ebiten.Image, b data.BuildingSpec, s float64, e *Editor) {
	tw, th := b.W, b.H
	if tw < 1 {
		tw = 2
	}
	if th < 1 {
		th = 2
	}
	x, y := e.worldToScreen(float64(b.X*sim.TileSize), float64(b.Y*sim.TileSize))
	col := color.RGBA{R: 0x6a, G: 0x42, B: 0x28, A: 0xff}
	if b.Kind == "tent" {
		col = color.RGBA{R: 0xe8, G: 0xdc, B: 0xc0, A: 0xff}
	}
	if b.Kind == "handover" {
		col = color.RGBA{R: 0xc4, G: 0xb4, B: 0x88, A: 0xff}
	}
	if b.Kind == "home" {
		col = color.RGBA{R: 0xf0, G: 0xe6, B: 0xc8, A: 0xff}
	}
	if b.Kind == "reinforced" {
		col = color.RGBA{R: 0x6a, G: 0x72, B: 0x78, A: 0xff}
	}
	if b.Kind == "factory" {
		col = color.RGBA{R: 0x8a, G: 0x6a, B: 0x3a, A: 0xff}
	}
	if b.Kind == "computer" {
		col = color.RGBA{R: 0x3a, G: 0xc0, B: 0x88, A: 0xff}
	}
	bw := int(float64(tw*sim.TileSize) * s)
	bh := int(float64(th*sim.TileSize) * s)
	fillRect(dst, int(x), int(y), bw, bh, col)
	if b.Kind == "tent" && bw > 4 && bh > 4 {
		fillRect(dst, int(x)+bw/2-1, int(y)+bh/2-bh/4, 2, bh/2, color.RGBA{R: 0xc8, G: 0x28, B: 0x28, A: 0xff})
		fillRect(dst, int(x)+bw/2-bw/5, int(y)+bh/2-1, bw/3, 2, color.RGBA{R: 0xc8, G: 0x28, B: 0x28, A: 0xff})
	}
}

// drawBarrel is a red drum. Mine cells and exploding barrels both use it.
func drawBarrel(dst *ebiten.Image, x, y, s float64) {
	w, h := 5*s, 8*s
	if w < 3 {
		w = 3
	}
	if h < 5 {
		h = 5
	}
	body := &vector.Path{}
	body.MoveTo(float32(x-w/2), float32(y-h/2+w*0.25))
	body.LineTo(float32(x+w/2), float32(y-h/2+w*0.25))
	body.LineTo(float32(x+w/2), float32(y+h/2-w*0.15))
	body.LineTo(float32(x-w/2), float32(y+h/2-w*0.15))
	body.Close()
	fill := &vector.DrawPathOptions{AntiAlias: true}
	fill.ColorScale.ScaleWithColor(color.RGBA{R: 0xc4, G: 0x28, B: 0x1c, A: 0xff})
	vector.FillPath(dst, body, nil, fill)
	band := &vector.Path{}
	by := y - h*0.05
	band.MoveTo(float32(x-w/2), float32(by-s))
	band.LineTo(float32(x+w/2), float32(by-s))
	band.LineTo(float32(x+w/2), float32(by+s))
	band.LineTo(float32(x-w/2), float32(by+s))
	band.Close()
	bf := &vector.DrawPathOptions{AntiAlias: true}
	bf.ColorScale.ScaleWithColor(color.RGBA{R: 0xf0, G: 0xc4, B: 0x30, A: 0xff})
	vector.FillPath(dst, band, nil, bf)
	top := &vector.Path{}
	top.Arc(float32(x), float32(y-h/2+w*0.25), float32(w/2), 0, 6.3, vector.Clockwise)
	tf := &vector.DrawPathOptions{AntiAlias: true}
	tf.ColorScale.ScaleWithColor(color.RGBA{R: 0x8e, G: 0x18, B: 0x12, A: 0xff})
	vector.FillPath(dst, top, nil, tf)
}

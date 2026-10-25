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
	once                         sync.Once
	door, plain, grenade, rocket, barrel *propCel
}

func sheets() (door, plain, grenade, rocket, barrel *propCel) {
	propLib.once.Do(func() {
		propLib.door, _ = loadCel("hut/door")
		propLib.plain, _ = loadCel("hut/plain")
		propLib.grenade, _ = loadCel("crate/grenade")
		propLib.rocket, _ = loadCel("crate/rocket")
		propLib.barrel, _ = loadCel("barrel/drum")
	})
	return propLib.door, propLib.plain, propLib.grenade, propLib.rocket, propLib.barrel
}

type celManifest struct {
	FrameW  int `json:"frameW"`
	FrameH  int `json:"frameH"`
	AnchorX int `json:"anchorX"`
	AnchorY int `json:"anchorY"`
}

func loadCel(key string) (*propCel, error) {
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
	frame := image.NewRGBA(image.Rect(0, 0, m.FrameW, m.FrameH))
	for y := 0; y < m.FrameH; y++ {
		for x := 0; x < m.FrameW; x++ {
			frame.Set(x, y, src.At(src.Bounds().Min.X+x, src.Bounds().Min.Y+y))
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
		if b.Door {
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
	fillRect(dst, int(x), int(y), int(float64(tw*sim.TileSize)*s), int(float64(th*sim.TileSize)*s), color.RGBA{R: 0x6a, G: 0x42, B: 0x28, A: 0xff})
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

package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

var (
	overviewBG    = color.RGBA{R: 0x10, G: 0x14, B: 0x10, A: 0xff}
	overviewGrass = color.RGBA{R: 0x1c, G: 0x4a, B: 0x1c, A: 0xff}
	overviewView  = color.RGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff}
	overviewHut   = color.RGBA{R: 0x6a, G: 0x42, B: 0x28, A: 0xff}
	overviewCrate = color.RGBA{R: 0xd0, G: 0xd0, B: 0xd0, A: 0xff}
)

// Overview draws a fog-free schematic of the whole map on the playfield.
func Overview(dst *ebiten.Image, w *sim.World) {
	if dst == nil || w == nil {
		return
	}
	const (
		screenW = 320
		screenH = 256
		pad     = 4
	)
	fillRect(dst, HUDWidth, 0, screenW-HUDWidth, screenH, overviewBG)
	mw, mh := w.Map.W, w.Map.H
	pw, ph := w.Map.PixelSize()
	if mw <= 0 || mh <= 0 {
		mw, mh = 20, 16
	}
	availW := screenW - HUDWidth - pad*2
	availH := screenH - pad*2
	scale := availW / mw
	if s := availH / mh; s < scale {
		scale = s
	}
	if scale < 1 {
		scale = 1
	}
	drawW := scale * mw
	drawH := scale * mh
	ox := HUDWidth + pad + (availW-drawW)/2
	oy := pad + (availH-drawH)/2
	for ty := 0; ty < mh; ty++ {
		for tx := 0; tx < mw; tx++ {
			tile := sim.TileGrass
			if w.Map.W > 0 {
				tile = w.Map.At(tx, ty)
			}
			fillRect(dst, ox+tx*scale, oy+ty*scale, scale, scale, overviewTile(tile))
		}
	}
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if !b.Alive {
			continue
		}
		x := ox + int(b.X/pw*float64(drawW))
		y := oy + int(b.Y/ph*float64(drawH))
		bw := int(b.W / pw * float64(drawW))
		bh := int(b.H / ph * float64(drawH))
		if bw < 2 {
			bw = 2
		}
		if bh < 2 {
			bh = 2
		}
		fillRect(dst, x, y, bw, bh, overviewHut)
	}
	for i := range w.Pickups {
		p := &w.Pickups[i]
		if !p.Alive {
			continue
		}
		fillRect(dst,
			ox+int(p.X/pw*float64(drawW))-1,
			oy+int(p.Y/ph*float64(drawH))-1,
			3, 3, overviewCrate)
	}
	dot := scale / 2
	if dot < 2 {
		dot = 2
	}
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Living() {
			continue
		}
		fillRect(dst,
			ox+int(u.X/pw*float64(drawW))-dot/2,
			oy+int(u.Y/ph*float64(drawH))-dot/2,
			dot, dot, overviewUnit(u))
	}
	strokeView(dst, ox, oy, drawW, drawH, pw, ph, w.Camera)
}

func overviewTile(t sim.Tile) color.RGBA {
	switch t {
	case sim.TileTree:
		return treeFill
	case sim.TileWaterShallow:
		return shallowFill
	case sim.TileWaterDeep:
		return deepFill
	case sim.TileBridge:
		return bridgeFill
	case sim.TileIce:
		return color.RGBA{R: 0xd8, G: 0xe4, B: 0xea, A: 0xff}
	case sim.TileCliff:
		return color.RGBA{R: 0x6a, G: 0x6a, B: 0x70, A: 0xff}
	default:
		return overviewGrass
	}
}

func overviewUnit(u *sim.Unit) color.RGBA {
	if u.Side == sim.SideEnemy {
		return color.RGBA{R: 0xe0, G: 0x30, B: 0x30, A: 0xff}
	}
	if u.Side == sim.SideCivilian {
		return color.RGBA{R: 0xe0, G: 0xe0, B: 0x40, A: 0xff}
	}
	return squadFill(u.SquadID, false)
}

func strokeView(dst *ebiten.Image, ox, oy, drawW, drawH int, pw, ph float64, cam sim.Camera) {
	if pw <= 0 || ph <= 0 {
		return
	}
	vw, vh := cam.ViewW, cam.ViewH
	if vw <= 0 {
		vw = 320
	}
	if vh <= 0 {
		vh = 256
	}
	x := ox + int(cam.X/pw*float64(drawW))
	y := oy + int(cam.Y/ph*float64(drawH))
	w := int(vw / pw * float64(drawW))
	h := int(vh / ph * float64(drawH))
	if w < 2 {
		w = 2
	}
	if h < 2 {
		h = 2
	}
	fillRect(dst, x, y, w, 1, overviewView)
	fillRect(dst, x, y+h-1, w, 1, overviewView)
	fillRect(dst, x, y, 1, h, overviewView)
	fillRect(dst, x+w-1, y, 1, h, overviewView)
}

package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
	ground "fannon-codder/internal/terrain"
)

var (
	overviewBG          = color.RGBA{R: 0x10, G: 0x14, B: 0x10, A: 0xff}
	overviewGrass       = color.RGBA{R: 44, G: 81, B: 64, A: 0xff}
	overviewForest      = color.RGBA{R: 28, G: 72, B: 40, A: 0xff}
	overviewDesert      = color.RGBA{R: 196, G: 164, B: 96, A: 0xff}
	overviewSnow        = color.RGBA{R: 181, G: 199, B: 215, A: 0xff}
	overviewView        = color.RGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff}
	overviewHut         = color.RGBA{R: 80, G: 62, B: 52, A: 0xff}
	overviewCrate       = color.RGBA{R: 0xd0, G: 0xd0, B: 0xd0, A: 0xff}
	overviewShallow     = color.RGBA{R: 93, G: 156, B: 198, A: 0xff}
	overviewDeep        = color.RGBA{R: 39, G: 69, B: 153, A: 0xff}
	overviewForestWater = color.RGBA{R: 42, G: 108, B: 78, A: 0xff}
	overviewForestDeep  = color.RGBA{R: 16, G: 58, B: 52, A: 0xff}
	overviewDesertWater = color.RGBA{R: 150, G: 196, B: 184, A: 0xff}
	overviewDesertDeep  = color.RGBA{R: 46, G: 118, B: 138, A: 0xff}
	overviewSnowWater   = color.RGBA{R: 150, G: 176, B: 190, A: 0xff}
	overviewSnowDeep    = color.RGBA{R: 70, G: 108, B: 140, A: 0xff}
	overviewSand        = color.RGBA{R: 197, G: 162, B: 74, A: 0xff}
	overviewIce         = color.RGBA{R: 196, G: 214, B: 226, A: 0xff}
	overviewBridge      = color.RGBA{R: 140, G: 108, B: 88, A: 0xff}
	overviewCliff       = color.RGBA{R: 133, G: 136, B: 148, A: 0xff}
	overviewRamp        = color.RGBA{R: 168, G: 169, B: 177, A: 0xff}
	overviewTree        = color.RGBA{R: 66, G: 81, B: 58, A: 0xff}
	overviewForestTree  = color.RGBA{R: 24, G: 64, B: 48, A: 0xff}
	overviewDesertTree  = color.RGBA{R: 122, G: 102, B: 48, A: 0xff}
	overviewMine        = color.RGBA{R: 48, G: 82, B: 64, A: 0xff}
)

// Overview draws a fog-free schematic of the whole map on the playfield.
// terrain is jungle, forest, desert, or snow. Open ground and trees use
// that kind's colour. Snow grass and snow trees stay the snow colour.
func Overview(dst *ebiten.Image, w *sim.World, terrain string) {
	if dst == nil || w == nil {
		return
	}
	const pad = 4
	s := PictureScale()
	if s <= 0 {
		s = 1
	}
	b := dst.Bounds()
	screenW := int(math.Round(float64(b.Dx()) / s))
	screenH := int(math.Round(float64(b.Dy()) / s))
	if screenW < 320 {
		screenW = 320
	}
	if screenH < 256 {
		screenH = 256
	}
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
			fillRect(dst, ox+tx*scale, oy+ty*scale, scale, scale, overviewTile(tile, terrain))
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
		if u.Dead() || u.VehicleID != 0 {
			continue
		}
		fillRect(dst,
			ox+int(u.X/pw*float64(drawW))-dot/2,
			oy+int(u.Y/ph*float64(drawH))-dot/2,
			dot, dot, overviewUnit(u))
	}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		c := color.RGBA{R: 0xd8, G: 0xd8, B: 0xdc, A: 0xff}
		if v.Side == sim.SideEnemy {
			c = color.RGBA{R: 0xe0, G: 0x30, B: 0x30, A: 0xff}
		}
		fillRect(dst,
			ox+int(v.X/pw*float64(drawW))-dot/2,
			oy+int(v.Y/ph*float64(drawH))-dot/2,
			dot, dot, c)
	}
	strokeView(dst, ox, oy, drawW, drawH, pw, ph, w.Camera)
}

func overviewTile(t sim.Tile, terrain string) color.RGBA {
	kind := ground.KindOf(terrain)
	switch t {
	case sim.TileTree:
		switch kind {
		case ground.Snow:
			return overviewSnow
		case ground.Forest:
			return overviewForestTree
		case ground.Desert:
			return overviewDesertTree
		default:
			return overviewTree
		}
	case sim.TileWaterShallow:
		switch kind {
		case ground.Snow:
			return overviewSnowWater
		case ground.Forest:
			return overviewForestWater
		case ground.Desert:
			return overviewDesertWater
		default:
			return overviewShallow
		}
	case sim.TileWaterDeep:
		switch kind {
		case ground.Snow:
			return overviewSnowDeep
		case ground.Forest:
			return overviewForestDeep
		case ground.Desert:
			return overviewDesertDeep
		default:
			return overviewDeep
		}
	case sim.TileBridge:
		return overviewBridge
	case sim.TileIce:
		return overviewIce
	case sim.TileCliff:
		return overviewCliff
	case sim.TileRamp:
		return overviewRamp
	case sim.TileQuicksand:
		return overviewSand
	case sim.TileMine:
		return overviewMine
	default:
		switch kind {
		case ground.Snow:
			return overviewSnow
		case ground.Forest:
			return overviewForest
		case ground.Desert:
			return overviewDesert
		default:
			return overviewGrass
		}
	}
}

func overviewUnit(u *sim.Unit) color.RGBA {
	if u.Side == sim.SideEnemy {
		if u.Kind == sim.KindGrenadier {
			return color.RGBA{R: 0xe0, G: 0x60, B: 0x18, A: 0xff}
		}
		if u.Kind == sim.KindRocketeer {
			return color.RGBA{R: 0x78, G: 0x18, B: 0x38, A: 0xff}
		}
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

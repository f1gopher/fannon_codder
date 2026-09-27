package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

var (
	treeFill    = color.RGBA{R: 0x1e, G: 0x5a, B: 0x1e, A: 0xff}
	shallowFill = color.RGBA{R: 0x4a, G: 0xa0, B: 0xc8, A: 0xff}
	deepFill    = color.RGBA{R: 0x1a, G: 0x3a, B: 0x88, A: 0xff}
	bridgeFill  = color.RGBA{R: 0x8a, G: 0x70, B: 0x40, A: 0xff}
	iceFill     = color.RGBA{R: 0xd4, G: 0xe6, B: 0xf0, A: 0xff}
	cliffFill   = color.RGBA{R: 0x4a, G: 0x50, B: 0x58, A: 0xff}
	rampFill    = color.RGBA{R: 0x9a, G: 0xa6, B: 0xb0, A: 0xff}
	sandFill    = color.RGBA{R: 0xc2, G: 0xa0, B: 0x4a, A: 0xff}
	mineFill    = color.RGBA{R: 0x6b, G: 0x55, B: 0x32, A: 0xff}
	mineMark    = color.RGBA{R: 0x2a, G: 0x22, B: 0x18, A: 0xff}
	tileSprites = map[sim.Tile]*ebiten.Image{}
)

func tileImage(t sim.Tile) *ebiten.Image {
	if img, ok := tileSprites[t]; ok {
		return img
	}
	var fill color.RGBA
	switch t {
	case sim.TileTree:
		fill = treeFill
	case sim.TileWaterShallow:
		fill = shallowFill
	case sim.TileWaterDeep:
		fill = deepFill
	case sim.TileBridge:
		fill = bridgeFill
	case sim.TileIce:
		fill = iceFill
	case sim.TileCliff:
		fill = cliffFill
	case sim.TileRamp:
		fill = rampFill
	case sim.TileQuicksand:
		fill = sandFill
	case sim.TileMine:
		fill = mineFill
	default:
		return nil
	}
	img := ebiten.NewImage(sim.TileSize, sim.TileSize)
	img.Fill(fill)
	if t == sim.TileMine {
		mark := ebiten.NewImage(6, 6)
		mark.Fill(mineMark)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(5, 5)
		img.DrawImage(mark, op)
	}
	tileSprites[t] = img
	return img
}

// Tiles draws non-grass tiles. Grass is the scene background.
func Tiles(dst *ebiten.Image, m sim.Map, cam sim.Camera) {
	if m.W == 0 {
		return
	}
	tx0, ty0, tx1, ty1 := visibleTiles(m, cam)
	for ty := ty0; ty < ty1; ty++ {
		for tx := tx0; tx < tx1; tx++ {
			img := tileImage(m.At(tx, ty))
			if img == nil {
				continue
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				cam.ScreenX(float64(tx*sim.TileSize)),
				cam.ScreenY(float64(ty*sim.TileSize)),
			)
			dst.DrawImage(img, op)
		}
	}
}

func visibleTiles(m sim.Map, cam sim.Camera) (tx0, ty0, tx1, ty1 int) {
	vw, vh := cam.ViewW, cam.ViewH
	if vw <= 0 {
		vw = 320
	}
	if vh <= 0 {
		vh = 256
	}
	tx0 = int(cam.X) / sim.TileSize
	ty0 = int(cam.Y) / sim.TileSize
	if cam.X < 0 {
		tx0 = -1
	}
	if cam.Y < 0 {
		ty0 = -1
	}
	tx1 = int(cam.X+vw)/sim.TileSize + 2
	ty1 = int(cam.Y+vh)/sim.TileSize + 2
	if tx0 < 0 {
		tx0 = 0
	}
	if ty0 < 0 {
		ty0 = 0
	}
	if tx1 > m.W {
		tx1 = m.W
	}
	if ty1 > m.H {
		ty1 = m.H
	}
	return tx0, ty0, tx1, ty1
}

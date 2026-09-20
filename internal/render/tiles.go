package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

var (
	treeFill   = color.RGBA{R: 0x1e, G: 0x5a, B: 0x1e, A: 0xff}
	treeSprite *ebiten.Image
)

func treeImage() *ebiten.Image {
	if treeSprite == nil {
		treeSprite = ebiten.NewImage(sim.TileSize, sim.TileSize)
		treeSprite.Fill(treeFill)
	}
	return treeSprite
}

// Tiles draws solid tiles (trees). Grass is the scene background.
func Tiles(dst *ebiten.Image, m sim.Map, cam sim.Camera) {
	if m.W == 0 {
		return
	}
	img := treeImage()
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			if m.At(tx, ty) != sim.TileTree {
				continue
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				float64(tx*sim.TileSize)-cam.X,
				float64(ty*sim.TileSize)-cam.Y,
			)
			dst.DrawImage(img, op)
		}
	}
}

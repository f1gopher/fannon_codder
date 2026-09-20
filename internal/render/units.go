package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

var playerFill = color.RGBA{R: 0x40, G: 0xc8, B: 0x40, A: 0xff}

var playerSprite *ebiten.Image

func playerImage() *ebiten.Image {
	if playerSprite == nil {
		playerSprite = ebiten.NewImage(sim.UnitSize, sim.UnitSize)
		playerSprite.Fill(playerFill)
	}
	return playerSprite
}

// Units draws placeholder troopers in screen space (world minus camera).
func Units(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	img := playerImage()
	half := float64(sim.UnitSize) / 2
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != sim.SidePlayer {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-w.Camera.X-half, u.Y-w.Camera.Y-half)
		dst.DrawImage(img, op)
	}
}

package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

var (
	playerFill  = color.RGBA{R: 0x40, G: 0xc8, B: 0x40, A: 0xff}
	enemyFill   = color.RGBA{R: 0xc8, G: 0x30, B: 0x30, A: 0xff}
	corpseFill  = color.RGBA{R: 0x50, G: 0x30, B: 0x30, A: 0xff}
	playerDead  = color.RGBA{R: 0x20, G: 0x50, B: 0x20, A: 0xff}
	tracerFill  = color.RGBA{R: 0xff, G: 0xff, B: 0xa0, A: 0xff}
	spriteCache = map[color.RGBA]*ebiten.Image{}
	tracerSprite *ebiten.Image
)

func unitSprite(c color.RGBA) *ebiten.Image {
	if img, ok := spriteCache[c]; ok {
		return img
	}
	img := ebiten.NewImage(sim.UnitSize, sim.UnitSize)
	img.Fill(c)
	spriteCache[c] = img
	return img
}

func unitColor(u *sim.Unit) color.RGBA {
	if u.Dead() {
		if u.Side == sim.SidePlayer {
			return playerDead
		}
		return corpseFill
	}
	if u.Side == sim.SideEnemy {
		return enemyFill
	}
	return playerFill
}

// Units draws placeholder troopers in screen space (world minus camera).
func Units(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	half := float64(sim.UnitSize) / 2
	for i := range w.Units {
		u := &w.Units[i]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-w.Camera.X-half, u.Y-w.Camera.Y-half)
		dst.DrawImage(unitSprite(unitColor(u)), op)
	}
}

// Projectiles draws MG tracers.
func Projectiles(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	if tracerSprite == nil {
		tracerSprite = ebiten.NewImage(2, 2)
		tracerSprite.Fill(tracerFill)
	}
	for i := range w.Projectiles {
		p := &w.Projectiles[i]
		if !p.Alive {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-w.Camera.X-1, p.Y-w.Camera.Y-1)
		dst.DrawImage(tracerSprite, op)
	}
}

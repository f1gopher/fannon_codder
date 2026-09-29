package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

var (
	hutFill   = color.RGBA{R: 0x6a, G: 0x42, B: 0x28, A: 0xff}
	doorFill  = color.RGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xff}
	crateFill = color.RGBA{R: 0x8a, G: 0x8a, B: 0x8a, A: 0xff}
	bombFill  = color.RGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xff}
	blastFill = color.RGBA{R: 0xe0, G: 0x70, B: 0x20, A: 0xff}
)

var (
	skidooFill  = color.RGBA{R: 0xd0, G: 0xd4, B: 0xd8, A: 0xff}
	vehicleTrim = color.RGBA{R: 0x68, G: 0x70, B: 0x78, A: 0xff}
	enemyLamp   = color.RGBA{R: 0xe0, G: 0x28, B: 0x28, A: 0xff}
)

func drawBuildingRect(dst *ebiten.Image, cam sim.Camera, b sim.Building) {
	x := cam.ScreenX(b.X)
	y := cam.ScreenY(b.Y)
	fillRectF(dst, x, y, b.W, b.H, hutFill)
	if b.HasDoor {
		fillRectF(dst, x+b.W/2-3, y+b.H-8, 6, 6, doorFill)
	}
}

func drawCrateRect(dst *ebiten.Image, cam sim.Camera, p sim.Pickup) {
	x := cam.ScreenX(p.X) - 6
	y := cam.ScreenY(p.Y) - 6
	fillRectF(dst, x, y, 12, 12, crateFill)
	label := "G"
	if p.Kind == sim.PickupRockets {
		label = "R"
	}
	Text(dst, label, x+3, y+2)
}

// drawVehicleRect is the placeholder skidoo. The enemy lamp is separate so a
// sheet keeps the same blink.
func drawVehicleRect(dst *ebiten.Image, cam sim.Camera, v *sim.Vehicle) {
	x := cam.ScreenX(v.X) - 10
	y := cam.ScreenY(v.Y) - 6
	fillRectF(dst, x, y, 20, 12, vehicleTrim)
	fillRectF(dst, x+1, y+1, 18, 10, skidooFill)
}

func drawEnemyLamp(dst *ebiten.Image, cam sim.Camera, v *sim.Vehicle) {
	if v.Side != sim.SideEnemy || int(v.Blink*6)%2 != 0 {
		return
	}
	x := cam.ScreenX(v.X) - 10
	y := cam.ScreenY(v.Y) - 6
	fillRectF(dst, x+2, y+2, 3, 3, enemyLamp)
}

// Bombs draws grenades in the air.
func Bombs(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	cam := w.Camera
	for i := range w.Grenades {
		g := &w.Grenades[i]
		if !g.Alive {
			continue
		}
		fillRectF(dst, cam.ScreenX(g.X)-1, cam.ScreenY(g.Y-g.Height)-1, 3, 3, bombFill)
	}
}

// Blasts draws the explosion flash.
func Blasts(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	cam := w.Camera
	for i := range w.Explosions {
		e := &w.Explosions[i]
		r := e.R
		x := cam.ScreenX(e.X) - r
		y := cam.ScreenY(e.Y) - r
		s := r * 2
		fillRectF(dst, x, y, s, 1, blastFill)
		fillRectF(dst, x, y+s, s, 1, blastFill)
		fillRectF(dst, x, y, 1, s, blastFill)
		fillRectF(dst, x+s, y, 1, s, blastFill)
	}
}

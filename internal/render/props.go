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

// Solids draws huts and grenade crates under the troopers.
func Solids(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	cam := w.Camera
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if !b.Alive {
			continue
		}
		x := cam.ScreenX(b.X)
		y := cam.ScreenY(b.Y)
		fillRectF(dst, x, y, b.W, b.H, hutFill)
		if b.HasDoor {
			fillRectF(dst, x+b.W/2-3, y+b.H-8, 6, 6, doorFill)
		}
	}
	for i := range w.Pickups {
		p := &w.Pickups[i]
		if !p.Alive {
			continue
		}
		x := cam.ScreenX(p.X) - 6
		y := cam.ScreenY(p.Y) - 6
		fillRectF(dst, x, y, 12, 12, crateFill)
		label := "G"
		if p.Kind == sim.PickupRockets {
			label = "R"
		}
		Text(dst, label, x+3, y+2)
	}
}

// Vehicles draws skidoos. Enemy ones blink a red lamp.
func Vehicles(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	cam := w.Camera
	body := color.RGBA{R: 0xd0, G: 0xd4, B: 0xd8, A: 0xff}
	trim := color.RGBA{R: 0x68, G: 0x70, B: 0x78, A: 0xff}
	lamp := color.RGBA{R: 0xe0, G: 0x28, B: 0x28, A: 0xff}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		x := cam.ScreenX(v.X) - 10
		y := cam.ScreenY(v.Y) - 6
		fillRectF(dst, x, y, 20, 12, trim)
		fillRectF(dst, x+1, y+1, 18, 10, body)
		if v.Side == sim.SideEnemy && int(v.Blink*6)%2 == 0 {
			fillRectF(dst, x+2, y+2, 3, 3, lamp)
		}
	}
}

// Grenades draws bombs in flight and the blast flash.
func Grenades(dst *ebiten.Image, w *sim.World) {
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

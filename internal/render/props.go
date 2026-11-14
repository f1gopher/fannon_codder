package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"fannon-codder/internal/sim"
)

var (
	hutFill      = color.RGBA{R: 0x6a, G: 0x42, B: 0x28, A: 0xff}
	steelFill    = color.RGBA{R: 0x6a, G: 0x72, B: 0x78, A: 0xff}
	tentFill     = color.RGBA{R: 0xe8, G: 0xdc, B: 0xc0, A: 0xff}
	handoverFill = color.RGBA{R: 0xc4, G: 0xb4, B: 0x88, A: 0xff}
	homeFill     = color.RGBA{R: 0xf0, G: 0xe6, B: 0xc8, A: 0xff}
	homeDoor     = color.RGBA{R: 0xf0, G: 0xc0, B: 0x40, A: 0xff}
	crossFill    = color.RGBA{R: 0xc8, G: 0x28, B: 0x28, A: 0xff}
	doorFill     = color.RGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xff}
	crateFill    = color.RGBA{R: 0x8a, G: 0x8a, B: 0x8a, A: 0xff}
	bombFill     = color.RGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xff}
	blastFill    = color.RGBA{R: 0xe0, G: 0x70, B: 0x20, A: 0xff}
)

var (
	skidooFill  = color.RGBA{R: 0xd0, G: 0xd4, B: 0xd8, A: 0xff}
	vehicleTrim = color.RGBA{R: 0x68, G: 0x70, B: 0x78, A: 0xff}
	enemyLamp   = color.RGBA{R: 0xe0, G: 0x28, B: 0x28, A: 0xff}
)

// blastRectHook counts the missing-sheet blast outline. Tests use it.
var blastRectHook func()

func drawBuildingRect(dst *ebiten.Image, cam sim.Camera, b sim.Building) {
	x := cam.ScreenX(b.X)
	y := cam.ScreenY(b.Y)
	fill := hutFill
	if b.Kind == sim.HutTent {
		fill = tentFill
	}
	if b.Kind == sim.HutHandover {
		fill = handoverFill
	}
	if b.Kind == sim.HutHome {
		fill = homeFill
	}
	if b.Kind == sim.HutReinforced {
		fill = steelFill
	}
	if b.Kind == sim.HutFactory {
		fill = color.RGBA{R: 0x8a, G: 0x6a, B: 0x3a, A: 0xff}
	}
	if b.Kind == sim.HutComputer {
		fill = color.RGBA{R: 0x3a, G: 0xc0, B: 0x88, A: 0xff}
	}
	fillRectF(dst, x, y, b.W, b.H, fill)
	if b.Kind == sim.HutTent {
		fillRectF(dst, x+b.W/2-1, y+b.H/2-6, 2, 12, crossFill)
		fillRectF(dst, x+b.W/2-6, y+b.H/2-1, 12, 2, crossFill)
	}
	if b.Kind == sim.HutHome {
		fillRectF(dst, x+b.W/2-3, y+b.H-8, 6, 8, homeDoor)
	}
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

// lampSeat is the roof beacon in world pixels, aft of the hull centre so it
// sits on the body. rad is the glass radius.
func lampSeat(v *sim.Vehicle) (dx, dy, rad float64) {
	rad = 1.35
	aft := 1.5
	switch v.Kind {
	case sim.KindTank:
		rad = 2.05
		aft = 5
	case sim.KindJeep:
		rad = 1.65
		aft = 3
	case sim.KindTurret:
		rad = 1.45
		aft = 0
	}
	dx = -math.Cos(v.Facing) * aft
	dy = -math.Sin(v.Facing) * aft
	return dx, dy, rad
}

func drawEnemyLamp(dst *ebiten.Image, cam sim.Camera, v *sim.Vehicle) {
	if dst == nil || v == nil || v.Side != sim.SideEnemy || !v.Alive {
		return
	}
	dx, dy, rad := lampSeat(v)
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	cx := float32((cam.ScreenX(v.X) + dx) * ps)
	cy := float32((cam.ScreenY(v.Y) + dy) * ps)
	on := int(v.Blink*6)%2 == 0
	r := float32(rad * ps)
	if on {
		vector.FillCircle(dst, cx, cy, r+float32(1.35*ps), color.RGBA{R: 0xe0, G: 0x18, B: 0x10, A: 0x48}, true)
	}
	vector.FillCircle(dst, cx, cy, r+float32(0.55*ps), color.RGBA{R: 0x24, G: 0x12, B: 0x10, A: 0xff}, true)
	glass := color.RGBA{R: 0x4a, G: 0x12, B: 0x0e, A: 0xff}
	if on {
		glass = enemyLamp
	}
	vector.FillCircle(dst, cx, cy, r, glass, true)
	if on {
		vector.FillCircle(dst, cx-r*0.28, cy-r*0.28, r*0.32, color.RGBA{R: 0xff, G: 0xe4, B: 0xc8, A: 0xff}, true)
	}
}

// Bombs draws grenades in the air.
func Bombs(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	cam := w.Camera
	sh := activeSheets().Get("fx/grenade")
	for i := range w.Grenades {
		g := &w.Grenades[i]
		if !g.Alive {
			continue
		}
		x := cam.ScreenX(g.X)
		y := cam.ScreenY(g.Y - g.Height)
		if sh == nil || !drawStill(dst, sh, FrameAt(animTime, sh.FPS, sh.Frames, sh.Loop), x, y, 0, false) {
			fillRectF(dst, x-1, y-1, 3, 3, bombFill)
		}
	}
}

// Blasts draws the explosion flash.
func Blasts(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	cam := w.Camera
	noteGrenadeBooms(w)
	drawBoomLayer(dst, cam, true)
	sh := activeSheets().Get("fx/blast")
	sheetDur := 0.0
	if sh != nil && sh.FPS > 0 {
		sheetDur = float64(sh.Frames) / sh.FPS
	}
	for i := range w.Explosions {
		e := &w.Explosions[i]
		if sh != nil && sheetDur > 0 {
			life := e.Life
			if life <= 0 {
				life = sheetDur
			}
			// A plain marker outlives the painting. A shed blast keeps the
			// sheet going for Life, looping when asked.
			if e.Life <= 0 && e.Age >= sheetDur {
				continue
			}
			if e.Age >= life {
				continue
			}
			rate := e.Rate
			if rate <= 0 {
				rate = 1
			}
			sx := cam.ScreenX(e.X)
			sy := cam.ScreenY(e.Y - e.Age*e.Lift)
			frame := FrameAt(e.Age*rate, sh.FPS, sh.Frames, e.Loop)
			scale := e.Scale
			if scale <= 0 {
				scale = 1
			}
			if drawBlast(dst, sh, frame, sx, sy, scale) {
				if scale > 1 {
					// A second, smaller boil so the shed fireball reads thicker.
					inner := FrameAt(e.Age*rate+0.08, sh.FPS, sh.Frames, true)
					drawBlast(dst, sh, inner, sx, sy-6, scale*0.62)
				}
				continue
			}
		}
		r := e.R
		x := cam.ScreenX(e.X) - r
		y := cam.ScreenY(e.Y) - r
		s := r * 2
		if blastRectHook != nil {
			blastRectHook()
		}
		fillRectF(dst, x, y, s, 1, blastFill)
		fillRectF(dst, x, y+s, s, 1, blastFill)
		fillRectF(dst, x, y, 1, s, blastFill)
		fillRectF(dst, x+s, y, 1, s, blastFill)
	}
	drawBoomLayer(dst, cam, false)
}

// drawBlast paints one explosion frame. scale is 1 for a grenade.
func drawBlast(dst *ebiten.Image, sh *Sheet, frame int, x, y, scale float64) bool {
	if sh == nil || scale <= 0 {
		return false
	}
	img, ok := sh.still(frame)
	if !ok || img == nil {
		return false
	}
	if spriteHook != nil {
		spriteHook(img, false)
	}
	sc := SpriteScale() * scale
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(-float64(sh.AnchorX), -float64(sh.AnchorY))
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x*ps, y*ps)
	dst.DrawImage(img, op)
	return true
}

// drawStill paints one stored frame. spin rotates the cel clockwise so a
// right-facing painting follows a shot.
func drawStill(dst *ebiten.Image, sh *Sheet, frame int, x, y, angle float64, spin bool) bool {
	if sh == nil {
		return false
	}
	img, ok := sh.still(frame)
	if !ok || img == nil {
		return false
	}
	if spin {
		DrawSpriteAngle(dst, img, sh.AnchorX, sh.AnchorY, x, y, angle)
	} else {
		DrawSprite(dst, img, sh.AnchorX, sh.AnchorY, false, x, y)
	}
	return true
}

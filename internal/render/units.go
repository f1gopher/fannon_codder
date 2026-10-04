package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

var (
	playerFill    = color.RGBA{R: 0x40, G: 0xc8, B: 0x40, A: 0xff}
	enemyFill     = color.RGBA{R: 0xc8, G: 0x30, B: 0x30, A: 0xff}
	playerSwim    = color.RGBA{R: 0x28, G: 0x88, B: 0x70, A: 0xff}
	enemySwim     = color.RGBA{R: 0x88, G: 0x28, B: 0x58, A: 0xff}
	corpseFill    = color.RGBA{R: 0x50, G: 0x30, B: 0x30, A: 0xff}
	playerDead    = color.RGBA{R: 0x20, G: 0x50, B: 0x20, A: 0xff}
	civilianFill  = color.RGBA{R: 0xe6, G: 0xd2, B: 0x3a, A: 0xff}
	civilianDead  = color.RGBA{R: 0x6a, G: 0x5a, B: 0x20, A: 0xff}
	grenadierFill = color.RGBA{R: 0xe0, G: 0x60, B: 0x18, A: 0xff}
	rocketeerFill = color.RGBA{R: 0x78, G: 0x18, B: 0x38, A: 0xff}
	grenadeWindup = color.RGBA{R: 0xff, G: 0xd0, B: 0x30, A: 0xff}
	rocketShot    = color.RGBA{R: 0xf0, G: 0x78, B: 0x20, A: 0xff}
	tracerFill    = color.RGBA{R: 0xff, G: 0xff, B: 0xa0, A: 0xff}
	facingFill    = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	spriteCache   = map[color.RGBA]*ebiten.Image{}
	tracerSprite  *ebiten.Image
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
	if u.Dead() || u.Wounded() {
		switch u.Side {
		case sim.SidePlayer:
			return playerDead
		case sim.SideCivilian:
			return civilianDead
		default:
			return corpseFill
		}
	}
	if u.Side == sim.SideCivilian {
		return civilianFill
	}
	if u.Side == sim.SideEnemy {
		if u.Kind == sim.KindGrenadier {
			return grenadierFill
		}
		if u.Kind == sim.KindRocketeer {
			return rocketeerFill
		}
		if u.InWater {
			return enemySwim
		}
		return enemyFill
	}
	return squadFill(u.SquadID, u.InWater)
}

// unitPoseScale shrinks a sinking trooper into the pool. Everyone else is full size.
func unitPoseScale(u *sim.Unit) (scale, drop float64) {
	if !u.Sinking || sim.SinkTime <= 0 {
		return 1, 0
	}
	p := u.Sink / sim.SinkTime
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	return 1 - 0.7*p, p * 4
}

func squadFill(id sim.SquadID, swim bool) color.RGBA {
	switch id {
	case sim.SquadEagle:
		if swim {
			return color.RGBA{R: 0x28, G: 0x48, B: 0x88, A: 0xff}
		}
		return color.RGBA{R: 0x48, G: 0x88, B: 0xe8, A: 0xff}
	case sim.SquadPanther:
		if swim {
			return color.RGBA{R: 0x88, G: 0x58, B: 0x20, A: 0xff}
		}
		return color.RGBA{R: 0xe0, G: 0xa0, B: 0x30, A: 0xff}
	default:
		if swim {
			return playerSwim
		}
		return playerFill
	}
}

// rectHook counts placeholder troopers. Tests use it with spriteHook.
var rectHook func()

// drawUnitRect is the placeholder trooper, used when no sheet matches.
func drawUnitRect(dst *ebiten.Image, cam sim.Camera, u *sim.Unit) {
	if rectHook != nil {
		rectHook()
	}
	half := float64(sim.UnitSize) / 2
	pose, drop := unitPoseScale(u)
	dx, dy := bodyShift(u)
	blit(dst, unitSprite(unitColor(u)),
		cam.ScreenX(u.X)+dx-half*pose,
		cam.ScreenY(u.Y)+dy-half*pose+drop,
		pose, pose,
	)
	if u.Living() {
		// Nose so a grunt's turn is visible before he fires.
		nx, ny := math.Cos(u.Facing), math.Sin(u.Facing)
		fillRectF(dst, cam.ScreenX(u.X)+nx*5, cam.ScreenY(u.Y)+ny*5, 2, 2, facingFill)
	}
	if u.Living() && (u.GrenadeWind > 0 || u.RocketWind > 0) {
		fillRectF(dst, cam.ScreenX(u.X)-2, cam.ScreenY(u.Y)-7, 4, 2, grenadeWindup)
	}
}

// Muzzles draws the shot flash while SinceShot is inside the 0.12s window.
// A man on foot flashes at his muzzle. A mounted gun flashes on the vehicle,
// because the troopers are hidden while they ride.
func Muzzles(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	sh := activeSheets().Get("fx/flash")
	if sh == nil {
		return
	}
	cam := w.Camera
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Living() || u.VehicleID != 0 || u.SinceShot >= shotWindow {
			continue
		}
		nx, ny := math.Cos(u.Facing), math.Sin(u.Facing)
		x := cam.ScreenX(u.X + nx*float64(sim.UnitSize)/2)
		y := cam.ScreenY(u.Y + ny*float64(sim.UnitSize)/2)
		drawStill(dst, sh, FrameAt(u.SinceShot, sh.FPS, sh.Frames, false), x, y, 0, false)
	}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		age := shotWindow
		for _, id := range v.Occupants {
			u := w.Unit(id)
			if u == nil || !u.Living() || u.SinceShot >= age {
				continue
			}
			age = u.SinceShot
		}
		if age >= shotWindow {
			continue
		}
		nx, ny := math.Cos(v.Facing), math.Sin(v.Facing)
		x := cam.ScreenX(v.X + nx*10)
		y := cam.ScreenY(v.Y + ny*10)
		drawStill(dst, sh, FrameAt(age, sh.FPS, sh.Frames, false), x, y, 0, false)
	}
}

// Projectiles draws MG tracers and rockets.
func Projectiles(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	if tracerSprite == nil {
		tracerSprite = ebiten.NewImage(2, 2)
		tracerSprite.Fill(tracerFill)
	}
	tracer := activeSheets().Get("fx/tracer")
	rocket := activeSheets().Get("fx/rocket")
	for i := range w.Projectiles {
		p := &w.Projectiles[i]
		if !p.Alive {
			continue
		}
		ang := math.Atan2(p.VY, p.VX)
		x := w.Camera.ScreenX(p.X)
		y := w.Camera.ScreenY(p.Y)
		if p.Kind == sim.ProjRocket {
			if rocket != nil && drawStill(dst, rocket, 0, x, y, ang, true) {
				continue
			}
			fillRectF(dst, x-2, y-1, 5, 3, rocketShot)
			continue
		}
		if tracer != nil && drawStill(dst, tracer, 0, x, y, ang, true) {
			continue
		}
		blit(dst, tracerSprite, x-1, y-1, 1, 1)
	}
}

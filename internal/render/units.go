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
	if u.Dead() {
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

// unitPose shrinks a sinking trooper into the pool. Everyone else is full size.
func unitPose(u *sim.Unit) (scale, drop float64) {
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

// Units draws placeholder troopers in screen space.
func Units(dst *ebiten.Image, w *sim.World) {
	if w == nil {
		return
	}
	half := float64(sim.UnitSize) / 2
	for i := range w.Units {
		u := &w.Units[i]
		if u.VehicleID != 0 {
			continue
		}
		pose, drop := unitPose(u)
		blit(dst, unitSprite(unitColor(u)),
			w.Camera.ScreenX(u.X)-half*pose,
			w.Camera.ScreenY(u.Y)-half*pose+drop,
			pose, pose,
		)
		if u.Living() {
			// Nose so a grunt's turn is visible before he fires.
			nx, ny := math.Cos(u.Facing), math.Sin(u.Facing)
			fillRectF(dst, w.Camera.ScreenX(u.X)+nx*5, w.Camera.ScreenY(u.Y)+ny*5, 2, 2, facingFill)
		}
		if u.Living() && (u.GrenadeWind > 0 || u.RocketWind > 0) {
			fillRectF(dst, w.Camera.ScreenX(u.X)-2, w.Camera.ScreenY(u.Y)-7, 4, 2, grenadeWindup)
		}
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
		if p.Kind == sim.ProjRocket {
			fillRectF(dst, w.Camera.ScreenX(p.X)-2, w.Camera.ScreenY(p.Y)-1, 5, 3, rocketShot)
			continue
		}
		blit(dst, tracerSprite, w.Camera.ScreenX(p.X)-1, w.Camera.ScreenY(p.Y)-1, 1, 1)
	}
}

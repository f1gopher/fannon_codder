package render

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

func chopperShadow(cam sim.Camera, v sim.Vehicle) body {
	sx := cam.ScreenX(v.X)
	sy := cam.ScreenY(v.Y)
	sh := activeSheets().Get(chopperSheet(v))
	var img *ebiten.Image
	if sh != nil {
		img, _, _ = sh.image(Direction(v.Facing), 0)
	}
	return body{
		y: v.Y,
		shadow: func(dst *ebiten.Image) {
			if img != nil {
				castShadow(dst, sx, sy, img)
			}
		},
	}
}

func drawChoppers(dst *ebiten.Image, w *sim.World) {
	if dst == nil || w == nil {
		return
	}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if v.Kind != sim.KindChopper {
			continue
		}
		drawOneChopper(dst, w.Camera, v)
	}
}

func chopperSheet(v sim.Vehicle) string {
	if !v.Alive {
		return "chopper/wreck"
	}
	if v.Armed {
		return "chopper/attack"
	}
	return "chopper/plain"
}

func drawOneChopper(dst *ebiten.Image, cam sim.Camera, v *sim.Vehicle) {
	sx := cam.ScreenX(v.X)
	sy := cam.ScreenY(v.Y) - v.Alt
	key := chopperSheet(*v)
	sh := activeSheets().Get(key)
	if sh != nil {
		img, mirror, ok := sh.image(Direction(v.Facing), 0)
		if ok {
			DrawSprite(dst, img, sh.AnchorX, sh.AnchorY, mirror, sx, sy)
		}
	} else {
		fillRectF(dst, sx-14, sy-6, 28, 10, rocketShot)
	}
	if !v.Alive {
		return
	}
	fx, fy := math.Cos(v.Facing), math.Sin(v.Facing)
	drawRotor(dst, sx, sy-8, 16, 6, animTime*16, 0.55)
	drawRotor(dst, sx-fx*12, sy-fy*12, 4, 4, animTime*28, 0.85)
}

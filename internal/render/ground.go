package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
	ground "fannon-codder/internal/terrain"
)

func drawTerrain(dst *ebiten.Image, m sim.Map, cam sim.Camera, terrain string) {
	ox := cam.ScreenX(0)
	oy := cam.ScreenY(0)
	ground.Prepare(m, terrain).DrawGround(dst, ox, oy, pictureScale, animTime)
}

func terrainTreeBodies(m sim.Map, cam sim.Camera, terrain string) []body {
	if m.W == 0 {
		return nil
	}
	cans := ground.Prepare(m, terrain).Canopies()
	ox := cam.ScreenX(0)
	oy := cam.ScreenY(0)
	bodies := make([]body, 0, len(cans))
	ps := pictureScale
	for i := range cans {
		c := cans[i]
		bodies = append(bodies, body{
			y: c.FootY,
			shadow: func(dst *ebiten.Image) {
				c.DrawShadow(dst, ox, oy, ps)
			},
			draw: func(dst *ebiten.Image) {
				c.Draw(dst, ox, oy, ps)
			},
		})
	}
	return bodies
}

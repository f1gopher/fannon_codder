package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
	ground "fannon-codder/internal/terrain"
)

// UseChart paints the battle from a freeform region map. ClearChart
// returns the next draw to the tile grid. Sandboxes leave this unset.
func UseChart(c ground.Chart) {
	battleChart.on = true
	battleChart.c = c
}

// ClearChart drops the region map. The next draw uses the tile grid.
func ClearChart() {
	battleChart.on = false
}

var battleChart struct {
	on bool
	c  ground.Chart
}

func groundView(m sim.Map, terrain string) *ground.View {
	if battleChart.on {
		return ground.PrepareChartLive(battleChart.c, m)
	}
	return ground.Prepare(m, terrain)
}

func drawTerrain(dst *ebiten.Image, m sim.Map, cam sim.Camera, terrain string) {
	ox := cam.ScreenX(0)
	oy := cam.ScreenY(0)
	groundView(m, terrain).DrawGround(dst, ox, oy, pictureScale, animTime)
}

func terrainTreeBodies(m sim.Map, cam sim.Camera, terrain string) []body {
	if m.W == 0 {
		return nil
	}
	cans := groundView(m, terrain).Canopies()
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

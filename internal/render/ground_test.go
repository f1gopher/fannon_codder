package render

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
	ground "fannon-codder/internal/terrain"
)

func TestFieldPaintsTheRegionChart(t *testing.T) {
	w := sim.NewEmpty()
	w.Map = sim.Map{W: 4, H: 4, Tiles: make([]sim.Tile, 16)}
	ch := ground.Chart{
		W: 64, H: 64, Terrain: "jungle",
		Regions: []ground.ChartRegion{{
			Tile:   sim.TileTree,
			Points: [][2]float64{{0, 0}, {48, 0}, {48, 48}, {0, 48}},
		}},
	}
	UseChart(ch)
	defer ClearChart()
	if n := len(terrainTreeBodies(w.Map, w.Camera, "jungle")); n == 0 {
		t.Fatal("region trees did not paint")
	}
	ClearChart()
	if n := len(terrainTreeBodies(w.Map, w.Camera, "jungle")); n != 0 {
		t.Fatalf("tile grass grew %d canopies", n)
	}
	dst := ebiten.NewImage(32, 32)
	UseChart(ch)
	Field(dst, w, "jungle")
}

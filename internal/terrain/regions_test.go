package terrain

import (
	"testing"
	"time"

	"fannon-codder/internal/sim"
)

func TestChartTreeIsCanopy(t *testing.T) {
	ch := Chart{
		W: 128, H: 128, Terrain: "jungle",
		Regions: []ChartRegion{{
			Tile: sim.TileTree,
			Points: [][2]float64{
				{32, 32}, {96, 32}, {96, 96}, {32, 96},
			},
		}},
	}
	paint := paintChart(ch)
	if canopyCover(paint, 56, 56) < 200 {
		t.Fatalf("crown alpha %d", canopyCover(paint, 56, 56))
	}
	if canopyCover(paint, 4, 4) != 0 {
		t.Fatal("open ground grew a canopy")
	}
	o := paint.ground.PixOffset(int(4*terrainPx), int(4*terrainPx))
	r, g, b := paint.ground.Pix[o], paint.ground.Pix[o+1], paint.ground.Pix[o+2]
	if g <= r || g <= b {
		t.Fatalf("open ground rgb %d,%d,%d", r, g, b)
	}
}

func TestChartWaterReplacesGround(t *testing.T) {
	ch := Chart{
		W: 128, H: 96, Terrain: "jungle",
		Regions: []ChartRegion{{
			Tile: sim.TileWaterShallow,
			Points: [][2]float64{
				{40, 40}, {90, 40}, {90, 80}, {40, 80},
			},
		}},
	}
	paint := paintChart(ch)
	grass := pixAt(paint, 10, 10)
	water := pixAt(paint, 60, 60)
	if water[2] <= grass[2] {
		t.Fatalf("water %v grass %v", water, grass)
	}
	// The shore waves, so the sand is somewhere on the bank, not on one pixel.
	found := false
	for x := 20; x < 48; x++ {
		c := pixAt(paint, float64(x), 60)
		if c[0] > c[1]+12 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no sand on the west bank, grass %v", grass)
	}
}

func TestChartPaintFitsALoad(t *testing.T) {
	ch := Chart{
		W: 24 * 16, H: 16 * 16, Terrain: "jungle",
		Regions: []ChartRegion{
			{Tile: sim.TileTree, Points: [][2]float64{{180, 8}, {260, 8}, {260, 56}, {180, 56}}},
			{Tile: sim.TileTree, Points: [][2]float64{{80, 50}, {170, 50}, {170, 120}, {80, 120}}},
			{Tile: sim.TileWaterDeep, Points: [][2]float64{{20, 140}, {300, 140}, {300, 200}, {20, 200}}},
		},
	}
	start := time.Now()
	paintChart(ch)
	if time.Since(start) > time.Second {
		t.Fatalf("chart paint %s", time.Since(start))
	}
}

func TestSpentMineDropsTheMark(t *testing.T) {
	ch := Chart{
		W: 64, H: 64, Terrain: "jungle",
		Regions: []ChartRegion{{
			Tile:   sim.TileMine,
			Points: [][2]float64{{20, 20}, {28, 20}, {28, 28}, {20, 28}},
		}},
	}
	tiles := make([]sim.Tile, 16)
	tiles[1*4+1] = sim.TileMine
	ch.Live = sim.Map{W: 4, H: 4, Tiles: tiles}
	armed := pixAt(paintChart(ch), 23, 23)
	tiles[1*4+1] = sim.TileGrass
	spent := pixAt(paintChart(ch), 23, 23)
	if armed == spent {
		t.Fatalf("spent mine still matches the armed mark %v", armed)
	}
	if !ch.Regions[0].Contains(24, 24) {
		t.Fatal("mine ring missed its cell centre")
	}
}

func canopyCover(p terrainPaint, wx, wy float64) uint8 {
	var a uint8
	for _, tree := range p.trees {
		for _, run := range tree.runs {
			if run.sprite == nil {
				continue
			}
			lx := int((wx - run.world) * terrainPx)
			ly := int((wy - run.north) * terrainPx)
			b := run.sprite.Bounds()
			if lx < 0 || ly < 0 || lx >= b.Dx() || ly >= b.Dy() {
				continue
			}
			v := run.sprite.Pix[run.sprite.PixOffset(lx, ly)+3]
			if v > a {
				a = v
			}
		}
	}
	return a
}

func pixAt(p terrainPaint, wx, wy float64) [3]uint8 {
	o := p.ground.PixOffset(int(wx*terrainPx), int(wy*terrainPx))
	return [3]uint8{p.ground.Pix[o], p.ground.Pix[o+1], p.ground.Pix[o+2]}
}

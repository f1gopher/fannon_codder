package terrain

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

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

func TestRasterMatchesTheRay(t *testing.T) {
	const w, h = 48, 36
	rings := [][][2]float64{
		{{4, 4}, {40, 4}, {40, 20}, {4, 20}},
		{{2, 2}, {30, 28}, {8, 30}},
		{{10, 8.5}, {36, 8.5}, {36, 22}, {10, 22}},
		{{6, 6}, {20, 6}, {20, 14}, {28, 14}, {28, 6}, {40, 6}, {40, 30}, {6, 30}},
		{{8, 10}, {24, 4}, {18, 16}, {30, 28}, {8, 22}},
		{{-4, -4}, {10, -4}, {10, 8}, {-4, 8}},
		{{0, 0}, {47, 0}, {47, 35}, {0, 35}, {0, 0}},
		{{1.5, 3}, {20, 3}, {20, 3}, {12, 18}},
	}
	open := surfGrass
	for i, pts := range rings {
		ch := Chart{
			W: w, H: h, Terrain: "jungle",
			Regions: []ChartRegion{{Tile: sim.TileWaterDeep, Points: pts}},
		}
		got := rasterChart(ch, open)
		for y := 0; y < h; y++ {
			py := float64(y) + 0.5
			for x := 0; x < w; x++ {
				want := open
				if pointInRing(float64(x)+0.5, py, pts) {
					want = surfDeep
				}
				if got[y*w+x] != want {
					t.Fatalf("ring %d at %d,%d got %v want %v", i, x, y, got[y*w+x], want)
				}
			}
		}
	}
}

func TestPrepareUsesTheWarmPaint(t *testing.T) {
	ch := Chart{
		W: 64, H: 48, Terrain: "jungle",
		Regions: []ChartRegion{{
			Tile:   sim.TileTree,
			Points: [][2]float64{{8, 8}, {40, 8}, {40, 36}, {8, 36}},
		}},
	}
	chartNow.hash = 0
	chartNow.view = nil
	chartUp = chartUpload{}
	chartWarm.mu.Lock()
	chartWarm.hash = 0
	chartWarm.paint = terrainPaint{}
	chartWarm.ready = false
	chartWarm.wait = nil
	chartWarm.mu.Unlock()

	chartPaintMu.Lock()
	before := paintCharts
	chartPaintMu.Unlock()
	WarmChart(ch)
	h := hashChart(ch)
	deadline := time.Now().Add(5 * time.Second)
	for {
		chartWarm.mu.Lock()
		ready := chartWarm.ready && chartWarm.hash == h && chartWarm.paint.ground != nil
		chartWarm.mu.Unlock()
		if ready {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("warm paint did not finish")
		}
		time.Sleep(10 * time.Millisecond)
	}
	chartPaintMu.Lock()
	painted := paintCharts
	chartPaintMu.Unlock()
	if painted != before+1 {
		t.Fatalf("warm paints %d, want 1", painted-before)
	}
	v := PrepareChart(ch)
	chartPaintMu.Lock()
	after := paintCharts
	chartPaintMu.Unlock()
	if after != painted {
		t.Fatal("PrepareChart painted again")
	}
	if v == nil || v.ground == nil || len(v.canopies) == 0 {
		t.Fatal("warm chart did not upload")
	}
	c := v.canopies[0]
	if !c.padded || c.img == nil || c.img.Bounds().Min.X != 1 || c.img.Bounds().Min.Y != 1 {
		t.Fatalf("canopy pad %+v bounds %v", c.padded, c.img.Bounds())
	}
	// Drawing uses the padded rect and does not read the texture back.
	dst := ebiten.NewImage(32, 32)
	c.Draw(dst, 0, 0, 1)
}

func TestPumpWarmFinishesOnTheGameThread(t *testing.T) {
	ch := Chart{
		W: 64, H: 48, Terrain: "jungle",
		Regions: []ChartRegion{{
			Tile:   sim.TileTree,
			Points: [][2]float64{{8, 8}, {40, 8}, {40, 36}, {8, 36}},
		}},
	}
	chartNow.hash = 0
	chartNow.view = nil
	chartUp = chartUpload{}
	chartWarm.mu.Lock()
	chartWarm.hash = 0
	chartWarm.paint = terrainPaint{}
	chartWarm.ready = false
	chartWarm.wait = nil
	chartWarm.mu.Unlock()

	WarmChart(ch)
	h := hashChart(ch)
	deadline := time.Now().Add(5 * time.Second)
	for !PumpWarm(ch, time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("warm upload did not finish")
		}
		time.Sleep(5 * time.Millisecond)
	}
	chartPaintMu.Lock()
	painted := paintCharts
	chartPaintMu.Unlock()
	v := PrepareChart(ch)
	chartPaintMu.Lock()
	after := paintCharts
	chartPaintMu.Unlock()
	if after != painted {
		t.Fatal("PrepareChart painted after the pump")
	}
	if v == nil || v.ground == nil || len(v.canopies) == 0 || chartNow.hash != h {
		t.Fatal("pumped chart was not cached")
	}
}

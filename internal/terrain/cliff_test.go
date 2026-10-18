package terrain

import (
	"image/color"
	"testing"
	"time"

	"fannon-codder/internal/sim"
)

func cliffBand() Chart {
	return Chart{
		W: 520, H: 160, Terrain: "snow",
		Regions: []ChartRegion{{
			Tile:   sim.TileCliff,
			Points: [][2]float64{{8, 80}, {512, 80}, {512, 108}, {8, 108}},
		}},
	}
}

// edgeY is how far the cliff colour has come down from the snow above it.
func cliffNorth(p terrainPaint, x int) float64 {
	ref := pixAt(p, float64(x), 60)
	for y := 68.0; y < 120; y += 0.5 {
		c := pixAt(p, float64(x), y)
		if colorDist(color.RGBA{R: c[0], G: c[1], B: c[2], A: 255}, color.RGBA{R: ref[0], G: ref[1], B: ref[2], A: 255}) > 28 {
			return y
		}
	}
	return 0
}

func TestSnowCliffEdgeWanders(t *testing.T) {
	p := paintChart(cliffBand())
	minY, maxY := 999.0, 0.0
	for x := 40; x <= 480; x += 4 {
		y := cliffNorth(p, x)
		if y == 0 {
			t.Fatalf("x=%d never leaves the snow", x)
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	if maxY-minY < 3 {
		t.Fatalf("north edge spans %.1f to %.1f, want a wave like the shore", minY, maxY)
	}
	if maxY-minY > 22 {
		t.Fatalf("north edge spans %.1f to %.1f, want a small wave", minY, maxY)
	}
	// The foot is darker than the lip, and cooler than the rock plate.
	lip := pixAt(p, 200, minY+1)
	foot := pixAt(p, 200, minY+18)
	if int(foot[0])+int(foot[1])+int(foot[2])+24 > int(lip[0])+int(lip[1])+int(lip[2]) {
		t.Fatalf("foot %v is not darker than the lip %v", foot, lip)
	}
	wx, wy := 200.5, minY+12
	rock := rockPlate.atScale(wx, wy, 5)
	got := pixAt(p, 200, minY+12)
	if colorDist(color.RGBA{R: got[0], G: got[1], B: got[2], A: 255}, rock) < 36 {
		t.Fatalf("face %v still matches the rock plate %v", got, rock)
	}
	far := pixAt(p, 200, 20)
	if int(far[0])+int(far[1])+int(far[2]) < 500 {
		t.Fatalf("open snow %v went dark", far)
	}
}

func TestSnowCliffRampStaysOpen(t *testing.T) {
	ch := cliffBand()
	ch.Regions = append(ch.Regions, ChartRegion{
		Tile:   sim.TileRamp,
		Points: [][2]float64{{240, 76}, {280, 76}, {280, 112}, {240, 112}},
	})
	p := paintChart(ch)
	c := pixAt(p, 260, 94)
	if c[0] < c[2]+12 {
		t.Fatalf("ramp %v, want the warm dirt cut", c)
	}
	// The face is still there beside the cut.
	side := pixAt(p, 200, 94)
	if side[2]+8 < side[0] {
		t.Fatalf("cliff beside the ramp %v, want cool rock", side)
	}
}

func TestJungleCliffKeepsTheRockPlate(t *testing.T) {
	tiles := make([]sim.Tile, 8*8)
	for x := 0; x < 8; x++ {
		tiles[3*8+x] = sim.TileCliff
	}
	p := paintMap(mapOf(8, 8, tiles), "jungle")
	px := (3*sim.TileSize + 4) * terrainPx
	py := (3*sim.TileSize + 4) * terrainPx
	wx := (float64(px) + 0.5) / float64(terrainPx)
	wy := (float64(py) + 0.5) / float64(terrainPx)
	got := p.ground.RGBAAt(px, py)
	want := rockPlate.atScale(wx, wy, 5)
	if got != want {
		t.Fatalf("jungle cliff %+v, rock plate %+v", got, want)
	}
}

func TestSnowCliffPaintFitsALoad(t *testing.T) {
	ch := Chart{
		W: 36 * sim.TileSize, H: 26 * sim.TileSize, Terrain: "snow",
		Regions: []ChartRegion{{
			Tile: sim.TileCliff,
			Points: [][2]float64{
				{0, 120}, {570, 120}, {570, 150}, {0, 150},
			},
		}},
	}
	start := time.Now()
	paintChart(ch)
	if d := time.Since(start); d > time.Second {
		t.Fatalf("snow cliff paint took %s", d)
	}
}

func TestSnowTileCliffWanders(t *testing.T) {
	const w, h = 24, 8
	tiles := make([]sim.Tile, w*h)
	for x := 0; x < w; x++ {
		tiles[3*w+x] = sim.TileCliff
	}
	p := paintMap(mapOf(w, h, tiles), "snow")
	minY, maxY := 999.0, 0.0
	for x := 2 * sim.TileSize; x < 22*sim.TileSize; x += 4 {
		ref := pixAt(p, float64(x), float64(2*sim.TileSize+4))
		yHit := 0.0
		for y := float64(3*sim.TileSize) - 8; y < float64(3*sim.TileSize)+8; y += 0.5 {
			c := pixAt(p, float64(x), y)
			if colorDist(color.RGBA{R: c[0], G: c[1], B: c[2], A: 255}, color.RGBA{R: ref[0], G: ref[1], B: ref[2], A: 255}) > 28 {
				yHit = y
				break
			}
		}
		if yHit == 0 {
			t.Fatalf("x=%d never leaves the snow", x)
		}
		if yHit < minY {
			minY = yHit
		}
		if yHit > maxY {
			maxY = yHit
		}
	}
	if maxY-minY < 3 {
		t.Fatalf("tile cliff spans %.1f to %.1f, want a wave", minY, maxY)
	}
	if maxY-minY > 24 {
		t.Fatalf("tile cliff wave %.1f is too large", maxY-minY)
	}
}

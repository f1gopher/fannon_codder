package terrain

import (
	"image"
	"math"
	"testing"

	"fannon-codder/internal/sim"
)

func TestCanopyAlphaFastPath(t *testing.T) {
	noise := buildNoise(16, 16, 11)
	for i := 0; i < 4000; i++ {
		wx := float64(i%180)*0.37 + float64(i%7)
		wy := float64(i%130)*0.41 - float64(i%5)
		if a := canopyAlpha(-6.8, wx, wy, noise); a != 1 {
			t.Fatalf("deep alpha %v", a)
		}
		if a := canopyAlpha(10.1, wx, wy, noise); a != 0 {
			t.Fatalf("outer alpha %v", a)
		}
	}
}

func TestAtHalfMatchesAtScale(t *testing.T) {
	p := lookOf("jungle").canopy
	for _, origin := range []int{-80, -1, 0, 3, 40, 500, 2000} {
		for k := 0; k < 32; k++ {
			wx := float64(origin) + (float64(k)+0.5)/4
			wy := float64(origin-7) + (float64(k)+0.5)/4
			ix := origin*4 + k
			iy := (origin-7)*4 + k
			got := p.atHalf(ix, iy)
			want := p.atScale(wx, wy, 4)
			if got != want {
				t.Fatalf("origin %d k %d got %v want %v", origin, k, got, want)
			}
		}
	}
}

func TestCrownNearMatchesTheScan(t *testing.T) {
	crowns := []crown{
		{tx: 2, ty: 2, cx: 40, cy: 36, mainR: 14},
		{tx: 3, ty: 2, cx: 58, cy: 40, mainR: 12},
		{tx: 8, ty: 6, cx: 130, cy: 100, mainR: 15},
		{tx: 9, ty: 7, cx: 150, cy: 120, mainR: 11},
	}
	near := newCrownNear(crowns)
	for wy := 0.0; wy < 180; wy += 5.5 {
		for wx := 0.0; wx < 200; wx += 4.5 {
			got := near.dist(wx, wy)
			want := shadeDistScan(crowns, wx, wy)
			if math.Abs(got-want) > 1e-9 {
				t.Fatalf("at %v,%v got %v want %v", wx, wy, got, want)
			}
		}
	}
}

func TestCellNearMatchesTheScan(t *testing.T) {
	cells := []image.Point{
		{X: 1, Y: 1},
		{X: 2, Y: 1},
		{X: 4, Y: 3},
		{X: 5, Y: 4},
	}
	pts := append([]image.Point(nil), cells...)
	near := newCellNear(pts)
	for wy := 8.0; wy < 90; wy += 3.25 {
		for wx := 8.0; wx < 110; wx += 2.75 {
			got, gotGap := near.nearest(wx, wy)
			want, wantGap := cellGap(pts, wx, wy)
			if got != want || near.index(wx, wy) != want {
				t.Fatalf("at %v,%v got %d index %d want %d", wx, wy, got, near.index(wx, wy), want)
			}
			if gotGap > wantGap+1e-6 {
				t.Fatalf("gap overstated at %v,%v got %v want %v", wx, wy, gotGap, wantGap)
			}
		}
	}
	// The neighbour sits inside the 3×3, so the gap is exact.
	idx, gap := near.nearest(24, 24)
	if idx != 0 || math.Abs(gap-16) > 1e-6 {
		t.Fatalf("centre owner %d gap %v", idx, gap)
	}
}

// cellGap is the full four-tile gap nearestTree would imply.
func cellGap(cells []image.Point, wx, wy float64) (int, float64) {
	best := 0
	bestD := math.MaxFloat64
	second := math.MaxFloat64
	gx := int(math.Floor(wx / float64(sim.TileSize)))
	gy := int(math.Floor(wy / float64(sim.TileSize)))
	for i, cell := range cells {
		if absInt(cell.X-gx) > 4 || absInt(cell.Y-gy) > 4 {
			continue
		}
		dx := float64(cell.X*sim.TileSize) + float64(sim.TileSize)/2 - wx
		dy := float64(cell.Y*sim.TileSize) + float64(sim.TileSize)/2 - wy
		d := dx*dx + dy*dy
		if d < bestD-1e-6 {
			second = bestD
			bestD = d
			best = i
			continue
		}
		if math.Abs(d-bestD) <= 1e-6 {
			prev := cells[best]
			if cell.Y > prev.Y || (cell.Y == prev.Y && cell.X > prev.X) {
				best = i
			}
			if bestD < second {
				second = bestD
			}
			continue
		}
		if d < second {
			second = d
		}
	}
	if second == math.MaxFloat64 {
		return best, 1e9
	}
	return best, math.Sqrt(second) - math.Sqrt(bestD)
}

func TestUniformBlockMatchesSamples(t *testing.T) {
	ch := Chart{
		W: 96, H: 80, Terrain: "jungle",
		Regions: []ChartRegion{
			{Tile: sim.TileWaterDeep, Points: [][2]float64{{10, 40}, {80, 40}, {80, 70}, {10, 70}}},
			{Tile: sim.TileTree, Points: [][2]float64{{20, 8}, {50, 8}, {50, 28}, {20, 28}}},
		},
	}
	lk := lookOf(ch.Terrain)
	prepareGrass(ch.W, ch.H, lk.kind)
	grid := rasterChart(ch, lk.open)
	toWater, toLand := chartWaterDist(grid, ch.W, ch.H)
	bindShore(toWater, toLand, ch.W, ch.H)
	for y := 0; y < ch.H; y++ {
		for x := 0; x < ch.W; x++ {
			wx := float64(x) + 0.5
			wy := float64(y) + 0.5
			col, k, ok := uniformPixel(lk, grid[y*ch.W+x], wx, wy)
			if !ok {
				continue
			}
			for sy := 0; sy < terrainPx; sy++ {
				py := y*terrainPx + sy
				wyy := (float64(py) + 0.5) / float64(terrainPx)
				for sx := 0; sx < terrainPx; sx++ {
					px := x*terrainPx + sx
					wxx := (float64(px) + 0.5) / float64(terrainPx)
					got, gk := chartPixel(lk, grid[y*ch.W+x], wxx, wyy)
					if got != col || gk != k {
						t.Fatalf("world %d,%d sample %d,%d got %v/%d want %v/%d", x, y, sx, sy, got, gk, col, k)
					}
				}
			}
		}
	}
}

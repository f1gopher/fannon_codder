package terrain

import (
	"testing"

	"fannon-codder/internal/sim"
)

func TestSandStaysWarm(t *testing.T) {
	for y := 0.0; y < 256; y += 3.5 {
		for x := 0.0; x < 384; x += 3.7 {
			c := sandColor(x, y)
			if c.R < c.G || c.G < c.B {
				t.Fatalf("sand at %.1f,%.1f is %+v", x, y, c)
			}
			if c.R < 110 {
				t.Fatalf("sand at %.1f,%.1f is %+v, want a lit bank", x, y, c)
			}
		}
	}
}

func TestSandDoesNotRepeat(t *testing.T) {
	same := true
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			a := sandColor(float64(x)+0.5, float64(y)+0.5)
			b := sandColor(float64(x)+200.5, float64(y)+0.5)
			if a != b {
				same = false
			}
		}
	}
	if same {
		t.Fatal("a 32px window matched the window 200px east")
	}
}

func TestSandHasNoCellGrid(t *testing.T) {
	maxStep := 0
	prev := sandColor(0.5, 40.5)
	for x := 1; x < 96; x++ {
		c := sandColor(float64(x)+0.5, 40.5)
		d := colorDist(c, prev)
		if d > maxStep {
			maxStep = d
		}
		prev = c
	}
	if maxStep == 0 || maxStep > 18 {
		t.Fatalf("neighbour step %d", maxStep)
	}
}

func TestSandOpenGroundIsProcedural(t *testing.T) {
	px, py := 40*sim.TileSize*terrainPx/16+5, 24*terrainPx+3
	wx := (float64(px) + 0.5) / float64(terrainPx)
	wy := (float64(py) + 0.5) / float64(terrainPx)
	paint := paintMap(mapOf(8, 8, nil), "desert")
	got := paint.ground.RGBAAt(px, py)
	want := sandSample(wx, wy)
	if colorDist(got, want) > 2 {
		t.Fatalf("desert pixel %+v, want %+v", got, want)
	}
	if paint.mask.RGBAAt(px, py).R != maskNone {
		t.Fatalf("desert sand mask %d", paint.mask.RGBAAt(px, py).R)
	}
	chart := paintChart(Chart{W: 128, H: 96, Terrain: "desert"})
	got = chart.ground.RGBAAt(px, py)
	if colorDist(got, sandSample(wx, wy)) > 2 {
		t.Fatalf("chart pixel %+v, want %+v", got, sandSample(wx, wy))
	}
}

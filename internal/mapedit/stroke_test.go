package mapedit

import (
	"testing"

	"fannon-codder/internal/sim"
)

func TestStrokeAddsATreeInTheRegionsView(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if ed.beginStroke(10, 10) {
		t.Fatal("terrain view accepted a stroke")
	}
	ed.gridView = true
	if !ed.beginStroke(10, 10) {
		t.Fatal("regions view rejected a stroke")
	}
	ed.extendStroke(80, 10)
	ed.extendStroke(80, 70)
	ed.extendStroke(10, 70)
	if !ed.endStroke() {
		t.Fatal("square was dropped")
	}
	if len(ed.chart.Regions) != 5 {
		t.Fatalf("regions = %d", len(ed.chart.Regions))
	}
	got := ed.chart.Regions[4]
	if got.Name != "tree" || got.Tile != sim.TileTree {
		t.Fatalf("region = %q %v", got.Name, got.Tile)
	}
	if len(got.Points) != 4 || got.Points[0] != [2]float64{10, 10} || got.Points[3] != [2]float64{10, 70} {
		t.Fatalf("points = %v", got.Points)
	}

	ed.freeform = false
	if ed.beginStroke(10, 10) {
		t.Fatal("tile grid accepted a stroke")
	}
	ed.load(3)
	ed.load(0)
	if !ed.gridView || len(ed.chart.Regions) != 5 {
		t.Fatalf("restored view=%v regions=%d", ed.gridView, len(ed.chart.Regions))
	}
	if ed.chart.Regions[4].Points[1] != [2]float64{80, 10} {
		t.Fatalf("restored = %v", ed.chart.Regions[4].Points)
	}
}

func TestShortStrokeIsDropped(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ed.gridView = true
	if ed.beginStroke(-1, 10) || ed.beginStroke(float64(ed.chart.W), 10) {
		t.Fatal("off-map point started a stroke")
	}
	if !ed.beginStroke(10, 10) {
		t.Fatal("begin")
	}
	ed.extendStroke(12, 11)
	ed.extendStroke(11, 14)
	if ed.endStroke() {
		t.Fatal("speck became a region")
	}
	if len(ed.chart.Regions) != 4 || len(ed.stroke) != 0 {
		t.Fatalf("regions=%d stroke=%d", len(ed.chart.Regions), len(ed.stroke))
	}
}

func TestStrokeClampsToTheMap(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ed.gridView = true
	w, h := float64(ed.chart.W), float64(ed.chart.H)
	if !ed.beginStroke(20, 20) {
		t.Fatal("begin")
	}
	ed.extendStroke(w+40, 20)
	ed.extendStroke(w+40, h+40)
	ed.extendStroke(20, h+40)
	if !ed.finishStroke(20, 20) {
		t.Fatal("close")
	}
	pts := ed.chart.Regions[len(ed.chart.Regions)-1].Points
	for _, p := range pts {
		if p[0] < 0 || p[1] < 0 || p[0] > w || p[1] > h {
			t.Fatalf("point %v outside %v×%v", p, w, h)
		}
	}
	if pts[1] != [2]float64{w, 20} || pts[2] != [2]float64{w, h} {
		t.Fatalf("clamped = %v", pts)
	}
}

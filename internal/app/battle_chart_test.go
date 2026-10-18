package app

import (
	"testing"

	"fannon-codder/internal/sim"
)

func TestBattlePaintsThePhaseRegions(t *testing.T) {
	p := NewProgress()
	if p.Phase == nil || !p.Phase.Map.Freeform() {
		t.Fatal("campaign phase is not a region map")
	}
	b := NewBattle(p)
	if !b.chartOn || b.chart.W != p.Phase.Map.W*sim.TileSize || len(b.chart.Regions) != len(p.Phase.Map.Regions) {
		t.Fatalf("chart on=%v size=%d regions=%d", b.chartOn, b.chart.W, len(b.chart.Regions))
	}
	sand := NewCoverBattle(p)
	if sand.chartOn {
		t.Fatal("sandbox used the campaign regions")
	}
}

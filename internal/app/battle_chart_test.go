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

func TestPhaseGridIsTheBattleMap(t *testing.T) {
	p := NewProgress()
	w, err := p.Phase.World()
	if err != nil {
		t.Fatal(err)
	}
	g, err := p.Phase.Grid()
	if err != nil {
		t.Fatal(err)
	}
	if g.W != w.Map.W || g.H != w.Map.H || len(g.Tiles) != len(w.Map.Tiles) {
		t.Fatalf("grid %d×%d (%d) battle %d×%d (%d)", g.W, g.H, len(g.Tiles), w.Map.W, w.Map.H, len(w.Map.Tiles))
	}
	for i, tile := range g.Tiles {
		if tile != w.Map.Tiles[i] {
			t.Fatalf("cell %d grid %v battle %v", i, tile, w.Map.Tiles[i])
		}
	}
}

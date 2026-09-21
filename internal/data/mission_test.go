package data

import (
	"testing"

	"fannon-codder/internal/sim"
)

func TestLoadMission1(t *testing.T) {
	p, err := LoadPhase("m01p01.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.Deploy != 2 {
		t.Fatalf("deploy=%d, want 2", p.Deploy)
	}
	if len(p.Enemies) != 3 {
		t.Fatalf("enemies=%d, want 3", len(p.Enemies))
	}
	w, err := p.World()
	if err != nil {
		t.Fatal(err)
	}
	sx, sy := sim.TileCenter(10, 13).X, sim.TileCenter(10, 13).Y
	if !w.Map.Walkable(sx, sy, 4) {
		t.Fatal("player start should be walkable")
	}
	tx, ty := sim.TileCenter(11, 1).X, sim.TileCenter(11, 1).Y
	if w.Map.Walkable(tx, ty, 4) {
		t.Fatal("tree tile should block")
	}
	nP, nE := 0, 0
	for i := range w.Units {
		if !w.Units[i].Living() {
			continue
		}
		switch w.Units[i].Side {
		case sim.SidePlayer:
			nP++
		case sim.SideEnemy:
			nE++
		}
	}
	if nP != 2 || nE != 3 {
		t.Fatalf("units player=%d enemy=%d, want 2 and 3", nP, nE)
	}
}

func TestParseWaterTiles(t *testing.T) {
	cases := []struct {
		ch rune
		t  sim.Tile
	}{
		{'~', sim.TileWaterShallow},
		{'W', sim.TileWaterDeep},
		{'B', sim.TileBridge},
		{'=', sim.TileBridge},
	}
	for _, c := range cases {
		got, err := parseTile(c.ch)
		if err != nil {
			t.Fatalf("parse %q: %v", string(c.ch), err)
		}
		if got != c.t {
			t.Fatalf("parse %q = %v, want %v", string(c.ch), got, c.t)
		}
	}
}

package sim

import "testing"

func TestEmptyMapHasLOS(t *testing.T) {
	var m Map
	if !m.HasLOS(0, 0, 400, 400) {
		t.Fatal("inactive map must not block LOS")
	}
}

func TestTreeBlocksLOS(t *testing.T) {
	m := Map{W: 6, H: 1, Tiles: []Tile{
		TileGrass, TileGrass, TileTree, TileTree, TileGrass, TileGrass,
	}}
	if m.HasLOS(8, 8, 72, 8) {
		t.Fatal("tree wall should block LOS")
	}
	if !m.HasLOS(8, 8, 24, 8) {
		t.Fatal("open grass should have LOS")
	}
}

func TestCoverWorldEastGruntBehindTrees(t *testing.T) {
	w := NewCoverWorld()
	p := w.Unit(w.ActiveSquad().LeaderID)
	var east *Unit
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side == SideEnemy && u.X > 320 {
			east = u
			break
		}
	}
	if p == nil || east == nil {
		t.Fatal("need player and an off-screen east grunt")
	}
	if w.Map.HasLOS(p.X, p.Y, east.X, east.Y) {
		t.Fatal("tree line should hide the east grunt from spawn")
	}
}

func TestFirstSolidStopsInTree(t *testing.T) {
	m := Map{W: 4, H: 1, Tiles: []Tile{TileGrass, TileTree, TileGrass, TileGrass}}
	hx, _, hit := m.FirstSolid(8, 8, 56, 8)
	if !hit {
		t.Fatal("expected a tree hit")
	}
	if hx < 16 || hx > 32 {
		t.Fatalf("hit x=%v, want inside tree tile [16,32]", hx)
	}
}

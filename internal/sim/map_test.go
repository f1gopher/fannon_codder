package sim

import "testing"

func TestTreeTileNotWalkable(t *testing.T) {
	m := Map{W: 4, H: 1, Tiles: []Tile{TileGrass, TileTree, TileGrass, TileGrass}}
	// Tile 1 occupies x 16–32.
	if m.Walkable(24, 8, 4) {
		t.Fatal("centre of tree tile should be blocked")
	}
	if !m.Walkable(8, 8, 4) {
		t.Fatal("grass should be walkable")
	}
}

func TestEmptyMapAlwaysWalkable(t *testing.T) {
	var m Map
	if !m.Walkable(999, 999, 4) {
		t.Fatal("inactive map must not block (keeps old tests working)")
	}
}

func TestSlideAlongTree(t *testing.T) {
	w := NewEmpty()
	w.Map = Map{W: 4, H: 2, Tiles: []Tile{
		TileGrass, TileTree, TileGrass, TileGrass,
		TileGrass, TileTree, TileGrass, TileGrass,
	}}
	u := w.SpawnUnit(SidePlayer, Vec2{X: 8, Y: 8})
	// Walk east into the tree at tile (1,*).
	for i := 0; i < 120; i++ {
		w.steerToward(u, 40, 8, WalkSpeed, 1.0/60, ArrivalRadius)
	}
	if u.X >= 16 {
		t.Fatalf("should slide/stop before tree, x=%v", u.X)
	}
}

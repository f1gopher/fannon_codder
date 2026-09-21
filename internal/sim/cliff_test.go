package sim

import "testing"

func TestIceWalksLikeGrass(t *testing.T) {
	grass := travelX(TileGrass)
	ice := travelX(TileIce)
	if ice < grass-1 || ice > grass+1 {
		t.Fatalf("ice %v should match grass %v", ice, grass)
	}
}

func TestCliffDropsSouthAndBlocksNorth(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	tiles := make([]Tile, 8*6)
	for i := range tiles {
		tiles[i] = TileIce
	}
	for x := 0; x < 8; x++ {
		tiles[2*8+x] = TileCliff
	}
	w.Map = Map{W: 8, H: 6, Tiles: tiles}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 24}})
	w.CommandMove(40, 90, true)
	for i := 0; i < 180; i++ {
		w.Step(1.0 / 60)
	}
	leader := w.Unit(w.ActiveSquad().LeaderID)
	if leader.Y < float64(3*TileSize) {
		t.Fatalf("should have dropped south of the cliff, y=%v", leader.Y)
	}
	if w.Map.TileAtPixel(leader.X, leader.Y) == TileCliff {
		t.Fatal("landed on the cliff face")
	}

	w = NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.Map = Map{W: 8, H: 6, Tiles: tiles}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 64}})
	w.CommandMove(40, 8, true)
	for i := 0; i < 180; i++ {
		w.Step(1.0 / 60)
	}
	leader = w.Unit(w.ActiveSquad().LeaderID)
	if leader.Y < float64(3*TileSize) {
		t.Fatalf("climbed the cliff from below, y=%v", leader.Y)
	}
}

func TestRampClimbsTheCliff(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	tiles := make([]Tile, 8*6)
	for i := range tiles {
		tiles[i] = TileIce
	}
	for x := 0; x < 8; x++ {
		tiles[2*8+x] = TileCliff
	}
	tiles[2*8+5] = TileRamp
	w.Map = Map{W: 8, H: 6, Tiles: tiles}
	// Ramp tile (5, 2) centre is (88, 40). Start south of it.
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 88, Y: 64}})
	w.CommandMove(88, 16, true)
	for i := 0; i < 240; i++ {
		w.Step(1.0 / 60)
	}
	leader := w.Unit(w.ActiveSquad().LeaderID)
	if leader.Y > float64(2*TileSize) {
		t.Fatalf("ramp should reach the high ground, y=%v", leader.Y)
	}
}

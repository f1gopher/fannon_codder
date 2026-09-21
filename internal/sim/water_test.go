package sim

import "testing"

func TestDeepWaterCannotFire(t *testing.T) {
	w := gunWorld([]Vec2{{X: 8, Y: 8}}, Vec2{X: 200, Y: 8})
	w.Map = strip(16, TileWaterDeep)
	w.SetFire(80, 8, true)
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("unit in deep water must not fire")
	}
}

func TestShallowWaterCanFire(t *testing.T) {
	w := gunWorld([]Vec2{{X: 8, Y: 8}}, Vec2{X: 40, Y: 8})
	w.Map = strip(8, TileWaterShallow)
	fireAt(w, 80, 8, 20)
	if enemyOf(w).Living() {
		t.Fatal("wading unit should still be able to kill with MG")
	}
}

func TestBridgeTreatsAsLand(t *testing.T) {
	w := gunWorld([]Vec2{{X: 8, Y: 8}}, Vec2{X: 40, Y: 8})
	w.Map = strip(8, TileBridge)
	w.SetFire(80, 8, true)
	w.Step(1.0 / 60)
	if len(w.Projectiles) == 0 {
		t.Fatal("bridge is land: MG should fire")
	}
	if w.Unit(w.ActiveSquad().LeaderID).InWater {
		t.Fatal("bridge must not set InWater")
	}
}

func TestDeepWaterSlowerThanShallow(t *testing.T) {
	grass := travelX(TileGrass)
	shallow := travelX(TileWaterShallow)
	deep := travelX(TileWaterDeep)
	if shallow >= grass {
		t.Fatalf("shallow %v should be slower than grass %v", shallow, grass)
	}
	if deep >= shallow {
		t.Fatalf("deep %v should be slower than shallow %v", deep, shallow)
	}
}

func TestEnemyInDeepWaterIsSittingDuck(t *testing.T) {
	w := aiWorld(Vec2{X: 8, Y: 8}, Vec2{X: 56, Y: 8})
	tiles := []Tile{TileGrass, TileGrass, TileWaterDeep, TileWaterDeep, TileWaterDeep, TileGrass}
	w.Map = Map{W: 6, H: 1, Tiles: tiles}
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("swimming grunt must not fire")
	}
}

func TestWaterAndBridgeWalkable(t *testing.T) {
	m := Map{W: 3, H: 1, Tiles: []Tile{TileWaterShallow, TileWaterDeep, TileBridge}}
	for x, name := range []string{"shallow", "deep", "bridge"} {
		cx := float64(x*TileSize) + 8
		if !m.Walkable(cx, 8, 4) {
			t.Fatalf("%s should be walkable", name)
		}
	}
}

func TestRiverWorldBridgeAndSwimmer(t *testing.T) {
	w := NewRiverWorld()
	if w.Map.At(10, 6) != TileBridge {
		t.Fatal("expected a bridge over the river")
	}
	if w.Map.At(4, 6) != TileWaterDeep {
		t.Fatal("expected deep water beside the bridge")
	}
	nSwim := 0
	nPlayer := 0
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side == SidePlayer {
			nPlayer++
		}
		if u.Side == SideEnemy && u.InWater {
			nSwim++
		}
	}
	if nSwim < 1 {
		t.Fatal("expected at least one swimming grunt")
	}
	if nPlayer != 3 {
		t.Fatalf("river sandbox wants 3 troopers to split, got %d", nPlayer)
	}
}

func strip(n int, t Tile) Map {
	tiles := make([]Tile, n)
	for i := range tiles {
		tiles[i] = t
	}
	return Map{W: n, H: 1, Tiles: tiles}
}

func travelX(t Tile) float64 {
	w := NewEmpty()
	if t != TileGrass {
		w.Map = strip(20, t)
	}
	u := w.SpawnUnit(SidePlayer, Vec2{X: 8, Y: 8})
	for i := 0; i < 60; i++ {
		w.steerToward(u, 400, 8, WalkSpeed, 1.0/60, ArrivalRadius)
	}
	return u.X
}

package sim

import "testing"

func TestEnemyShootsWhenPlayerInRange(t *testing.T) {
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 40, Y: 0})
	w.Step(1.0 / 60)
	if len(w.Projectiles) == 0 {
		t.Fatal("grunt in range should fire")
	}
}

func TestEnemyIdleWhenPlayerFar(t *testing.T) {
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 200, Y: 0})
	ex := enemyOf(w).X
	for i := 0; i < 20; i++ {
		w.Step(1.0 / 60)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("grunt out of approach range should not fire")
	}
	if enemyOf(w).X != ex {
		t.Fatal("grunt should idle, not walk")
	}
}

func TestEnemyApproachesWhenCloseButOutOfShot(t *testing.T) {
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 110, Y: 0})
	start := enemyOf(w).X
	for i := 0; i < 40; i++ {
		w.Step(1.0 / 60)
	}
	got := enemyOf(w).X
	if got >= start {
		t.Fatalf("grunt should walk toward player, x %v -> %v", start, got)
	}
}

func TestKillAllEnemiesWins(t *testing.T) {
	w := NewDemoWorld()
	w.Spread = 0
	if w.Status != Playing {
		t.Fatal("demo should start Playing")
	}
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy {
			w.kill(&w.Units[i])
		}
	}
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatalf("status=%v, want Won", w.Status)
	}
}

func TestAllPlayersDeadLoses(t *testing.T) {
	w := NewDemoWorld()
	for i := range w.Units {
		if w.Units[i].Side == SidePlayer {
			w.kill(&w.Units[i])
		}
	}
	w.evaluateObjectives()
	if w.Status != Lost {
		t.Fatalf("status=%v, want Lost", w.Status)
	}
}

func TestDemoHasThreeGrunts(t *testing.T) {
	w := NewDemoWorld()
	n := 0
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy && w.Units[i].Living() {
			n++
		}
	}
	if n != 3 {
		t.Fatalf("got %d grunts, want 3", n)
	}
}

func TestEnemyDoesNotShootThroughTree(t *testing.T) {
	w := aiWorld(Vec2{X: 8, Y: 8}, Vec2{X: 72, Y: 8})
	w.Map = Map{W: 6, H: 1, Tiles: []Tile{
		TileGrass, TileGrass, TileTree, TileTree, TileGrass, TileGrass,
	}}
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("grunt in range but blocked by trees must not fire")
	}
}

func aiWorld(player, enemy Vec2) *World {
	w := &World{Spread: 0, nextID: 1, rng: newRNG(), AI: true}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{player})
	w.SpawnUnit(SideEnemy, enemy)
	return w
}

package sim

import "testing"

func TestMGDoesNotHurtLivingFriendly(t *testing.T) {
	w := gunWorld(
		[]Vec2{{X: 0, Y: 0}, {X: 20, Y: 0}},
		Vec2{X: 50, Y: 0},
	)
	fireAt(w, 80, 0, 30)
	friendly := w.Unit(w.ActiveSquad().MemberIDs[1])
	if !friendly.Living() {
		t.Fatal("player MG must not kill a living friendly in the line of fire")
	}
	if dummy := enemyOf(w); dummy.Living() {
		t.Fatal("dummy behind the friendly should still die (shots pass through friendlies)")
	}
}

func TestMGKillsEnemyInOneHit(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 40, Y: 0})
	fireAt(w, 80, 0, 20)
	d := enemyOf(w)
	if d.Living() {
		t.Fatalf("dummy still alive at (%v,%v)", d.X, d.Y)
	}
	if !d.Dead() {
		t.Fatalf("want HP Dead, got %v", d.HP)
	}
}

func TestCorpseStaysOnMap(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 40, Y: 0})
	id := enemyOf(w).ID
	fireAt(w, 80, 0, 20)
	u := w.Unit(id)
	if u == nil {
		t.Fatal("corpse was removed from the world")
	}
	if !u.Dead() {
		t.Fatal("dummy should be a corpse")
	}
	if u.X != 40 || u.Y != 0 {
		t.Fatalf("corpse moved to (%v,%v)", u.X, u.Y)
	}
}

func TestDeadRemovedFromFile(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}, {X: -FileSpacing, Y: 0}}, Vec2{X: 200, Y: 200})
	s := w.ActiveSquad()
	follower := w.Unit(s.MemberIDs[1])
	w.kill(follower)
	s = w.ActiveSquad()
	if len(s.MemberIDs) != 1 {
		t.Fatalf("file still has %d members", len(s.MemberIDs))
	}
	if w.Unit(follower.ID) == nil || !w.Unit(follower.ID).Dead() {
		t.Fatal("dead follower should remain as a corpse")
	}
}

func TestAllLivingSquadMembersFire(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}, {X: 0, Y: 12}}, Vec2{X: 200, Y: 200})
	w.SetFire(80, 0, true)
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 2 {
		t.Fatalf("got %d projectiles, want 2 (both troopers)", len(w.Projectiles))
	}
}

func TestCorporalOutrangesPrivate(t *testing.T) {
	pvt := GunStatsFor(0)
	cpl := GunStatsFor(1)
	if cpl.Range <= pvt.Range {
		t.Fatalf("corporal range %v should exceed private %v", cpl.Range, pvt.Range)
	}
	if cpl.RoF <= pvt.RoF {
		t.Fatalf("corporal RoF %v should exceed private %v", cpl.RoF, pvt.RoF)
	}
	if cpl.Spread >= pvt.Spread {
		t.Fatalf("corporal spread %v should be tighter than private %v", cpl.Spread, pvt.Spread)
	}
	gen := GunStatsFor(15)
	if gen.Range != generalRange || gen.RoF != generalRoF {
		t.Fatalf("general stats %+v", gen)
	}
}

func TestCorporalProjectileOutrangesPrivate(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}, {X: 0, Y: 16}}, Vec2{X: 400, Y: 400})
	w.Unit(w.ActiveSquad().MemberIDs[0]).Rank = 0
	w.Unit(w.ActiveSquad().MemberIDs[1]).Rank = 1
	w.SetFire(80, 0, true)
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 2 {
		t.Fatalf("shots=%d", len(w.Projectiles))
	}
	a, b := w.Projectiles[0].Left, w.Projectiles[1].Left
	if a < b {
		a, b = b, a
	}
	if a <= b {
		t.Fatalf("remaining travel %v and %v; corporal shot should have more range left", a, b)
	}
}

func gunWorld(players []Vec2, dummy Vec2) *World {
	w := &World{Spread: 0, nextID: 1, rng: newRNG()}
	w.SpawnPlayerSquad(SquadSnake, players)
	w.SpawnUnit(SideEnemy, dummy)
	return w
}

func fireAt(w *World, x, y float64, frames int) {
	w.SetFire(x, y, true)
	for i := 0; i < frames; i++ {
		w.Step(1.0 / 60)
	}
}

func enemyOf(w *World) *Unit {
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy {
			return &w.Units[i]
		}
	}
	return nil
}

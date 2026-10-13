package sim

import (
	"math"
	"testing"
)

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

func TestMGCanWoundThenFinish(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 40, Y: 0})
	w.WoundChance = 1
	id := enemyOf(w).ID
	fireAt(w, 80, 0, 8)
	d := w.Unit(id)
	if !d.Wounded() {
		t.Fatalf("HP=%v, want wounded", d.HP)
	}
	if d.X != 40 || d.Y != 0 {
		t.Fatalf("wounded man moved to (%v,%v)", d.X, d.Y)
	}
	shooter := w.Unit(w.ActiveSquad().LeaderID)
	if shooter.Kills != 0 {
		t.Fatalf("a wound scored %d", shooter.Kills)
	}
	var yelled bool
	for _, c := range w.TakeCues() {
		if c.Kind == CueDeath {
			yelled = true
		}
	}
	if yelled {
		t.Fatal("a wound played the death yell")
	}
	fireAt(w, 80, 0, 12)
	d = w.Unit(id)
	if !d.Dead() {
		t.Fatalf("finish left HP=%v", d.HP)
	}
	if w.Unit(w.ActiveSquad().LeaderID).Kills != 1 {
		t.Fatalf("finish scored %d, want 1", w.Unit(w.ActiveSquad().LeaderID).Kills)
	}
}

func TestPlayerMGKillsInsteadOfWounding(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.WoundChance = 1
	vid := w.SpawnUnit(SidePlayer, Vec2{X: 40, Y: 0}).ID
	sid := w.SpawnUnit(SideEnemy, Vec2{X: 0, Y: 0}).ID
	shot := func() {
		w.Projectiles = append(w.Projectiles, Projectile{
			X: 30, Y: 0, VX: MGSpeed, VY: 0, Left: 80,
			Alive: true, OwnerID: sid, OwnerSide: SideEnemy,
		})
	}
	shot()
	w.Step(1.0 / 60)
	body := w.Unit(vid)
	if !body.Dead() {
		t.Fatalf("HP=%v, want dead", body.HP)
	}
	for i := 0; i < 90 && (body.Hop != 0 || body.VX != 0 || body.VY != 0 || body.VZ != 0); i++ {
		w.Step(1.0 / 60)
		body = w.Unit(vid)
	}
	dx := body.X - 40
	if dx < 10 || dx > 22 || body.Y != 0 || body.Hop != 0 {
		t.Fatalf("killing shot left him at (%v,%v) hop=%v", body.X, body.Y, body.Hop)
	}
	x, y := body.X, body.Y
	shot()
	for i := 0; i < 20; i++ {
		w.Step(1.0 / 60)
	}
	body = w.Unit(vid)
	if body.X != x || body.Y != y || body.Hop != 0 || body.VX != 0 || body.VY != 0 || body.VZ != 0 {
		t.Fatalf("later round moved a dead player: x=%v y=%v hop=%v v=(%v,%v,%v)", body.X, body.Y, body.Hop, body.VX, body.VY, body.VZ)
	}
}

func TestFinishWoundedFriendly(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}, {X: 20, Y: 0}}, Vec2{X: 200, Y: 200})
	friend := w.Unit(w.ActiveSquad().MemberIDs[1])
	w.wound(friend)
	if len(w.ActiveSquad().MemberIDs) != 1 {
		t.Fatal("a wounded man should drop out of the file")
	}
	fireAt(w, 80, 0, 12)
	if !w.Unit(friend.ID).Dead() {
		t.Fatal("player MG should finish a wounded friendly")
	}
}

func TestWoundedEnemyBlocksKillAll(t *testing.T) {
	w := NewDemoWorld()
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side == SideEnemy {
			w.wound(u)
		}
	}
	w.evaluateObjectives()
	if w.Status != Playing {
		t.Fatalf("status=%v, want Playing while an enemy is down", w.Status)
	}
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side == SideEnemy {
			w.kill(u)
		}
	}
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatalf("status=%v, want Won", w.Status)
	}
}

func TestBlastFinishesWounded(t *testing.T) {
	w := NewEmpty()
	u := w.SpawnUnit(SideEnemy, Vec2{X: 10, Y: 10})
	w.wound(u)
	w.explode(10, 10, 0)
	if !w.Unit(u.ID).Dead() {
		t.Fatal("a blast should finish a wounded man")
	}
}

func TestEnemyFliesBackWhenShot(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 40, Y: 0})
	id := enemyOf(w).ID
	w.Projectiles = append(w.Projectiles, Projectile{
		X: 20, Y: 0, VX: MGSpeed, VY: 0, Left: 80,
		Alive: true, OwnerSide: SidePlayer,
	})
	for i := 0; i < 6; i++ {
		w.Step(1.0 / 60)
	}
	body := w.Unit(id)
	if !body.Dead() {
		t.Fatal("setup should leave a corpse")
	}
	var hopped bool
	for i := 0; i < 90; i++ {
		if body.Hop > 1 {
			hopped = true
		}
		if body.VX == 0 && body.VY == 0 && body.VZ == 0 && body.Hop == 0 && hopped {
			break
		}
		w.Step(1.0 / 60)
		body = w.Unit(id)
	}
	if !hopped {
		t.Fatal("the killing shot should lift him off the ground")
	}
	dx := body.X - 40
	if dx < 10 || dx > 22 || math.Abs(body.Y) > 1 {
		t.Fatalf("flew to (%v,%v), want about one tile along the shot", body.X, body.Y)
	}
	if body.Hop != 0 || body.VX != 0 || body.VY != 0 || body.VZ != 0 {
		t.Fatalf("still moving hop=%v v=(%v,%v,%v)", body.Hop, body.VX, body.VY, body.VZ)
	}
}

func TestBlastThrowsMuchFarther(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	const bx, by = 160.0, 128.0
	on := w.SpawnUnit(SideEnemy, Vec2{X: bx, Y: by}).ID
	w.Unit(on).Facing = 0 // thrown west, back from the way he faced
	mid := w.SpawnUnit(SidePlayer, Vec2{X: bx + GrenadeRadius/2, Y: by}).ID
	edge := w.SpawnUnit(SideEnemy, Vec2{X: bx + GrenadeRadius*0.75, Y: by}).ID
	civil := w.SpawnUnit(SideCivilian, Vec2{X: bx + 8, Y: by}).ID
	w.explode(bx, by, 0)
	for i := 0; i < 180; i++ {
		w.Step(1.0 / 60)
	}
	flew := func(id int, x0, y0 float64) float64 {
		u := w.Unit(id)
		if u.Hop != 0 || u.VX != 0 || u.VY != 0 || u.VZ != 0 {
			t.Fatalf("id %d still moving at (%v,%v)", id, u.X, u.Y)
		}
		return hypot(u.X-x0, u.Y-y0)
	}
	full := flew(on, bx, by)
	if full < 80 || full > 130 || w.Unit(on).X >= bx {
		t.Fatalf("on the blast he landed at (%v,%v), travel %v", w.Unit(on).X, w.Unit(on).Y, full)
	}
	half := flew(mid, bx+GrenadeRadius/2, by)
	rim := flew(edge, bx+GrenadeRadius*0.75, by)
	if half >= full || rim >= half || rim < 1 {
		t.Fatalf("travel on=%v half=%v rim=%v", full, half, rim)
	}
	c := w.Unit(civil)
	if !c.Dead() || c.X != bx+8 || c.Y != by {
		t.Fatalf("civilian moved to (%v,%v)", c.X, c.Y)
	}
}

func TestCorpseJuggle(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 40, Y: 0})
	id := enemyOf(w).ID
	fireAt(w, 80, 0, 8)
	body := w.Unit(id)
	if !body.Dead() {
		t.Fatal("setup should leave a corpse")
	}
	x := body.X
	w.SetFire(200, 0, true)
	for i := 0; i < 40; i++ {
		w.Step(1.0 / 60)
		body = w.Unit(id)
		if body.Hop > 1 || body.X > x+1 {
			return
		}
	}
	t.Fatalf("corpse stayed at x=%v hop=%v", body.X, body.Hop)
}

func TestCorpseStopsReacting(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 40, Y: 0})
	id := enemyOf(w).ID
	w.kill(w.Unit(id))
	for w.Unit(id).DeadFor <= juggleLife {
		w.Step(1.0 / 30)
	}
	body := w.Unit(id)
	x, y := body.X, body.Y
	w.SetFire(200, 0, true)
	for i := 0; i < 40; i++ {
		w.Step(1.0 / 60)
	}
	body = w.Unit(id)
	if body.Hop != 0 || body.VX != 0 || body.VY != 0 || body.VZ != 0 || body.X != x || body.Y != y {
		t.Fatalf("settled corpse still moved: x=%v y=%v hop=%v v=(%v,%v,%v)", body.X, body.Y, body.Hop, body.VX, body.VY, body.VZ)
	}
}

func TestCorpseStaysOnMap(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 40, Y: 0})
	id := enemyOf(w).ID
	w.SetFire(80, 0, true)
	for i := 0; i < 20 && !w.Unit(id).Dead(); i++ {
		w.Step(1.0 / 60)
	}
	w.Firing = false
	w.Step(1.0 / 60)
	u := w.Unit(id)
	if u == nil {
		t.Fatal("corpse was removed from the world")
	}
	if !u.Dead() {
		t.Fatal("dummy should be a corpse")
	}
	if u.X < 40 || u.X > 40+22 || u.Y != 0 {
		t.Fatalf("corpse left the map at (%v,%v)", u.X, u.Y)
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

func TestMGStoppedByTree(t *testing.T) {
	w := gunWorld([]Vec2{{X: 8, Y: 8}}, Vec2{X: 72, Y: 8})
	w.Map = Map{W: 6, H: 1, Tiles: []Tile{
		TileGrass, TileGrass, TileTree, TileTree, TileGrass, TileGrass,
	}}
	fireAt(w, 200, 8, 40)
	if enemyOf(w).Living() {
		return
	}
	t.Fatal("MG must not kill through a tree line")
}

func TestMGFlankAroundTreeHits(t *testing.T) {
	// Wall in cols 2–3, rows 0–1; row 2 is open so a shot from the south is clear.
	tiles := make([]Tile, 6*3)
	for ty := 0; ty < 2; ty++ {
		tiles[ty*6+2] = TileTree
		tiles[ty*6+3] = TileTree
	}
	w := gunWorld([]Vec2{{X: 72, Y: 40}}, Vec2{X: 72, Y: 8})
	w.Map = Map{W: 6, H: 3, Tiles: tiles}
	fireAt(w, 72, 0, 40)
	if enemyOf(w).Living() {
		t.Fatal("flanking shot with clear LOS should kill")
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

package sim

import "testing"

func TestQuicksandSinkDeath(t *testing.T) {
	const W, H = 10, 10
	tiles := make([]Tile, W*H)
	tiles[4*W+4] = TileQuicksand
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = []Objective{KillAllEnemy}
	w.Map = Map{W: W, H: H, Tiles: tiles}
	c := TileCenter(4, 4)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{c})
	w.SpawnUnit(SideEnemy, TileCenter(8, 8))
	w.CommandMove(c.X+80, c.Y, true)
	id := w.ActiveSquad().LeaderID

	const dt = 1.0 / 60
	elapsed := 0.0
	for elapsed < SinkTime-0.05 {
		w.SetFire(c.X+40, c.Y, true)
		w.Step(dt)
		elapsed += dt
	}
	u := w.Unit(id)
	if !u.Living() {
		t.Fatal("sank to death before 2 seconds")
	}
	if !u.Sinking {
		t.Fatal("unit standing in quicksand should be trapped")
	}
	if hypot(u.X-c.X, u.Y-c.Y) > 1 {
		t.Fatalf("trapped unit walked to (%v,%v)", u.X, u.Y)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("a sinking unit must not fire")
	}
	for elapsed < SinkTime+0.1 {
		w.Step(dt)
		elapsed += dt
	}
	u = w.Unit(id)
	if !u.Dead() {
		t.Fatalf("want sink death, hp=%v sink=%v", u.HP, u.Sink)
	}
	if w.Status != Lost {
		t.Fatal("the last trooper sinking should fail the phase")
	}
}

func TestWalkingOntoQuicksandTraps(t *testing.T) {
	const (
		W, H = 10, 12
		col  = 5
		row  = 5
		dt   = 1.0 / 60
	)
	tiles := make([]Tile, W*H)
	tiles[row*W+col] = TileQuicksand
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.Map = Map{W: W, H: H, Tiles: tiles}
	start := TileCenter(col, 8)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{start})
	w.CommandMove(TileCenter(col, 2).X, TileCenter(col, 2).Y, true)
	id := w.ActiveSquad().LeaderID

	south := float64((row + 1) * TileSize) // y where the pool begins (walking north)
	elapsed := 0.0
	var sinkStart float64
	sinking := false
	for i := 0; i < 60*6; i++ {
		w.Step(dt)
		elapsed += dt
		u := w.Unit(id)
		if u.Y < float64(row*TileSize)-1 {
			t.Fatalf("walked through quicksand to y=%v", u.Y)
		}
		if u.Sinking && !sinking {
			sinking = true
			sinkStart = elapsed
		}
		if u.Dead() {
			if !sinking {
				t.Fatal("died without sinking")
			}
			held := elapsed - sinkStart
			if held < SinkTime-0.1 || held > SinkTime+0.1 {
				t.Fatalf("sink lasted %v, want ~%v", held, SinkTime)
			}
			if u.Y > south {
				t.Fatalf("died before entering the pool, y=%v", u.Y)
			}
			return
		}
	}
	t.Fatal("unit never sank")
}

func TestMineExplodesLikeGrenade(t *testing.T) {
	const W, H = 20, 16
	tiles := make([]Tile, W*H)
	tiles[8*W+8] = TileMine
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.Map = Map{W: W, H: H, Tiles: tiles}
	mine := TileCenter(8, 8)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{TileCenter(8, 12)})
	nearID := w.SpawnUnit(SideEnemy, Vec2{X: mine.X + 16, Y: mine.Y}).ID
	farID := w.SpawnUnit(SideEnemy, Vec2{X: mine.X + 80, Y: mine.Y}).ID
	w.AddBuilding(8, 6, 2, 2, false, 0)
	id := w.ActiveSquad().LeaderID
	w.CommandMove(mine.X, mine.Y, true)

	for i := 0; i < 180; i++ {
		w.Step(1.0 / 60)
		if w.Unit(id).Dead() {
			break
		}
	}
	if w.Unit(id).Living() {
		t.Fatal("walker should die on the mine")
	}
	if w.Unit(nearID).Living() {
		t.Fatal("mine blast should kill a unit inside grenade radius")
	}
	if !w.Unit(farID).Living() {
		t.Fatal("mine blast should spare a unit outside the radius")
	}
	if w.Map.At(8, 8) == TileMine {
		t.Fatal("mine tile should be spent")
	}
	if w.Buildings[0].Alive {
		t.Fatal("mine blast should destroy a hut inside the radius")
	}
	again := w.SpawnUnit(SideEnemy, mine).ID
	for i := 0; i < 10; i++ {
		w.Step(1.0 / 60)
	}
	if !w.Unit(again).Living() {
		t.Fatal("a spent mine must not explode again")
	}
}

func TestShootingCivilianDoesNotFailPhase(t *testing.T) {
	w := gunWorld([]Vec2{{X: 8, Y: 8}}, Vec2{X: 80, Y: 80})
	w.Objectives = []Objective{KillAllEnemy}
	civID := w.SpawnUnit(SideCivilian, Vec2{X: 40, Y: 8}).ID
	fireAt(w, 80, 8, 30)
	if c := w.Unit(civID); c == nil || !c.Dead() {
		t.Fatal("MG should kill a civilian")
	}
	if w.Status != Playing {
		t.Fatalf("killing a civilian must not end the phase, status=%v", w.Status)
	}
	if !enemyOf(w).Living() {
		t.Fatal("enemy off the shot should live")
	}
	w.kill(enemyOf(w))
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatal("kill-all still wins after a civilian dies")
	}
}

func TestProtectCiviliansIsReserved(t *testing.T) {
	o, ok := ParseObjective("protect_civilians")
	if !ok || o != ProtectCivilians {
		t.Fatalf("parse protect_civilians = %v ok=%v", o, ok)
	}
	w := gunWorld([]Vec2{{X: 8, Y: 8}}, Vec2{X: 200, Y: 8})
	w.Objectives = []Objective{KillAllEnemy, ProtectCivilians}
	civID := w.SpawnUnit(SideCivilian, Vec2{X: 40, Y: 80}).ID
	w.kill(w.Unit(civID))
	w.evaluateObjectives()
	if w.Status != Playing {
		t.Fatalf("protect_civilians must not change the phase, status=%v", w.Status)
	}
}

func TestCivilianWandersAndDoesNotFight(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	id := w.SpawnUnit(SideCivilian, Vec2{X: 160, Y: 120}).ID
	x0, y0 := w.Unit(id).X, w.Unit(id).Y
	for i := 0; i < 180; i++ {
		w.Step(1.0 / 60)
	}
	u := w.Unit(id)
	if !u.Living() {
		t.Fatal("civilian should still be alive")
	}
	if hypot(u.X-x0, u.Y-y0) < 4 {
		t.Fatalf("civilian should wander, stayed at (%v,%v)", u.X, u.Y)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("a civilian must not shoot")
	}
	if !w.Unit(w.ActiveSquad().LeaderID).Living() {
		t.Fatal("a civilian must not hurt the squad")
	}
}

package sim

import "testing"

func TestPlayerShotEmitsOneGunCue(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 200, Y: 200})
	w.SetFire(80, 0, true)
	w.Step(1.0 / 60)
	cues := w.TakeCues()
	if len(cues) != 1 {
		t.Fatalf("cues=%d, want 1", len(cues))
	}
	if cues[0].Kind != CueGun {
		t.Fatalf("kind=%v, want gun", cues[0].Kind)
	}
	// Muzzle is half a trooper in front of the unit, along the aim.
	if cues[0].X != float64(UnitSize)/2 || cues[0].Y != 0 {
		t.Fatalf("cue at (%v,%v), want the muzzle", cues[0].X, cues[0].Y)
	}
	w.Step(1.0 / 60)
	if again := w.TakeCues(); len(again) != 0 {
		t.Fatalf("cooldown frame emitted %d cues", len(again))
	}
}

func TestEachLivingTrooperEmitsAGunCue(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}, {X: 0, Y: 12}}, Vec2{X: 200, Y: 200})
	w.SetFire(80, 0, true)
	w.Step(1.0 / 60)
	if len(w.TakeCues()) != 2 {
		t.Fatal("each living trooper should emit one gun cue")
	}
}

func TestVehicleGunEmitsCue(t *testing.T) {
	w := NewEmpty()
	v := w.AddSkidoo(Vec2{X: 10, Y: 20}, SidePlayer, true)
	w.vehicleShoot(v, 80, 20, 1)
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueGun {
		t.Fatalf("cues=%v, want one gun", cues)
	}
	if cues[0].X != 20 || cues[0].Y != 20 {
		t.Fatalf("mounted muzzle (%v,%v), want (20,20)", cues[0].X, cues[0].Y)
	}
	if len(w.TakeCues()) != 0 {
		t.Fatal("TakeCues should empty the queue")
	}
}

func TestExplodeEmitsOneBoom(t *testing.T) {
	w := NewEmpty()
	w.explode(8, 12, 0)
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueBoom {
		t.Fatalf("cues=%v, want one boom", cues)
	}
	if cues[0].X != 8 || cues[0].Y != 12 {
		t.Fatalf("boom at (%v,%v), want the blast point", cues[0].X, cues[0].Y)
	}
}

func TestGrenadeLandingEmitsBoom(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.launchGrenade(&Unit{X: 0, Y: 0}, 40, 0)
	if len(w.TakeCues()) != 0 {
		t.Fatal("the throw stays silent")
	}
	var booms []Cue
	for i := 0; i < 120; i++ {
		w.Step(1.0 / 60)
		for _, c := range w.TakeCues() {
			if c.Kind != CueBoom {
				t.Fatalf("in-flight cue %v", c.Kind)
			}
			booms = append(booms, c)
		}
	}
	if len(booms) != 1 || booms[0].X != 40 || booms[0].Y != 0 {
		t.Fatalf("booms=%v, want one at the landing point", booms)
	}
}

func TestCrateChainBoomsForEachCrate(t *testing.T) {
	w := NewEmpty()
	w.AddGrenadeCrate(Vec2{X: 0, Y: 0})
	w.AddGrenadeCrate(Vec2{X: 10, Y: 0})
	w.Pickups[0].Alive = false // the round already destroyed this crate
	w.explode(0, 0, 0)
	cues := w.TakeCues()
	if len(cues) != 2 {
		t.Fatalf("cues=%d, want the blast and the crate it sets off", len(cues))
	}
	for _, c := range cues {
		if c.Kind != CueBoom {
			t.Fatalf("kind=%v", c.Kind)
		}
	}
	if cues[0].X != 0 || cues[1].X != 10 {
		t.Fatalf("chain at %v then %v", cues[0].X, cues[1].X)
	}
}

func TestKillEmitsOneDeathCue(t *testing.T) {
	w := NewEmpty()
	u := w.SpawnUnit(SideCivilian, Vec2{X: 10, Y: 20})
	id := u.ID
	w.kill(u)
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueDeath || cues[0].ID != id {
		t.Fatalf("cues=%v, want one death for unit %d", cues, id)
	}
	if cues[0].X != 10 || cues[0].Y != 20 {
		t.Fatalf("yell at (%v,%v)", cues[0].X, cues[0].Y)
	}
	w.kill(u)
	if again := w.TakeCues(); len(again) != 0 {
		t.Fatalf("second kill emitted %d cues", len(again))
	}
}

func TestMGKillEmitsGunThenDeath(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 48, Y: 0})
	enemy := enemyOf(w)
	id := enemy.ID
	w.SetFire(48, 0, true)
	var kinds []CueKind
	var death Cue
	for i := 0; i < 30; i++ {
		w.Step(1.0 / 60)
		for _, c := range w.TakeCues() {
			kinds = append(kinds, c.Kind)
			if c.Kind == CueDeath {
				death = c
			}
		}
		if death.Kind == CueDeath {
			break
		}
	}
	if len(kinds) < 2 || kinds[0] != CueGun {
		t.Fatalf("kinds=%v, want the crack before the yell", kinds)
	}
	if death.Kind != CueDeath || death.ID != id || death.X != 48 {
		t.Fatalf("death=%v, want the grunt at x=48", death)
	}
	nDeath := 0
	for _, k := range kinds {
		if k == CueDeath {
			nDeath++
		}
	}
	if nDeath != 1 {
		t.Fatalf("deaths=%d in %v", nDeath, kinds)
	}
}

func TestExplosionBoomsOnceAndYellsPerMan(t *testing.T) {
	w := NewEmpty()
	a := w.SpawnUnit(SideEnemy, Vec2{X: 0, Y: 0})
	b := w.SpawnUnit(SideCivilian, Vec2{X: 8, Y: 0})
	w.explode(4, 0, 0)
	cues := w.TakeCues()
	if len(cues) != 3 || cues[0].Kind != CueBoom || cues[1].Kind != CueDeath || cues[2].Kind != CueDeath {
		t.Fatalf("cues=%v, want one boom and two yells", cues)
	}
	if cues[0].X != 4 || cues[1].ID != a.ID || cues[2].ID != b.ID {
		t.Fatalf("cues=%v", cues)
	}
}

func TestQuicksandDeathYells(t *testing.T) {
	tiles := make([]Tile, 4)
	tiles[0] = TileQuicksand
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.Map = Map{W: 2, H: 2, Tiles: tiles}
	c := TileCenter(0, 0)
	u := w.SpawnUnit(SideEnemy, c)
	id := u.ID
	deaths := 0
	for i := 0; i < int(SinkTime*60)+8; i++ {
		w.Step(1.0 / 60)
		for _, cue := range w.TakeCues() {
			if cue.Kind != CueDeath || cue.ID != id {
				t.Fatalf("cue %+v", cue)
			}
			deaths++
		}
	}
	if deaths != 1 || !w.Unit(id).Dead() {
		t.Fatalf("deaths=%d dead=%v", deaths, w.Unit(id).Dead())
	}
}

func TestThrowAndLaunchStaySilent(t *testing.T) {
	w := NewEmpty()
	w.launchGrenade(&Unit{X: 0, Y: 0}, 40, 0)
	w.launchRocket(0, SidePlayer, 0, 0, 40, 0)
	if len(w.Cues) != 0 {
		t.Fatalf("leaving the hand emitted %d cues", len(w.Cues))
	}
}

func TestCueQueueDropsOldest(t *testing.T) {
	w := NewEmpty()
	n := maxCues + 5
	for i := 0; i < n; i++ {
		w.emit(CueGun, float64(i), 0)
	}
	if len(w.Cues) != maxCues {
		t.Fatalf("len=%d, want %d", len(w.Cues), maxCues)
	}
	if w.Cues[0].X != 5 || w.Cues[maxCues-1].X != float64(n-1) {
		t.Fatalf("queue spans %v..%v", w.Cues[0].X, w.Cues[maxCues-1].X)
	}
}

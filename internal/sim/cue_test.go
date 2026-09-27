package sim

import "testing"

func TestSceneStingsAreNotWorldCues(t *testing.T) {
	if CueClick == CueGun || CueWin == CueGun || CueFail == CueGun {
		t.Fatal("a scene sting shares the gun kind")
	}
	if CueClick == CueWin || CueWin == CueFail || CueClick == CueFail || CueClick == CueNone {
		t.Fatal("click, win, and fail should be three kinds")
	}
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 20, Y: 20}, {X: 8, Y: 20}})
	w.SpawnUnit(SideEnemy, Vec2{X: 80, Y: 20})
	w.SetFire(80, 20, true)
	w.Step(1.0 / 60)
	w.explode(80, 20, 0)
	w.launchGrenade(&Unit{X: 20, Y: 20}, 90, 20)
	v := w.AddSkidoo(Vec2{X: 20, Y: 20}, SidePlayer, true)
	if !w.CommandBoard(v.ID) || !w.CommandExit(v.ID) {
		t.Fatal("board and exit should emit their own cues")
	}
	for _, c := range w.TakeCues() {
		switch c.Kind {
		case CueClick, CueWin, CueFail:
			t.Fatalf("world emitted %v", c.Kind)
		}
	}
}

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
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueThrow || cues[0].X != 0 || cues[0].Y != 0 {
		t.Fatalf("cues=%v, want one throw at the hand", cues)
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
	deaths, sinks := 0, 0
	for i := 0; i < int(SinkTime*60)+8; i++ {
		w.Step(1.0 / 60)
		for _, cue := range w.TakeCues() {
			switch cue.Kind {
			case CueSink:
				sinks++
			case CueDeath:
				if cue.ID != id {
					t.Fatalf("death cue %+v", cue)
				}
				deaths++
			default:
				t.Fatalf("cue %+v", cue)
			}
		}
	}
	if sinks != 1 || deaths != 1 || !w.Unit(id).Dead() {
		t.Fatalf("sinks=%d deaths=%d dead=%v", sinks, deaths, w.Unit(id).Dead())
	}
}

func TestSplashOnEnteringWater(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Map = Map{W: 3, H: 1, Tiles: []Tile{TileGrass, TileWaterShallow, TileWaterDeep}}
	u := w.SpawnUnit(SideEnemy, TileCenter(0, 0))
	w.refreshTerrain()
	if len(w.TakeCues()) != 0 {
		t.Fatal("the first reading of dry ground should be silent")
	}
	u.X, u.Y = TileCenter(1, 0).X, TileCenter(1, 0).Y
	w.refreshTerrain()
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueSplash || cues[0].X != u.X {
		t.Fatalf("wade cues=%v", cues)
	}
	w.refreshTerrain()
	if len(w.TakeCues()) != 0 {
		t.Fatal("standing in the water should be silent")
	}
	u.X = TileCenter(2, 0).X
	w.refreshTerrain()
	if len(w.TakeCues()) != 0 {
		t.Fatal("shallow to deep is still in the water")
	}
	u.X = TileCenter(0, 0).X
	w.refreshTerrain()
	if len(w.TakeCues()) != 0 {
		t.Fatal("leaving the water should be silent")
	}
	u.X = TileCenter(2, 0).X
	w.refreshTerrain()
	cues = w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueSplash {
		t.Fatalf("swim cues=%v", cues)
	}
}

func TestQuicksandEmitsOneSink(t *testing.T) {
	tiles := make([]Tile, 4)
	tiles[1] = TileQuicksand
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.Map = Map{W: 2, H: 2, Tiles: tiles}
	u := w.SpawnUnit(SideEnemy, TileCenter(0, 0))
	w.Step(1.0 / 60)
	if len(w.TakeCues()) != 0 {
		t.Fatal("dry ground should not sink")
	}
	c := TileCenter(1, 0)
	u.X, u.Y = c.X, c.Y
	w.Step(1.0 / 60)
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueSink || cues[0].X != c.X || cues[0].Y != c.Y {
		t.Fatalf("cues=%v", cues)
	}
	w.Step(1.0 / 60)
	if again := w.TakeCues(); len(again) != 0 {
		t.Fatalf("still sinking emitted %v", again)
	}
}

func TestCratePickupEmitsOnce(t *testing.T) {
	for _, kind := range []PickupKind{PickupGrenades, PickupRockets} {
		w := NewEmpty()
		w.AI = false
		w.Objectives = nil
		at := Vec2{X: 40, Y: 40}
		w.SpawnPlayerSquad(SquadSnake, []Vec2{at})
		if kind == PickupRockets {
			w.AddRocketCrate(at)
		} else {
			w.AddGrenadeCrate(at)
		}
		w.Step(1.0 / 60)
		cues := w.TakeCues()
		if len(cues) != 1 || cues[0].Kind != CuePickup || cues[0].X != at.X || cues[0].Y != at.Y {
			t.Fatalf("kind %v cues=%v", kind, cues)
		}
		w.Step(1.0 / 60)
		if again := w.TakeCues(); len(again) != 0 {
			t.Fatalf("kind %v emitted again %v", kind, again)
		}
	}
}

func TestBoardAndExitEmitOnce(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 20, Y: 20}, {X: 10, Y: 20}})
	v := w.AddSkidoo(Vec2{X: 20, Y: 20}, SidePlayer, true)
	if !w.CommandBoard(v.ID) {
		t.Fatal("board")
	}
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueBoard || cues[0].X != v.X || cues[0].Y != v.Y {
		t.Fatalf("board cues=%v", cues)
	}
	if !w.CommandExit(v.ID) {
		t.Fatal("exit")
	}
	cues = w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueExit || cues[0].X != 20 || cues[0].Y != 20 {
		t.Fatalf("exit cues=%v", cues)
	}
}

func TestRocketWhooshThenBoom(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.launchRocket(0, SidePlayer, 0, 0, 40, 0)
	cues := w.TakeCues()
	tube := float64(UnitSize)
	if len(cues) != 1 || cues[0].Kind != CueRocket || cues[0].X != tube || cues[0].Y != 0 {
		t.Fatalf("cues=%v, want the whoosh at the tube", cues)
	}
	var booms []Cue
	for i := 0; i < 40; i++ {
		w.Step(1.0 / 60)
		for _, c := range w.TakeCues() {
			if c.Kind != CueBoom {
				t.Fatalf("in-flight cue %v", c.Kind)
			}
			booms = append(booms, c)
		}
	}
	if len(booms) != 1 || booms[0].X != 40 || booms[0].Y != 0 {
		t.Fatalf("booms=%v, want one at the impact", booms)
	}
}

func TestGrenadierWindupIsSilent(t *testing.T) {
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 60, Y: 0})
	e := enemyOf(w)
	e.Kind = KindGrenadier
	e.Bombs = 1
	e.GrenadeCD = 0
	w.Step(1.0 / 60)
	if e.GrenadeWind <= 0 {
		t.Fatal("expected the yellow telegraph")
	}
	if cues := w.TakeCues(); len(cues) != 0 {
		t.Fatalf("windup emitted %v", cues)
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

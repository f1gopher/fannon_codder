package sim

import "testing"

func TestChopperHoldsFiresAndFalls(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	w.AI = true
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 160, Y: 40}})
	v := w.AddChopper(Vec2{X: 40, Y: 40}, true, nil)
	if v.Kind != KindChopper || v.Side != SideEnemy || !v.Armed || len(v.Occupants) != 0 {
		t.Fatalf("chopper %+v", v)
	}
	if w.CommandBoard(v.ID) {
		t.Fatal("nothing boards a chopper")
	}
	x, y := v.X, v.Y
	w.Step(0.2)
	v = w.vehicleByID(v.ID)
	if v.X != x || v.Y != y {
		t.Fatal("with no route the chopper holds its tile")
	}
	homing := func() int {
		n := 0
		for _, p := range w.Projectiles {
			if p.Kind == ProjHoming {
				n++
			}
		}
		return n
	}
	if homing() != 0 {
		t.Fatal("the reaction should hold the first rocket")
	}
	for i := 0; i < 50; i++ {
		w.Step(1.0 / 60)
	}
	if homing() < 1 {
		t.Fatal("expected a homing rocket")
	}
	var whistle, boom int
	for _, c := range w.Cues {
		if c.Kind == CueHoming {
			whistle++
		}
		if c.Kind == CueBoom {
			boom++
		}
	}
	if whistle < 1 || boom != 0 {
		t.Fatalf("whistle %d boom %d", whistle, boom)
	}
	if w.EngineHumAt(v.X, v.Y, 320).Run {
		t.Fatal("a chopper does not use the engine hum")
	}
	if !w.ChopperNear(v.X, v.Y, 320) {
		t.Fatal("a living chopper should be heard")
	}
	w.addMG(v.X-30, v.Y, 0, 80, 1, SidePlayer)
	w.Step(1.0 / 60)
	if !w.vehicleByID(v.ID).Alive {
		t.Fatal("machine-gun fire should miss the chopper")
	}
	w.explodeRocket(v.X, v.Y, 0)
	v = w.vehicleByID(v.ID)
	if v.Alive || v.Alt <= 0 {
		t.Fatalf("the wreck should be falling %+v", v)
	}
	var crash int
	for _, c := range w.TakeCues() {
		if c.Kind == CueCrash {
			crash++
		}
	}
	if crash != 1 {
		t.Fatalf("crash cues %d", crash)
	}
	for i := 0; i < 80; i++ {
		w.Step(1.0 / 60)
	}
	v = w.vehicleByID(v.ID)
	if v.Alt != 0 {
		t.Fatalf("alt %v", v.Alt)
	}
	landed := 0
	for _, c := range w.Cues {
		if c.Kind == CueBoom {
			landed++
		}
	}
	if landed < 1 {
		t.Fatal("the boom plays when the wreck hits")
	}
	if w.ChopperNear(v.X, v.Y, 320) {
		t.Fatal("a wreck should not spin the rotor up")
	}
}

func TestUnarmedChopperDoesNotShoot(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	w.AI = true
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 80, Y: 40}})
	w.AddChopper(Vec2{X: 40, Y: 40}, false, nil)
	w.Step(2)
	for _, p := range w.Projectiles {
		if p.Kind == ProjHoming {
			t.Fatal("an unarmed chopper should not launch")
		}
	}
	for _, c := range w.Cues {
		if c.Kind == CueHoming {
			t.Fatal("an unarmed chopper has no launch cue")
		}
	}
}

func TestChopperPassAndRouteIgnoreWater(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	w.AI = true
	tiles := make([]Tile, 20*8)
	for i := range tiles {
		tiles[i] = TileWaterDeep
	}
	w.Map = Map{W: 20, H: 8, Tiles: tiles}
	man := w.SpawnUnit(SidePlayer, Vec2{X: 40, Y: 40})
	v := w.AddChopper(Vec2{X: 40, Y: 40}, false, [][2]int{{12, 2}})
	w.Step(1.0 / 60)
	if man.Living() {
		t.Fatal("a low pass should kill the man underneath")
	}
	x := v.X
	w.Step(0.4)
	v = w.vehicleByID(v.ID)
	if v.X <= x {
		t.Fatalf("water should not stop the chopper, x %v -> %v", x, v.X)
	}
}

func TestHomingTurnsTowardTheMan(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	w.AI = false
	man := w.SpawnUnit(SidePlayer, Vec2{X: 40, Y: 10})
	v := w.AddChopper(Vec2{X: 40, Y: 80}, true, nil)
	w.launchHoming(v, man.ID, man.X, 80)
	w.Step(0.35)
	var p *Projectile
	for i := range w.Projectiles {
		if w.Projectiles[i].Kind == ProjHoming && w.Projectiles[i].Alive {
			p = &w.Projectiles[i]
		}
	}
	if p == nil {
		t.Fatal("the rocket should still be in the air")
	}
	if p.VY >= 0 {
		t.Fatalf("the rocket should be turning north, vy %v", p.VY)
	}
}

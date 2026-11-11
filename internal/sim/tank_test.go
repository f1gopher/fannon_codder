package sim

import "testing"

func TestTankIsSlowAndFiresAShell(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	v := w.AddTank(Vec2{X: 48, Y: 40}, SidePlayer, true)
	if !w.CommandBoard(v.ID) {
		t.Fatal("the squad should board the tank")
	}
	w.SetDrive(400, 40, true)
	for i := 0; i < 100; i++ {
		w.SetDrive(400, 40, true)
		w.Step(1.0 / 60)
	}
	v = w.vehicleByID(v.ID)
	speed := hypot(v.VX, v.VY)
	top := VehicleMaxSpeed * TankSpeedMul
	if speed > top+1 || speed < top*0.85 {
		t.Fatalf("tank speed %v, top %v", speed, top)
	}
	h := w.EngineHumAt(v.X, v.Y, 320)
	if !h.Run || h.Voice != KindTank {
		t.Fatalf("hum %+v", h)
	}
	if h.Pitch < EngineTopPitch-0.05 {
		t.Fatalf("pitch should use the tank's own top speed, got %v", h.Pitch)
	}
	w.SetFire(v.X+80, v.Y, true)
	w.Step(1.0 / 60)
	var shells int
	for _, p := range w.Projectiles {
		if p.Kind == ProjShell {
			shells++
		}
		if p.Kind == ProjMG {
			t.Fatal("a tank should not fire the mounted gun")
		}
	}
	if shells != 1 {
		t.Fatalf("shells %d", shells)
	}
	cues := w.TakeCues()
	var shell, boom int
	for _, c := range cues {
		if c.Kind == CueShell {
			shell++
		}
		if c.Kind == CueBoom {
			boom++
		}
	}
	if shell != 1 || boom != 0 {
		t.Fatalf("launch cues shell %d boom %d", shell, boom)
	}
}

func TestGrenadeClanksOnATankAndAShellKillsIt(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	w.AI = false
	tank := w.AddTank(Vec2{X: 80, Y: 40}, SideEnemy, true)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 20, Y: 40}})
	w.explode(80, 40, w.ActiveSquad().LeaderID)
	if !w.vehicleByID(tank.ID).Alive {
		t.Fatal("a grenade should leave the tank")
	}
	var clank bool
	for _, c := range w.TakeCues() {
		if c.Kind == CueClank {
			clank = true
		}
	}
	if !clank {
		t.Fatal("a grenade on a tank should clank")
	}
	ski := w.AddSkidoo(Vec2{X: 200, Y: 40}, SideEnemy, false)
	w.AddFactory(12, 2, 4, 3)
	w.explodeShell(200, 40, w.ActiveSquad().LeaderID)
	w.explodeShell(80, 40, w.ActiveSquad().LeaderID)
	if w.vehicleByID(ski.ID).Alive || w.vehicleByID(tank.ID).Alive {
		t.Fatal("a shell should destroy a skidoo and a tank")
	}
	if w.Buildings[0].Alive {
		t.Fatal("a shell should destroy a factory")
	}
	w.AddComputer(2, 2, 1, 1)
	w.explode(2*TileSize+8, 2*TileSize+8, w.ActiveSquad().LeaderID)
	if !w.Buildings[len(w.Buildings)-1].Alive {
		t.Fatal("a grenade should leave a computer")
	}
}

func TestEnemyTankHoldsAndFiresShells(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 220, Y: 40}})
	v := w.AddTank(Vec2{X: 40, Y: 40}, SideEnemy, true)
	id := v.ID
	var held bool
	for i := 0; i < 400; i++ {
		keepPlayerAlive(w)
		w.Step(1.0 / 60)
		v = w.vehicleByID(id)
		d := hypot(v.X-220, v.Y-40)
		if d <= TankStandoff+8 {
			held = true
			if d <= VehicleRamR {
				t.Fatalf("the tank rammed, dist %v", d)
			}
			break
		}
		for _, p := range w.Projectiles {
			if p.Kind == ProjShell {
				t.Fatal("fired while still closing")
			}
		}
	}
	if !held {
		t.Fatal("the tank should reach standoff")
	}
	var shells int
	for i := 0; i < 180; i++ {
		keepPlayerAlive(w)
		w.Step(1.0 / 60)
		for _, p := range w.Projectiles {
			if p.Kind == ProjMG {
				t.Fatal("enemy tank used the infantry gun")
			}
			if p.Kind == ProjShell {
				shells++
			}
		}
		w.Projectiles = nil
	}
	if shells < 1 {
		t.Fatal("the tank should fire a shell from the hold")
	}
	v = w.vehicleByID(id)
	if d := hypot(v.X-220, v.Y-40); d <= VehicleRamR {
		t.Fatalf("held inside the ram, dist %v", d)
	}
}

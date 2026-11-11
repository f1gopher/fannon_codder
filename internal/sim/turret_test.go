package sim

import "testing"

func TestTurretHoldsAndFiresAShell(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	w.AI = true
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 180, Y: 40}})
	v := w.AddTurret(Vec2{X: 40, Y: 40})
	if v.Side != SideEnemy || !v.Armed || len(v.Occupants) != 0 {
		t.Fatalf("turret %+v", v)
	}
	if w.CommandBoard(v.ID) {
		t.Fatal("a turret cannot be boarded")
	}
	x, y := v.X, v.Y
	w.Step(0.3)
	v = w.vehicleByID(v.ID)
	if v.X != x || v.Y != y || hypot(v.VX, v.VY) != 0 {
		t.Fatal("a turret should stay put")
	}
	shells := func() int {
		n := 0
		for _, p := range w.Projectiles {
			if p.Kind == ProjShell {
				n++
			}
			if p.Kind == ProjMG {
				t.Fatal("a turret should not fire the mounted gun")
			}
		}
		return n
	}
	if shells() != 0 {
		t.Fatal("the reaction should hold the first shell")
	}
	w.Step(0.45)
	w.Step(1.0 / 60)
	if shells() < 1 {
		t.Fatal("expected a shell after the reaction")
	}
	var launch int
	for _, c := range w.Cues {
		if c.Kind == CueShell {
			launch++
		}
		if c.Kind == CueBoard || c.Kind == CueBoom {
			t.Fatalf("unexpected cue %v", c.Kind)
		}
	}
	if launch < 1 {
		t.Fatal("launch should be CueShell")
	}
	h := w.EngineHumAt(v.X, v.Y, 320)
	if h.Run {
		t.Fatalf("a turret has no engine %+v", h)
	}
	w.explode(v.X, v.Y, 0)
	if !w.vehicleByID(v.ID).Alive {
		t.Fatal("a grenade should leave the turret")
	}
	var clank bool
	for _, c := range w.Cues {
		if c.Kind == CueClank {
			clank = true
		}
	}
	if !clank {
		t.Fatal("a grenade on a turret should clank")
	}
	w.addMG(v.X-20, v.Y, 0, 40, 1, SidePlayer)
	w.Step(1.0 / 60)
	if !w.vehicleByID(v.ID).Alive {
		t.Fatal("machine-gun fire should leave the turret")
	}
	w.explodeShell(v.X, v.Y, 0)
	if w.vehicleByID(v.ID).Alive {
		t.Fatal("a shell should destroy the turret")
	}
}

func TestTurretStaysEnemyWhenLoaded(t *testing.T) {
	w := NewEmpty()
	w.Objectives = nil
	v := w.AddTurret(Vec2{X: 16, Y: 16})
	if v.Kind != KindTurret || v.Side != SideEnemy || !v.Armed {
		t.Fatalf("%+v", v)
	}
}

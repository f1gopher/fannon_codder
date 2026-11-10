package sim

import "testing"

func TestHostageStandsUntilTouched(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Objectives = []Objective{RescueHostages}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	id := w.SpawnUnit(SideHostage, Vec2{X: 120, Y: 40}).ID
	w.CommandMove(200, 40, true)
	for i := 0; i < 30; i++ {
		w.Step(1.0 / 60)
	}
	h := w.Unit(id)
	if h.Escorting || h.X != 120 {
		t.Fatalf("hostage moved before a touch: escort=%v x=%v", h.Escorting, h.X)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("a hostage fired")
	}
	leader := w.Unit(w.ActiveSquad().LeaderID)
	leader.X, leader.Y = h.X, h.Y
	w.Step(1.0 / 60)
	if !w.Unit(id).Escorting {
		t.Fatal("a touch should start the escort")
	}
}

func TestHostageEscortsThenStops(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 80}})
	id := w.SpawnUnit(SideHostage, Vec2{X: 40, Y: 80}).ID
	w.Step(1.0 / 60)
	if !w.Unit(id).Escorting {
		t.Fatal("touch")
	}
	w.CommandMove(200, 80, true)
	for i := 0; i < 90; i++ {
		w.Step(1.0 / 60)
	}
	leader := w.Unit(w.ActiveSquad().LeaderID)
	h := w.Unit(id)
	if h.X >= leader.X-FileSpacing*0.5 {
		t.Fatalf("escort should stay behind the leader: h=%v lead=%v", h.X, leader.X)
	}
	hx := h.X
	w.ActiveSquad().HasDest = false
	w.Step(1.0 / 60)
	if w.Unit(id).X != hx {
		t.Fatalf("a stopped squad should hold the escort, x=%v want %v", w.Unit(id).X, hx)
	}
}

func TestRescueHostagesNeedsTheTent(t *testing.T) {
	o, ok := ParseObjective("rescue_hostages")
	if !ok || o != RescueHostages {
		t.Fatalf("parse = %v ok=%v", o, ok)
	}
	w := NewEmpty()
	w.AI = false
	w.Objectives = []Objective{RescueHostages}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	h := w.SpawnUnit(SideHostage, Vec2{X: 80, Y: 40})
	h.Escorting = true
	w.AddTent(4, 2, 2, 2)
	w.evaluateObjectives()
	if w.Status != Playing {
		t.Fatal("an unsaved hostage must keep the phase open")
	}
	c := TileCenter(4, 2)
	h.X, h.Y = c.X, c.Y+4
	w.Step(1.0 / 60)
	if !h.Saved || w.Status != Won {
		t.Fatalf("saved=%v status=%v", h.Saved, w.Status)
	}
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueSave {
		t.Fatalf("cues=%v", cues)
	}
}

func TestHostageDeathLosesThePhase(t *testing.T) {
	w := NewEmpty()
	w.Objectives = []Objective{RescueHostages, KillAllEnemy}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	id := w.SpawnUnit(SideHostage, Vec2{X: 40, Y: 40}).ID
	w.SpawnUnit(SideEnemy, Vec2{X: 200, Y: 40})
	w.kill(w.Unit(id))
	w.evaluateObjectives()
	if w.Status != Lost {
		t.Fatalf("status=%v", w.Status)
	}
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueCry {
		t.Fatalf("cues=%v", cues)
	}
}

func TestTentIgnoresAGrenade(t *testing.T) {
	w := NewEmpty()
	id := w.AddTent(2, 2, 2, 2)
	b := &w.Buildings[id]
	w.explode(b.X+b.W/2, b.Y+b.H/2, 0)
	if !b.Alive || len(w.Roofs) != 0 {
		t.Fatalf("tent took the blast alive=%v roofs=%d", b.Alive, len(w.Roofs))
	}
	if w.anyDoorBuilding() {
		t.Fatal("a tent counts as a door hut")
	}
}

func TestHostageDoesNotBlockKillAll(t *testing.T) {
	w := NewEmpty()
	w.Objectives = []Objective{KillAllEnemy}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	w.SpawnUnit(SideHostage, Vec2{X: 40, Y: 40})
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatalf("a hostage blocked kill-all, status=%v", w.Status)
	}
}

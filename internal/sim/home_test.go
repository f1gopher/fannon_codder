package sim

import "testing"

func TestHomeCivilianStandsUntilTouched(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Objectives = []Objective{GetCivilianHome}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	u := w.SpawnUnit(SideCivilian, Vec2{X: 120, Y: 40})
	u.MustHome = true
	id := u.ID
	w.CommandMove(200, 40, true)
	for i := 0; i < 30; i++ {
		w.Step(1.0 / 60)
	}
	h := w.Unit(id)
	if h.Escorting || h.X != 120 {
		t.Fatalf("home civilian moved before a touch: escort=%v x=%v", h.Escorting, h.X)
	}
	leader := w.Unit(w.ActiveSquad().LeaderID)
	leader.X, leader.Y = h.X, h.Y
	w.Step(1.0 / 60)
	if !w.Unit(id).Escorting {
		t.Fatal("a touch should start the escort")
	}
}

func TestGetCivilianHomeNeedsTheCottage(t *testing.T) {
	o, ok := ParseObjective("get_civilian_home")
	if !ok || o != GetCivilianHome {
		t.Fatalf("parse = %v ok=%v", o, ok)
	}
	w := NewEmpty()
	w.AI = false
	w.Objectives = []Objective{GetCivilianHome}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	h := w.SpawnUnit(SideCivilian, Vec2{X: 80, Y: 40})
	h.MustHome = true
	h.Escorting = true
	wander := w.SpawnUnit(SideCivilian, Vec2{X: 200, Y: 40})
	w.AddHome(4, 2, 2, 2)
	w.evaluateObjectives()
	if w.Status != Playing {
		t.Fatal("an unsaved home civilian must keep the phase open")
	}
	c := TileCenter(4, 2)
	w.Unit(h.ID).X, w.Unit(h.ID).Y = c.X, c.Y+4
	w.Step(1.0 / 60)
	h = w.Unit(h.ID)
	if !h.Saved || w.Status != Won {
		t.Fatalf("saved=%v status=%v wander=%v", h.Saved, w.Status, wander.Saved)
	}
	if wander.Saved {
		t.Fatal("a wandering civilian is not delivered")
	}
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueSave {
		t.Fatalf("cues=%v", cues)
	}
}

func TestHomeCivilianDeathLosesThePhase(t *testing.T) {
	w := NewEmpty()
	w.Objectives = []Objective{GetCivilianHome}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	id := w.SpawnUnit(SideCivilian, Vec2{X: 40, Y: 40}).ID
	w.Unit(id).MustHome = true
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

func TestWanderingDeathDoesNotFailTheHomeObjective(t *testing.T) {
	w := NewEmpty()
	w.Objectives = []Objective{GetCivilianHome}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	home := w.SpawnUnit(SideCivilian, Vec2{X: 40, Y: 40})
	home.MustHome = true
	home.Saved = true
	other := w.SpawnUnit(SideCivilian, Vec2{X: 80, Y: 40})
	w.kill(other)
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatalf("status=%v", w.Status)
	}
}

func TestHomeIgnoresAGrenade(t *testing.T) {
	w := NewEmpty()
	id := w.AddHome(2, 2, 2, 2)
	b := &w.Buildings[id]
	w.explode(b.X+b.W/2, b.Y+b.H/2, 0)
	if !b.Alive || len(w.Roofs) != 0 {
		t.Fatalf("home took the blast alive=%v roofs=%d", b.Alive, len(w.Roofs))
	}
}

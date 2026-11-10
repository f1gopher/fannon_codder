package sim

import "testing"

func TestLeaderFightsUntilTouched(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Objectives = []Objective{KillAllEnemy, KidnapLeader}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	id := w.SpawnUnit(SideEnemy, Vec2{X: 200, Y: 40}).ID
	w.Unit(id).Kind = KindLeader
	gid := w.SpawnUnit(SideEnemy, Vec2{X: 300, Y: 40}).ID
	for i := 0; i < 20; i++ {
		w.Step(1.0 / 60)
	}
	leader := w.Unit(id)
	if leader.Captured || leader.Side != SideEnemy || leader.X != 200 {
		t.Fatalf("leader left his post: %+v", leader)
	}
	man := w.Unit(w.ActiveSquad().LeaderID)
	man.X, man.Y = leader.X, leader.Y
	w.Step(1.0 / 60)
	leader = w.Unit(id)
	if !leader.Captured || leader.Side != SideCaptured {
		t.Fatalf("touch should capture: side=%v captured=%v", leader.Side, leader.Captured)
	}
	n := len(w.Projectiles)
	for i := 0; i < 40; i++ {
		w.Step(1.0 / 60)
	}
	if len(w.Projectiles) != n {
		t.Fatal("a captured leader fired")
	}
	w.kill(w.Unit(gid))
	w.TakeCues()
	w.Step(1.0 / 60)
	if w.anyRemaining(SideEnemy) {
		t.Fatal("the captured leader still blocks kill_all_enemy")
	}
}

func TestLeaderHandoverSaves(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = []Objective{KidnapLeader}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 80}})
	id := w.SpawnUnit(SideEnemy, Vec2{X: 40, Y: 80}).ID
	w.Unit(id).Kind = KindLeader
	w.AddHandover(4, 4, 2, 2)
	w.Step(1.0 / 60)
	leader := w.Unit(id)
	if !leader.Captured {
		t.Fatal("touch")
	}
	b := &w.Buildings[0]
	leader.X, leader.Y = b.X+4, b.Y+4
	w.Step(1.0 / 60)
	leader = w.Unit(id)
	if !leader.Saved {
		t.Fatal("handover should save him")
	}
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueSave {
		t.Fatalf("cues=%v", cues)
	}
	w.Step(1.0 / 60)
	if w.Status != Won {
		t.Fatalf("status=%v", w.Status)
	}
}

func TestLeaderDeathLosesThePhase(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = []Objective{KidnapLeader}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	id := w.SpawnUnit(SideEnemy, Vec2{X: 80, Y: 40}).ID
	w.Unit(id).Kind = KindLeader
	w.kill(w.Unit(id))
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueDeath {
		t.Fatalf("cues=%v", cues)
	}
	w.Step(1.0 / 60)
	if w.Status != Lost {
		t.Fatalf("status=%v", w.Status)
	}
}

func TestLeaderDeathWithoutTheObjective(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = []Objective{KillAllEnemy}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	id := w.SpawnUnit(SideEnemy, Vec2{X: 80, Y: 40}).ID
	w.Unit(id).Kind = KindLeader
	w.kill(w.Unit(id))
	w.Step(1.0 / 60)
	if w.Status != Won {
		t.Fatalf("status=%v, a phase with no kidnap objective should not lose", w.Status)
	}
}

func TestHandoverDoesNotBurn(t *testing.T) {
	w := NewEmpty()
	w.AddHandover(2, 2, 2, 2)
	w.explode(w.Buildings[0].X+8, w.Buildings[0].Y+8, 0)
	if !w.Buildings[0].Alive {
		t.Fatal("a blast destroyed the handover")
	}
}

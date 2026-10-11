package sim

import "testing"

func TestLateShotAfterTheEnemyIsDownIsAWipe(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 120, Y: 80}})
	foe := w.SpawnUnit(SideEnemy, Vec2{X: 240, Y: 200})
	w.kill(foe)
	w.addMG(40, 80, 0, 200, foe.ID, SideEnemy)
	w.evaluateObjectives()
	if w.Status != Playing {
		t.Fatalf("status %v while a round is still in the air", w.Status)
	}
	for i := 0; i < 40; i++ {
		w.Step(1.0 / 60)
	}
	if w.anyLiving(SidePlayer) {
		t.Fatal("the round should still kill the player")
	}
	if w.Status != Lost {
		t.Fatalf("status %v, want Lost", w.Status)
	}
}

func TestLateShotThatMissesStillWins(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 120, Y: 80}})
	foe := w.SpawnUnit(SideEnemy, Vec2{X: 240, Y: 200})
	w.kill(foe)
	w.addMG(40, 200, 0, 30, foe.ID, SideEnemy)
	if !w.HoldOutcome() {
		t.Fatal("a round still flying should hold the result")
	}
	for i := 0; i < 20; i++ {
		w.Step(1.0 / 60)
	}
	if !w.anyLiving(SidePlayer) {
		t.Fatal("the round was not aimed at the squad")
	}
	if w.Status != Won {
		t.Fatalf("status %v, want Won", w.Status)
	}
}

func TestBlastHoldsTheResult(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 120, Y: 80}})
	foe := w.SpawnUnit(SideEnemy, Vec2{X: 240, Y: 200})
	w.kill(foe)
	w.Explosions = append(w.Explosions, Explosion{X: 10, Y: 10, R: GrenadeRadius})
	w.evaluateObjectives()
	if w.Status != Playing {
		t.Fatalf("status %v while the blast is still playing", w.Status)
	}
	w.Step(blastTime)
	if w.Status != Won {
		t.Fatalf("status %v after the blast, want Won", w.Status)
	}
}

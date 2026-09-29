package sim

import (
	"math"
	"testing"
)

func quietWorld() *World {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	return w
}

func TestAddMGZeroesSinceShot(t *testing.T) {
	w := quietWorld()
	u := w.SpawnUnit(SidePlayer, Vec2{X: 10, Y: 10})
	if u.SinceShot < 0.12 {
		t.Fatalf("fresh SinceShot=%v, want it past the shot window", u.SinceShot)
	}
	w.addMG(u.X, u.Y, 0, 40, u.ID, u.Side)
	if u.SinceShot != 0 {
		t.Fatalf("SinceShot=%v after addMG, want 0", u.SinceShot)
	}
	const dt = 1.0 / 60
	w.Step(dt)
	if math.Abs(u.SinceShot-dt) > 1e-12 {
		t.Fatalf("SinceShot=%v after one step, want %v", u.SinceShot, dt)
	}
}

func TestLaunchGrenadeZeroesSinceThrow(t *testing.T) {
	w := quietWorld()
	s := w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 20, Y: 20}})
	u := w.Unit(s.LeaderID)
	if u.SinceThrow < 0.25 {
		t.Fatalf("fresh SinceThrow=%v, want it past the throw flourish", u.SinceThrow)
	}
	w.launchGrenade(u, 80, 20)
	if u.SinceThrow != 0 {
		t.Fatalf("SinceThrow=%v after launchGrenade, want 0", u.SinceThrow)
	}
	const dt = 1.0 / 60
	w.Step(dt)
	if math.Abs(u.SinceThrow-dt) > 1e-12 {
		t.Fatalf("SinceThrow=%v after one step, want %v", u.SinceThrow, dt)
	}
}

func TestLaunchRocketZeroesSinceThrow(t *testing.T) {
	w := quietWorld()
	u := w.SpawnUnit(SidePlayer, Vec2{X: 20, Y: 20})
	w.launchRocket(u.ID, u.Side, u.X, u.Y, 80, 20)
	if u.SinceThrow != 0 {
		t.Fatalf("SinceThrow=%v after launchRocket, want 0", u.SinceThrow)
	}
}

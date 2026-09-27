package app

import (
	"testing"

	"fannon-codder/internal/sim"
)

func TestHearPointIsTheLeader(t *testing.T) {
	w := sim.NewDemoWorld()
	u := w.Unit(w.ActiveSquad().LeaderID)
	x, y := hearPoint(w)
	if x != u.X || y != u.Y {
		t.Fatalf("hear point (%v,%v), leader (%v,%v)", x, y, u.X, u.Y)
	}
}

func TestHearPointFallsBackToCameraCentre(t *testing.T) {
	w := sim.NewDemoWorld()
	w.Unit(w.ActiveSquad().LeaderID).HP = sim.Dead
	w.Camera.X = 10
	w.Camera.Y = 20
	w.Camera.ViewW = 268
	w.Camera.ViewH = 256
	x, y := hearPoint(w)
	if x != 10+134 || y != 20+128 {
		t.Fatalf("hear point (%v,%v), want camera centre (144,148)", x, y)
	}
}

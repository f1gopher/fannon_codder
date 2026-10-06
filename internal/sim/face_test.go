package sim

import (
	"math"
	"testing"
)

func TestActiveSquadFacesThePointer(t *testing.T) {
	w := twoManWorld(Vec2{X: 40, Y: 40}, Vec2{X: 40 - FileSpacing, Y: 40})
	w.SetFire(40, 8, false)
	w.Step(1.0 / 60)
	s := w.ActiveSquad()
	for _, id := range s.MemberIDs {
		u := w.Unit(id)
		want := math.Atan2(8-u.Y, 40-u.X)
		if math.Abs(wrapAngle(u.Facing-want)) > 1e-9 {
			t.Fatalf("unit %d facing %v, want %v toward the pointer", id, u.Facing, want)
		}
	}
}

func TestActiveSquadFacesThePointerWhileWalking(t *testing.T) {
	w := twoManWorld(Vec2{X: 40, Y: 40}, Vec2{X: 40 - FileSpacing, Y: 40})
	w.CommandMove(120, 40, true)
	w.SetFire(40, 8, false)
	w.Step(1.0 / 60)
	s := w.ActiveSquad()
	leader := w.Unit(s.LeaderID)
	if leader.X <= 40 {
		t.Fatal("leader did not walk east")
	}
	want := math.Atan2(8-leader.Y, 40-leader.X)
	if math.Abs(wrapAngle(leader.Facing-want)) > 1e-9 {
		t.Fatalf("leader facing %v, want the pointer %v while walking east", leader.Facing, want)
	}
}

func TestParkedSquadDoesNotTrackThePointer(t *testing.T) {
	w := twoManWorld(Vec2{X: 40, Y: 40}, Vec2{X: 40 - FileSpacing, Y: 40})
	w.Selected = []int{w.ActiveSquad().MemberIDs[1]}
	if !w.Split() {
		t.Fatal("split")
	}
	var parked *Squad
	for i := range w.Squads {
		if !w.Squads[i].Active {
			parked = &w.Squads[i]
		}
	}
	if parked == nil {
		t.Fatal("no parked squad")
	}
	u := w.Unit(parked.MemberIDs[0])
	u.Facing = 0
	w.SetFire(40, 8, false)
	w.Step(1.0 / 60)
	u = w.Unit(parked.MemberIDs[0])
	if u.Facing != 0 {
		t.Fatalf("parked facing %v, want east", u.Facing)
	}
}

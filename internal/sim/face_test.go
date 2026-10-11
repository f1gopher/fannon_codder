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

func TestParkedSquadFansTheWatch(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	w.SpawnPlayerSquad(SquadEagle, []Vec2{
		{X: 80, Y: 40},
		{X: 80 - FileSpacing, Y: 40},
		{X: 80 - 2*FileSpacing, Y: 40},
	})
	w.Step(1.0 / 60)
	s := w.SquadByID(SquadEagle)
	if !s.WatchSet {
		t.Fatal("a parked file should take a watch heading")
	}
	var prev float64
	for i, id := range s.MemberIDs {
		u := w.Unit(id)
		want := wrapAngle(float64(i) * 2 * math.Pi / 3)
		if math.Abs(wrapAngle(u.Facing-want)) > 1e-9 {
			t.Fatalf("man %d facing %v, want %v", i, u.Facing, want)
		}
		if i > 0 && math.Abs(wrapAngle(u.Facing-prev)) < 0.2 {
			t.Fatalf("man %d faces the same way as the man ahead", i)
		}
		prev = u.Facing
	}
	w.Step(1.0 / 60)
	for i, id := range s.MemberIDs {
		u := w.Unit(id)
		want := wrapAngle(float64(i) * 2 * math.Pi / 3)
		if math.Abs(wrapAngle(u.Facing-want)) > 1e-9 {
			t.Fatalf("watch drifted on man %d to %v", i, u.Facing)
		}
	}
}

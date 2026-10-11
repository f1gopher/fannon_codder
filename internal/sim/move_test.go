package sim

import "testing"

func TestLeaderReachesDestination(t *testing.T) {
	w := twoManWorld(Vec2{X: 0, Y: 0}, Vec2{X: -FileSpacing, Y: 0})
	w.CommandMove(90, 0, true)
	stepUntilIdle(w, 400)
	leader := w.Unit(w.ActiveSquad().LeaderID)
	if hypot(leader.X-90, leader.Y-0) > ArrivalRadius {
		t.Fatalf("leader at (%v,%v), want near (90,0)", leader.X, leader.Y)
	}
	if leader.VX != 0 || leader.VY != 0 {
		t.Fatalf("leader still moving: vx=%v vy=%v", leader.VX, leader.VY)
	}
}

func TestFollowerStaysBehindNotStacked(t *testing.T) {
	w := twoManWorld(Vec2{X: 0, Y: 0}, Vec2{X: -FileSpacing, Y: 0})
	w.CommandMove(120, 0, true)
	stepUntilIdle(w, 500)
	s := w.ActiveSquad()
	leader := w.Unit(s.LeaderID)
	follower := w.Unit(s.MemberIDs[1])
	d := hypot(follower.X-leader.X, follower.Y-leader.Y)
	if d < FileSpacing*0.6 {
		t.Fatalf("stacked: distance=%v, want at least %v", d, FileSpacing*0.6)
	}
	if d > FileSpacing*1.8 {
		t.Fatalf("too far back: distance=%v", d)
	}
	if follower.X >= leader.X-1 {
		t.Fatalf("follower should remain west of leader, follower=%v leader=%v", follower.X, leader.X)
	}
	destDist := hypot(follower.X-120, follower.Y-0)
	if destDist < ArrivalRadius {
		t.Fatal("follower sat on the destination instead of trailing")
	}
}

func TestSpawnFileHoldsWithoutAMove(t *testing.T) {
	w := twoManWorld(Vec2{X: 40, Y: 40}, Vec2{X: 40 - FileSpacing, Y: 40})
	for i := 0; i < 120; i++ {
		w.Step(1.0 / 60)
	}
	s := w.ActiveSquad()
	leader := w.Unit(s.LeaderID)
	follower := w.Unit(s.MemberIDs[1])
	if hypot(leader.X-40, leader.Y-40) > 0.1 {
		t.Fatalf("leader drifted to (%v,%v)", leader.X, leader.Y)
	}
	d := hypot(follower.X-leader.X, follower.Y-leader.Y)
	if d < FileSpacing*0.8 {
		t.Fatalf("spawn stacked the file, distance=%v", d)
	}
}

func TestShortTrailStillSpacesTheFile(t *testing.T) {
	head := Vec2{X: 50, Y: 20}
	trail := []Vec2{{X: 50, Y: 20}}
	a, _ := trailPoint(head, trail, FileSpacing)
	b, _ := trailPoint(head, trail, FileSpacing*2)
	if hypot(a.X-head.X, a.Y-head.Y) < FileSpacing*0.9 {
		t.Fatalf("first slot stacked on the leader: %+v", a)
	}
	if hypot(a.X-b.X, a.Y-b.Y) < FileSpacing*0.9 {
		t.Fatalf("slots share a point: %+v %+v", a, b)
	}
}

func TestJoinedSquadDoesNotStack(t *testing.T) {
	w := threeMan()
	guard := w.ActiveSquad().MemberIDs[2]
	w.ToggleSelect(guard)
	if !w.Split() {
		t.Fatal("split")
	}
	if !w.SetActiveSquad(SquadSnake) {
		t.Fatal("snake")
	}
	g := w.Unit(guard)
	w.CommandMove(g.X, g.Y, true)
	for i := 0; i < 300; i++ {
		w.Step(1.0 / 60)
		if len(w.Squads) == 1 && !w.ActiveSquad().HasDest {
			break
		}
	}
	if len(w.Squads) != 1 {
		t.Fatal("squads should have joined")
	}
	for i := 0; i < 180; i++ {
		w.Step(1.0 / 60)
	}
	s := w.ActiveSquad()
	for i := 0; i < len(s.MemberIDs); i++ {
		a := w.Unit(s.MemberIDs[i])
		for j := i + 1; j < len(s.MemberIDs); j++ {
			b := w.Unit(s.MemberIDs[j])
			if d := hypot(a.X-b.X, a.Y-b.Y); d < FileSpacing*0.6 {
				t.Fatalf("men %d and %d share a spot, distance=%v", a.ID, b.ID, d)
			}
		}
	}
}

func TestFileAlongPathNotJustOffset(t *testing.T) {
	// Walk south, then the follower should also be on that north-south line.
	w := twoManWorld(Vec2{X: 40, Y: 40}, Vec2{X: 40, Y: 40 - FileSpacing})
	w.CommandMove(40, 160, true)
	stepUntilIdle(w, 500)
	s := w.ActiveSquad()
	leader := w.Unit(s.LeaderID)
	follower := w.Unit(s.MemberIDs[1])
	if hypot(follower.X-leader.X, 0) > 3 {
		t.Fatalf("file should stay on the path (x≈40): follower=(%v,%v) leader=(%v,%v)",
			follower.X, follower.Y, leader.X, leader.Y)
	}
	if follower.Y >= leader.Y-1 {
		t.Fatalf("follower should be north of leader, fy=%v ly=%v", follower.Y, leader.Y)
	}
}

func twoManWorld(a, b Vec2) *World {
	w := &World{nextID: 1}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{a, b})
	return w
}

func stepUntilIdle(w *World, maxSteps int) {
	dt := 1.0 / 60
	for i := 0; i < maxSteps; i++ {
		w.Step(dt)
		s := w.ActiveSquad()
		if s == nil || s.HasDest {
			continue
		}
		allStopped := true
		for _, id := range s.MemberIDs {
			u := w.Unit(id)
			if u.VX != 0 || u.VY != 0 {
				allStopped = false
				break
			}
		}
		if allStopped {
			return
		}
	}
}

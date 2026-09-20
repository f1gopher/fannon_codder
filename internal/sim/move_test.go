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

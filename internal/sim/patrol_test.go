package sim

import (
	"math"
	"testing"
)

func TestPatrolWalksTheLineAndLooksAround(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	e := w.SpawnUnit(SideEnemy, Vec2{})
	w.setPatrol(e, []Vec2{{}, {X: 120}})
	id := e.ID
	var reached, returned, paused bool
	var still float64
	for i := 0; i < 60*40; i++ {
		w.Step(1.0 / 60)
		e = w.Unit(id)
		if math.Abs(e.Y) > 1 {
			t.Fatalf("left the line, at (%v,%v)", e.X, e.Y)
		}
		if e.X > 110 {
			reached = true
		}
		if reached && e.X < 16 {
			returned = true
		}
		if e.VX == 0 && e.VY == 0 && e.X > 12 && e.X < 108 {
			still += 1.0 / 60
			if still > 0.6 {
				paused = true
			}
		} else {
			still = 0
		}
	}
	if !reached || !returned {
		t.Fatalf("route reached=%v returned=%v x=%v dir=%d", reached, returned, e.X, e.PatrolDir)
	}
	if !paused {
		t.Fatal("he never stopped along the line to look around")
	}
}

func TestPatrolPauseLooksBothWays(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	e := w.SpawnUnit(SideEnemy, Vec2{})
	w.setPatrol(e, []Vec2{{}, {X: 80}})
	e.PatrolWait = 2
	e.PatrolSpan = 2
	e.PatrolHead = 0
	e.Facing = 0
	id := e.ID
	w.Step(0.5)
	e = w.Unit(id)
	if e.X != 0 || e.VX != 0 || e.VY != 0 {
		t.Fatalf("a look-around holds him, x=%v v=(%v,%v)", e.X, e.VX, e.VY)
	}
	if e.Facing < 0.7 {
		t.Fatalf("he should look off the route, facing %v", e.Facing)
	}
	w.Step(1)
	e = w.Unit(id)
	if e.Facing > -0.5 {
		t.Fatalf("he should look the other way, facing %v", e.Facing)
	}
	if e.PatrolWait <= 0 {
		t.Fatal("the pause should still be running")
	}
}

func TestPatrolStopsToShoot(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Spread = 0
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 50}})
	e := w.SpawnUnit(SideEnemy, Vec2{})
	w.setPatrol(e, []Vec2{{}, {X: 200}})
	id := e.ID
	for i := 0; i < 120; i++ {
		keepPlayerAlive(w)
		w.Step(1.0 / 60)
	}
	e = w.Unit(id)
	if e.X > 4 {
		t.Fatalf("gun range should take him off the line, x=%v", e.X)
	}
	if e.BurstN == 0 && len(w.Projectiles) == 0 {
		t.Fatal("he should fire once he has stopped")
	}
}

func TestPatrolStaysOnTheLinePastSight(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Mission = 1
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 260}})
	e := w.SpawnUnit(SideEnemy, Vec2{})
	w.setPatrol(e, []Vec2{{}, {Y: 80}})
	id := e.ID
	w.Step(1)
	e = w.Unit(id)
	if math.Abs(e.X) > 1 {
		t.Fatalf("past sight he keeps the line, x=%v", e.X)
	}
	if e.Y < patrolSpeed-4 || e.Y > patrolSpeed+4 {
		t.Fatalf("he should stroll the line at %v px/s, y=%v", patrolSpeed, e.Y)
	}
}

func TestPatrolInvestigatesASighting(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Mission = 1
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 120}})
	e := w.SpawnUnit(SideEnemy, Vec2{})
	w.setPatrol(e, []Vec2{{}, {Y: 80}})
	id := e.ID
	w.Step(1)
	e = w.Unit(id)
	if math.Abs(e.X-WalkSpeed) > 2 {
		t.Fatalf("a sighting should pull him off the line at chase speed, x=%v", e.X)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("he is not in gun range yet")
	}
}

func TestPatrolCurveLeavesTheStraightLeg(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	e := w.SpawnUnit(SideEnemy, Vec2{})
	w.setPatrol(e, []Vec2{{}, {X: 100}, {X: 100, Y: 100}})
	if len(e.Patrol) < 8 {
		t.Fatalf("a curve should be sampled, points=%d", len(e.Patrol))
	}
	e.PatrolGap = 30
	id := e.ID
	var bowed bool
	for i := 0; i < 120; i++ {
		w.Step(1.0 / 60)
		e = w.Unit(id)
		if e.X > 15 && e.X < 80 && math.Abs(e.Y) > 2 {
			bowed = true
		}
	}
	if !bowed {
		t.Fatalf("the walk should bow off the straight leg, at (%v,%v)", e.X, e.Y)
	}
	last := e.Patrol[len(e.Patrol)-1]
	if last.X != 100 || last.Y != 100 {
		t.Fatalf("the curve should still end on the last point, %+v", last)
	}
}

func TestPatrolChasesThenReturnsToTheLine(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 150}})
	e := w.SpawnUnit(SideEnemy, Vec2{})
	e.Aggressive = true
	w.setPatrol(e, []Vec2{{}, {Y: 90}})
	id := e.ID
	w.Step(1)
	e = w.Unit(id)
	if e.X < 20 {
		t.Fatalf("an acquired patrol should close, x=%v", e.X)
	}
	for i := range w.Units {
		if w.Units[i].Side == SidePlayer {
			w.Units[i].X = 800
		}
	}
	left := e.X
	var back bool
	for i := 0; i < 180; i++ {
		w.Step(1.0 / 60)
		e = w.Unit(id)
		if e.X < left-10 {
			back = true
			break
		}
	}
	if !back {
		t.Fatalf("once the squad is out of sight he should return, x %v -> %v", left, e.X)
	}
}

func TestPatrolStopsForAHeardShot(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	tiles := make([]Tile, 4*6)
	for i := range tiles {
		tiles[i] = TileGrass
	}
	tiles[2*4+1] = TileTree
	w.Map = Map{W: 4, H: 6, Tiles: tiles}
	player := w.SpawnUnit(SidePlayer, TileCenter(2, 2))
	e := w.SpawnUnit(SideEnemy, TileCenter(0, 2))
	w.setPatrol(e, []Vec2{TileCenter(0, 2), TileCenter(0, 5)})
	e.HearID = player.ID
	id := e.ID
	x0, y0 := e.X, e.Y
	for i := 0; i < 24; i++ {
		w.Step(1.0 / 60)
	}
	e = w.Unit(id)
	if e.X < x0+2 {
		t.Fatalf("he should walk toward the shot until the tree, x %v -> %v", x0, e.X)
	}
	if math.Abs(e.Y-y0) > 1 {
		t.Fatalf("the tree blocks the straight walk, y %v -> %v", y0, e.Y)
	}
	if e.Facing > 0.4 {
		t.Fatalf("he should face the shot, facing %v", e.Facing)
	}
	if e.BurstN != 0 || len(w.Projectiles) != 0 {
		t.Fatal("hearing through a tree does not fire")
	}
}

func TestPatrolPausesAndLooksAtTheEnd(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	e := w.SpawnUnit(SideEnemy, Vec2{X: 118})
	w.setPatrol(e, []Vec2{{}, {X: 120}})
	e.PatrolGap = 10
	id := e.ID
	for i := 0; i < 30 && w.Unit(id).PatrolWait == 0; i++ {
		w.Step(1.0 / 60)
	}
	e = w.Unit(id)
	if e.PatrolDir != -1 || e.PatrolTo != 0 {
		t.Fatalf("the end of the line sends him back, dir=%d to=%d", e.PatrolDir, e.PatrolTo)
	}
	if e.PatrolWait < patrolPauseMin || e.PatrolWait > patrolPauseMax {
		t.Fatalf("end pause %v, want %v–%v s", e.PatrolWait, patrolPauseMin, patrolPauseMax)
	}
	wait := e.PatrolWait
	var looked bool
	face0 := e.Facing
	// Stay on the end for the whole look. The last frame is allowed to step off.
	frames := int(wait*60) - 1
	if frames < 1 {
		t.Fatalf("pause too short: %v", wait)
	}
	for i := 0; i < frames; i++ {
		w.Step(1.0 / 60)
		e = w.Unit(id)
		if math.Abs(e.X-120) > 1 || e.VX != 0 || e.VY != 0 {
			t.Fatalf("left the end during the look, frame %d x=%v v=(%v,%v)", i, e.X, e.VX, e.VY)
		}
		if math.Abs(e.Facing-face0) > 0.4 {
			looked = true
		}
	}
	if !looked {
		t.Fatal("he should look around before walking back")
	}
	for e.PatrolWait > 0 {
		w.Step(1.0 / 60)
		e = w.Unit(id)
		if math.Abs(e.X-120) > 1 {
			t.Fatalf("left the end before the pause finished, x=%v", e.X)
		}
	}
	w.Step(1)
	e = w.Unit(id)
	if e.X > 120-patrolSpeed+2 {
		t.Fatalf("after the pause he should walk back, x=%v", e.X)
	}
}

func TestPatrolWalksSlowerThanAChase(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	e := w.SpawnUnit(SideEnemy, Vec2{})
	w.setPatrol(e, []Vec2{{}, {X: 200}})
	e.PatrolGap = 30
	id := e.ID
	w.Step(1)
	e = w.Unit(id)
	if math.Abs(e.X-patrolSpeed) > 1 {
		t.Fatalf("moved %v px in one second, want %v", e.X, patrolSpeed)
	}
	if patrolSpeed >= WalkSpeed {
		t.Fatal("patrol speed should be slower than a chase")
	}
}

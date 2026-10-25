package sim

import (
	"fmt"
	"math"
)

const (
	patrolSpeed     = 16   // px/s. Slower than WalkSpeed, so the route reads as a stroll.
	patrolArrive    = 0.75 // px. Tight, so a short curve sample does not get snapped across.
	patrolCurveStep = 8    // world px between samples of a curved line
	patrolPauseMin  = 1.5
	patrolPauseMax  = 4.0
	patrolLook      = 1.05 // radians either side of the route while he pauses
	patrolGapMin    = 2.6  // seconds of walking between mid-route pauses
	patrolGapMax    = 5.2
)

// SetPatrolFromTiles makes u walk tile centres back and forth.
// Two points are a straight line. Three or more are a curve through those
// centres. The samples he actually walks must be walkable. He leaves the
// line to investigate a player or a shot, then comes back.
func (w *World) SetPatrolFromTiles(u *Unit, tiles [][2]int) error {
	if u == nil {
		return fmt.Errorf("patrol missing soldier")
	}
	if len(tiles) < 2 {
		return fmt.Errorf("patrol needs a line")
	}
	pts := make([]Vec2, len(tiles))
	for i, t := range tiles {
		if !w.Map.active() || t[0] < 0 || t[1] < 0 || t[0] >= w.Map.W || t[1] >= w.Map.H {
			return fmt.Errorf("patrol point %d,%d is off the map", t[0], t[1])
		}
		if i > 0 && t == tiles[i-1] {
			return fmt.Errorf("patrol line repeats %d,%d", t[0], t[1])
		}
		if !w.Map.WalkableTile(w.Map.At(t[0], t[1])) {
			return fmt.Errorf("patrol point %d,%d is not walkable", t[0], t[1])
		}
		pts[i] = TileCenter(t[0], t[1])
	}
	if len(pts) == 2 && !w.pathClear(pts[0].X, pts[0].Y, pts[1].X, pts[1].Y) {
		a, b := tiles[0], tiles[1]
		return fmt.Errorf("patrol line %d,%d to %d,%d is blocked", a[0], a[1], b[0], b[1])
	}
	curve := PatrolCurve(pts)
	for _, p := range curve {
		if !w.patrolSampleOK(p) {
			tx := int(math.Floor(p.X / float64(TileSize)))
			ty := int(math.Floor(p.Y / float64(TileSize)))
			return fmt.Errorf("patrol curve through %d,%d is blocked", tx, ty)
		}
	}
	w.assignPatrol(u, curve)
	return nil
}

func (w *World) patrolSampleOK(p Vec2) bool {
	if !w.Map.active() {
		return true
	}
	mw, mh := w.Map.PixelSize()
	if p.X < 0 || p.Y < 0 || p.X >= mw || p.Y >= mh {
		return false
	}
	return w.tileWalkable(p.X, p.Y) && w.Map.Walkable(p.X, p.Y, unitHalf)
}

// PatrolCurve samples a route through pts. Two points stay a straight segment.
// More points are a Catmull-Rom curve, so the walk bends instead of cutting
// each corner. The first and last points are the ends he pauses on.
func PatrolCurve(pts []Vec2) []Vec2 {
	if len(pts) < 3 {
		return append([]Vec2(nil), pts...)
	}
	out := []Vec2{pts[0]}
	for i := 0; i < len(pts)-1; i++ {
		p0 := pts[i]
		if i > 0 {
			p0 = pts[i-1]
		}
		p1 := pts[i]
		p2 := pts[i+1]
		p3 := p2
		if i+2 < len(pts) {
			p3 = pts[i+2]
		}
		seg := hypot(p2.X-p1.X, p2.Y-p1.Y)
		n := int(math.Ceil(seg / patrolCurveStep))
		if n < 2 {
			n = 2
		}
		for s := 1; s <= n; s++ {
			if s == n {
				out = append(out, p2)
				continue
			}
			out = append(out, catmull(p0, p1, p2, p3, float64(s)/float64(n)))
		}
	}
	return out
}

func catmull(p0, p1, p2, p3 Vec2, t float64) Vec2 {
	t2 := t * t
	t3 := t2 * t
	b := func(a, b, c, d float64) float64 {
		return 0.5 * ((2 * b) + (-a+c)*t + (2*a-5*b+4*c-d)*t2 + (-a+3*b-3*c+d)*t3)
	}
	return Vec2{X: b(p0.X, p1.X, p2.X, p3.X), Y: b(p0.Y, p1.Y, p2.Y, p3.Y)}
}

func (w *World) setPatrol(u *Unit, controls []Vec2) {
	w.assignPatrol(u, PatrolCurve(controls))
}

func (w *World) assignPatrol(u *Unit, pts []Vec2) {
	u.Patrol = append([]Vec2(nil), pts...)
	if len(u.Patrol) < 2 {
		return
	}
	u.PatrolTo = 1
	u.PatrolDir = 1
	u.PatrolWait = 0
	u.PatrolGap = w.patrolGap()
	dx := u.Patrol[1].X - u.X
	dy := u.Patrol[1].Y - u.Y
	if dx != 0 || dy != 0 {
		u.Facing = math.Atan2(dy, dx)
	}
}

// patrolBreak is a player he can shoot, a player inside the sight check,
// or a heard infantry shot. Any of those takes him off the line.
func (w *World) patrolBreak(u *Unit) bool {
	if _, _, see := w.gruntContact(u); see {
		return true
	}
	if _, _, ok := w.patrolDetect(u); ok {
		return true
	}
	if u.HearID != 0 {
		if s := w.Unit(u.HearID); s != nil && s.Living() {
			return true
		}
	}
	return false
}

// patrolDetect is the nearest living player inside the sight check.
// Mission 1 still counts. Past EnemySight he is ignored.
func (w *World) patrolDetect(u *Unit) (x, y float64, ok bool) {
	best := 1e12
	for i := range w.Units {
		o := &w.Units[i]
		if o.Side != SidePlayer || !o.Living() {
			continue
		}
		if !w.acquired(u.X, u.Y, o.X, o.Y) {
			continue
		}
		d := hypot(o.X-u.X, o.Y-u.Y)
		if d < best {
			best = d
			x, y = o.X, o.Y
			ok = true
		}
	}
	return x, y, ok
}

// stepPatrolAlert leaves the route. A clear shot is an attack. A player he
// can see, or a shot he heard, is a walk toward that spot.
func (w *World) stepPatrolAlert(u *Unit, dt float64) {
	if _, _, see := w.gruntContact(u); see {
		w.stepGruntGun(u, dt)
		return
	}
	if x, y, ok := w.patrolDetect(u); ok {
		w.steerToward(u, x, y, WalkSpeed, dt, ArrivalRadius)
		return
	}
	if x, y, ok := w.gruntNoise(u); ok {
		w.steerToward(u, x, y, WalkSpeed, dt, ArrivalRadius)
		return
	}
	w.stepPatrol(u, dt)
}

// stepPatrol walks the line, or stands and looks along it.
func (w *World) stepPatrol(u *Unit, dt float64) {
	if len(u.Patrol) < 2 || u.PatrolTo < 0 || u.PatrolTo >= len(u.Patrol) {
		u.VX = 0
		u.VY = 0
		return
	}
	if u.PatrolWait > 0 {
		w.stepPatrolLook(u, dt)
		return
	}
	if u.PatrolGap <= 0 {
		u.PatrolGap = w.patrolGap()
	}
	u.PatrolGap -= dt
	tx := u.Patrol[u.PatrolTo].X
	ty := u.Patrol[u.PatrolTo].Y
	if w.steerToward(u, tx, ty, patrolSpeed, dt, patrolArrive) {
		w.advancePatrol(u)
		return
	}
	if u.PatrolGap <= 0 && w.farFromPatrolEnds(u) {
		w.beginPatrolPause(u)
	}
}

func (w *World) farFromPatrolEnds(u *Unit) bool {
	n := len(u.Patrol)
	if n < 2 {
		return false
	}
	a := u.Patrol[0]
	b := u.Patrol[n-1]
	return hypot(u.X-a.X, u.Y-a.Y) > float64(TileSize) && hypot(u.X-b.X, u.Y-b.Y) > float64(TileSize)
}

// advancePatrol aims him at the next point, reversing at either end.
// Reaching an end always starts a look-around. He does not leave until it ends.
func (w *World) advancePatrol(u *Unit) {
	n := len(u.Patrol)
	atEnd := u.PatrolTo == 0 || u.PatrolTo == n-1
	next := u.PatrolTo + u.PatrolDir
	if next < 0 || next >= n {
		u.PatrolDir = -u.PatrolDir
		next = u.PatrolTo + u.PatrolDir
	}
	if next < 0 || next >= n || next == u.PatrolTo {
		u.VX = 0
		u.VY = 0
		return
	}
	u.PatrolTo = next
	u.PatrolGap = w.patrolGap()
	u.VX = 0
	u.VY = 0
	if atEnd {
		w.beginPatrolPause(u)
	}
}

func (w *World) beginPatrolPause(u *Unit) {
	if w.rng == nil {
		w.rng = newRNG()
	}
	u.PatrolWait = patrolPauseMin + w.rng.Float64()*(patrolPauseMax-patrolPauseMin)
	u.PatrolSpan = u.PatrolWait
	tx := u.Patrol[u.PatrolTo].X
	ty := u.Patrol[u.PatrolTo].Y
	dx, dy := tx-u.X, ty-u.Y
	if dx == 0 && dy == 0 {
		u.PatrolHead = u.Facing
	} else {
		u.PatrolHead = math.Atan2(dy, dx)
	}
	u.VX = 0
	u.VY = 0
}

// stepPatrolLook turns him off the route and back while he stands.
// The sweep is one cycle: ahead, one side, the other side, ahead.
func (w *World) stepPatrolLook(u *Unit, dt float64) {
	u.PatrolWait -= dt
	if u.PatrolWait < 0 {
		u.PatrolWait = 0
	}
	span := u.PatrolSpan
	if span < 1e-6 {
		span = patrolPauseMin
	}
	frac := 1 - u.PatrolWait/span
	if frac < 0 {
		frac = 0
	} else if frac > 1 {
		frac = 1
	}
	ang := u.PatrolHead + math.Sin(frac*2*math.Pi)*patrolLook
	w.turnToward(u, u.X+math.Cos(ang)*32, u.Y+math.Sin(ang)*32, dt)
	u.VX = 0
	u.VY = 0
}

func (w *World) patrolGap() float64 {
	if w.rng == nil {
		w.rng = newRNG()
	}
	return patrolGapMin + w.rng.Float64()*(patrolGapMax-patrolGapMin)
}

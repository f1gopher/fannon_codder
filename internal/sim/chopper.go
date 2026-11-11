package sim

import "math"

const (
	KindChopper = "chopper"

	// ChopperSpeed is the patrol-style cruise. Ground, water, trees, and
	// huts do not slow it. ChopperPassR is the low pass that kills a man.
	ChopperSpeed = 64.0
	ChopperPassR = 11.0
	// ChopperFall is the drop after a hit. The boom waits until Alt hits 0.
	ChopperFall      = 56.0
	ChopperFallSpeed = 72.0

	// HomingRange is how far an armed chopper looks. The rocket turns
	// toward that man and does not care about cover.
	HomingRange  = 180.0
	HomingSpeed  = 120.0
	HomingTurn   = 2.4
	HomingReload = 1.25
)

// AddChopper parks an enemy aircraft. It is never boarded.
// line is later route tiles. With none, it holds p and only turns.
func (w *World) AddChopper(p Vec2, armed bool, line [][2]int) *Vehicle {
	if w.nextVID == 0 {
		w.nextVID = 1
	}
	v := Vehicle{
		ID:       w.nextVID,
		X:        p.X,
		Y:        p.Y,
		Side:     SideEnemy,
		Kind:     KindChopper,
		Armed:    armed,
		Alive:    true,
		Facing:   0,
		RouteDir: 1,
	}
	w.nextVID++
	pts := []Vec2{p}
	for _, t := range line {
		pts = append(pts, TileCenter(t[0], t[1]))
	}
	if len(pts) >= 2 {
		v.Route = PatrolCurve(pts)
		v.RouteTo = 1
	}
	w.Vehicles = append(w.Vehicles, v)
	return &w.Vehicles[len(w.Vehicles)-1]
}

// ChopperNear reports a living combat chopper inside radius of the listener.
// The opening drop does not count. A falling wreck does not either.
func (w *World) ChopperNear(x, y, radius float64) bool {
	if w == nil || radius <= 0 {
		return false
	}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive || v.Kind != KindChopper || v.Alt > 0 {
			continue
		}
		if hypot(v.X-x, v.Y-y) <= radius {
			return true
		}
	}
	return false
}

func (w *World) stepChopper(v *Vehicle, dt float64) {
	if v == nil || !v.Alive {
		return
	}
	w.chopperPass(v)
	px, py, ok := w.nearestLiving(SidePlayer, v.X, v.Y)
	dist := 0.0
	if ok {
		dist = hypot(px-v.X, py-v.Y)
	}
	engage := ok && v.Armed && dist <= HomingRange
	if engage {
		v.VX, v.VY = 0, 0
		turnVehicle(v, px, py, dt)
		w.stepChopperGun(v, px, py, dt)
		return
	}
	w.idleChopperGun(v)
	w.flyChopper(v, dt)
}

func (w *World) chopperPass(v *Vehicle) {
	for i := range w.Units {
		u := &w.Units[i]
		if u.Dead() || u.Saved || u.VehicleID != 0 {
			continue
		}
		if hypot(u.X-v.X, u.Y-v.Y) <= ChopperPassR {
			w.kill(u)
		}
	}
}

func (w *World) flyChopper(v *Vehicle, dt float64) {
	if len(v.Route) < 2 {
		v.VX, v.VY = 0, 0
		return
	}
	if v.RouteTo < 0 || v.RouteTo >= len(v.Route) {
		v.RouteTo = 0
	}
	dest := v.Route[v.RouteTo]
	dx, dy := dest.X-v.X, dest.Y-v.Y
	dist := hypot(dx, dy)
	if dist < 4 {
		next := v.RouteTo + v.RouteDir
		if next < 0 || next >= len(v.Route) {
			v.RouteDir = -v.RouteDir
			next = v.RouteTo + v.RouteDir
		}
		if next >= 0 && next < len(v.Route) {
			v.RouteTo = next
			dest = v.Route[v.RouteTo]
			dx, dy = dest.X-v.X, dest.Y-v.Y
			dist = hypot(dx, dy)
		}
	}
	if dist < 1 {
		v.VX, v.VY = 0, 0
		return
	}
	v.Facing = math.Atan2(dy, dx)
	step := ChopperSpeed * dt
	if step > dist {
		step = dist
	}
	v.X += dx / dist * step
	v.Y += dy / dist * step
	v.VX = dx / dist * ChopperSpeed
	v.VY = dy / dist * ChopperSpeed
	w.clampChopper(v)
}

func (w *World) clampChopper(v *Vehicle) {
	mw, mh := w.Map.PixelSize()
	if v.X < 8 {
		v.X = 8
	}
	if v.Y < 8 {
		v.Y = 8
	}
	if v.X > mw-8 {
		v.X = mw - 8
	}
	if v.Y > mh-8 {
		v.Y = mh - 8
	}
}

func (w *World) stepChopperGun(v *Vehicle, px, py, dt float64) {
	if v.ReactAt == 0 {
		if w.rng == nil {
			w.rng = newRNG()
		}
		v.ReactAt = EnemyReact + (w.rng.Float64()*2-1)*EnemyReactJitter
	}
	v.SpotT += dt
	if v.SpotT+1e-9 < v.ReactAt {
		return
	}
	if vehicleFacingError(v, px, py) > EnemyFaceTol {
		return
	}
	v.FireCD -= dt
	if v.FireCD > 0 {
		return
	}
	v.FireCD = HomingReload
	id := 0
	if u := w.nearestLivingUnit(SidePlayer, v.X, v.Y); u != nil {
		id = u.ID
	}
	w.launchHoming(v, id, px, py)
}

func (w *World) idleChopperGun(v *Vehicle) {
	v.SpotT = 0
	v.ReactAt = 0
	v.FireCD = 0
}

func (w *World) nearestLivingUnit(side Side, x, y float64) *Unit {
	var best *Unit
	bestD := 1e12
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != side || !u.Living() {
			continue
		}
		d := hypot(u.X-x, u.Y-y)
		if d < bestD {
			bestD = d
			best = u
		}
	}
	return best
}

// launchHoming fires the chopper rocket. CueHoming is the whistle.
// The impact later is CueBoom, from the explosion, not this launch.
func (w *World) launchHoming(v *Vehicle, target int, tx, ty float64) {
	dx, dy := tx-v.X, ty-v.Y
	if dx == 0 && dy == 0 {
		dx = 1
	}
	ang := math.Atan2(dy, dx)
	nx, ny := math.Cos(ang), math.Sin(ang)
	w.Projectiles = append(w.Projectiles, Projectile{
		Kind:      ProjHoming,
		X:         v.X + nx*12,
		Y:         v.Y + ny*12,
		VX:        nx * HomingSpeed,
		VY:        ny * HomingSpeed,
		OwnerID:   v.ID,
		OwnerSide: v.Side,
		TargetID:  target,
		Left:      HomingRange,
		Alive:     true,
	})
	w.emit(CueHoming, v.X, v.Y)
}

func (w *World) stepHoming(p *Projectile, dt float64) {
	if tgt := w.Unit(p.TargetID); tgt != nil && tgt.Living() {
		want := math.Atan2(tgt.Y-p.Y, tgt.X-p.X)
		have := math.Atan2(p.VY, p.VX)
		d := wrapAngle(want - have)
		step := HomingTurn * dt
		if d > step {
			d = step
		} else if d < -step {
			d = -step
		}
		ang := have + d
		p.VX = math.Cos(ang) * HomingSpeed
		p.VY = math.Sin(ang) * HomingSpeed
	}
	x0, y0 := p.X, p.Y
	speed := hypot(p.VX, p.VY)
	if speed < 1 {
		speed = HomingSpeed
	}
	step := speed * dt
	if step > p.Left {
		step = p.Left
	}
	p.X += p.VX / speed * step
	p.Y += p.VY / speed * step
	p.Left -= step
	if hx, hy, hit := w.homingHit(x0, y0, p.X, p.Y); hit || p.Left <= 0 {
		if !hit {
			hx, hy = p.X, p.Y
		}
		w.explodeRocket(hx, hy, 0)
		p.Alive = false
	}
}

// homingHit is a man. Cover does not stop the rocket.
func (w *World) homingHit(x0, y0, x1, y1 float64) (float64, float64, bool) {
	best := 1e12
	var hx, hy float64
	ok := false
	for i := range w.Units {
		u := &w.Units[i]
		if u.Dead() || u.Saved || u.VehicleID != 0 {
			continue
		}
		if segmentHitsCircle(x0, y0, x1, y1, u.X, u.Y, MGHitR) {
			d := hypot(u.X-x0, u.Y-y0)
			if d < best {
				best = d
				hx, hy = u.X, u.Y
				ok = true
			}
		}
	}
	return hx, hy, ok
}

func (w *World) stepChopperFalls(dt float64) {
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if v.Alt > 0 {
			w.stepChopperFall(v, dt)
		}
	}
}

func (w *World) stepChopperFall(v *Vehicle, dt float64) {
	v.Alt -= ChopperFallSpeed * dt
	v.VX, v.VY = 0, 0
	if v.Alt > 0 {
		return
	}
	v.Alt = 0
	w.explodeRocket(v.X, v.Y, 0)
}

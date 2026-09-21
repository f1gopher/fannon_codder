package sim

import "math"

const (
	WalkSpeed      = 30 // world px / second
	ArrivalRadius  = 4
	FollowerArrive = 2
	FileSpacing    = 10
	trailSample    = 2
	trailMax       = 256
)

func hypot(dx, dy float64) float64 {
	return math.Hypot(dx, dy)
}

const unitHalf = float64(UnitSize) / 2

// steerToward walks u toward (tx,ty), sliding along solid tiles.
// Reports whether it is inside arrival and the dest is standable.
func (w *World) steerToward(u *Unit, tx, ty, speed, dt, arrival float64) bool {
	speed *= w.speedMul(u)
	dx := tx - u.X
	dy := ty - u.Y
	dist := hypot(dx, dy)
	if dist <= arrival {
		if w.walkableUnit(tx, ty) {
			u.X = tx
			u.Y = ty
			u.VX = 0
			u.VY = 0
			return true
		}
		u.VX = 0
		u.VY = 0
		return false
	}
	u.Facing = math.Atan2(dy, dx)
	step := speed * dt
	if step > dist {
		step = dist
	}
	nx, ny := dx/dist, dy/dist
	u.VX = nx * speed
	u.VY = ny * speed
	w.slide(u, nx*step, ny*step)
	return false
}

func (w *World) walkableUnit(x, y float64) bool {
	if !w.Map.Walkable(x, y, unitHalf) {
		return false
	}
	return !w.unitHitsBuilding(x, y, unitHalf)
}

func (w *World) slide(u *Unit, dx, dy float64) {
	nx, ny := u.X+dx, u.Y+dy
	if w.walkableUnit(nx, ny) {
		u.X, u.Y = nx, ny
		return
	}
	if w.tryCliffDrop(u, nx, ny) {
		return
	}
	if w.walkableUnit(nx, u.Y) {
		u.X = nx
		u.VY = 0
		return
	}
	if w.walkableUnit(u.X, ny) {
		u.Y = ny
		u.VX = 0
		return
	}
	u.VX = 0
	u.VY = 0
}

// tryCliffDrop lands a unit on the south side of a cliff it is stepping off.
// The face stays solid from below; TileRamp is the way back up.
func (w *World) tryCliffDrop(u *Unit, nx, ny float64) bool {
	if ny <= u.Y || !w.Map.active() {
		return false
	}
	tx0, ty0, tx1, ty1 := footprintTiles(nx, ny, unitHalf)
	for ty := ty0; ty <= ty1; ty++ {
		top := float64(ty * TileSize)
		if u.Y+unitHalf > top+1 {
			continue
		}
		for tx := tx0; tx <= tx1; tx++ {
			if w.Map.At(tx, ty) != TileCliff {
				continue
			}
			landX := float64(tx*TileSize) + float64(TileSize)/2
			landY := float64((ty+1)*TileSize) + unitHalf + 0.5
			if !w.walkableUnit(landX, landY) {
				continue
			}
			u.X, u.Y = landX, landY
			u.VX, u.VY = 0, 0
			return true
		}
	}
	return false
}

func footprintTiles(x, y, half float64) (tx0, ty0, tx1, ty1 int) {
	x0 := int(x - half)
	y0 := int(y - half)
	x1 := int(x + half - 0.001)
	y1 := int(y + half - 0.001)
	tx0 = x0 / TileSize
	ty0 = y0 / TileSize
	tx1 = x1 / TileSize
	ty1 = y1 / TileSize
	if x0 < 0 {
		tx0 = -1
	}
	if y0 < 0 {
		ty0 = -1
	}
	return tx0, ty0, tx1, ty1
}

func recordTrail(s *Squad, p Vec2) {
	if len(s.Trail) == 0 {
		s.Trail = append(s.Trail, p)
		return
	}
	last := s.Trail[len(s.Trail)-1]
	if hypot(p.X-last.X, p.Y-last.Y) < trailSample {
		return
	}
	s.Trail = append(s.Trail, p)
	if len(s.Trail) > trailMax {
		s.Trail = s.Trail[len(s.Trail)-trailMax:]
	}
}

// trailPoint is `back` pixels behind head (the leader), walking the crumb trail.
func trailPoint(head Vec2, trail []Vec2, back float64) (Vec2, bool) {
	if back <= 0 {
		return head, true
	}
	remaining := back
	prev := head
	for i := len(trail) - 1; i >= 0; i-- {
		b := trail[i]
		seg := hypot(prev.X-b.X, prev.Y-b.Y)
		if seg <= 0 {
			continue
		}
		if remaining <= seg {
			t := remaining / seg
			return Vec2{
				X: prev.X + (b.X-prev.X)*t,
				Y: prev.Y + (b.Y-prev.Y)*t,
			}, true
		}
		remaining -= seg
		prev = b
	}
	if len(trail) == 0 {
		return head, true
	}
	return trail[0], true
}

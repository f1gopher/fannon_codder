package sim

import "math"

const (
	RocketRange = 200.0
	RocketSpeed = 280.0
	VehicleHitR = 12.0
)

// AddRocketCrate drops a crate of CrateAmount rockets.
func (w *World) AddRocketCrate(p Vec2) {
	w.Pickups = append(w.Pickups, Pickup{
		Kind:   PickupRockets,
		X:      p.X,
		Y:      p.Y,
		Amount: CrateAmount,
		Alive:  true,
	})
}

// UseSpecial throws a grenade or fires the bazooka, whichever is selected.
// Neither works from inside a vehicle.
func (w *World) UseSpecial(x, y float64) bool {
	if w.LeaderInVehicle() {
		return false
	}
	if w.Special == SpecialRocket {
		return w.FireRocket(x, y)
	}
	return w.ThrowGrenade(x, y)
}

// FireRocket sends the active leader's bazooka round toward (x, y).
func (w *World) FireRocket(x, y float64) bool {
	if w.Special != SpecialRocket {
		return false
	}
	s := w.ActiveSquad()
	if s == nil || s.Rockets <= 0 {
		return false
	}
	u := w.Unit(s.LeaderID)
	if !w.CanShoot(u) {
		return false
	}
	s.Rockets--
	w.launchRocket(u.ID, u.Side, u.X, u.Y, x, y)
	return true
}

// launchRocket fires a fast explosive round. It detonates on impact or at max range.
func (w *World) launchRocket(owner int, side Side, x, y, tx, ty float64) {
	dx, dy := tx-x, ty-y
	dist := hypot(dx, dy)
	if dist < 1 {
		dx, dy = 1, 0
		dist = 1
	}
	if dist > RocketRange {
		dist = RocketRange
	}
	nx, ny := math.Cos(math.Atan2(dy, dx)), math.Sin(math.Atan2(dy, dx))
	if u := w.Unit(owner); u != nil {
		u.Facing = math.Atan2(ny, nx)
		u.SinceThrow = 0
	}
	muzzle := float64(UnitSize)
	travel := dist - muzzle
	if travel < 8 {
		travel = 8
	}
	w.Projectiles = append(w.Projectiles, Projectile{
		Kind:      ProjRocket,
		X:         x + nx*muzzle,
		Y:         y + ny*muzzle,
		VX:        nx * RocketSpeed,
		VY:        ny * RocketSpeed,
		OwnerID:   owner,
		OwnerSide: side,
		Left:      travel,
		Alive:     true,
	})
	w.emit(CueRocket, x+nx*muzzle, y+ny*muzzle)
}

// rocketHit is the first thing a bazooka round meets along the segment.
// Crew inside a vehicle are not a separate target; the hull is.
func (w *World) rocketHit(owner int, x0, y0, x1, y1 float64) (float64, float64, bool) {
	best := 1e12
	var hx, hy float64
	ok := false
	consider := func(x, y float64) {
		d := hypot(x-x0, y-y0)
		if d < best {
			best = d
			hx, hy = x, y
			ok = true
		}
	}
	if x, y, hit := w.Map.FirstSolid(x0, y0, x1, y1); hit {
		consider(x, y)
	}
	if x, y, hit := w.buildingHit(x0, y0, x1, y1); hit {
		consider(x, y)
	}
	for i := range w.Units {
		u := &w.Units[i]
		if u.ID == owner || u.Dead() || u.VehicleID != 0 {
			continue
		}
		if segmentHitsCircle(x0, y0, x1, y1, u.X, u.Y, MGHitR) {
			consider(u.X, u.Y)
		}
	}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		if segmentHitsCircle(x0, y0, x1, y1, v.X, v.Y, VehicleHitR) {
			consider(v.X, v.Y)
		}
	}
	for i := range w.Pickups {
		p := &w.Pickups[i]
		if !p.Alive {
			continue
		}
		if segmentHitsCircle(x0, y0, x1, y1, p.X, p.Y, CrateHitR) {
			consider(p.X, p.Y)
		}
	}
	return hx, hy, ok
}

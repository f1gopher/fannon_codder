package sim

import "math"

const (
	GrenadeRange  = 96.0
	GrenadeRadius = 24.0 // reaches a grunt standing in the doorway
	GrenadeSpeed  = 70.0 // px/s along the ground
	GrenadeArc    = 18.0 // peak height, visual only
	CrateAmount   = 4
	CrateHitR     = 6.0
	PickupRadius  = 10.0
	blastTime     = 0.35

	// Grenadiers carry a couple of bombs and throw rarely.
	// The windup is the telegraph: they stop and face you before the bomb leaves.
	GrenadierBombs      = 2
	GrenadierWindup     = 0.8
	GrenadierCooldown   = 5.0
	GrenadierFirstDelay = 3.0
	// Inside this range the blast would catch the thrower, so they shoot instead.
	GrenadierMinRange = GrenadeRadius + 8
)

// PickupKind is a crate on the ground.
type PickupKind int

const (
	PickupGrenades PickupKind = iota
	PickupRockets
)

// Pickup is a crate. Walking over a grenade crate adds ammo; shooting it explodes.
type Pickup struct {
	Kind   PickupKind
	X, Y   float64
	Amount int
	Alive  bool
}

// Grenade is a thrown bomb. It flies over trees and huts and lands at the aim point.
type Grenade struct {
	X, Y    float64
	SX, SY  float64
	TX, TY  float64
	Height  float64
	T, Dur  float64
	OwnerID int
	Alive   bool
}

// Explosion is a short blast marker for drawing. Damage is applied immediately.
type Explosion struct {
	X, Y float64
	R    float64
	Age  float64
}

// AddGrenadeCrate drops a crate of CrateAmount grenades at a world point.
func (w *World) AddGrenadeCrate(p Vec2) {
	w.Pickups = append(w.Pickups, Pickup{
		Kind:   PickupGrenades,
		X:      p.X,
		Y:      p.Y,
		Amount: CrateAmount,
		Alive:  true,
	})
}

// Special is the weapon a chord or Space uses.
type Special int

const (
	SpecialGrenade Special = iota
	SpecialRocket
)

// UseAmmoIcon clicks the grenade or rocket count.
// Highlighted men cycle how much of that ammo a split takes.
// With nobody highlighted, the click selects that special.
func (w *World) UseAmmoIcon(kind Special) {
	if len(w.Selected) > 0 {
		if kind == SpecialRocket {
			w.CycleRocketShare()
		} else {
			w.CycleGrenadeShare()
		}
		return
	}
	w.Special = kind
}

// ToggleSpecial flips grenade and bazooka (the C key).
func (w *World) ToggleSpecial() {
	if w.Special == SpecialRocket {
		w.Special = SpecialGrenade
		return
	}
	w.Special = SpecialRocket
}

// ThrowGrenade sends the active leader's grenade toward (x, y).
// Right-held + left click, or Space. Deep water cannot throw. Leader only.
// A selected bazooka does not spend grenades.
func (w *World) ThrowGrenade(x, y float64) bool {
	if w.Special != SpecialGrenade {
		return false
	}
	s := w.ActiveSquad()
	if s == nil || s.Grenades <= 0 {
		return false
	}
	u := w.Unit(s.LeaderID)
	if !w.CanShoot(u) {
		return false
	}
	s.Grenades--
	w.launchGrenade(u, x, y)
	return true
}

// launchGrenade arcs a bomb from u toward (x, y), clamped to GrenadeRange.
func (w *World) launchGrenade(u *Unit, x, y float64) {
	dx, dy := x-u.X, y-u.Y
	dist := hypot(dx, dy)
	if dist < 1 {
		dx, dy = 1, 0
		dist = 1
	}
	if dist > GrenadeRange {
		scale := GrenadeRange / dist
		x = u.X + dx*scale
		y = u.Y + dy*scale
		dist = GrenadeRange
	}
	dur := dist / GrenadeSpeed
	if dur < 0.2 {
		dur = 0.2
	}
	u.Facing = math.Atan2(y-u.Y, x-u.X)
	w.Grenades = append(w.Grenades, Grenade{
		X:       u.X,
		Y:       u.Y,
		SX:      u.X,
		SY:      u.Y,
		TX:      x,
		TY:      y,
		Dur:     dur,
		OwnerID: u.ID,
		Alive:   true,
	})
}

func (w *World) stepPickups() {
	for pi := range w.Pickups {
		p := &w.Pickups[pi]
		if !p.Alive {
			continue
		}
		if p.Kind != PickupGrenades && p.Kind != PickupRockets {
			continue
		}
		for i := range w.Units {
			u := &w.Units[i]
			if u.Side != SidePlayer || !u.Living() {
				continue
			}
			if hypot(u.X-p.X, u.Y-p.Y) > PickupRadius {
				continue
			}
			if s := w.SquadByID(u.SquadID); s != nil {
				if p.Kind == PickupRockets {
					s.Rockets += p.Amount
				} else {
					s.Grenades += p.Amount
				}
			}
			p.Alive = false
			break
		}
	}
}

func (w *World) stepGrenades(dt float64) {
	out := w.Grenades[:0]
	for i := range w.Grenades {
		g := w.Grenades[i]
		if !g.Alive {
			continue
		}
		g.T += dt
		u := g.T / g.Dur
		if u > 1 {
			u = 1
		}
		g.X = g.SX + (g.TX-g.SX)*u
		g.Y = g.SY + (g.TY-g.SY)*u
		g.Height = 4 * GrenadeArc * u * (1 - u)
		if g.T >= g.Dur {
			w.explode(g.TX, g.TY, g.OwnerID)
			continue
		}
		out = append(out, g)
	}
	w.Grenades = out
}

func (w *World) stepBlasts(dt float64) {
	out := w.Explosions[:0]
	for i := range w.Explosions {
		e := w.Explosions[i]
		e.Age += dt
		if e.Age < blastTime {
			out = append(out, e)
		}
	}
	w.Explosions = out
}

// explode hurts every living unit and destroys huts and crates in the radius.
// Grenades do not spare friendlies. A destroyed crate explodes too.
func (w *World) explode(x, y float64, ownerID int) {
	w.emit(CueBoom, x, y)
	w.Explosions = append(w.Explosions, Explosion{X: x, Y: y, R: GrenadeRadius})
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Living() {
			continue
		}
		if hypot(u.X-x, u.Y-y) > GrenadeRadius {
			continue
		}
		w.kill(u)
		if owner := w.Unit(ownerID); owner != nil && owner.Living() && owner.ID != u.ID {
			owner.Kills++
		}
	}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		if hypot(v.X-x, v.Y-y) > GrenadeRadius+VehicleHitR {
			continue
		}
		w.destroyVehicle(v)
	}
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if !b.Alive {
			continue
		}
		if circleHitsRect(x, y, GrenadeRadius, b.X, b.Y, b.W, b.H) {
			b.Alive = false
			b.HP = 0
		}
	}
	for i := range w.Pickups {
		p := &w.Pickups[i]
		if !p.Alive {
			continue
		}
		if hypot(p.X-x, p.Y-y) > GrenadeRadius {
			continue
		}
		p.Alive = false
		w.explode(p.X, p.Y, ownerID)
	}
}

func (w *World) tryHitCrate(p *Projectile, x0, y0, x1, y1 float64) {
	for i := range w.Pickups {
		c := &w.Pickups[i]
		if !c.Alive {
			continue
		}
		if !segmentHitsCircle(x0, y0, x1, y1, c.X, c.Y, CrateHitR) {
			continue
		}
		c.Alive = false
		p.Alive = false
		w.explode(c.X, c.Y, p.OwnerID)
		return
	}
}

func circleHitsRect(cx, cy, r, x, y, w, h float64) bool {
	nx, ny := cx, cy
	if nx < x {
		nx = x
	} else if nx > x+w {
		nx = x + w
	}
	if ny < y {
		ny = y
	} else if ny > y+h {
		ny = y + h
	}
	dx, dy := cx-nx, cy-ny
	return dx*dx+dy*dy <= r*r
}

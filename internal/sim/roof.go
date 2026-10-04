package sim

import "math"

const (
	// RoofReach is how far a shed roof travels from the hut, in world pixels.
	RoofReach = 64
	// RoofFlight is the tumble, in seconds. Longer than a grenade flash.
	RoofFlight = 1.2
	// RoofArc is the peak height of that tumble.
	RoofArc = 42
	// RoofHitR is the crush radius once the roof is low enough to hit.
	RoofHitR = 16
	// RoofCrushH is the height under which the slab kills.
	RoofCrushH = 12
	// RoofTurns is how many times the slab flips before it settles upright.
	RoofTurns = 2

	hutBlastLife  = 1.25
	hutBlastScale = 2.15
	hutBlastRate  = 0.42
	hutBlastLift  = 14
)

// Roof is a shed top blown off its walls. It tumbles, then crushes men it lands on.
type Roof struct {
	X, Y    float64
	SX, SY  float64
	TX, TY  float64
	Height  float64
	Angle   float64
	T       float64
	Door    bool
	Flying  bool
	OwnerID int
	risen   bool
}

func (w *World) launchRoof(b *Building, blastX, blastY float64, ownerID int) {
	cx := b.X + b.W/2
	footY := b.Y + b.H
	dx, dy := cx-blastX, (b.Y+b.H/2)-blastY
	tx, ty := w.roofLanding(cx, footY, dx, dy)
	w.Roofs = append(w.Roofs, Roof{
		X: cx, Y: footY,
		SX: cx, SY: footY,
		TX: tx, TY: ty,
		Door:    b.HasDoor,
		Flying:  true,
		OwnerID: ownerID,
	})
	w.Explosions = append(w.Explosions, Explosion{
		X: cx, Y: b.Y + b.H/2, R: GrenadeRadius * hutBlastScale,
		Scale: hutBlastScale,
		Life:  hutBlastLife,
		Rate:  hutBlastRate,
		Lift:  hutBlastLift,
		Loop:  true,
	})
}

// roofLanding is a point RoofReach away from the hut, on the side opposite the blast.
func (w *World) roofLanding(sx, sy, dx, dy float64) (float64, float64) {
	dist := math.Hypot(dx, dy)
	if dist < 4 {
		dx, dy = 0.55, 1
		dist = math.Hypot(dx, dy)
	}
	nx, ny := dx/dist, dy/dist
	for _, reach := range []float64{RoofReach, RoofReach * 0.7, RoofReach * 0.45} {
		x := sx + nx*reach
		y := sy + ny*reach
		if w.Map.W == 0 || w.Map.Walkable(x, y, unitHalf) {
			return x, y
		}
	}
	return sx + nx*RoofReach, sy + ny*RoofReach
}

func (w *World) stepRoofs(dt float64) {
	for i := range w.Roofs {
		r := &w.Roofs[i]
		if !r.Flying {
			continue
		}
		r.T += dt
		u := r.T / RoofFlight
		if u > 1 {
			u = 1
		}
		r.X = r.SX + (r.TX-r.SX)*u
		r.Y = r.SY + (r.TY-r.SY)*u
		r.Height = 4 * RoofArc * u * (1 - u)
		// Two full turns while it is in the air. The settled slab is upright.
		r.Angle = u * RoofTurns * 2 * math.Pi
		if r.Height > RoofCrushH {
			r.risen = true
		}
		landed := r.T >= RoofFlight
		if landed {
			r.T = RoofFlight
			r.Height = 0
			r.X, r.Y = r.TX, r.TY
			r.Angle = 0
			r.Flying = false
		}
		// The launch frame is also at ground height. Crush only after the
		// slab has been up, then on the way down, so it does not scrape the hut.
		if landed || (r.risen && r.Height <= RoofCrushH) {
			w.crushUnderRoof(r)
		}
	}
}

// crushUnderRoof kills soldiers under the slab. It runs on each low frame
// of the descent, so a man at the landing is caught even if the first
// sample is still a few pixels short.
func (w *World) crushUnderRoof(r *Roof) {
	for i := range w.Units {
		u := &w.Units[i]
		if u.Dead() || u.VehicleID != 0 {
			continue
		}
		if math.Hypot(u.X-r.X, u.Y-r.Y) > RoofHitR {
			continue
		}
		w.kill(u)
		w.creditKill(r.OwnerID, u)
	}
}

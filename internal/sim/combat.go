package sim

import (
	"math"
	"math/rand/v2"
)

const (
	MGSpeed  = 500.0
	MGHitR   = 4.0
	MGSpread = 0.05 // default world spread enable (0 in tests = none)

	privateRange  = 80.0
	generalRange  = 150.0
	privateRoF    = 8.0
	generalRoF    = 14.0
	privateSpread = 0.12 // radians, wide
	generalSpread = 0.02 // tight
	rankMax       = 15   // General
)

// GunStats is MG performance for a player rank (0=Private … 15=General).
type GunStats struct {
	Range  float64
	RoF    float64
	Spread float64
}

// GunStatsFor interpolates the Amiga-ish curve by rank index.
func GunStatsFor(rank int) GunStats {
	if rank < 0 {
		rank = 0
	}
	if rank > rankMax {
		rank = rankMax
	}
	t := float64(rank) / float64(rankMax)
	return GunStats{
		Range:  privateRange + (generalRange-privateRange)*t,
		RoF:    privateRoF + (generalRoF-privateRoF)*t,
		Spread: privateSpread + (generalSpread-privateSpread)*t,
	}
}

// SetFire aims the active squad at a world-space point. firing is right-held (or Ctrl).
func (w *World) SetFire(aimX, aimY float64, firing bool) {
	w.AimX = aimX
	w.AimY = aimY
	w.Firing = firing
}

func (w *World) stepFire(dt float64) {
	s := w.ActiveSquad()
	if s == nil || !w.Firing {
		return
	}
	for _, id := range s.MemberIDs {
		u := w.Unit(id)
		if u == nil || !u.Living() {
			continue
		}
		st := GunStatsFor(u.Rank)
		dx := w.AimX - u.X
		dy := w.AimY - u.Y
		if dx == 0 && dy == 0 {
			dx = 1
		}
		u.Facing = math.Atan2(dy, dx)
		u.FireCD -= dt
		if u.FireCD > 0 {
			continue
		}
		u.FireCD = 1.0 / st.RoF
		ang := u.Facing
		spread := st.Spread
		if w.Spread == 0 {
			spread = 0
		}
		if spread > 0 {
			ang += (w.rng.Float64()*2 - 1) * spread
		}
		w.spawnMG(u, ang, st.Range)
	}
}

func (w *World) spawnMG(u *Unit, ang float64, maxRange float64) {
	nx, ny := math.Cos(ang), math.Sin(ang)
	muzzle := float64(UnitSize) / 2
	w.Projectiles = append(w.Projectiles, Projectile{
		X:         u.X + nx*muzzle,
		Y:         u.Y + ny*muzzle,
		VX:        nx * MGSpeed,
		VY:        ny * MGSpeed,
		OwnerID:   u.ID,
		OwnerSide: u.Side,
		Left:      maxRange,
		Alive:     true,
	})
}

func (w *World) stepProjectiles(dt float64) {
	out := w.Projectiles[:0]
	for i := range w.Projectiles {
		p := &w.Projectiles[i]
		if !p.Alive {
			continue
		}
		x0, y0 := p.X, p.Y
		step := MGSpeed * dt
		if step > p.Left {
			step = p.Left
		}
		dist := hypot(p.VX, p.VY)
		if dist == 0 {
			continue
		}
		nx, ny := p.VX/dist, p.VY/dist
		p.X += nx * step
		p.Y += ny * step
		p.Left -= step
		w.tryHit(p, x0, y0, p.X, p.Y)
		if p.Alive && p.Left > 0 {
			out = append(out, *p)
		}
	}
	w.Projectiles = out
}

func (w *World) tryHit(p *Projectile, x0, y0, x1, y1 float64) {
	for i := range w.Units {
		u := &w.Units[i]
		if u.ID == p.OwnerID || u.Dead() {
			continue
		}
		if !w.mgCanHurt(p, u) {
			continue
		}
		if !segmentHitsCircle(x0, y0, x1, y1, u.X, u.Y, MGHitR) {
			continue
		}
		w.kill(u)
		if owner := w.Unit(p.OwnerID); owner != nil {
			owner.Kills++
		}
		p.Alive = false
		return
	}
}

// Player MG does not harm living friendlies. Explosives later ignore this.
func (w *World) mgCanHurt(p *Projectile, u *Unit) bool {
	if p.OwnerSide == SidePlayer && u.Side == SidePlayer && u.Living() {
		return false
	}
	return true
}

func (w *World) kill(u *Unit) {
	if u.Dead() {
		return
	}
	u.HP = Dead
	u.VX = 0
	u.VY = 0
	w.dropFromFile(u.ID)
}

func (w *World) dropFromFile(id int) {
	for i := range w.Squads {
		s := &w.Squads[i]
		n := 0
		wasLeader := s.LeaderID == id
		for _, m := range s.MemberIDs {
			if m != id {
				s.MemberIDs[n] = m
				n++
			}
		}
		s.MemberIDs = s.MemberIDs[:n]
		if wasLeader {
			s.LeaderID = 0
			s.HasDest = false
			s.Trail = nil
			if n > 0 {
				s.LeaderID = s.MemberIDs[0]
			}
		}
	}
}

func segmentHitsCircle(x0, y0, x1, y1, cx, cy, r float64) bool {
	dx := x1 - x0
	dy := y1 - y0
	fx := x0 - cx
	fy := y0 - cy
	a := dx*dx + dy*dy
	if a < 1e-12 {
		return fx*fx+fy*fy <= r*r
	}
	t := -(fx*dx + fy*dy) / a
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	px := x0 + t*dx - cx
	py := y0 + t*dy - cy
	return px*px+py*py <= r*r
}

func newRNG() *rand.Rand {
	return rand.New(rand.NewPCG(1, 2))
}

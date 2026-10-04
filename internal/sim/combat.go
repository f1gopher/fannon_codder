package sim

import (
	"math"
	"math/rand/v2"
)

const (
	MGSpeed  = 500.0
	MGHitR   = 4.0
	MGSpread = 0.05 // default world spread enable (0 in tests = none)

	// WoundOdds is how often a standing man's MG hit drops him instead of killing him.
	// A burst still finishes him. Tests leave WoundChance at 0.
	WoundOdds = 0.34

	juggleKick    = 36.0 // px/s along the shot; extra rounds do not stack past this
	juggleHop     = 64.0 // px/s upward
	juggleGravity = 420.0
	juggleBounce  = 0.28
	juggleDrag    = 200.0 // px/s^2 once the body is on the ground
	juggleRest    = 12.0  // px/s; slower than this, the hop ends
	juggleLife    = 3.0   // seconds after death; later MG hits leave the body still

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
		if !w.CanShoot(u) {
			continue
		}
		st := GunStatsFor(u.Rank)
		dx := w.AimX - u.X
		dy := w.AimY - u.Y
		if dx == 0 && dy == 0 {
			dx = 1
		}
		w.shootAt(u, dx, dy, st, dt)
	}
}

// stepInactiveFire: a squad you are not controlling holds, turns, and bursts.
// Range, rate, and cone stay that man's rank gun. He does not hear shots,
// and he does not throw. The active squad is stepFire.
func (w *World) stepInactiveFire(dt float64) {
	active := w.ActiveSquad()
	for i := range w.Squads {
		s := &w.Squads[i]
		if active != nil && s.ID == active.ID {
			continue
		}
		for _, id := range s.MemberIDs {
			u := w.Unit(id)
			if !w.CanShoot(u) {
				continue
			}
			w.stepParkedGun(u, dt)
		}
	}
}

// stepParkedGun is the grunt contact clock with GunStatsFor. Losing the
// enemy, the range, or the line resets the clock. He stays on his tile.
func (w *World) stepParkedGun(u *Unit, dt float64) {
	st := GunStatsFor(u.Rank)
	tx, ty, see := w.nearestEnemy(u, st.Range)
	if !see {
		resetGruntContact(u)
		return
	}
	if u.ReactAt == 0 {
		if w.rng == nil {
			w.rng = newRNG()
		}
		u.ReactAt = EnemyReact + (w.rng.Float64()*2-1)*EnemyReactJitter
	}
	u.SpotT += dt
	w.turnToward(u, tx, ty, dt)
	if u.SpotT+1e-9 < u.ReactAt {
		return
	}
	// The pause is real time under contact. Facing and the trigger come after,
	// so a man still turning does not freeze the gap or fire through it.
	if u.BurstGap > 0 {
		u.BurstGap -= dt
		if u.BurstGap > 0 {
			return
		}
		u.BurstGap = 0
		u.BurstN = 0
		u.FireCD = 0
	}
	if facingError(u, tx, ty) > EnemyFaceTol {
		return
	}
	u.FireCD -= dt
	if u.FireCD > 0 {
		return
	}
	u.BurstN++
	if u.BurstN >= EnemyBurst {
		u.BurstGap = EnemyBurstPause
		u.FireCD = 0
	} else {
		u.FireCD = 1.0 / st.RoF
	}
	ang := u.Facing
	spread := st.Spread
	if w.Spread == 0 {
		spread = 0
	}
	if spread > 0 {
		if w.rng == nil {
			w.rng = newRNG()
		}
		ang += (w.rng.Float64()*2 - 1) * spread
	}
	w.spawnMG(u, ang, st.Range)
}

func (w *World) nearestEnemy(u *Unit, maxRange float64) (tx, ty float64, ok bool) {
	best := maxRange
	for i := range w.Units {
		o := &w.Units[i]
		if o.Side != SideEnemy || !o.Living() {
			continue
		}
		d := hypot(o.X-u.X, o.Y-u.Y)
		if d > best || !w.lineClear(u.X, u.Y, o.X, o.Y) {
			continue
		}
		best = d
		tx, ty = o.X, o.Y
		ok = true
	}
	return tx, ty, ok
}

func (w *World) shootAt(u *Unit, dx, dy float64, st GunStats, dt float64) {
	if dx == 0 && dy == 0 {
		dx = 1
	}
	u.Facing = math.Atan2(dy, dx)
	u.FireCD -= dt
	if u.FireCD > 0 {
		return
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

func (w *World) spawnMG(u *Unit, ang float64, maxRange float64) {
	nx, ny := math.Cos(ang), math.Sin(ang)
	muzzle := float64(UnitSize) / 2
	w.addMG(u.X+nx*muzzle, u.Y+ny*muzzle, ang, maxRange, u.ID, u.Side)
}

// addMG spawns one round and records a gun cue at the muzzle.
func (w *World) addMG(x, y, ang, travel float64, owner int, side Side) {
	nx, ny := math.Cos(ang), math.Sin(ang)
	w.Projectiles = append(w.Projectiles, Projectile{
		Kind:      ProjMG,
		X:         x,
		Y:         y,
		VX:        nx * MGSpeed,
		VY:        ny * MGSpeed,
		OwnerID:   owner,
		OwnerSide: side,
		Left:      travel,
		Alive:     true,
	})
	w.emit(CueGun, x, y)
	if u := w.Unit(owner); u != nil {
		u.SinceShot = 0
	}
	w.wakeFromMG(x, y, owner)
}

func (w *World) stepProjectiles(dt float64) {
	out := w.Projectiles[:0]
	for i := range w.Projectiles {
		p := &w.Projectiles[i]
		if !p.Alive {
			continue
		}
		x0, y0 := p.X, p.Y
		speed := hypot(p.VX, p.VY)
		if speed < 1 {
			continue
		}
		step := speed * dt
		if step > p.Left {
			step = p.Left
		}
		nx, ny := p.VX/speed, p.VY/speed
		p.X += nx * step
		p.Y += ny * step
		p.Left -= step
		if p.Kind == ProjRocket {
			if hx, hy, hit := w.rocketHit(p.OwnerID, x0, y0, p.X, p.Y); hit || p.Left <= 0 {
				if !hit {
					hx, hy = p.X, p.Y
				}
				w.explode(hx, hy, p.OwnerID)
				continue
			}
			out = append(out, *p)
			continue
		}
		blocked := false
		if hx, hy, hit := w.Map.FirstSolid(x0, y0, p.X, p.Y); hit {
			p.X, p.Y = hx, hy
			p.Left = 0
			blocked = true
		}
		if hx, hy, hit := w.buildingHit(x0, y0, p.X, p.Y); hit {
			p.X, p.Y = hx, hy
			p.Left = 0
			blocked = true
		}
		w.tryHit(p, x0, y0, p.X, p.Y)
		if p.Alive {
			w.tryHitCrate(p, x0, y0, p.X, p.Y)
		}
		if blocked {
			p.Alive = false
		}
		if p.Alive && p.Left > 0 {
			out = append(out, *p)
		}
	}
	w.Projectiles = out
}

func (w *World) tryHit(p *Projectile, x0, y0, x1, y1 float64) {
	for i := range w.Units {
		u := &w.Units[i]
		if u.ID == p.OwnerID || u.VehicleID != 0 {
			continue
		}
		if !w.mgCanHurt(p, u) {
			continue
		}
		if !segmentHitsCircle(x0, y0, x1, y1, u.X, u.Y, MGHitR) {
			continue
		}
		w.mgStrike(p, u)
		p.Alive = false
		return
	}
}

// mgStrike wounds or kills a man on his feet, finishes a wounded man, and
// launches a corpse. Player MG still ignores a friendly who is on his feet.
func (w *World) mgStrike(p *Projectile, u *Unit) {
	if u.Dead() {
		w.juggle(u, p)
		return
	}
	if u.Wounded() || u.Sinking || !w.rollWound() {
		w.kill(u)
		w.creditKill(p.OwnerID, u)
		return
	}
	w.wound(u)
}

func (w *World) rollWound() bool {
	if w.WoundChance <= 0 {
		return false
	}
	if w.WoundChance >= 1 {
		return true
	}
	if w.rng == nil {
		w.rng = newRNG()
	}
	return w.rng.Float64() < w.WoundChance
}

func (w *World) wound(u *Unit) {
	if u == nil || u.HP != Alive {
		return
	}
	u.HP = Wounded
	u.VX = 0
	u.VY = 0
	u.GrenadeWind = 0
	u.RocketWind = 0
	u.WindHeld = false
	w.dropFromFile(u.ID)
}

// juggle is the corpse easter egg: an MG round kicks the body along the shot
// and pops it off the ground. After juggleLife the body no longer reacts.
func (w *World) juggle(u *Unit, p *Projectile) {
	if u.DeadFor > juggleLife {
		return
	}
	sp := hypot(p.VX, p.VY)
	if sp < 1 {
		return
	}
	u.VX += p.VX / sp * juggleKick
	u.VY += p.VY / sp * juggleKick
	along := hypot(u.VX, u.VY)
	if along > juggleKick {
		u.VX = u.VX / along * juggleKick
		u.VY = u.VY / along * juggleKick
	}
	if u.VZ < juggleHop {
		u.VZ = juggleHop
	}
}

func (w *World) stepBodies(dt float64) {
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Dead() || u.VehicleID != 0 {
			continue
		}
		u.DeadFor += dt
		if u.Hop == 0 && u.VZ == 0 && u.VX == 0 && u.VY == 0 {
			continue
		}
		u.VZ -= juggleGravity * dt
		u.Hop += u.VZ * dt
		if u.Hop <= 0 {
			u.Hop = 0
			if u.VZ < 0 {
				u.VZ = -u.VZ * juggleBounce
				if u.VZ < juggleRest {
					u.VZ = 0
				}
			}
			sp := hypot(u.VX, u.VY)
			if sp > 0 {
				drop := juggleDrag * dt
				if drop >= sp {
					u.VX, u.VY = 0, 0
				} else {
					u.VX -= u.VX / sp * drop
					u.VY -= u.VY / sp * drop
				}
			}
		}
		w.shoveCorpse(u, u.VX*dt, u.VY*dt)
	}
}

// shoveCorpse slides a body and stops it on a solid tile. It does not sink,
// and it does not step off a cliff.
func (w *World) shoveCorpse(u *Unit, dx, dy float64) {
	nx, ny := u.X+dx, u.Y+dy
	if w.walkableUnit(nx, ny) {
		u.X, u.Y = nx, ny
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
	u.VX, u.VY = 0, 0
}

// Player MG does not harm living friendlies. Explosives later ignore this.
func (w *World) mgCanHurt(p *Projectile, u *Unit) bool {
	if p.OwnerSide == SidePlayer && u.Side == SidePlayer && u.Living() {
		return false
	}
	return true
}

// creditKill gives the owner one point when the victim is an enemy.
// Grenades, rockets, and the mounted gun already name the trooper who
// used them. A dead owner still scores: the blast can take him too.
func (w *World) creditKill(ownerID int, victim *Unit) {
	if victim == nil || victim.Side != SideEnemy {
		return
	}
	owner := w.Unit(ownerID)
	if owner == nil || owner.ID == victim.ID || owner.Side != SidePlayer {
		return
	}
	owner.Kills++
}

func (w *World) kill(u *Unit) {
	if u.Dead() {
		return
	}
	u.HP = Dead
	u.VX = 0
	u.VY = 0
	w.dropFromFile(u.ID)
	w.emitFrom(CueDeath, u.X, u.Y, u.ID)
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

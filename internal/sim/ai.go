package sim

import "math"

const (
	EnemyMGRoF       = 4.0  // slower than player 8/s
	EnemyMGRange     = 70.0 // shorter than a Private's 80
	EnemyReact       = 0.55 // seconds of contact before the first round
	EnemyReactJitter = 0.15 // added once per contact, in [-j, +j]
	EnemyTurnRate    = 4.0  // rad/s. A 180° turn takes about 0.8s
	EnemyFaceTol     = 0.35 // rad. No round until he is looking at the player
	EnemyBurst       = 3    // MG rounds, then silence
	EnemyBurstPause  = 0.75 // seconds with no MG round; he keeps turning
	EnemyMGSpread    = 0.20 // rad. Wider than a Private's 0.12
	EnemyHear        = 96.0 // px. An infantry MG round starts a reaction with no LOS
)

// stepAI posts grunts: they hold their tile, turn toward a player in gun
// range with clear LOS, withhold the first round until the reaction and
// the turn are both done, then fire a short burst and pause. Rocketeers
// and grenade throws keep their own windups.
func (w *World) stepAI(dt float64) {
	if !w.AI {
		return
	}
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != SideEnemy || !u.Living() || u.VehicleID != 0 {
			continue
		}
		if u.Sinking {
			u.VX = 0
			u.VY = 0
			u.HasPost = false
			continue
		}
		if u.HasPost {
			w.stepDoorPost(u, dt)
			continue
		}
		px, py, ok := w.nearestLiving(SidePlayer, u.X, u.Y)
		if !ok {
			u.VX = 0
			u.VY = 0
			u.GrenadeWind = 0
			u.RocketWind = 0
			resetGruntContact(u)
			continue
		}
		dist := hypot(px-u.X, py-u.Y)
		if u.Kind == KindRocketeer {
			w.stepRocketeer(u, px, py, dist, dt)
			continue
		}
		if w.stepGrenadier(u, px, py, dist, dt) {
			continue
		}
		w.stepGruntGun(u, dt)
	}
}

// stepDoorPost walks a door grunt straight to his post. He does not shoot
// and he does not follow the player. A step that cannot move posts him there.
func (w *World) stepDoorPost(u *Unit, dt float64) {
	prevX, prevY := u.X, u.Y
	arrived := w.steerToward(u, u.PostX, u.PostY, WalkSpeed, dt, ArrivalRadius)
	if arrived || (u.X == prevX && u.Y == prevY) {
		u.HasPost = false
		u.VX = 0
		u.VY = 0
	}
}

// stepGruntGun is the MG path for a grunt and for a grenadier who is not
// throwing. He never leaves his tile. A round needs a living player inside
// gun range with clear LOS. A heard infantry shot turns him toward the
// shooter and runs the same clock, and it never fires by itself.
func (w *World) stepGruntGun(u *Unit, dt float64) {
	u.VX = 0
	u.VY = 0
	px, py, see := w.gruntContact(u)
	nx, ny, hear := w.gruntNoise(u)
	if !see && !hear {
		resetGruntContact(u)
		return
	}
	if !see {
		px, py = nx, ny
	}
	if u.ReactAt == 0 {
		if w.rng == nil {
			w.rng = newRNG()
		}
		u.ReactAt = EnemyReact + (w.rng.Float64()*2-1)*EnemyReactJitter
	}
	u.SpotT += dt
	w.turnToward(u, px, py, dt)
	if !see {
		return
	}
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
	if facingError(u, px, py) > EnemyFaceTol {
		return
	}
	if !w.CanShoot(u) {
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
		u.FireCD = 1.0 / EnemyMGRoF
	}
	if w.rng == nil {
		w.rng = newRNG()
	}
	ang := u.Facing + (w.rng.Float64()*2-1)*EnemyMGSpread
	w.spawnMG(u, ang, EnemyMGRange)
}

// resetGruntContact drops a grunt back to idle. The next sighting owes a
// full reaction, including one new jitter roll.
func resetGruntContact(u *Unit) {
	u.SpotT = 0
	u.ReactAt = 0
	u.BurstN = 0
	u.BurstGap = 0
	u.HearID = 0
}

// gruntNoise is the living infantry shooter who woke this grunt. The shot
// that set HearID can be anywhere inside EnemyHear. Dead shooters drop it.
func (w *World) gruntNoise(u *Unit) (x, y float64, ok bool) {
	if u == nil || u.HearID == 0 {
		return 0, 0, false
	}
	s := w.Unit(u.HearID)
	if s == nil || !s.Living() {
		u.HearID = 0
		return 0, 0, false
	}
	return s.X, s.Y, true
}

// wakeFromMG starts a posted grunt's reaction when an infantry machine-gun
// round lands inside EnemyHear. He must be idle. Grenades, rockets, and
// vehicle guns never call this. The grunt does not move.
func (w *World) wakeFromMG(x, y float64, owner int) {
	src := w.Unit(owner)
	if src == nil || !src.Living() || src.VehicleID != 0 {
		return
	}
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != SideEnemy || u.Kind != KindInfantry || !u.Living() || u.VehicleID != 0 {
			continue
		}
		if u.ReactAt != 0 || u.HearID != 0 {
			continue
		}
		if hypot(u.X-x, u.Y-y) > EnemyHear {
			continue
		}
		u.HearID = owner
	}
}

// gruntContact is the nearest living player inside gun range with LOS.
func (w *World) gruntContact(u *Unit) (px, py float64, ok bool) {
	best := EnemyMGRange
	for i := range w.Units {
		o := &w.Units[i]
		if o.Side != SidePlayer || !o.Living() {
			continue
		}
		d := hypot(o.X-u.X, o.Y-u.Y)
		if d > EnemyMGRange || (ok && d >= best) {
			continue
		}
		if !w.lineClear(u.X, u.Y, o.X, o.Y) {
			continue
		}
		best = d
		px, py = o.X, o.Y
		ok = true
	}
	return px, py, ok
}

func (w *World) turnToward(u *Unit, px, py, dt float64) {
	dx, dy := px-u.X, py-u.Y
	if dx == 0 && dy == 0 {
		return
	}
	d := wrapAngle(math.Atan2(dy, dx) - u.Facing)
	step := EnemyTurnRate * dt
	if d > step {
		d = step
	} else if d < -step {
		d = -step
	}
	u.Facing = wrapAngle(u.Facing + d)
}

func facingError(u *Unit, px, py float64) float64 {
	dx, dy := px-u.X, py-u.Y
	if dx == 0 && dy == 0 {
		return 0
	}
	return math.Abs(wrapAngle(math.Atan2(dy, dx) - u.Facing))
}

func wrapAngle(d float64) float64 {
	for d > math.Pi {
		d -= 2 * math.Pi
	}
	for d < -math.Pi {
		d += 2 * math.Pi
	}
	return d
}

// stepGrenadier spends the frame on a telegraphed throw.
// Between bombs the grenadier fights as a grunt.
func (w *World) stepGrenadier(u *Unit, px, py, dist, dt float64) bool {
	if u.Kind != KindGrenadier {
		return false
	}
	if u.GrenadeWind > 0 {
		u.VX = 0
		u.VY = 0
		u.Facing = math.Atan2(py-u.Y, px-u.X)
		u.GrenadeWind -= dt
		if u.GrenadeWind > 0 {
			return true
		}
		u.GrenadeWind = 0
		if u.Bombs > 0 && w.CanShoot(u) {
			w.launchGrenade(u, px, py)
			u.Bombs--
			u.GrenadeCD = GrenadierCooldown
		}
		return true
	}
	if u.GrenadeCD > 0 {
		u.GrenadeCD -= dt
		if u.GrenadeCD < 0 {
			u.GrenadeCD = 0
		}
	}
	if u.Bombs <= 0 || u.GrenadeCD > 0 || !w.CanShoot(u) {
		return false
	}
	if dist < GrenadierMinRange || dist > GrenadeRange {
		return false
	}
	u.GrenadeWind = GrenadierWindup
	u.VX = 0
	u.VY = 0
	u.Facing = math.Atan2(py-u.Y, px-u.X)
	return true
}

const (
	RocketeerRange      = RocketRange
	RocketeerWindup     = 0.55
	RocketeerCooldown   = 3.2
	RocketeerFirstDelay = 2.0
	RocketeerApproach   = 90.0 // a rocketeer shuffles in; a grunt holds his post
	RocketeerSpeed      = 12.0
	RocketeerMinRange   = GrenadeRadius + 8
)

// stepRocketeer fires a slow rocket. Beside a tree they hold still instead of chasing.
func (w *World) stepRocketeer(u *Unit, px, py, dist, dt float64) {
	hiding := w.treeAdjacent(u)
	if u.RocketWind > 0 {
		u.VX = 0
		u.VY = 0
		u.Facing = math.Atan2(py-u.Y, px-u.X)
		u.RocketWind -= dt
		if u.RocketWind > 0 {
			return
		}
		u.RocketWind = 0
		if w.CanShoot(u) {
			w.launchRocket(u.ID, u.Side, u.X, u.Y, px, py)
			u.RocketCD = RocketeerCooldown
		}
		return
	}
	if u.RocketCD > 0 {
		u.RocketCD -= dt
		if u.RocketCD < 0 {
			u.RocketCD = 0
		}
	}
	if hiding {
		u.VX = 0
		u.VY = 0
	} else if dist <= RocketeerApproach && dist > RocketeerMinRange {
		w.steerToward(u, px, py, RocketeerSpeed, dt, ArrivalRadius)
	} else {
		u.VX = 0
		u.VY = 0
	}
	if u.RocketCD > 0 || !w.CanShoot(u) {
		return
	}
	if dist < RocketeerMinRange || dist > RocketeerRange {
		return
	}
	if !w.lineClear(u.X, u.Y, px, py) {
		return
	}
	u.RocketWind = RocketeerWindup
	u.VX = 0
	u.VY = 0
	u.Facing = math.Atan2(py-u.Y, px-u.X)
}

// treeAdjacent reports a tree in the eight neighbouring cells.
// Off-map cells are solid but are not cover.
func (w *World) treeAdjacent(u *Unit) bool {
	if u == nil || !w.Map.active() {
		return false
	}
	tx := int(math.Floor(u.X / float64(TileSize)))
	ty := int(math.Floor(u.Y / float64(TileSize)))
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := tx+dx, ty+dy
			if nx < 0 || ny < 0 || nx >= w.Map.W || ny >= w.Map.H {
				continue
			}
			if w.Map.At(nx, ny) == TileTree {
				return true
			}
		}
	}
	return false
}

func (w *World) nearestLiving(side Side, x, y float64) (px, py float64, ok bool) {
	best := 1e12
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != side || !u.Living() {
			continue
		}
		d := hypot(u.X-x, u.Y-y)
		if d < best {
			best = d
			px, py = u.X, u.Y
			ok = true
		}
	}
	return px, py, ok
}

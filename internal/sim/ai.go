package sim

import "math"

const (
	EnemyMGRoF     = 4.0  // slower than player 8/s
	EnemyMGRange   = 70.0 // shorter than a Private's 80
	EnemyWalkSpeed = 18.0
	EnemyApproach  = 140.0
)

// stepAI: shoot if a player is in range and has LOS (trees block);
// otherwise walk slowly toward a nearby player; else idle.
func (w *World) stepAI(dt float64) {
	if !w.AI {
		return
	}
	interval := 1.0 / EnemyMGRoF
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != SideEnemy || !u.Living() || u.VehicleID != 0 {
			continue
		}
		if u.Sinking {
			u.VX = 0
			u.VY = 0
			continue
		}
		px, py, ok := w.nearestLiving(SidePlayer, u.X, u.Y)
		if !ok {
			u.VX = 0
			u.VY = 0
			u.GrenadeWind = 0
			u.RocketWind = 0
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
		if dist <= EnemyMGRange && w.CanShoot(u) && w.lineClear(u.X, u.Y, px, py) {
			u.VX = 0
			u.VY = 0
			dx, dy := px-u.X, py-u.Y
			if dx == 0 && dy == 0 {
				dx = 1
			}
			u.Facing = math.Atan2(dy, dx)
			u.FireCD -= dt
			if u.FireCD > 0 {
				continue
			}
			u.FireCD = interval
			ang := u.Facing
			if w.Spread > 0 {
				ang += (w.rng.Float64()*2 - 1) * w.Spread
			}
			w.spawnMG(u, ang, EnemyMGRange)
			continue
		}
		if dist <= EnemyApproach {
			w.steerToward(u, px, py, EnemyWalkSpeed, dt, ArrivalRadius)
			continue
		}
		u.VX = 0
		u.VY = 0
	}
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
	RocketeerApproach   = 90.0 // shorter than a grunt's 140, so they stay in cover
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

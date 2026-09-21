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
		if u.Side != SideEnemy || !u.Living() {
			continue
		}
		px, py, ok := w.nearestLiving(SidePlayer, u.X, u.Y)
		if !ok {
			u.VX = 0
			u.VY = 0
			continue
		}
		dist := hypot(px-u.X, py-u.Y)
		if dist <= EnemyMGRange && w.CanShoot(u) && w.Map.HasLOS(u.X, u.Y, px, py) {
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

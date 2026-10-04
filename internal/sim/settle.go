package sim

// Settle keeps pictures that were already moving after the phase ends:
// shots and rockets in the air, grenades, shed roofs, blasts, a throw that had started,
// a skidding vehicle, hopping bodies, and the shoot and throw clocks.
// It does not take orders, and it does not start new fire or a new windup.
func (w *World) Settle(dt float64) {
	if w == nil || dt <= 0 || w.Status == Playing {
		return
	}
	w.finishWindups(dt)
	w.stepProjectiles(dt)
	w.stepGrenades(dt)
	w.stepRoofs(dt)
	w.stepBlasts(dt)
	w.stepBodies(dt)
	w.coastVehicles(dt)
	w.stepAnimClocks(dt)
}

// Pending reports effects that can still change who is alive. The battle
// waits for these before it tallies the phase. A blast that has already
// landed, and a shoot pose, are not pending.
func (w *World) Pending() bool {
	if w == nil || w.Status == Playing {
		return false
	}
	if len(w.Grenades) > 0 || len(w.Projectiles) > 0 {
		return true
	}
	for i := range w.Roofs {
		if w.Roofs[i].Flying {
			return true
		}
	}
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Living() {
			continue
		}
		if u.GrenadeWind > 0 || u.RocketWind > 0 || u.WindHeld {
			return true
		}
	}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if v.Alive && hypot(v.VX, v.VY) > 6 {
			return true
		}
	}
	return false
}

// finishWindups releases a grenade or rocket whose telegraph had already
// started. A man who had not begun one is left alone.
func (w *World) finishWindups(dt float64) {
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != SideEnemy || !u.Living() || u.VehicleID != 0 {
			continue
		}
		if u.GrenadeWind <= 0 && u.RocketWind <= 0 && !u.WindHeld {
			continue
		}
		px, py, ok := w.nearestLiving(SidePlayer, u.X, u.Y)
		if !ok {
			u.GrenadeWind = 0
			u.RocketWind = 0
			u.WindHeld = false
			continue
		}
		dist := hypot(px-u.X, py-u.Y)
		if u.Kind == KindRocketeer || u.RocketWind > 0 {
			w.stepRocketeer(u, px, py, dist, dt)
			continue
		}
		w.stepGrenadier(u, px, py, dist, dt)
	}
}

// coastVehicles lets a hull that is already moving bleed off. Nothing new
// drives or shoots.
func (w *World) coastVehicles(dt float64) {
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		v.Blink += dt
		w.applyDrive(v, v.X, v.Y, false, dt)
		w.syncCrew(v)
		w.ram(v)
	}
}

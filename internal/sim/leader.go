package sim

// stepLeaders keeps a leader on his tile, fighting as a posted grunt, until
// a living player on foot touches him. After that he stops shooting, leaves
// the enemy side, and walks behind the active squad. A handover saves him.
func (w *World) stepLeaders(dt float64) {
	slot := 0
	for i := range w.Units {
		u := &w.Units[i]
		if u.Kind != KindLeader || u.Dead() || u.Saved || u.Sinking {
			continue
		}
		if !u.Captured && w.playerTouches(u) {
			u.Captured = true
			u.Side = SideCaptured
			u.VX = 0
			u.VY = 0
			u.SinceShot = animRest
			u.GrenadeWind = 0
			u.RocketWind = 0
		}
		if !u.Captured {
			continue
		}
		w.stepEscort(u, slot, dt)
		slot++
		w.saveLeader(u)
	}
}

// saveLeader removes a captured leader whose centre is inside a handover.
func (w *World) saveLeader(u *Unit) {
	if u == nil || !u.Captured || u.Saved || u.Dead() {
		return
	}
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if b.Kind != HutHandover || !b.Holds(u.X, u.Y) {
			continue
		}
		u.Saved = true
		u.VX = 0
		u.VY = 0
		u.Escorting = false
		w.emit(CueSave, u.X, u.Y)
		return
	}
}

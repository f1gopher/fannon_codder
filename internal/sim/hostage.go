package sim

// stepHostages keeps a hostage on his tile until a living player on foot
// touches him. After that he walks one file slot behind the active squad
// and stops when that squad stops. He does not shoot, board, or take orders.
// Standing on a tent saves him.
func (w *World) stepHostages(dt float64) {
	slot := 0
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != SideHostage || u.Dead() || u.Saved || u.Sinking {
			continue
		}
		if !u.Escorting && w.playerTouches(u) {
			u.Escorting = true
		}
		if u.Escorting {
			w.stepEscort(u, slot, dt)
			slot++
		} else {
			u.VX = 0
			u.VY = 0
		}
		w.saveHostage(u)
	}
}

// playerTouches is a living player on foot whose centre meets u.
func (w *World) playerTouches(u *Unit) bool {
	for i := range w.Units {
		p := &w.Units[i]
		if p.Side != SidePlayer || !p.Living() || p.VehicleID != 0 {
			continue
		}
		if hypot(p.X-u.X, p.Y-u.Y) <= float64(UnitSize) {
			return true
		}
	}
	return false
}

// stepEscort walks u behind the active squad. slot 0 is one FileSpacing
// behind the last man. A stopped squad, including a kneel, holds him still.
func (w *World) stepEscort(u *Unit, slot int, dt float64) {
	s := w.ActiveSquad()
	if s == nil || s.Kneel || !s.HasDest {
		u.VX = 0
		u.VY = 0
		return
	}
	leader := w.Unit(s.LeaderID)
	if leader == nil || !leader.Living() || leader.VehicleID != 0 {
		u.VX = 0
		u.VY = 0
		return
	}
	behind := FileSpacing * float64(len(s.MemberIDs)+slot)
	target, _ := trailPoint(Vec2{X: leader.X, Y: leader.Y}, s.Trail, behind)
	w.steerToward(u, target.X, target.Y, WalkSpeed, dt, FollowerArrive)
}

// saveHostage removes a man whose centre is inside a tent.
func (w *World) saveHostage(u *Unit) {
	if u == nil || u.Saved || u.Dead() {
		return
	}
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if b.Kind != HutTent || !b.Holds(u.X, u.Y) {
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

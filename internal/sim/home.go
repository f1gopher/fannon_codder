package sim

// saveHome removes a marked civilian whose centre is inside a home.
func (w *World) saveHome(u *Unit) {
	if u == nil || !u.MustHome || u.Saved || u.Dead() {
		return
	}
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if b.Kind != HutHome || !b.Holds(u.X, u.Y) {
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

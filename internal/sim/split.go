package sim

// ToggleSelect marks or unmarks a living member of the active squad for a split.
func (w *World) ToggleSelect(id int) bool {
	s := w.ActiveSquad()
	if s == nil || !w.memberOf(s, id) {
		return false
	}
	u := w.Unit(id)
	if u == nil || !u.Living() {
		return false
	}
	for i, sel := range w.Selected {
		if sel == id {
			w.Selected = append(w.Selected[:i], w.Selected[i+1:]...)
			return true
		}
	}
	w.Selected = append(w.Selected, id)
	return true
}

// CycleGrenadeShare / CycleRocketShare step the three-way split toggle.
func (w *World) CycleGrenadeShare() { w.GrenadeShare = w.GrenadeShare.Next() }

func (w *World) CycleRocketShare() { w.RocketShare = w.RocketShare.Next() }

// SetActiveSquad makes id the controlled squad. The one left behind holds.
func (w *World) SetActiveSquad(id SquadID) bool {
	idx := -1
	for i := range w.Squads {
		if w.Squads[i].ID == id && squadLiving(w, &w.Squads[i]) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	for i := range w.Squads {
		if w.Squads[i].Active && i != idx {
			w.Squads[i].HasDest = false
		}
		w.Squads[i].Active = i == idx
	}
	w.Selected = nil
	return true
}

// Split moves the highlighted men into a new squad (Eagle, then Panther).
// At least one man must stay, and there are never more than three squads.
func (w *World) Split() bool {
	newID, ok := w.freeSquadID()
	if !ok {
		return false
	}
	active := w.ActiveSquad()
	if active == nil {
		return false
	}
	sel := map[int]bool{}
	for _, id := range w.Selected {
		sel[id] = true
	}
	var stay, moving []int
	for _, mid := range active.MemberIDs {
		u := w.Unit(mid)
		if u == nil || !u.Living() {
			continue
		}
		if sel[mid] {
			moving = append(moving, mid)
		} else {
			stay = append(stay, mid)
		}
	}
	if len(moving) == 0 || len(stay) == 0 {
		return false
	}
	gStay, gGo := shareAmmo(active.Grenades, w.GrenadeShare)
	rStay, rGo := shareAmmo(active.Rockets, w.RocketShare)
	active.Grenades = gStay
	active.Rockets = rStay
	active.MemberIDs = stay
	leaderStays := false
	for _, id := range stay {
		if id == active.LeaderID {
			leaderStays = true
			break
		}
	}
	if !leaderStays {
		active.LeaderID = stay[0]
		active.HasDest = false
	}
	for _, mid := range moving {
		if u := w.Unit(mid); u != nil {
			u.SquadID = newID
		}
	}
	ns := Squad{
		ID:        newID,
		LeaderID:  moving[0],
		MemberIDs: moving,
		Grenades:  gGo,
		Rockets:   rGo,
	}
	if l := w.Unit(ns.LeaderID); l != nil {
		ns.Trail = []Vec2{{X: l.X, Y: l.Y}}
	}
	w.Squads = append(w.Squads, ns)
	w.Selected = nil
	return true
}

func (w *World) freeSquadID() (SquadID, bool) {
	if len(w.Squads) >= 3 {
		return 0, false
	}
	used := [3]bool{}
	for i := range w.Squads {
		id := w.Squads[i].ID
		if id >= 0 && int(id) < len(used) {
			used[id] = true
		}
	}
	for _, id := range []SquadID{SquadEagle, SquadPanther, SquadSnake} {
		if !used[id] {
			return id, true
		}
	}
	return 0, false
}

func shareAmmo(total int, mode AmmoShare) (stay, give int) {
	if total < 0 {
		total = 0
	}
	switch mode {
	case ShareAll:
		return 0, total
	case ShareHalf:
		give = total / 2
		return total - give, give
	default:
		return total, 0
	}
}

func (w *World) stepMerge() {
	for {
		active := w.ActiveSquad()
		if active == nil {
			return
		}
		dstID := active.ID
		var srcID SquadID
		found := false
		for i := range w.Squads {
			s := &w.Squads[i]
			if s.ID == dstID {
				continue
			}
			if w.squadsTouch(active, s) {
				srcID = s.ID
				found = true
				break
			}
		}
		if !found {
			return
		}
		w.absorbInto(dstID, srcID)
	}
}

func (w *World) squadsTouch(a, b *Squad) bool {
	for _, ida := range a.MemberIDs {
		ua := w.Unit(ida)
		if ua == nil || !ua.Living() {
			continue
		}
		for _, idb := range b.MemberIDs {
			ub := w.Unit(idb)
			if ub == nil || !ub.Living() {
				continue
			}
			if hypot(ua.X-ub.X, ua.Y-ub.Y) <= MergeRadius {
				return true
			}
		}
	}
	return false
}

func (w *World) absorbInto(dstID, srcID SquadID) {
	dstIdx, srcIdx := -1, -1
	for i := range w.Squads {
		if w.Squads[i].ID == dstID {
			dstIdx = i
		}
		if w.Squads[i].ID == srcID {
			srcIdx = i
		}
	}
	if dstIdx < 0 || srcIdx < 0 {
		return
	}
	src := w.Squads[srcIdx]
	for _, id := range src.MemberIDs {
		u := w.Unit(id)
		if u == nil || !u.Living() {
			continue
		}
		u.SquadID = dstID
		w.Squads[dstIdx].MemberIDs = append(w.Squads[dstIdx].MemberIDs, id)
	}
	w.Squads[dstIdx].Grenades += src.Grenades
	w.Squads[dstIdx].Rockets += src.Rockets
	w.removeSquad(srcID)
}

func (w *World) removeSquad(id SquadID) {
	n := 0
	for _, s := range w.Squads {
		if s.ID == id {
			continue
		}
		w.Squads[n] = s
		n++
	}
	w.Squads = w.Squads[:n]
}

func (w *World) pruneSquads() {
	n := 0
	for i := range w.Squads {
		if squadLiving(w, &w.Squads[i]) {
			w.Squads[n] = w.Squads[i]
			n++
		}
	}
	w.Squads = w.Squads[:n]
	active := false
	for i := range w.Squads {
		if w.Squads[i].Active {
			active = true
			break
		}
	}
	if !active && len(w.Squads) > 0 {
		w.Squads[0].Active = true
	}
	w.scrubSelection()
}

func (w *World) scrubSelection() {
	s := w.ActiveSquad()
	if s == nil {
		w.Selected = nil
		return
	}
	dst := w.Selected[:0]
	for _, id := range w.Selected {
		if w.memberOf(s, id) {
			dst = append(dst, id)
		}
	}
	w.Selected = dst
}

func (w *World) memberOf(s *Squad, id int) bool {
	for _, m := range s.MemberIDs {
		if m == id {
			return true
		}
	}
	return false
}

func squadLiving(w *World, s *Squad) bool {
	for _, id := range s.MemberIDs {
		if u := w.Unit(id); u != nil && u.Living() {
			return true
		}
	}
	return false
}

// SquadByID returns the player squad with that troop id, or nil.
func (w *World) SquadByID(id SquadID) *Squad {
	for i := range w.Squads {
		if w.Squads[i].ID == id {
			return &w.Squads[i]
		}
	}
	return nil
}

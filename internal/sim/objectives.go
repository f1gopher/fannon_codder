package sim

// Status is the phase outcome.
type Status int

const (
	Playing Status = iota
	Won
	Lost
)

// Objective is a win condition. Both may be required on later maps.
type Objective int

const (
	KillAllEnemy Objective = iota
	DestroyEnemyBuildings
	// ProtectCivilians loses the phase when a civilian dies. A wound does not.
	ProtectCivilians
)

func ParseObjective(s string) (Objective, bool) {
	switch s {
	case "kill_all_enemy":
		return KillAllEnemy, true
	case "destroy_enemy_buildings":
		return DestroyEnemyBuildings, true
	case "protect_civilians":
		return ProtectCivilians, true
	default:
		return 0, false
	}
}

func (w *World) anyLiving(side Side) bool {
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side == side && u.Living() {
			return true
		}
	}
	return false
}

// anyRemaining is a man still on the field, including one who is down and
// squirming. A wounded enemy blocks kill-all until he is finished.
func (w *World) anyRemaining(side Side) bool {
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side == side && !u.Dead() {
			return true
		}
	}
	return false
}

// HoldOutcome is true when the phase would already be over, but a round,
// rocket, grenade, blast, or flying roof can still kill someone. The result
// waits until those finish.
func (w *World) HoldOutcome() bool {
	if w == nil || w.Status != Playing || len(w.Objectives) == 0 || !w.unresolvedHarm() {
		return false
	}
	if !w.anyLiving(SidePlayer) || w.protectFailed() {
		return true
	}
	return w.objectivesMet()
}

// protectFailed is a dead civilian on a phase that must keep them alive.
// A man who is only wounded has not failed it yet.
func (w *World) protectFailed() bool {
	if !w.hasObjective(ProtectCivilians) {
		return false
	}
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side == SideCivilian && u.Dead() {
			return true
		}
	}
	return false
}

func (w *World) hasObjective(o Objective) bool {
	for _, have := range w.Objectives {
		if have == o {
			return true
		}
	}
	return false
}

// unresolvedHarm is something already started that can still change who lives.
// A blast's damage landed when it opened; it still holds the result so the
// phase is not called until the explosion has finished.
func (w *World) unresolvedHarm() bool {
	if len(w.Grenades) > 0 || len(w.Projectiles) > 0 || len(w.Explosions) > 0 {
		return true
	}
	for i := range w.Roofs {
		if w.Roofs[i].Flying {
			return true
		}
	}
	return false
}

func (w *World) objectivesMet() bool {
	for _, o := range w.Objectives {
		switch o {
		case KillAllEnemy:
			if w.anyRemaining(SideEnemy) {
				return false
			}
		case DestroyEnemyBuildings:
			if w.anyDoorBuilding() {
				return false
			}
		case ProtectCivilians:
			if w.protectFailed() {
				return false
			}
		}
	}
	return true
}

func (w *World) evaluateObjectives() {
	if w.Status != Playing || len(w.Objectives) == 0 {
		return
	}
	if w.unresolvedHarm() {
		return
	}
	if !w.anyLiving(SidePlayer) || w.protectFailed() {
		w.Status = Lost
		w.Firing = false
		return
	}
	if !w.objectivesMet() {
		return
	}
	w.Status = Won
	w.Firing = false
}

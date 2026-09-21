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
)

func ParseObjective(s string) (Objective, bool) {
	switch s {
	case "kill_all_enemy":
		return KillAllEnemy, true
	case "destroy_enemy_buildings":
		return DestroyEnemyBuildings, true
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

func (w *World) evaluateObjectives() {
	if w.Status != Playing || len(w.Objectives) == 0 {
		return
	}
	if !w.anyLiving(SidePlayer) {
		w.Status = Lost
		w.Firing = false
		return
	}
	for _, o := range w.Objectives {
		switch o {
		case KillAllEnemy:
			if w.anyLiving(SideEnemy) {
				return
			}
		case DestroyEnemyBuildings:
			if w.anyDoorBuilding() {
				return
			}
		}
	}
	w.Status = Won
	w.Firing = false
}

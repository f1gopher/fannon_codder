package mapedit

import (
	"fmt"

	"fannon-codder/internal/data"
)

// memo is one phase's placements for this session, including unsaved edits.
type memo struct {
	phase     *data.Phase
	start     [2]int
	hasStart  bool
	enemies   []data.Enemy
	buildings []data.BuildingSpec
	pickups   []data.PickupSpec
	barrels   []data.BarrelSpec
	civilians []data.Enemy
	hostages  []data.Enemy
	vehicles  []data.VehicleSpec
	dirty     bool
}

func cloneEnemies(in []data.Enemy) []data.Enemy {
	out := append([]data.Enemy(nil), in...)
	for i := range out {
		out[i].Line = append([][2]int(nil), in[i].Line...)
	}
	return out
}

func cloneVehicles(in []data.VehicleSpec) []data.VehicleSpec {
	out := append([]data.VehicleSpec(nil), in...)
	for i := range out {
		out[i].Line = append([][2]int(nil), in[i].Line...)
	}
	return out
}

func (e *Editor) takeMemo() memo {
	return memo{
		phase:     e.phase,
		start:     e.start,
		hasStart:  e.hasStart,
		enemies:   cloneEnemies(e.enemies),
		buildings: append([]data.BuildingSpec(nil), e.buildings...),
		pickups:   append([]data.PickupSpec(nil), e.pickups...),
		barrels:   append([]data.BarrelSpec(nil), e.barrels...),
		civilians: append([]data.Enemy(nil), e.civilians...),
		hostages:  append([]data.Enemy(nil), e.hostages...),
		vehicles:  cloneVehicles(e.vehicles),
		dirty:     e.dirty,
	}
}

func (e *Editor) remember() {
	if e.phase == nil || e.err != nil {
		return
	}
	if e.memos == nil {
		e.memos = map[int]memo{}
	}
	e.memos[e.index] = e.takeMemo()
}

func (e *Editor) useMemo(m memo) {
	e.phase = m.phase
	e.start = m.start
	e.hasStart = m.hasStart
	e.enemies = cloneEnemies(m.enemies)
	e.buildings = append([]data.BuildingSpec(nil), m.buildings...)
	e.pickups = append([]data.PickupSpec(nil), m.pickups...)
	e.barrels = append([]data.BarrelSpec(nil), m.barrels...)
	e.civilians = append([]data.Enemy(nil), m.civilians...)
	e.hostages = append([]data.Enemy(nil), m.hostages...)
	e.vehicles = cloneVehicles(m.vehicles)
	e.dirty = m.dirty
}

// phaseForSave copies the loaded phase and replaces the editable lists.
// Tree regions drawn this session stay out of the file.
func (e *Editor) phaseForSave() *data.Phase {
	p := *e.phase
	p.PlayerStart = e.start
	p.Enemies = cloneEnemies(e.enemies)
	p.Buildings = append([]data.BuildingSpec(nil), e.buildings...)
	p.Pickups = append([]data.PickupSpec(nil), e.pickups...)
	p.Barrels = append([]data.BarrelSpec(nil), e.barrels...)
	p.Civilians = append([]data.Enemy(nil), e.civilians...)
	p.Hostages = append([]data.Enemy(nil), e.hostages...)
	p.Vehicles = cloneVehicles(e.vehicles)
	return &p
}

func (e *Editor) writePhase(path string) error {
	if e.phase == nil || e.err != nil {
		return fmt.Errorf("nothing to save")
	}
	return data.WritePhase(path, e.phaseForSave())
}

func (e *Editor) save() {
	if e.phase == nil || e.err != nil || len(e.phases) == 0 {
		e.status = "nothing to save"
		return
	}
	if !e.dirty {
		e.status = "no changes"
		return
	}
	p := e.phaseForSave()
	if err := data.SavePhase(e.phases[e.index].file, p); err != nil {
		e.status = err.Error()
		return
	}
	e.phase = p
	e.dirty = false
	e.status = "saved " + e.phases[e.index].file
}

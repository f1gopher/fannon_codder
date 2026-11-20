package mapedit

import (
	"fmt"
	"path/filepath"
	"strings"

	"fannon-codder/internal/data"
)

func (e *Editor) phaseMission() int {
	if e.index < 0 || e.index >= len(e.phases) {
		return -1
	}
	return e.phases[e.index].mission
}

func (e *Editor) missionName() string {
	mi := e.nameMission
	if mi < 0 || mi >= len(e.missions) {
		return ""
	}
	return e.missions[mi].Name
}

func (e *Editor) renameMission(name string) {
	mi := e.nameMission
	if mi < 0 || mi >= len(e.missions) || name == "" || name == e.missions[mi].Name {
		return
	}
	e.missions[mi].Name = name
	e.campaignDirty = true
	e.rebuildRows()
	e.status = "mission name"
}

// movePhase swaps the selected part with its neighbour in the same mission.
func (e *Editor) movePhase(dir int) {
	mi := e.phaseMission()
	if mi < 0 || dir == 0 || e.index < 0 || e.index >= len(e.phases) {
		return
	}
	selected := e.phases[e.index].file
	parts := e.missions[mi].Phases
	at := -1
	for i, name := range parts {
		if name == selected {
			at = i
			break
		}
	}
	j := at + dir
	if at < 0 || j < 0 || j >= len(parts) {
		return
	}
	parts[at], parts[j] = parts[j], parts[at]
	e.missions[mi].Phases = parts
	ia, ib := -1, -1
	for i, row := range e.phases {
		if row.file == parts[at] {
			ia = i
		}
		if row.file == parts[j] {
			ib = i
		}
	}
	if ia >= 0 && ib >= 0 {
		e.phases[ia], e.phases[ib] = e.phases[ib], e.phases[ia]
	}
	for i, row := range e.phases {
		if row.file == selected {
			e.index = i
			break
		}
	}
	e.renumber(mi)
	e.campaignDirty = true
	e.rebuildRows()
	e.status = "phase order"
}

// addPhase inserts an empty part after the selected one. The file is written on S.
func (e *Editor) addPhase() {
	mi := e.phaseMission()
	if mi < 0 || e.phase == nil {
		return
	}
	file := e.nextPhaseFile(e.missions[mi].Number)
	blank := data.BlankPhase(e.missions[mi].Number, 1, e.phase.Terrain)
	at := 0
	for i, name := range e.missions[mi].Phases {
		if name == e.phases[e.index].file {
			at = i + 1
			break
		}
	}
	e.missions[mi].Phases = insertFile(e.missions[mi].Phases, at, file)
	row := phaseRow{
		file:    file,
		name:    strings.TrimSuffix(file, ".json"),
		title:   blank.Title,
		mission: mi,
	}
	insertAt := e.index + 1
	e.phases = append(e.phases, phaseRow{})
	copy(e.phases[insertAt+1:], e.phases[insertAt:])
	e.phases[insertAt] = row
	if e.memos == nil {
		e.memos = map[string]memo{}
	}
	if e.fresh == nil {
		e.fresh = map[string]bool{}
	}
	e.fresh[file] = true
	e.memos[file] = memo{phase: blank, start: blank.PlayerStart, hasStart: true}
	e.renumber(mi)
	e.campaignDirty = true
	e.rebuildRows()
	e.load(insertAt)
	e.dirty = true
	e.status = "new phase " + file
}

// removePhase drops the selected part from the mission. The file stays on disk.
func (e *Editor) removePhase() {
	mi := e.phaseMission()
	if mi < 0 || len(e.missions[mi].Phases) < 2 {
		e.status = "a mission keeps one phase"
		return
	}
	file := e.phases[e.index].file
	var kept []string
	for _, name := range e.missions[mi].Phases {
		if name != file {
			kept = append(kept, name)
		}
	}
	e.missions[mi].Phases = kept
	e.phases = append(e.phases[:e.index], e.phases[e.index+1:]...)
	delete(e.memos, file)
	delete(e.added, file)
	delete(e.fresh, file)
	if e.index >= len(e.phases) {
		e.index = len(e.phases) - 1
	}
	e.renumber(mi)
	e.campaignDirty = true
	e.rebuildRows()
	e.load(e.index)
	e.status = "removed " + file
}

func (e *Editor) renumber(mi int) {
	if mi < 0 || mi >= len(e.missions) {
		return
	}
	for i, file := range e.missions[mi].Phases {
		if m, ok := e.memos[file]; ok && m.phase != nil {
			m.phase.Mission = e.missions[mi].Number
			m.phase.Phase = i + 1
			e.memos[file] = m
		}
		if e.phase != nil && e.index >= 0 && e.index < len(e.phases) && e.phases[e.index].file == file {
			e.phase.Mission = e.missions[mi].Number
			e.phase.Phase = i + 1
		}
	}
}

func (e *Editor) nextPhaseFile(number int) string {
	used := map[string]bool{}
	for _, row := range e.phases {
		used[row.file] = true
	}
	for n := 1; n < 100; n++ {
		name := fmt.Sprintf("m%02dp%02d.json", number, n)
		if !used[name] {
			return name
		}
	}
	return fmt.Sprintf("m%02dp99.json", number)
}

func insertFile(files []string, at int, file string) []string {
	if at < 0 || at > len(files) {
		at = len(files)
	}
	files = append(files, "")
	copy(files[at+1:], files[at:])
	files[at] = file
	return files
}

func (e *Editor) writeCampaign() error {
	for mi := range e.missions {
		for pi, file := range e.missions[mi].Phases {
			p, err := e.phaseByFile(file)
			if err != nil {
				return err
			}
			if p.Mission == e.missions[mi].Number && p.Phase == pi+1 && !e.fresh[file] {
				continue
			}
			p.Mission = e.missions[mi].Number
			p.Phase = pi + 1
			if err := data.SavePhase(file, p); err != nil {
				return err
			}
			delete(e.fresh, file)
		}
	}
	return data.SaveCampaign(&data.Campaign{Missions: e.missions})
}

func (e *Editor) phaseByFile(file string) (*data.Phase, error) {
	if e.phase != nil && e.index >= 0 && e.index < len(e.phases) && e.phases[e.index].file == file {
		return e.phaseForSave(), nil
	}
	if m, ok := e.memos[file]; ok && m.phase != nil {
		return m.phase, nil
	}
	if e.fresh[file] {
		return nil, fmt.Errorf("new phase %s is not loaded", file)
	}
	p, err := data.LoadPhase(filepath.Base(file))
	if err != nil {
		return nil, err
	}
	return p, nil
}

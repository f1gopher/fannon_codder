package app

import (
	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/campaign"
	"fannon-codder/internal/input"
	"fannon-codder/internal/render"
	"fannon-codder/internal/sim"
)

// Pause-menu rows, in the order they are drawn.
const (
	menuRestart = iota
	menuTitle
	menuSave
	menuLoad
	menuQuit
	menuCount
)

// battleMenu is the Escape menu on a live battle. It freezes the phase
// until the player continues or picks a row.
type battleMenu struct {
	open      bool
	sel       int
	heldPause bool
	note      string
}

func (m battleMenu) lines() []string {
	labels := [menuCount]string{
		"Restart level",
		"Main menu",
		"Save game",
		"Load game",
		"Quit",
	}
	lines := []string{"PAUSED"}
	for i, lab := range labels {
		if i == m.sel {
			lab = "> " + lab
		} else {
			lab = "  " + lab
		}
		lines = append(lines, lab)
	}
	if m.note != "" {
		lines = append(lines, m.note)
	}
	lines = append(lines, "Up Down Enter    Escape continues")
	return lines
}

// rowAt is the pause-menu row under a world-frame point. The headline,
// the save note, and the key hint are not rows.
func (m battleMenu) rowAt(frameW, frameH, x, y float64) (int, bool) {
	i, ok := render.NoticeLineAt(frameW, frameH, x, y, m.lines()...)
	if !ok || i < 1 || i > menuCount {
		return 0, false
	}
	return i - 1, true
}

func (m *battleMenu) move(delta int) {
	if !m.open {
		return
	}
	m.sel = (m.sel + delta) % menuCount
	if m.sel < 0 {
		m.sel += menuCount
	}
}

// saveSnapshot is the campaign file written from a battle. Deployed men go
// back on the queue in the file, so a later load does not lose the squad
// that was out on the field. The live pool is left as it is.
func (p *Progress) saveSnapshot(deployed []campaign.Soldier) campaign.SaveGame {
	s := p.ToSave()
	if len(deployed) == 0 {
		return s
	}
	s.Recruits = append(append([]campaign.Soldier{}, deployed...), s.Recruits...)
	return s
}

func (p *Progress) SaveSnapshot(deployed []campaign.Soldier) error {
	if p == nil {
		return nil
	}
	path, err := campaign.DefaultSavePath()
	if err != nil {
		return err
	}
	return campaign.Save(path, p.saveSnapshot(deployed))
}

// ownsEscape is true when this battle, not the process quit prompt, should
// see Escape. That is a live phase, or a pause menu that is already open.
func (b *Battle) ownsEscape(escape bool) bool {
	if b == nil || b.world == nil || b.world.Status != sim.Playing {
		return false
	}
	return b.menu.open || escape
}

func (b *Battle) openMenu() {
	b.menu.heldPause = b.paused
	b.menu.open = true
	b.menu.sel = 0
	b.menu.note = ""
	b.paused = true
	if b.world != nil {
		b.world.SetFire(b.world.AimX, b.world.AimY, false)
		b.world.SetDrive(0, 0, false)
	}
}

func (b *Battle) closeMenu() {
	b.paused = b.menu.heldPause
	b.menu.open = false
	b.menu.note = ""
}

// stepMenu owns Escape while the phase is being played, and owns every key
// while the menu is up. It reports whether the rest of the battle should wait.
func (b *Battle) stepMenu(h Host, escape, up, down, confirm bool, p input.Pointer) (bool, error) {
	if b.world == nil || b.world.Status != sim.Playing {
		b.menu.open = false
		return false, nil
	}
	if !b.menu.open {
		if escape {
			b.openMenu()
		}
		return b.menu.open, nil
	}
	if escape {
		b.closeMenu()
		return true, nil
	}
	if up {
		b.menu.move(-1)
	}
	if down {
		b.menu.move(1)
	}
	if row, ok := b.menu.rowAt(float64(ScreenWidth), float64(ScreenHeight), p.X, p.Y); ok {
		b.menu.sel = row
		if p.LeftDown {
			confirm = true
		}
	}
	if !confirm {
		return true, nil
	}
	switch b.menu.sel {
	case menuRestart:
		b.restoreSquad()
		h.Switch(b.restartScene())
	case menuTitle:
		b.restoreSquad()
		h.Switch(NewTitle())
	case menuSave:
		if b.prog == nil {
			b.menu.note = "No campaign"
			return true, nil
		}
		if err := b.prog.SaveSnapshot(b.deployed); err != nil {
			b.menu.note = "Save failed"
			return true, nil
		}
		b.menu.note = "Saved"
	case menuLoad:
		if b.prog == nil {
			b.menu.note = "No save"
			return true, nil
		}
		if err := b.prog.Load(); err != nil {
			b.menu.note = "No save"
			return true, nil
		}
		b.deployed = nil
		b.prog.SaveNotice = "Loaded"
		h.Switch(NewBootHill(b.prog))
	case menuQuit:
		return true, ebiten.Termination
	}
	return true, nil
}

// restoreSquad puts the men who started this attempt back on the queue.
// Deaths and pickups from the attempt are dropped. A sandbox has no queue.
func (b *Battle) restoreSquad() {
	if b.sandbox || b.prog == nil || b.prog.Pool == nil {
		b.deployed = nil
		return
	}
	b.prog.Pool.Return(b.deployed)
	b.deployed = nil
}

func (b *Battle) restartScene() Scene {
	if b.again != nil {
		return b.again(b.prog)
	}
	return NewBattle(b.prog)
}

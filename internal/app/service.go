package app

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fannon-codder/internal/campaign"
	"fannon-codder/internal/render"
	"fannon-codder/internal/sim"
)

// Service is the debrief after a cleared phase: the living, then the dead.
type Service struct {
	prog  *Progress
	pages [][]string
	i     int
}

func NewService(prog *Progress, alive, dead []campaign.Soldier) *Service {
	var pages [][]string
	if lines := rosterLines("ACHIEVEMENTS", alive); lines != nil {
		pages = append(pages, lines)
	}
	if lines := rosterLines("REMEMBRANCE", dead); lines != nil {
		pages = append(pages, lines)
	}
	return &Service{prog: prog, pages: pages}
}

func rosterLines(title string, men []campaign.Soldier) []string {
	if len(men) == 0 {
		return nil
	}
	lines := []string{title}
	for _, m := range men {
		lines = append(lines, fmt.Sprintf("%-12s %-4s %d", m.Name, m.Rank.Abbrev(), m.Kills))
	}
	lines = append(lines, "Click or Enter")
	return lines
}

func (s *Service) Enter() {}
func (s *Service) Leave() {}

func (s *Service) Update(h Host) error {
	if s.i >= len(s.pages) {
		h.Switch(NewBootHill(s.prog))
		return nil
	}
	p := h.Pointer()
	if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		h.PlayKind(sim.CueClick)
		s.i++
		if s.i >= len(s.pages) {
			h.Switch(NewBootHill(s.prog))
		}
	}
	return nil
}

func (s *Service) Draw(screen *ebiten.Image) {
	screen.Fill(titleColor)
	if s.i < 0 || s.i >= len(s.pages) {
		return
	}
	render.Notice(screen, s.pages[s.i]...)
}

package app

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var briefBG = color.RGBA{R: 0x10, G: 0x18, B: 0x10, A: 0xff}

// Briefing shows the phase title and objective, then deploys.
type Briefing struct {
	prog *Progress
}

func NewBriefing(prog *Progress) *Briefing { return &Briefing{prog: prog} }

func (b *Briefing) Enter() {}
func (b *Briefing) Leave() {}

func (b *Briefing) Update(h Host) error {
	p := h.Pointer()
	if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if b.prog.CanStart() {
			h.Switch(NewBattle(b.prog))
		} else {
			h.Switch(NewBootHill(b.prog))
		}
	}
	return nil
}

func (b *Briefing) Draw(screen *ebiten.Image) {
	screen.Fill(briefBG)
	title := "Mission"
	body := "Kill all enemy"
	if b.prog.Phase != nil {
		title = b.prog.Phase.Title
		body = b.prog.Phase.Briefing
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"BRIEFING\nMission %d  Phase %d\n\n%s\n%s\n\nClick or Enter to deploy",
		b.prog.MissionNumber(),
		phaseNum(b.prog),
		title,
		body,
	))
}

func phaseNum(p *Progress) int {
	if p.Phase != nil && p.Phase.Phase > 0 {
		return p.Phase.Phase
	}
	return 1
}

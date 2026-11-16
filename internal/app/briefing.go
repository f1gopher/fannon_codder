package app

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/render"
	"fannon-codder/internal/sim"
	"fannon-codder/internal/terrain"
)

var briefBG = color.RGBA{R: 0x10, G: 0x18, B: 0x10, A: 0xff}

// Briefing shows the phase title and objective, then deploys.
type Briefing struct {
	prog  *Progress
	chart terrain.Chart
	warm  bool
}

func NewBriefing(prog *Progress) *Briefing { return &Briefing{prog: prog} }

func (b *Briefing) Enter() {
	if b == nil {
		return
	}
	b.chart, b.warm = warmPhase(b.prog)
}

func (b *Briefing) Leave() {}

func (b *Briefing) Update(h Host) error {
	if b.warm {
		terrain.PumpWarm(b.chart, 6*time.Millisecond)
	}
	p := h.Pointer()
	if p.LeftDown || h.Pressed(ebiten.KeyEnter) {
		if b.prog.CanStart() {
			h.PlayKind(sim.CueClick)
			h.Switch(NewBattle(b.prog))
		} else {
			h.Switch(NewBootHill(b.prog))
		}
	}
	return nil
}

func (b *Briefing) Draw(screen *ebiten.Image) {
	if !render.DrawMenu(screen, "menu/briefing") {
		screen.Fill(briefBG)
	}
	title := "Mission"
	body := "Kill all enemy"
	if b.prog.Phase != nil {
		title = b.prog.Phase.Title
		body = b.prog.Phase.Briefing
	}
	// The blank green panel sits in the middle of the wide map.
	fw := float64(screen.Bounds().Dx())
	if s := render.PictureScale(); s > 0 {
		fw /= s
	}
	const boxW, boxH = 220, 108
	const pad = 8
	x := int((fw - boxW) / 2)
	if x < pad {
		x = pad
	}
	render.Rect(screen, x, 100, boxW, boxH, color.RGBA{R: 0x10, G: 0x18, B: 0x10, A: 0xaa})
	inner := float64(boxW - 2*pad)
	title = render.WrapText(title, inner)
	body = render.WrapText(body, inner)
	render.Text(screen, fmt.Sprintf(
		"BRIEFING\nMission %d  Phase %d\n\n%s\n%s\n\nClick or Enter to deploy",
		b.prog.MissionNumber(),
		phaseNum(b.prog),
		title,
		body,
	), float64(x+pad), 104)
}

func phaseNum(p *Progress) int {
	if p.Phase != nil && p.Phase.Phase > 0 {
		return p.Phase.Phase
	}
	return 1
}

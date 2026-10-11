package app

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fannon-codder/internal/render"
	"fannon-codder/internal/sim"
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
	x := int((fw - boxW) / 2)
	if x < 8 {
		x = 8
	}
	render.Rect(screen, x, 100, boxW, boxH, color.RGBA{R: 0x10, G: 0x18, B: 0x10, A: 0xaa})
	render.Text(screen, fmt.Sprintf(
		"BRIEFING\nMission %d  Phase %d\n\n%s\n%s\n\nClick or Enter to deploy",
		b.prog.MissionNumber(),
		phaseNum(b.prog),
		title,
		body,
	), float64(x+8), 104)
}

func phaseNum(p *Progress) int {
	if p.Phase != nil && p.Phase.Phase > 0 {
		return p.Phase.Phase
	}
	return 1
}

package app

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fannon-codder/internal/campaign"
	"fannon-codder/internal/render"
)

var (
	hillSky   = color.RGBA{R: 0x6a, G: 0xb0, B: 0xe0, A: 0xff}
	hillFill  = color.RGBA{R: 0x3a, G: 0x8a, B: 0x3a, A: 0xff}
	graveFill = color.RGBA{R: 0x50, G: 0x50, B: 0x50, A: 0xff}
	manFill   = color.RGBA{R: 0x40, G: 0xc8, B: 0x40, A: 0xff}
	pathFill  = color.RGBA{R: 0x8a, G: 0x70, B: 0x40, A: 0xff}
	iconFill  = color.RGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xff}
)

const (
	loadX, loadY, loadW, loadH = 8, 8, 40, 14
	saveX, saveY, saveW, saveH = 272, 8, 40, 14
)

// BootHill is the between-mission hill: queue, graves, mission number.
type BootHill struct {
	prog *Progress
}

func NewBootHill(prog *Progress) *BootHill { return &BootHill{prog: prog} }

func (b *BootHill) Enter() {}
func (b *BootHill) Leave() {}

func (b *BootHill) Update(h Host) error {
	p := h.Pointer()
	click := p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	if !click {
		return nil
	}
	if p.LeftDown && inRect(p.X, p.Y, loadX, loadY, loadW, loadH) {
		if err := b.prog.Load(); err != nil {
			b.prog.SaveNotice = "No save"
		} else {
			b.prog.SaveNotice = "Loaded"
		}
		return nil
	}
	if p.LeftDown && inRect(p.X, p.Y, saveX, saveY, saveW, saveH) {
		if !b.prog.CanSave() {
			b.prog.SaveNotice = "Save after a mission"
			return nil
		}
		if err := b.prog.Save(); err != nil {
			b.prog.SaveNotice = err.Error()
		} else {
			b.prog.SaveNotice = "Saved"
		}
		return nil
	}
	if b.prog.AwaitingStub {
		h.Switch(NewMissionStub(b.prog))
		return nil
	}
	if b.prog.CanStart() {
		h.Switch(NewBriefing(b.prog))
	}
	return nil
}

func inRect(px, py, x, y, w, h float64) bool {
	return px >= x && py >= y && px < x+w && py < y+h
}

func (b *BootHill) Draw(screen *ebiten.Image) {
	screen.Fill(hillSky)
	render.Rect(screen, 0, 140, ScreenWidth, ScreenHeight-140, hillFill)
	render.Rect(screen, 40, 168, 240, 8, pathFill)
	n := b.prog.Pool.Remaining()
	if n > 24 {
		n = 24
	}
	for i := 0; i < n; i++ {
		render.Rect(screen, 48+i*9, 160, 6, 8, manFill)
	}
	for i := 0; i < b.prog.Graves && i < 40; i++ {
		render.Rect(screen, 20+(i%10)*12, 190+(i/10)*14, 4, 8, graveFill)
	}

	render.Rect(screen, loadX, loadY, loadW, loadH, iconFill)
	render.Rect(screen, saveX, saveY, saveW, saveH, iconFill)
	render.Text(screen, "LOAD", loadX+4, loadY+2)
	render.Text(screen, "SAVE", saveX+4, saveY+2)

	render.Text(screen, fmt.Sprintf(
		"BOOT HILL  Mission %d  Graves %d  Queue %d",
		b.prog.MissionNumber(), b.prog.Graves, b.prog.Pool.Remaining(),
	), 8, 24)

	y := 36
	shown := 0
	if b.prog.Pool != nil {
		for _, s := range b.prog.Pool.Recruits {
			if s.Rank <= campaign.Private {
				continue
			}
			render.Text(screen, s.Rank.Abbrev()+" "+s.Name, 8, float64(y))
			y += 8
			shown++
			if shown >= 4 {
				break
			}
		}
	}

	switch {
	case b.prog.GameOver:
		render.Text(screen, "GAME OVER — no recruits left", 8, 88)
	case b.prog.AwaitingStub && b.prog.MissionsCompleted >= 5:
		render.Text(screen, "THE CAMPAIGN CONTINUES ANOTHER DAY", 8, 88)
	case b.prog.AwaitingStub:
		render.Text(screen, "MISSION COMPLETE  Click — later missions", 8, 88)
	default:
		render.Text(screen, "Click or Enter to start", 8, 88)
	}
	if b.prog.SaveNotice != "" {
		render.Text(screen, b.prog.SaveNotice, 8, 108)
	}
}

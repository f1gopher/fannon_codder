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
	prog   *Progress
	t      float64
	heroes bool
}

func NewBootHill(prog *Progress) *BootHill { return &BootHill{prog: prog} }

func (b *BootHill) Enter() {}
func (b *BootHill) Leave() {}

func (b *BootHill) Update(h Host) error {
	b.t += 1.0 / 60
	p := h.Pointer()
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		b.heroes = !b.heroes
		return nil
	}
	click := p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	if b.heroes {
		if click {
			b.heroes = false
		}
		return nil
	}
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
	painted := render.DrawSheet(screen, "menu/hill", "E", 0, 0, 0)
	if !painted {
		screen.Fill(hillSky)
		render.Rect(screen, 0, 140, ScreenWidth, ScreenHeight-140, hillFill)
		render.Rect(screen, 40, 168, 240, 8, pathFill)
	}
	n := b.prog.Pool.Remaining()
	if n > 24 {
		n = 24
	}
	frame := render.FrameAt(b.t, 8, 4, true)
	for i := 0; i < n; i++ {
		x := 48 + float64(i)*9
		if !render.DrawSheet(screen, "snake/idle", "SE", frame, x+3, 176) {
			render.Rect(screen, int(x), 160, 6, 8, manFill)
		}
	}
	for i := 0; i < b.prog.Graves && i < 40; i++ {
		x := 20 + (i%10)*12
		y := 190 + (i/10)*14
		if !render.DrawSheet(screen, "menu/grave", "E", 0, float64(x+2), float64(y+8)) {
			render.Rect(screen, x, y, 4, 8, graveFill)
		}
	}

	if !render.DrawSheet(screen, "menu/load", "E", 0, loadX+loadW/2, loadY+loadH/2) {
		render.Rect(screen, loadX, loadY, loadW, loadH, iconFill)
		render.Text(screen, "LOAD", loadX+4, loadY+2)
	}
	if !render.DrawSheet(screen, "menu/save", "E", 0, saveX+saveW/2, saveY+saveH/2) {
		render.Rect(screen, saveX, saveY, saveW, saveH, iconFill)
		render.Text(screen, "SAVE", saveX+4, saveY+2)
	}
	render.Rect(screen, 4, 22, 312, 100, color.RGBA{R: 0x14, G: 0x28, B: 0x18, A: 0xaa})

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
		render.Text(screen, "Click or Enter to start    H heroes", 8, 88)
	}
	if b.heroes {
		b.drawHeroes(screen)
	}
	if b.prog.SaveNotice != "" {
		render.Text(screen, b.prog.SaveNotice, 8, 108)
	}
}

func (b *BootHill) drawHeroes(screen *ebiten.Image) {
	render.Rect(screen, 24, 36, 272, 176, color.RGBA{R: 0x10, G: 0x18, B: 0x10, A: 0xee})
	render.Text(screen, "HIGH SCORING HEROES", 36, 44)
	var living []campaign.Soldier
	if b.prog.Pool != nil {
		living = b.prog.Pool.Recruits
	}
	rows := campaign.HighScorers(living, b.prog.Heroes)
	if len(rows) == 0 {
		render.Text(screen, "No scores yet", 36, 64)
	}
	for i, h := range rows {
		render.Text(screen, fmt.Sprintf("%2d  %-4s %-12s %d", i+1, h.Rank.Abbrev(), h.Name, h.Kills), 36, float64(64+i*8))
	}
	render.Text(screen, "Click or Enter — Boot Hill", 36, 192)
}

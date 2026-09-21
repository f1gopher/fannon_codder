package app

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fannon-codder/internal/campaign"
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
	for y := 140; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			screen.Set(x, y, hillFill)
		}
	}
	for x := 40; x < 280; x++ {
		for y := 168; y < 176; y++ {
			screen.Set(x, y, pathFill)
		}
	}
	n := b.prog.Pool.Remaining()
	if n > 24 {
		n = 24
	}
	for i := 0; i < n; i++ {
		fillRect(screen, 48+i*9, 160, 6, 8, manFill)
	}
	for i := 0; i < b.prog.Graves && i < 40; i++ {
		fillRect(screen, 20+(i%10)*12, 190+(i/10)*14, 4, 8, graveFill)
	}

	fillRect(screen, loadX, loadY, loadW, loadH, iconFill)
	fillRect(screen, saveX, saveY, saveW, saveH, iconFill)
	ebitenutil.DebugPrintAt(screen, "LOAD", loadX+4, loadY+2)
	ebitenutil.DebugPrintAt(screen, "SAVE", saveX+4, saveY+2)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(
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
			ebitenutil.DebugPrintAt(screen, s.Rank.Abbrev()+" "+s.Name, 8, y)
			y += 10
			shown++
			if shown >= 4 {
				break
			}
		}
	}

	switch {
	case b.prog.GameOver:
		ebitenutil.DebugPrintAt(screen, "GAME OVER — no recruits left", 8, 88)
	case b.prog.AwaitingStub:
		ebitenutil.DebugPrintAt(screen, "MISSION COMPLETE  Click — later missions", 8, 88)
	default:
		ebitenutil.DebugPrintAt(screen, "Click or Enter to start", 8, 88)
	}
	if b.prog.SaveNotice != "" {
		ebitenutil.DebugPrintAt(screen, b.prog.SaveNotice, 8, 100)
	}
}

func fillRect(dst *ebiten.Image, x, y, w, h int, c color.Color) {
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			dst.Set(x+i, y+j, c)
		}
	}
}

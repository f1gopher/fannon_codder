package app

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	hillSky   = color.RGBA{R: 0x6a, G: 0xb0, B: 0xe0, A: 0xff}
	hillFill  = color.RGBA{R: 0x3a, G: 0x8a, B: 0x3a, A: 0xff}
	graveFill = color.RGBA{R: 0x50, G: 0x50, B: 0x50, A: 0xff}
	manFill   = color.RGBA{R: 0x40, G: 0xc8, B: 0x40, A: 0xff}
	pathFill  = color.RGBA{R: 0x8a, G: 0x70, B: 0x40, A: 0xff}
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
	if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if b.prog.CanStart() {
			h.Switch(NewBriefing(b.prog))
		}
	}
	return nil
}

func (b *BootHill) Draw(screen *ebiten.Image) {
	screen.Fill(hillSky)
	// Hill
	for y := 140; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			screen.Set(x, y, hillFill)
		}
	}
	// Path
	for x := 40; x < 280; x++ {
		for y := 168; y < 176; y++ {
			screen.Set(x, y, pathFill)
		}
	}
	// Recruit queue along the path
	n := b.prog.Pool.Remaining()
	if n > 24 {
		n = 24
	}
	for i := 0; i < n; i++ {
		x := 48 + i*9
		y := 160
		fillRect(screen, x, y, 6, 8, manFill)
	}
	// Graves
	for i := 0; i < b.prog.Graves && i < 40; i++ {
		x := 20 + (i%10)*12
		y := 190 + (i/10)*14
		fillRect(screen, x, y, 4, 8, graveFill)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"BOOT HILL\nMission %d\nGraves %d  Queue %d",
		b.prog.MissionNumber(), b.prog.Graves, b.prog.Pool.Remaining(),
	))
	switch {
	case b.prog.GameOver:
		ebitenutil.DebugPrintAt(screen, "GAME OVER — no recruits left", 8, 48)
	case b.prog.MissionDone:
		ebitenutil.DebugPrintAt(screen, "MISSION COMPLETE\n(promotions next chunk)", 8, 48)
	default:
		ebitenutil.DebugPrintAt(screen, "Click or Enter to start", 8, 48)
	}
}

func fillRect(dst *ebiten.Image, x, y, w, h int, c color.Color) {
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			dst.Set(x+i, y+j, c)
		}
	}
}

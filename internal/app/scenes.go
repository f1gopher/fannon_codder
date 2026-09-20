package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fannon-codder/internal/input"
	"fannon-codder/internal/sim"
)

// Scene is one full-screen mode (title, battle, later Boot Hill).
type Scene interface {
	Update(h Host) error
	Draw(screen *ebiten.Image)
	Enter()
	Leave()
}

// Host is what a scene may call on the game.
type Host interface {
	Pointer() input.Pointer
	Switch(Scene)
}

var (
	titleColor  = color.RGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xff}
	battleColor = color.RGBA{R: 0x3a, G: 0x8f, B: 0x3a, A: 0xff}
)

// Title is the click-through splash.
type Title struct{}

func NewTitle() *Title { return &Title{} }

func (t *Title) Enter() {}
func (t *Title) Leave() {}

func (t *Title) Update(h Host) error {
	p := h.Pointer()
	if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		h.Switch(NewBattle())
	}
	return nil
}

func (t *Title) Draw(screen *ebiten.Image) {
	screen.Fill(titleColor)
	ebitenutil.DebugPrint(screen, "FANNON CODDER\n\nClick or press Enter")
}

// Battle is an empty green field with a camera (no units yet).
type Battle struct {
	cam sim.Camera
}

func NewBattle() *Battle {
	return &Battle{
		cam: sim.Camera{
			ViewW: ScreenWidth,
			ViewH: ScreenHeight,
			MapW:  ScreenWidth,
			MapH:  ScreenHeight,
		},
	}
}

func (b *Battle) Enter() {}
func (b *Battle) Leave() {}

func (b *Battle) Update(h Host) error {
	p := h.Pointer()
	b.cam.ScrollToward(p.X, p.Y, ScreenWidth, ScreenHeight, 1.0/TPS)
	return nil
}

func (b *Battle) Draw(screen *ebiten.Image) {
	screen.Fill(battleColor)
}

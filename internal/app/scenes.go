package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fannon-codder/internal/data"
	"fannon-codder/internal/input"
	"fannon-codder/internal/render"
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

// Battle is a green field with a two-man squad (Chunk 03).
type Battle struct {
	world *sim.World
}

func NewBattle() *Battle {
	w, err := data.LoadFirstWorld()
	if err != nil {
		// Tests and missing data still get the empty-field sandbox.
		w = sim.NewDemoWorld()
	}
	return &Battle{world: w}
}

func (b *Battle) Enter() {}
func (b *Battle) Leave() {}

func (b *Battle) Update(h Host) error {
	p := h.Pointer()
	if b.world.Status != sim.Playing {
		if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			h.Switch(NewTitle())
		}
		return nil
	}
	wx := p.X + b.world.Camera.X
	wy := p.Y + b.world.Camera.Y
	if p.LeftDown {
		b.world.CommandMove(wx, wy, true)
	} else if p.Left {
		if s := b.world.ActiveSquad(); s != nil && s.HasDest {
			b.world.CommandMove(wx, wy, false)
		}
	}
	firing := p.Right ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
	b.world.SetFire(wx, wy, firing)
	b.world.Camera.ScrollToward(p.X, p.Y, ScreenWidth, ScreenHeight, 1.0/TPS)
	b.world.Step(1.0 / TPS)
	return nil
}

func (b *Battle) Draw(screen *ebiten.Image) {
	screen.Fill(battleColor)
	render.Tiles(screen, b.world.Map, b.world.Camera)
	render.Units(screen, b.world)
	render.Projectiles(screen, b.world)
	switch b.world.Status {
	case sim.Won:
		ebitenutil.DebugPrint(screen, "\n\n  PHASE COMPLETE\n  Click or Enter")
	case sim.Lost:
		ebitenutil.DebugPrint(screen, "\n\n  PHASE FAILED\n  Click or Enter")
	}
}

package app

import (
	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/audio"
	"fannon-codder/internal/input"
	"fannon-codder/internal/render"
	"fannon-codder/internal/sim"
)

const (
	ScreenWidth  = 320
	ScreenHeight = 256
	TPS          = 60

	DefaultWindowWidth  = ScreenWidth * 3
	DefaultWindowHeight = ScreenHeight * 3
)

// Game is the Ebitengine entry point: scale, input, and the current scene.
type Game struct {
	scene   Scene
	next    Scene
	tracker input.Tracker
	pointer input.Pointer
	prog    *Progress
	sound   *audio.Mixer
}

// newShell is the window shell shared by every entry point, including the sandboxes.
func newShell() *Game {
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	return &Game{
		prog:  NewProgress(),
		sound: audio.NewMixer(),
	}
}

func New(skipTitle bool) *Game {
	g := newShell()
	if skipTitle {
		g.scene = NewBootHill(g.prog)
	} else {
		g.scene = NewTitle()
	}
	g.scene.Enter()
	return g
}

// NewCover starts the oversized tree-cover sandbox (chunk 11).
func NewCover() *Game {
	g := newShell()
	g.scene = NewCoverBattle(g.prog)
	g.scene.Enter()
	return g
}

// NewRiver starts the river/bridge sandbox (chunk 12).
func NewRiver() *Game {
	g := newShell()
	g.scene = NewRiverBattle(g.prog)
	g.scene.Enter()
	return g
}

// NewHut starts the spawner-hut and grenade-crate sandbox (chunk 14).
func NewHut() *Game {
	g := newShell()
	g.scene = NewHutBattle(g.prog)
	g.scene.Enter()
	return g
}

// NewSkidoo starts the skidoo and bazooka sandbox (chunk 20).
func NewSkidoo() *Game {
	g := newShell()
	g.scene = NewSkidooBattle(g.prog)
	g.scene.Enter()
	return g
}

// NewHazard starts the civilian, quicksand, and mine sandbox (chunk 18).
func NewHazard() *Game {
	g := newShell()
	g.scene = NewHazardBattle(g.prog)
	g.scene.Enter()
	return g
}

func (g *Game) Pointer() input.Pointer { return g.pointer }

func (g *Game) Progress() *Progress { return g.prog }

func (g *Game) Switch(s Scene) { g.next = s }

// PlayCues plays the sounds the battle recorded this frame.
func (g *Game) PlayCues(cues []sim.Cue) {
	if g.sound == nil {
		return
	}
	g.sound.Play(cues)
}

// SetListener is the world point cue volume is measured from.
func (g *Game) SetListener(x, y float64) {
	if g.sound == nil {
		return
	}
	g.sound.SetListener(x, y)
}

// hearPoint is the active leader, including the vehicle he is driving.
// With no leader, volume is measured from the middle of the playfield.
func hearPoint(w *sim.World) (x, y float64) {
	if x, y, ok := w.CameraFocus(); ok {
		return x, y
	}
	c := w.Camera
	return c.X + c.ViewW/2, c.Y + c.ViewH/2
}

func (g *Game) Update() error {
	x, y := ebiten.CursorPosition()
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > ScreenWidth-1 {
		x = ScreenWidth - 1
	}
	if y > ScreenHeight-1 {
		y = ScreenHeight - 1
	}
	g.pointer = g.tracker.Update(
		float64(x),
		float64(y),
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight),
	)

	if g.next != nil {
		if g.scene != nil {
			g.scene.Leave()
		}
		g.scene = g.next
		g.next = nil
		g.scene.Enter()
	}
	if g.scene != nil {
		return g.scene.Update(g)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.scene != nil {
		g.scene.Draw(screen)
	}
	kind := render.PointerArrow
	if b, ok := g.scene.(*Battle); ok {
		kind = b.CursorKind(g.pointer)
	} else if g.pointer.Right ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight) {
		kind = render.PointerCrosshair
	}
	render.Pointer(screen, g.pointer.X, g.pointer.Y, kind)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// DrawFinalScreen integer-scales the 320×256 offscreen into the window with
// letterboxing. Nearest-neighbour; never a fractional scale.
func (g *Game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, _ ebiten.GeoM) {
	b := screen.Bounds()
	scale, ox, oy := integerScale(b.Dx(), b.Dy(), ScreenWidth, ScreenHeight)
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(ox), float64(oy))
	screen.DrawImage(offscreen, op)
}

func integerScale(outsideW, outsideH, logicalW, logicalH int) (scale, offsetX, offsetY int) {
	if logicalW <= 0 || logicalH <= 0 {
		return 1, 0, 0
	}
	sx := outsideW / logicalW
	sy := outsideH / logicalH
	scale = sx
	if sy < scale {
		scale = sy
	}
	if scale < 1 {
		scale = 1
	}
	offsetX = (outsideW - logicalW*scale) / 2
	offsetY = (outsideH - logicalH*scale) / 2
	return scale, offsetX, offsetY
}

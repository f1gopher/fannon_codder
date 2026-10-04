package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fannon-codder/internal/audio"
	"fannon-codder/internal/input"
	"fannon-codder/internal/render"
	"fannon-codder/internal/sim"
)

const (
	ScreenWidth  = 320
	ScreenHeight = 256
	TPS          = 60

	DefaultWindowWidth  = 1024
	DefaultWindowHeight = 768
)

// Game is the Ebitengine entry point: scale, input, and the current scene.
type Game struct {
	scene   Scene
	next    Scene
	tracker input.Tracker
	pointer input.Pointer
	prog    *Progress
	sound   *audio.Mixer
	quit    quitPrompt

	// scale is offscreen pixels per world pixel. offW/offH is the picture.
	// fbW/fbH is the final framebuffer. WindowSize stays at the size the game
	// requested, so a window-manager resize has to be read from the frame.
	scale      float64
	offW, offH int
	fbW, fbH   int
}

// newShell is the window shell shared by every entry point, including the sandboxes.
func newShell() *Game {
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	s, w, h := PictureSize(DefaultWindowWidth, DefaultWindowHeight, 1)
	return &Game{
		prog:  NewProgress(),
		sound: audio.NewMixer(),
		scale: s,
		offW:  w,
		offH:  h,
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

// SetEngine starts or stops the skidoo hum and sets its pitch.
func (g *Game) SetEngine(run bool, pitch float64) {
	if g.sound == nil {
		return
	}
	g.sound.SetEngine(run, pitch)
}

// PlayKind plays a sting that has no world position. Menus use it.
func (g *Game) PlayKind(kind sim.CueKind) {
	if g.sound == nil {
		return
	}
	g.sound.PlayKind(kind)
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
	x, y, inside := g.frameCursor()
	left := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	right := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	// The bars are outside the 320×256 frame. A click there is not a click.
	if !inside {
		left = false
		right = false
	}
	g.pointer = g.tracker.Update(x, y, left, right)

	if g.next != nil {
		// The hum belongs to the battle. A new scene starts it again if it
		// still has a skidoo to hear.
		g.SetEngine(false, 0)
		if g.scene != nil {
			g.scene.Leave()
		}
		g.scene = g.next
		g.next = nil
		g.scene.Enter()
	}
	// Escape asks to quit from every screen. The scene stays frozen until
	// the player confirms or backs out, so a battle press does not also
	// order a move.
	switch g.quit.decide(
		inpututil.IsKeyJustPressed(ebiten.KeyEscape),
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyY),
		inpututil.IsKeyJustPressed(ebiten.KeyN),
	) {
	case quitExit:
		return ebiten.Termination
	case quitHold, quitCancel:
		return nil
	}
	if g.scene != nil {
		return g.scene.Update(g)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	render.SetPictureScale(g.scale)
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
	if g.quit.open {
		render.Notice(screen, "QUIT THE GAME?", "Y or Enter    quit", "N or Escape    stay")
	}
	render.Pointer(screen, g.pointer.X, g.pointer.Y, kind)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.prepareLayout(float64(outsideWidth), float64(outsideHeight))
}

func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	w, h := g.prepareLayout(outsideWidth, outsideHeight)
	return float64(w), float64(h)
}

func (g *Game) prepareLayout(outsideWidth, outsideHeight float64) (int, int) {
	s, w, h := PictureSize(outsideWidth, outsideHeight, monitorScale())
	g.scale = s
	g.offW = w
	g.offH = h
	return w, h
}

// DrawFinalScreen clears the framebuffer and blits the offscreen at 1:1 in
// the centre. The scale is 1, so the filter is nearest. The margin is the bars.
func (g *Game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, _ ebiten.GeoM) {
	b := screen.Bounds()
	g.fbW = b.Dx()
	g.fbH = b.Dy()
	screen.Fill(color.Black)
	ob := offscreen.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(float64(b.Dx()-ob.Dx())/2, float64(b.Dy()-ob.Dy())/2)
	screen.DrawImage(offscreen, op)
}

// frameCursor returns the pointer in world-frame pixels. inside is false in the bars.
func (g *Game) frameCursor() (x, y float64, inside bool) {
	cx, cy := ebiten.CursorPosition()
	// CursorPosition is where Ebitengine's own letterbox would put the pointer.
	// The blit is 1:1 on the real framebuffer, which a resize can change while
	// WindowSize stays at 1024×768.
	ox, oy := OffscreenCursor(float64(cx), float64(cy), g.fbW, g.fbH, g.offW, g.offH)
	return FramePoint(ox, oy, g.scale, g.offW, g.offH)
}

func monitorScale() float64 {
	m := ebiten.Monitor()
	if m == nil {
		return 1
	}
	s := m.DeviceScaleFactor()
	if s <= 0 {
		return 1
	}
	return s
}

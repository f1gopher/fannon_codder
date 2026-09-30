package app

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fannon-codder/internal/audio"
	"fannon-codder/internal/campaign"
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
	Progress() *Progress
	PlayCues([]sim.Cue)
	SetListener(x, y float64)
	SetEngine(run bool, pitch float64)
	PlayKind(sim.CueKind)
}

var (
	titleColor  = color.RGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xff}
	battleColor = color.RGBA{R: 0x3a, G: 0x8f, B: 0x3a, A: 0xff}
	snowColor   = color.RGBA{R: 0xe6, G: 0xec, B: 0xf0, A: 0xff}
)

// Title is the click-through splash.
type Title struct {
	continues bool
}

func NewTitle() *Title { return &Title{} }

func (t *Title) Enter() {}
func (t *Title) Leave() {}

func (t *Title) Update(h Host) error {
	if prog := h.Progress(); prog != nil && prog.MissionsCompleted >= 5 {
		t.continues = true
	}
	p := h.Pointer()
	if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		h.PlayKind(sim.CueClick)
		h.Switch(NewBootHill(h.Progress()))
	}
	return nil
}

func (t *Title) Draw(screen *ebiten.Image) {
	screen.Fill(titleColor)
	msg := "FANNON CODDER\n\nClick or press Enter"
	if t.continues {
		msg += "\n\nThe campaign continues another day."
	}
	render.Text(screen, msg, 8, 8)
}

// Battle is one phase on the map.
type Battle struct {
	prog      *Progress
	world     *sim.World
	remaining int
	deployed  []campaign.Soldier
	unitIDs   []int
	settled   bool
	sandbox   bool
	mapOpen   bool
}

func NewBattle(prog *Progress) *Battle {
	var (
		w   *sim.World
		err error
	)
	if prog.Phase != nil {
		w, err = prog.Phase.World()
	} else {
		w, err = data.LoadFirstWorld()
	}
	if err != nil || w == nil {
		w = sim.NewDemoWorld()
	}
	return battleFromWorld(prog, w, false)
}

// NewCoverBattle is the chunk 11 sandbox: oversized map, tree cover, scrolling.
func NewCoverBattle(prog *Progress) *Battle {
	return battleFromWorld(prog, sim.NewCoverWorld(), true)
}

// NewRiverBattle is the chunk 12 sandbox: river, bridge, swimmers.
func NewRiverBattle(prog *Progress) *Battle {
	return battleFromWorld(prog, sim.NewRiverWorld(), true)
}

// NewHutBattle is the chunk 14 sandbox: a spawner hut and a grenade crate.
func NewHutBattle(prog *Progress) *Battle {
	return battleFromWorld(prog, sim.NewHutWorld(), true)
}

// NewSkidooBattle is the chunk 20 sandbox: skidoo, rocket crate, hut.
func NewSkidooBattle(prog *Progress) *Battle {
	return battleFromWorld(prog, sim.NewSkidooWorld(), true)
}

// NewHazardBattle is the chunk 18 sandbox: mine, quicksand, civilian, doorless hut.
func NewHazardBattle(prog *Progress) *Battle {
	return battleFromWorld(prog, sim.NewHazardWorld(), true)
}

// applyPlayfield sizes the camera to the area beside the status strip.
// A map as wide as the full screen can still pan by the strip's width,
// so the west edge is not stuck underneath it.
func applyPlayfield(w *sim.World) {
	w.Camera.ViewW = float64(ScreenWidth - render.HUDWidth)
	w.Camera.ViewH = float64(ScreenHeight)
	w.Camera.OriginX = float64(render.HUDWidth)
}

func battleFromWorld(prog *Progress, w *sim.World, sandbox bool) *Battle {
	applyPlayfield(w)
	if s := w.ActiveSquad(); s != nil {
		if l := w.Unit(s.LeaderID); l != nil {
			w.Camera.CenterOn(l.X, l.Y)
		}
	}
	ids := []int{}
	if s := w.ActiveSquad(); s != nil {
		ids = append(ids, s.MemberIDs...)
	}
	if sandbox {
		nameCoverSquad(w)
		return &Battle{
			prog:      prog,
			world:     w,
			remaining: 13,
			unitIDs:   ids,
			sandbox:   true,
		}
	}
	n := prog.DeployCount()
	if s := w.ActiveSquad(); s != nil && len(s.MemberIDs) < n {
		n = len(s.MemberIDs)
	}
	men := prog.Pool.Deploy(n)
	nameSquad(w, men)
	return &Battle{
		prog:      prog,
		world:     w,
		remaining: prog.Pool.Remaining(),
		deployed:  men,
		unitIDs:   ids,
	}
}

func nameCoverSquad(w *sim.World) {
	names := []string{"Jools", "Jops", "Stoo"}
	s := w.ActiveSquad()
	if s == nil {
		return
	}
	for i, id := range s.MemberIDs {
		u := w.Unit(id)
		if u == nil || i >= len(names) {
			continue
		}
		u.Name = names[i]
	}
}

func nameSquad(w *sim.World, men []campaign.Soldier) {
	s := w.ActiveSquad()
	if s == nil {
		return
	}
	for i, id := range s.MemberIDs {
		if i >= len(men) {
			break
		}
		u := w.Unit(id)
		if u == nil {
			continue
		}
		u.Name = men[i].Name
		u.Rank = int(men[i].Rank)
	}
}

func (b *Battle) Enter() { render.ResetAnim() }
func (b *Battle) Leave() {}

// CursorKind is the playfield pointer: board and exit sit on a skidoo.
func (b *Battle) CursorKind(p input.Pointer) int {
	if p.Right || ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight) {
		return render.PointerCrosshair
	}
	if b.world == nil || b.mapOpen || p.X < float64(render.HUDWidth) {
		return render.PointerArrow
	}
	hover, _ := b.world.VehicleHover(b.world.Camera.WorldX(p.X), b.world.Camera.WorldY(p.Y))
	switch hover {
	case sim.HoverBoard:
		return render.PointerBoard
	case sim.HoverExit:
		return render.PointerExit
	default:
		return render.PointerArrow
	}
}

func (b *Battle) Update(h Host) error {
	defer func() {
		render.Advance(1.0/TPS, b.world)
		if b.world != nil {
			x, y := hearPoint(b.world)
			h.SetListener(x, y)
			hum := b.world.EngineHumAt(x, y, audio.HearFar)
			h.SetEngine(hum.Run, hum.Pitch)
			h.PlayCues(b.world.TakeCues())
		}
	}()
	p := h.Pointer()
	if b.world.Status != sim.Playing {
		b.settleOnce(b.world.Status == sim.Won)
		if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			b.leaveBattle(h, b.world.Status == sim.Won)
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		b.world.SetActiveSquad(sim.SquadSnake)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		b.world.SetActiveSquad(sim.SquadEagle)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		b.world.SetActiveSquad(sim.SquadPanther)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		b.world.ToggleSpecial()
	}
	wx := b.world.Camera.WorldX(p.X)
	wy := b.world.Camera.WorldY(p.Y)
	onHUD := p.X < float64(render.HUDWidth)
	hover, veh := b.world.VehicleHover(wx, wy)
	inVeh := b.world.LeaderInVehicle()
	// Right-held + left click, or Space: leader's selected special. Not a move order.
	if !onHUD && !b.mapOpen && (p.ChordGrenade || inpututil.IsKeyJustPressed(ebiten.KeySpace)) {
		b.world.UseSpecial(wx, wy)
	}
	if p.LeftDown && onHUD {
		kind, id := render.HitHUD(b.world, p.X, p.Y)
		switch kind {
		case render.HitSplit:
			b.world.Split()
		case render.HitGrenade:
			b.world.UseAmmoIcon(sim.SpecialGrenade)
		case render.HitRocket:
			b.world.UseAmmoIcon(sim.SpecialRocket)
		case render.HitMember:
			b.world.ToggleSelect(id)
		case render.HitSquad:
			b.world.SetActiveSquad(sim.SquadID(id))
		case render.HitMap:
			b.mapOpen = !b.mapOpen
		}
	} else if p.LeftDown && b.mapOpen {
		b.mapOpen = false
	} else if p.LeftDown && !p.ChordGrenade && hover == sim.HoverBoard && veh != nil {
		b.world.CommandBoard(veh.ID)
	} else if p.LeftDown && !p.ChordGrenade && hover == sim.HoverExit && veh != nil {
		b.world.CommandExit(veh.ID)
	} else if !onHUD && !b.mapOpen && inVeh && (p.LeftDown || p.Left) && !p.ChordGrenade {
		b.world.SetDrive(wx, wy, true)
	} else if p.LeftDown && !p.ChordGrenade {
		b.world.CommandMove(wx, wy, true)
	} else if p.Left && !onHUD && !b.mapOpen {
		if s := b.world.ActiveSquad(); s != nil && s.HasDest {
			b.world.CommandMove(wx, wy, false)
		}
	} else if inVeh {
		b.world.SetDrive(wx, wy, false)
	}
	if onHUD || b.mapOpen {
		b.world.SetFire(b.world.AimX, b.world.AimY, false)
	} else {
		firing := p.Right ||
			ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
			ebiten.IsKeyPressed(ebiten.KeyControlRight)
		b.world.SetFire(wx, wy, firing)
		// The strip covers the left edge, so pan from the playfield beside it.
		b.world.Camera.ScrollToward(
			p.X-float64(render.HUDWidth), p.Y,
			ScreenWidth-render.HUDWidth, ScreenHeight,
			1.0/TPS,
		)
	}
	b.world.Step(1.0 / TPS)
	if x, y, ok := b.world.CameraFocus(); ok {
		b.world.Camera.Contain(x, y)
	}
	return nil
}

func (b *Battle) settleOnce(won bool) {
	if b.settled {
		return
	}
	b.settled = true
	if b.sandbox {
		return
	}
	graves, survivors := b.tally()
	if won {
		for i := range survivors {
			survivors[i].PhasesThisMission++
		}
	}
	b.prog.Graves += graves
	b.prog.Pool.Return(survivors)
	b.remaining = b.prog.Pool.Remaining()
}

func (b *Battle) tally() (graves int, survivors []campaign.Soldier) {
	for i, id := range b.unitIDs {
		u := b.world.Unit(id)
		if i >= len(b.deployed) || u == nil || !u.Living() {
			graves++
			continue
		}
		survivors = append(survivors, b.deployed[i])
	}
	return graves, survivors
}

func (b *Battle) leaveBattle(h Host, won bool) {
	if won {
		h.PlayKind(sim.CueWin)
	} else {
		h.PlayKind(sim.CueFail)
	}
	if b.sandbox {
		h.Switch(NewTitle())
		return
	}
	if won {
		b.prog.OnPhaseWon()
	} else {
		b.prog.MarkGameOverIfNeeded()
	}
	h.Switch(NewBootHill(b.prog))
}

func (b *Battle) arctic() bool {
	return !b.sandbox && b.prog != nil && b.prog.Phase != nil && b.prog.Phase.Terrain == "arctic"
}

func (b *Battle) battlefieldColor() color.Color {
	if b.arctic() {
		return snowColor
	}
	return battleColor
}

func (b *Battle) Draw(screen *ebiten.Image) {
	screen.Fill(b.battlefieldColor())
	s := render.PictureScale()
	left := int(math.Floor(float64(render.HUDWidth) * s))
	play := screen.SubImage(image.Rect(left, 0, screen.Bounds().Dx(), screen.Bounds().Dy())).(*ebiten.Image)
	terrain := "grass"
	if b.arctic() {
		terrain = "snow"
	}
	render.Field(play, b.world, terrain)
	if b.mapOpen {
		render.Overview(screen, b.world, terrain)
	}
	render.HUD(screen, b.world, b.remaining)
	switch b.world.Status {
	case sim.Won:
		render.Text(screen, "PHASE COMPLETE\nClick or Enter", 70, 96)
	case sim.Lost:
		render.Text(screen, "PHASE FAILED\nClick or Enter", 70, 96)
	}
}

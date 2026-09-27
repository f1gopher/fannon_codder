package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

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
}

var (
	titleColor  = color.RGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xff}
	battleColor = color.RGBA{R: 0x3a, G: 0x8f, B: 0x3a, A: 0xff}
	snowColor   = color.RGBA{R: 0xe6, G: 0xec, B: 0xf0, A: 0xff}
)

// Title is the click-through splash.
type Title struct{}

func NewTitle() *Title { return &Title{} }

func (t *Title) Enter() {}
func (t *Title) Leave() {}

func (t *Title) Update(h Host) error {
	p := h.Pointer()
	if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		h.Switch(NewBootHill(h.Progress()))
	}
	return nil
}

func (t *Title) Draw(screen *ebiten.Image) {
	screen.Fill(titleColor)
	ebitenutil.DebugPrint(screen, "FANNON CODDER\n\nClick or press Enter")
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

// NewHazardBattle is the chunk 18 sandbox: mine, quicksand, civilian, doorless hut.
func NewHazardBattle(prog *Progress) *Battle {
	return battleFromWorld(prog, sim.NewHazardWorld(), true)
}

func battleFromWorld(prog *Progress, w *sim.World, sandbox bool) *Battle {
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

func (b *Battle) Enter() {}
func (b *Battle) Leave() {}

func (b *Battle) Update(h Host) error {
	p := h.Pointer()
	if b.world.Status != sim.Playing {
		b.settleOnce(b.world.Status == sim.Won)
		if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			b.leaveBattle(h, b.world.Status == sim.Won)
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		b.settleOnce(false)
		b.leaveBattle(h, false)
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
	wx := p.X + b.world.Camera.X
	wy := p.Y + b.world.Camera.Y
	onHUD := p.X < float64(render.HUDWidth)
	// Right-held + left click, or Space: leader's selected special. Not a move order.
	if !onHUD && !b.mapOpen && (p.ChordGrenade || inpututil.IsKeyJustPressed(ebiten.KeySpace)) {
		b.world.ThrowGrenade(wx, wy)
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
	} else if p.LeftDown && !p.ChordGrenade {
		b.world.CommandMove(wx, wy, true)
	} else if p.Left && !onHUD && !b.mapOpen {
		if s := b.world.ActiveSquad(); s != nil && s.HasDest {
			b.world.CommandMove(wx, wy, false)
		}
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

func (b *Battle) battlefieldColor() color.Color {
	if !b.sandbox && b.prog != nil && b.prog.Phase != nil && b.prog.Phase.Terrain == "arctic" {
		return snowColor
	}
	return battleColor
}

func (b *Battle) Draw(screen *ebiten.Image) {
	screen.Fill(b.battlefieldColor())
	render.Tiles(screen, b.world.Map, b.world.Camera)
	render.Solids(screen, b.world)
	render.Units(screen, b.world)
	render.Projectiles(screen, b.world)
	render.Grenades(screen, b.world)
	if b.mapOpen {
		render.Overview(screen, b.world)
	}
	render.HUD(screen, b.world, b.remaining)
	switch b.world.Status {
	case sim.Won:
		ebitenutil.DebugPrint(screen, "\n\n  PHASE COMPLETE\n  Click or Enter")
	case sim.Lost:
		ebitenutil.DebugPrint(screen, "\n\n  PHASE FAILED\n  Click or Enter")
	}
}

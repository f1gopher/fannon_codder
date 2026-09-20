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
	n := prog.DeployCount()
	if s := w.ActiveSquad(); s != nil && len(s.MemberIDs) < n {
		n = len(s.MemberIDs)
	}
	men := prog.Pool.Deploy(n)
	nameSquad(w, men)
	ids := []int{}
	if s := w.ActiveSquad(); s != nil {
		ids = append(ids, s.MemberIDs...)
	}
	return &Battle{
		prog:      prog,
		world:     w,
		remaining: prog.Pool.Remaining(),
		deployed:  men,
		unitIDs:   ids,
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
		b.settleOnce()
		if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			b.leaveBattle(h, b.world.Status == sim.Won)
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		b.settleOnce()
		b.leaveBattle(h, false)
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

func (b *Battle) settleOnce() {
	if b.settled {
		return
	}
	b.settled = true
	graves, survivors := b.tally()
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
		b.prog.OnPhaseWon()
	} else {
		b.prog.MarkGameOverIfNeeded()
	}
	h.Switch(NewBootHill(b.prog))
}

func (b *Battle) Draw(screen *ebiten.Image) {
	screen.Fill(battleColor)
	render.Tiles(screen, b.world.Map, b.world.Camera)
	render.Units(screen, b.world)
	render.Projectiles(screen, b.world)
	render.HUD(screen, b.world, b.remaining)
	switch b.world.Status {
	case sim.Won:
		ebitenutil.DebugPrint(screen, "\n\n  PHASE COMPLETE\n  Click or Enter")
	case sim.Lost:
		ebitenutil.DebugPrint(screen, "\n\n  PHASE FAILED\n  Click or Enter")
	}
}

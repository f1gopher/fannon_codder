package app

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var stubBG = color.RGBA{R: 0x18, G: 0x18, B: 0x28, A: 0xff}

// MissionStub is shown when the next mission has no map yet (Chunk 09).
type MissionStub struct {
	prog *Progress
}

func NewMissionStub(prog *Progress) *MissionStub { return &MissionStub{prog: prog} }

func (s *MissionStub) Enter() {}
func (s *MissionStub) Leave() {}

func (s *MissionStub) Update(h Host) error {
	p := h.Pointer()
	if p.LeftDown || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		h.Switch(NewBootHill(s.prog))
	}
	return nil
}

func (s *MissionStub) Draw(screen *ebiten.Image) {
	screen.Fill(stubBG)
	n := s.prog.MissionsCompleted + 1
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"MISSION %d\n\nNot implemented yet.\n\nQueue %d\n\nClick or Enter — Boot Hill",
		n, s.prog.Pool.Remaining(),
	))
}

package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"fannon-codder/internal/campaign"
	"fannon-codder/internal/sim"
)

const HUDWidth = 52

var (
	hudFill   = color.RGBA{R: 0x08, G: 0x08, B: 0x08, A: 0xe8}
	logoFill  = color.RGBA{R: 0x3a, G: 0xb0, B: 0x3a, A: 0xff}
	hudPanel  *ebiten.Image
	logoPanel *ebiten.Image
)

func hudBG() *ebiten.Image {
	if hudPanel == nil {
		hudPanel = ebiten.NewImage(HUDWidth, 256)
		hudPanel.Fill(hudFill)
	}
	return hudPanel
}

func snakeLogo() *ebiten.Image {
	if logoPanel == nil {
		logoPanel = ebiten.NewImage(12, 12)
		logoPanel.Fill(logoFill)
	}
	return logoPanel
}

// HUD draws the left status strip: Snake logo and named troopers.
func HUD(dst *ebiten.Image, w *sim.World, remaining int) {
	if w == nil {
		return
	}
	dst.DrawImage(hudBG(), nil)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(4, 4)
	dst.DrawImage(snakeLogo(), op)
	ebitenutil.DebugPrintAt(dst, "S", 6, 5)
	if remaining >= 0 {
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("R%d", remaining), 20, 6)
	}
	s := w.ActiveSquad()
	if s == nil {
		return
	}
	y := 22
	for _, id := range s.MemberIDs {
		u := w.Unit(id)
		if u == nil || !u.Living() {
			continue
		}
		name := u.Name
		if name == "" {
			name = "?"
		}
		abbr := campaign.Rank(u.Rank).Abbrev()
		ebitenutil.DebugPrintAt(dst, abbr, 2, y)
		ebitenutil.DebugPrintAt(dst, name, 2, y+8)
		y += 20
	}
}

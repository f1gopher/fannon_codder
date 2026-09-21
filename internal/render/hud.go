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

// HUD click kinds. id is a unit id for HitMember and a sim.SquadID for HitSquad.
type HUDHit int

const (
	HitNone HUDHit = iota
	HitSplit
	HitGrenade
	HitRocket
	HitMember
	HitSquad
)

const (
	hudLogoX, hudLogoY, hudLogoS = 4, 4, 12
	hudGrenX, hudGrenY           = 2, 18
	hudRockX                     = 28
	hudIconW, hudIconH           = 22, 16
	hudSquadTop                  = 36
	hudHeadH                     = 12
	hudManH                      = 18
)

var (
	hudFill     = color.RGBA{R: 0x08, G: 0x08, B: 0x08, A: 0xe8}
	grenFill    = color.RGBA{R: 0x50, G: 0x80, B: 0x40, A: 0xff}
	rocketFill  = color.RGBA{R: 0x70, G: 0x60, B: 0x30, A: 0xff}
	selectFill  = color.RGBA{R: 0x40, G: 0x40, B: 0x10, A: 0xff}
	outlineFill = color.RGBA{R: 0xff, G: 0xee, B: 0x80, A: 0xff}
	hudPanel    *ebiten.Image
	grenPanel   *ebiten.Image
	rocketPanel *ebiten.Image
	whitePx     *ebiten.Image
)

func hudBG() *ebiten.Image {
	if hudPanel == nil {
		hudPanel = ebiten.NewImage(HUDWidth, 256)
		hudPanel.Fill(hudFill)
	}
	return hudPanel
}

type logoKey struct {
	c color.RGBA
	n int
}

var logoCache = map[logoKey]*ebiten.Image{}

func colorLogo(c color.RGBA, size int) *ebiten.Image {
	k := logoKey{c, size}
	if img, ok := logoCache[k]; ok {
		return img
	}
	img := ebiten.NewImage(size, size)
	img.Fill(c)
	logoCache[k] = img
	return img
}

type hudRect struct {
	x, y, w, h int
}

func (r hudRect) contains(px, py float64) bool {
	return px >= float64(r.x) && py >= float64(r.y) && px < float64(r.x+r.w) && py < float64(r.y+r.h)
}

type hudMan struct {
	hudRect
	id     int
	squad  sim.SquadID
	active bool
}

type hudHeader struct {
	hudRect
	id sim.SquadID
}

type hudLayout struct {
	logo, gren, rock hudRect
	headers          []hudHeader
	men              []hudMan
}

func layoutHUD(w *sim.World) hudLayout {
	lay := hudLayout{
		logo: hudRect{hudLogoX, hudLogoY, hudLogoS, hudLogoS},
		gren: hudRect{hudGrenX, hudGrenY, hudIconW, hudIconH},
		rock: hudRect{hudRockX, hudGrenY, hudIconW, hudIconH},
	}
	if w == nil {
		return lay
	}
	y := hudSquadTop
	active := w.ActiveSquad()
	for _, id := range []sim.SquadID{sim.SquadSnake, sim.SquadEagle, sim.SquadPanther} {
		s := w.SquadByID(id)
		if s == nil {
			continue
		}
		var living []int
		for _, mid := range s.MemberIDs {
			u := w.Unit(mid)
			if u != nil && u.Living() {
				living = append(living, mid)
			}
		}
		if len(living) == 0 {
			continue
		}
		on := active != nil && active.ID == id
		lay.headers = append(lay.headers, hudHeader{
			hudRect: hudRect{0, y, HUDWidth, hudHeadH},
			id:      id,
		})
		y += hudHeadH
		for _, mid := range living {
			lay.men = append(lay.men, hudMan{
				hudRect: hudRect{0, y, HUDWidth, hudManH},
				id:      mid,
				squad:   id,
				active:  on,
			})
			y += hudManH
		}
	}
	return lay
}

// HitHUD maps a logical screen pixel on the status strip to a control.
func HitHUD(w *sim.World, x, y float64) (HUDHit, int) {
	lay := layoutHUD(w)
	if lay.logo.contains(x, y) {
		return HitSplit, 0
	}
	if lay.gren.contains(x, y) {
		return HitGrenade, 0
	}
	if lay.rock.contains(x, y) {
		return HitRocket, 0
	}
	for _, h := range lay.headers {
		if h.contains(x, y) {
			if s := w.ActiveSquad(); s != nil && s.ID == h.id {
				return HitSplit, int(h.id)
			}
			return HitSquad, int(h.id)
		}
	}
	for _, m := range lay.men {
		if !m.contains(x, y) {
			continue
		}
		if m.active {
			return HitMember, m.id
		}
		return HitSquad, int(m.squad)
	}
	return HitNone, 0
}

// HUD draws the left status strip: troop logo, ammo, and every squad.
func HUD(dst *ebiten.Image, w *sim.World, remaining int) {
	if w == nil {
		return
	}
	dst.DrawImage(hudBG(), nil)
	lay := layoutHUD(w)
	active := w.ActiveSquad()
	logoID := sim.SquadSnake
	if active != nil {
		logoID = active.ID
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(lay.logo.x), float64(lay.logo.y))
	dst.DrawImage(colorLogo(squadFill(logoID, false), hudLogoS), op)
	ebitenutil.DebugPrintAt(dst, logoID.Letter(), lay.logo.x+2, lay.logo.y+1)
	if remaining >= 0 {
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("R%d", remaining), 20, 6)
	}
	gCount, rCount := 0, 0
	if active != nil {
		gCount, rCount = active.Grenades, active.Rockets
	}
	drawAmmoIcon(dst, lay.gren, "G", gCount, grenIcon(), w.GrenadeShare)
	drawAmmoIcon(dst, lay.rock, "R", rCount, rocketIcon(), w.RocketShare)

	selected := map[int]bool{}
	for _, id := range w.Selected {
		selected[id] = true
	}
	for _, h := range lay.headers {
		swatch := colorLogo(squadFill(h.id, false), 8)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(2, float64(h.y+2))
		dst.DrawImage(swatch, op)
		mark := h.id.Letter()
		if active != nil && active.ID == h.id {
			mark = ">" + mark
		}
		ebitenutil.DebugPrintAt(dst, mark, 12, h.y+2)
	}
	for _, m := range lay.men {
		u := w.Unit(m.id)
		if u == nil {
			continue
		}
		if selected[m.id] {
			fillRect(dst, 1, m.y, HUDWidth-2, hudManH-1, selectFill)
		}
		name := u.Name
		if name == "" {
			name = "?"
		}
		ebitenutil.DebugPrintAt(dst, campaign.Rank(u.Rank).Abbrev(), 2, m.y)
		ebitenutil.DebugPrintAt(dst, name, 2, m.y+8)
	}
}

func grenIcon() *ebiten.Image {
	if grenPanel == nil {
		grenPanel = ebiten.NewImage(hudIconW, hudIconH)
		grenPanel.Fill(grenFill)
	}
	return grenPanel
}

func rocketIcon() *ebiten.Image {
	if rocketPanel == nil {
		rocketPanel = ebiten.NewImage(hudIconW, hudIconH)
		rocketPanel.Fill(rocketFill)
	}
	return rocketPanel
}

func drawAmmoIcon(dst *ebiten.Image, r hudRect, label string, n int, img *ebiten.Image, mode sim.AmmoShare) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(r.x), float64(r.y))
	dst.DrawImage(img, op)
	strokeShare(dst, r, mode)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%s%d", label, n), r.x+2, r.y+4)
}

func strokeShare(dst *ebiten.Image, r hudRect, mode sim.AmmoShare) {
	switch mode {
	case sim.ShareNone:
		return
	case sim.ShareHalf:
		fillRect(dst, r.x, r.y, r.w, 1, outlineFill)
		fillRect(dst, r.x, r.y+r.h-1, r.w, 1, outlineFill)
	default:
		fillRect(dst, r.x, r.y, r.w, 1, outlineFill)
		fillRect(dst, r.x, r.y+r.h-1, r.w, 1, outlineFill)
		fillRect(dst, r.x, r.y, 1, r.h, outlineFill)
		fillRect(dst, r.x+r.w-1, r.y, 1, r.h, outlineFill)
	}
}

func fillRect(dst *ebiten.Image, x, y, w, h int, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w), float64(h))
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	dst.DrawImage(pixel(), op)
}

func pixel() *ebiten.Image {
	if whitePx == nil {
		whitePx = ebiten.NewImage(1, 1)
		whitePx.Fill(color.White)
	}
	return whitePx
}

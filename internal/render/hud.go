package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

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
	HitMap
)

const (
	hudLogoX, hudLogoY, hudLogoS = 4, 4, 12
	hudGrenX, hudGrenY           = 2, 18
	hudRockX                     = 28
	hudIconW, hudIconH           = 22, 16
	hudSquadTop                  = 36
	hudHeadH                     = 12
	hudManH                      = 18
	hudFootX, hudFootY, hudFootS = 40, 4, 8
	hudMapX, hudMapY, hudMapS    = 2, 238, 16
)

var (
	hudFill     = color.RGBA{R: 0x08, G: 0x08, B: 0x08, A: 0xff}
	grenFill    = color.RGBA{R: 0x50, G: 0x80, B: 0x40, A: 0xff}
	rocketFill  = color.RGBA{R: 0x70, G: 0x60, B: 0x30, A: 0xff}
	selectFill  = color.RGBA{R: 0x40, G: 0x40, B: 0x10, A: 0xff}
	outlineFill = color.RGBA{R: 0xff, G: 0xee, B: 0x80, A: 0xff}
	selectSpec  = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	activeRow   = color.RGBA{R: 0x14, G: 0x28, B: 0x18, A: 0xff}
	footFill    = color.RGBA{R: 0xc8, G: 0xa0, B: 0x60, A: 0xff}
	vehicleFill = color.RGBA{R: 0xc0, G: 0xc0, B: 0xc8, A: 0xff}
	mapIconFill = color.RGBA{R: 0x30, G: 0x48, B: 0x30, A: 0xff}
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

type hudBlock struct {
	hudRect
	id     sim.SquadID
	active bool
}

type hudLayout struct {
	logo, gren, rock hudRect
	foot, mapIcon    hudRect
	headers          []hudHeader
	men              []hudMan
	blocks           []hudBlock
}

func layoutHUD(w *sim.World) hudLayout {
	lay := hudLayout{
		logo:    hudRect{hudLogoX, hudLogoY, hudLogoS, hudLogoS},
		gren:    hudRect{hudGrenX, hudGrenY, hudIconW, hudIconH},
		rock:    hudRect{hudRockX, hudGrenY, hudIconW, hudIconH},
		foot:    hudRect{hudFootX, hudFootY, hudFootS, hudFootS},
		mapIcon: hudRect{hudMapX, hudMapY, hudMapS, hudMapS},
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
		if y+hudHeadH > hudMapY {
			break
		}
		on := active != nil && active.ID == id
		top := y
		lay.headers = append(lay.headers, hudHeader{
			hudRect: hudRect{0, y, HUDWidth, hudHeadH},
			id:      id,
		})
		y += hudHeadH
		for _, mid := range living {
			if y+hudManH > hudMapY {
				break
			}
			lay.men = append(lay.men, hudMan{
				hudRect: hudRect{0, y, HUDWidth, hudManH},
				id:      mid,
				squad:   id,
				active:  on,
			})
			y += hudManH
		}
		lay.blocks = append(lay.blocks, hudBlock{
			hudRect: hudRect{0, top, HUDWidth, y - top},
			id:      id,
			active:  on,
		})
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
	if lay.mapIcon.contains(x, y) {
		return HitMap, 0
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
	blit(dst, hudBG(), 0, 0, 1, 1)
	lay := layoutHUD(w)
	active := w.ActiveSquad()
	logoID := sim.SquadSnake
	if active != nil {
		logoID = active.ID
	}
	blit(dst, colorLogo(squadFill(logoID, false), hudLogoS), float64(lay.logo.x), float64(lay.logo.y), 1, 1)
	Text(dst, logoID.Letter(), float64(lay.logo.x+2), float64(lay.logo.y+1))
	if remaining >= 0 {
		Text(dst, fmt.Sprintf("R%d", remaining), 20, 6)
	}
	gCount, rCount := 0, 0
	if active != nil {
		gCount, rCount = active.Grenades, active.Rockets
	}
	drawAmmoIcon(dst, lay.gren, "G", gCount, grenIcon(), w.GrenadeShare, w.Special == sim.SpecialGrenade)
	drawAmmoIcon(dst, lay.rock, "R", rCount, rocketIcon(), w.RocketShare, w.Special == sim.SpecialRocket)
	drawStance(dst, lay.foot, leaderOnFoot(w, active))
	drawMapIcon(dst, lay.mapIcon)

	selected := map[int]bool{}
	for _, id := range w.Selected {
		selected[id] = true
	}
	for _, b := range lay.blocks {
		if !b.active {
			continue
		}
		fillRect(dst, b.x, b.y, b.w, b.h, activeRow)
		fillRect(dst, b.x, b.y, 3, b.h, squadFill(b.id, false))
	}
	for _, h := range lay.headers {
		swatch := colorLogo(squadFill(h.id, false), 8)
		blit(dst, swatch, 2, float64(h.y+2), 1, 1)
		mark := h.id.Letter()
		if active != nil && active.ID == h.id {
			mark = ">" + mark
		}
		Text(dst, mark, 12, float64(h.y+2))
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
		Text(dst, campaign.Rank(u.Rank).Abbrev(), 2, float64(m.y))
		Text(dst, name, 2, float64(m.y+8))
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

func drawAmmoIcon(dst *ebiten.Image, r hudRect, label string, n int, img *ebiten.Image, mode sim.AmmoShare, selected bool) {
	blit(dst, img, float64(r.x), float64(r.y), 1, 1)
	strokeShare(dst, r, mode)
	if selected {
		strokeRect(dst, r, selectSpec)
	}
	Text(dst, fmt.Sprintf("%s%d", label, n), float64(r.x+2), float64(r.y+4))
}

func strokeRect(dst *ebiten.Image, r hudRect, c color.Color) {
	fillRect(dst, r.x, r.y, r.w, 1, c)
	fillRect(dst, r.x, r.y+r.h-1, r.w, 1, c)
	fillRect(dst, r.x, r.y, 1, r.h, c)
	fillRect(dst, r.x+r.w-1, r.y, 1, r.h, c)
}

func leaderOnFoot(w *sim.World, s *sim.Squad) bool {
	if s == nil {
		return true
	}
	u := w.Unit(s.LeaderID)
	return u == nil || u.VehicleID == 0
}

func drawStance(dst *ebiten.Image, r hudRect, foot bool) {
	if foot {
		fillRect(dst, r.x, r.y+2, 3, r.h-2, footFill)
		fillRect(dst, r.x+5, r.y+2, 3, r.h-2, footFill)
		return
	}
	fillRect(dst, r.x, r.y+2, r.w, r.h-3, vehicleFill)
}

func drawMapIcon(dst *ebiten.Image, r hudRect) {
	fillRect(dst, r.x, r.y, r.w, r.h, mapIconFill)
	strokeRect(dst, r, overviewGrass)
	Text(dst, "M", float64(r.x+4), float64(r.y+4))
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
	fillRectF(dst, float64(x), float64(y), float64(w), float64(h), c)
}

func fillRectF(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	if dst == nil || w <= 0 || h <= 0 {
		return
	}
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(w*s, h*s)
	op.GeoM.Translate(x*s, y*s)
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

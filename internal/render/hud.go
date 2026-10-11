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
	HitPause
)

const (
	hudGrenX, hudGrenY              = 2, 2
	hudRockX                        = 28
	hudIconW, hudIconH              = 22, 16
	hudSquadTop                     = 20
	hudHeadH                        = 20
	hudMarkS                        = 16
	hudManH                         = 10
	hudRankS                        = 8
	hudFootS                        = 8
	hudMapX, hudMapY, hudMapS       = 2, 238, 16
	hudPauseX, hudPauseY, hudPauseS = 20, 238, 16
)

var (
	hudFill     = color.RGBA{R: 0x08, G: 0x08, B: 0x08, A: 0xff}
	grenFill    = color.RGBA{R: 0x50, G: 0x80, B: 0x40, A: 0xff}
	rocketFill  = color.RGBA{R: 0x70, G: 0x60, B: 0x30, A: 0xff}
	selectFill  = color.RGBA{R: 0x40, G: 0x40, B: 0x10, A: 0xff}
	outlineFill = color.RGBA{R: 0xff, G: 0xee, B: 0x80, A: 0xff}
	selectSpec  = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
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
	gren, rock     hudRect
	mapIcon, pause hudRect
	headers        []hudHeader
	men            []hudMan
	blocks         []hudBlock
}

func layoutHUD(w *sim.World) hudLayout {
	lay := hudLayout{
		gren:    hudRect{hudGrenX, hudGrenY, hudIconW, hudIconH},
		rock:    hudRect{hudRockX, hudGrenY, hudIconW, hudIconH},
		mapIcon: hudRect{hudMapX, hudMapY, hudMapS, hudMapS},
		pause:   hudRect{hudPauseX, hudPauseY, hudPauseS, hudPauseS},
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
	if lay.gren.contains(x, y) {
		return HitGrenade, 0
	}
	if lay.rock.contains(x, y) {
		return HitRocket, 0
	}
	if lay.mapIcon.contains(x, y) {
		return HitMap, 0
	}
	if lay.pause.contains(x, y) {
		return HitPause, 0
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

// HUD draws the left status strip: ammo and every squad.
// paused strokes the pause button.
func HUD(dst *ebiten.Image, w *sim.World, paused bool) {
	if w == nil {
		return
	}
	blit(dst, hudBG(), 0, 0, 1, 1)
	lay := layoutHUD(w)
	active := w.ActiveSquad()
	gCount, rCount := 0, 0
	if active != nil {
		gCount, rCount = active.Grenades, active.Rockets
	}
	drawAmmoIcon(dst, lay.gren, gCount, "ui/grenade", grenIcon(), w.GrenadeShare, w.Special == sim.SpecialGrenade)
	drawAmmoIcon(dst, lay.rock, rCount, "ui/rocket", rocketIcon(), w.RocketShare, w.Special == sim.SpecialRocket)
	drawMapIcon(dst, lay.mapIcon)
	drawPauseIcon(dst, lay.pause, paused)

	selected := map[int]bool{}
	for _, id := range w.Selected {
		selected[id] = true
	}
	for _, b := range lay.blocks {
		// One dark pixel between squads keeps the plates from running together.
		fillRect(dst, b.x, b.y, b.w, b.h-1, squadPlate(b.id, b.active))
		if b.active {
			strokeRect(dst, hudRect{b.x, b.y, b.w, b.h - 1}, squadEdge(b.id))
			fillRect(dst, b.x, b.y, 3, b.h-1, squadEdge(b.id))
		}
	}
	for _, h := range lay.headers {
		mx := float64(h.x) + float64(h.w-hudMarkS)/2
		my := float64(h.y) + float64(h.h-hudMarkS)/2
		// Marks are 64 source px and normally draw at 8 world px. Two times that fills the header.
		if !drawUIMul(dst, squadMark(h.id), mx, my, float64(hudMarkS)/8) {
			blit(dst, colorLogo(squadFill(h.id, false), hudMarkS), mx, my, 1, 1)
		}
		stance := hudRect{HUDWidth - hudFootS - 2, h.y + 2, hudFootS, hudFootS}
		drawStance(dst, stance, leaderOnFoot(w, w.SquadByID(h.id)))
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
		ry := float64(m.y + (hudManH-hudRankS)/2)
		drawRank(dst, campaign.Rank(u.Rank), 1, ry)
		// The face is 14/3 world pixels tall. Center it on the insignia.
		Text(dst, name, 1+float64(hudRankS)+2, ry+(float64(hudRankS)-TextPixels(1))/2)
	}
}

// rankKeys are the status-strip flashes, low rank to high. Each cell is
// 64×64 and draws at 8 world pixels to the left of the man's name.
var rankKeys = [...]string{
	campaign.Private:             "ui/rank-pte",
	campaign.Corporal:            "ui/rank-cpl",
	campaign.Sergeant:            "ui/rank-sgt",
	campaign.StaffSergeant:       "ui/rank-ssgt",
	campaign.SergeantFirstClass:  "ui/rank-sfc",
	campaign.MasterSergeant:      "ui/rank-msg",
	campaign.SergeantMajor:       "ui/rank-sgm",
	campaign.Specialist4:         "ui/rank-sp4",
	campaign.Specialist6:         "ui/rank-sp6",
	campaign.WarrantOfficer:      "ui/rank-wo",
	campaign.ChiefWarrantOfficer: "ui/rank-cwo",
	campaign.Captain:             "ui/rank-cpt",
	campaign.Major:               "ui/rank-maj",
	campaign.Colonel:             "ui/rank-col",
	campaign.BrigadierGeneral:    "ui/rank-bg",
	campaign.General:             "ui/rank-gen",
}

func drawRank(dst *ebiten.Image, rank campaign.Rank, x, y float64) {
	key := rankKeys[campaign.Private]
	if rank >= 0 && int(rank) < len(rankKeys) {
		key = rankKeys[rank]
	}
	if !drawUIAt(dst, key, x, y) {
		Text(dst, rank.Abbrev(), x, y)
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

// squadPlate is the header and roster fill. The hues follow the team marks:
// snake olive, eagle steel blue, panther bronze. Active lifts the same hue.
func squadPlate(id sim.SquadID, active bool) color.RGBA {
	switch id {
	case sim.SquadEagle:
		if active {
			return color.RGBA{R: 0x1c, G: 0x30, B: 0x48, A: 0xff}
		}
		return color.RGBA{R: 0x10, G: 0x18, B: 0x28, A: 0xff}
	case sim.SquadPanther:
		if active {
			return color.RGBA{R: 0x34, G: 0x22, B: 0x10, A: 0xff}
		}
		return color.RGBA{R: 0x1c, G: 0x12, B: 0x0c, A: 0xff}
	default:
		if active {
			return color.RGBA{R: 0x1c, G: 0x30, B: 0x1c, A: 0xff}
		}
		return color.RGBA{R: 0x10, G: 0x1c, B: 0x12, A: 0xff}
	}
}

// squadEdge is the bright rim on the selected team, sampled from the mark.
func squadEdge(id sim.SquadID) color.RGBA {
	switch id {
	case sim.SquadEagle:
		return color.RGBA{R: 0x48, G: 0x88, B: 0xc0, A: 0xff}
	case sim.SquadPanther:
		return color.RGBA{R: 0xc8, G: 0x88, B: 0x40, A: 0xff}
	default:
		return color.RGBA{R: 0x68, G: 0x98, B: 0x50, A: 0xff}
	}
}

func squadMark(id sim.SquadID) string {
	switch id {
	case sim.SquadEagle:
		return "ui/mark-eagle"
	case sim.SquadPanther:
		return "ui/mark-panther"
	default:
		return "ui/mark-snake"
	}
}

func drawAmmoIcon(dst *ebiten.Image, r hudRect, n int, key string, img *ebiten.Image, mode sim.AmmoShare, selected bool) {
	if !drawUIAt(dst, key, float64(r.x), float64(r.y)) {
		blit(dst, img, float64(r.x), float64(r.y), 1, 1)
	}
	strokeShare(dst, r, mode)
	if selected {
		strokeRect(dst, r, selectSpec)
	}
	TextCenter(dst, fmt.Sprintf("%d", n), float64(r.x)+float64(r.w)/2, float64(r.y)+float64(r.h)/2, pointerInk, 2)
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
	key := "ui/vehicle"
	if foot {
		key = "ui/foot"
	}
	if drawUIAt(dst, key, float64(r.x), float64(r.y)) {
		return
	}
	if foot {
		fillRect(dst, r.x, r.y+2, 3, r.h-2, footFill)
		fillRect(dst, r.x+5, r.y+2, 3, r.h-2, footFill)
		return
	}
	fillRect(dst, r.x, r.y+2, r.w, r.h-3, vehicleFill)
}

func drawMapIcon(dst *ebiten.Image, r hudRect) {
	if drawUIAt(dst, "ui/map", float64(r.x), float64(r.y)) {
		return
	}
	fillRect(dst, r.x, r.y, r.w, r.h, mapIconFill)
	strokeRect(dst, r, overviewGrass)
	Text(dst, "M", float64(r.x+4), float64(r.y+4))
}

func drawPauseIcon(dst *ebiten.Image, r hudRect, paused bool) {
	if !drawUIAt(dst, "ui/pause", float64(r.x), float64(r.y)) {
		fillRect(dst, r.x, r.y, r.w, r.h, mapIconFill)
		strokeRect(dst, r, overviewGrass)
		fillRect(dst, r.x+4, r.y+3, 3, r.h-6, outlineFill)
		fillRect(dst, r.x+9, r.y+3, 3, r.h-6, outlineFill)
	}
	if paused {
		strokeRect(dst, r, selectSpec)
	}
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

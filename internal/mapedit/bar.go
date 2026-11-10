package mapedit

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// barStamps is the first menu row, left to right. Keys 1–9 follow that order.
var barStamps = []string{
	stampGrunt, stampGrenadier, stampRocketeer, stampPatrol,
	stampBarrel, stampHut, stampDoor, stampGrenades, stampRockets,
}

// barRow2 is the second menu row. Shift+1 selects the first icon, and so on.
var barRow2 = []string{stampCivilian, stampHostage, stampTent, stampLeader, stampHandover, stampHome}

func menuRows() [][]string {
	return [][]string{barStamps, barRow2}
}

func allStamps() []string {
	var out []string
	for _, row := range menuRows() {
		out = append(out, row...)
	}
	return out
}

// barRowHeight is one icon row. BarHeight stacks every row.
func barRowHeight(scale float64) float64 {
	if scale <= 0 {
		scale = 1
	}
	h := 36 * scale / 3
	if h < 32 {
		h = 32
	}
	return h
}

// BarHeight is the menu strip in offscreen pixels.
func BarHeight(scale float64) float64 {
	return barRowHeight(scale) * float64(len(menuRows()))
}

func toolButtons(scale float64) (size, gap, x0, y0 float64) {
	size = barRowHeight(scale) - 10
	if size < 18 {
		size = 18
	}
	return size, 6, 8, 5
}

// ToolHit reports which menu icon is under an offscreen point.
// A point in a gap, or past the last icon, is not a tool.
func ToolHit(x, y, scale float64) (int, bool) {
	size, gap, x0, y0 := toolButtons(scale)
	rows := menuRows()
	rowH := barRowHeight(scale)
	if y < y0 || x < x0 {
		return 0, false
	}
	row := int((y - y0) / rowH)
	if row < 0 || row >= len(rows) {
		return 0, false
	}
	if y-y0-float64(row)*rowH >= size {
		return 0, false
	}
	step := size + gap
	rel := x - x0
	i := int(rel / step)
	if i < 0 || i >= len(rows[row]) {
		return 0, false
	}
	if rel-float64(i)*step > size {
		return 0, false
	}
	base := 0
	for r := 0; r < row; r++ {
		base += len(rows[r])
	}
	return base + i, true
}

var barBG = color.RGBA{R: 18, G: 24, B: 14, A: 255}

var iconLib struct {
	once sync.Once
	cels []*propCel
}

// iconCels matches barStamps. Patrol reuses the grunt and draws its own route.
func iconCels() []*propCel {
	iconLib.once.Do(func() {
		grunt, _ := loadCelRow("grunt/idle", 2)
		gren, _ := loadCelRow("grenadier/idle", 2)
		rocket, _ := loadCelRow("rocketeer/idle", 2)
		_, plain, grenade, rockets, barrel := sheets()
		door, _, _, _, _ := sheets()
		civil := civilianCel()
		iconLib.cels = []*propCel{grunt, gren, rocket, grunt, barrel, plain, door, grenade, rockets, civil, hostageCel(), tentCel(), leaderCel(), handoverCel(), homeCel()}
	})
	return iconLib.cels
}

func (e *Editor) drawBar(screen *ebiten.Image) {
	if screen == nil || e.offW < 1 {
		return
	}
	h := BarHeight(e.scale)
	fillRect(screen, 0, 0, e.offW, int(h), barBG)
	fillRect(screen, 0, int(h)-1, e.offW, 1, railSel)
	size, gap, x0, y0 := toolButtons(e.scale)
	cels := iconCels()
	rowH := barRowHeight(e.scale)
	n := 0
	for row, stamps := range menuRows() {
		for col, stamp := range stamps {
			x := x0 + float64(col)*(size+gap)
			y := y0 + float64(row)*rowH
			if stamp == e.stamp || (e.stamp == "" && stamp == stampGrunt) {
				fillRect(screen, int(x), int(y), int(size), int(size), railSel)
			}
			if n < len(cels) && cels[n] != nil {
				drawIcon(screen, cels[n], x, y, size)
			} else if stamp == stampHostage {
				fillRect(screen, int(x+size*0.25), int(y+size*0.2), int(size*0.5), int(size*0.6), color.RGBA{R: 0x3a, G: 0x6e, B: 0xc8, A: 255})
			} else if stamp == stampTent {
				fillRect(screen, int(x+4), int(y+size*0.35), int(size-8), int(size*0.45), color.RGBA{R: 0xe8, G: 0xdc, B: 0xc0, A: 255})
				fillRect(screen, int(x+size/2-1), int(y+size*0.42), 2, int(size*0.3), color.RGBA{R: 0xc8, G: 0x28, B: 0x28, A: 255})
			} else if stamp == stampLeader {
				fillRect(screen, int(x+size*0.28), int(y+size*0.28), int(size*0.44), int(size*0.5), color.RGBA{R: 0x2a, G: 0x32, B: 0x48, A: 255})
			} else if stamp == stampHandover {
				fillRect(screen, int(x+6), int(y+size*0.4), int(size-12), int(size*0.35), color.RGBA{R: 0xc4, G: 0xb4, B: 0x88, A: 255})
			} else if stamp == stampHome {
				fillRect(screen, int(x+6), int(y+size*0.32), int(size-12), int(size*0.48), color.RGBA{R: 0xf0, G: 0xe6, B: 0xc8, A: 255})
				fillRect(screen, int(x+size/2-2), int(y+size*0.52), 4, int(size*0.28), color.RGBA{R: 0xf0, G: 0xc0, B: 0x40, A: 255})
			}
			if stamp == stampPatrol {
				drawPatrolIcon(screen, x, y, size)
			}
			n++
		}
	}
	nameX := x0 + float64(len(barStamps))*(size+gap) + 8
	drawFace(screen, e.font, e.scale, stampName(e.stamp), nameX, y0+size*0.25, railInkSel)
	cx, cy := e.cursor()
	if i, ok := ToolHit(cx, cy, e.scale); ok {
		stamps := allStamps()
		col := i
		row := 0
		if i >= len(barStamps) {
			col = i - len(barStamps)
			row = 1
		}
		x := x0 + float64(col)*(size+gap)
		e.drawTip(screen, stampTip(stamps[i]), x, y0+float64(row)*rowH+size+4)
	}
}

func stampTip(stamp string) string {
	switch stamp {
	case stampGrunt:
		return "Grunt — left click places"
	case stampGrenadier:
		return "Grenadier — left click places"
	case stampRocketeer:
		return "Rocketeer — left click places"
	case stampPatrol:
		return "Patrol — left click places, W adds a point"
	case stampBarrel:
		return "Barrel — left click places"
	case stampHut:
		return "Hut — left click places"
	case stampDoor:
		return "Door hut — left click places"
	case stampGrenades:
		return "Grenade crate — left click places"
	case stampRockets:
		return "Rocket crate — left click places"
	case stampCivilian:
		return "Civilian — left click places"
	case stampHostage:
		return "Hostage — left click places"
	case stampTent:
		return "Tent — left click places"
	case stampLeader:
		return "Leader — left click places"
	case stampHandover:
		return "Handover — left click places"
	case stampHome:
		return "Home — left click places"
	default:
		return "Left click places"
	}
}

func (e *Editor) drawTip(dst *ebiten.Image, label string, x, y float64) {
	if dst == nil || label == "" {
		return
	}
	s := e.scale
	if s <= 0 {
		s = 1
	}
	face := 14 * s / 3
	tw := float64(len(label))*face*0.52 + 12
	th := face + 8
	if x+tw > float64(e.offW)-4 {
		x = float64(e.offW) - 4 - tw
	}
	if x < 4 {
		x = 4
	}
	fillRect(dst, int(x), int(y), int(tw), int(th), barBG)
	fillRect(dst, int(x), int(y), int(tw), 1, railSel)
	drawFace(dst, e.font, s, label, x+6, y+3, railInkSel)
}

func drawIcon(dst *ebiten.Image, cel *propCel, x, y, size float64) {
	if dst == nil || cel == nil || cel.img == nil || size < 4 {
		return
	}
	b := cel.img.Bounds()
	fw, fh := float64(b.Dx()), float64(b.Dy())
	if fw < 1 || fh < 1 {
		return
	}
	pad := size * 0.14
	sc := (size - pad*2) / fh
	if fw*sc > size-pad*2 {
		sc = (size - pad*2) / fw
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x+(size-fw*sc)/2, y+(size-fh*sc)/2)
	dst.DrawImage(cel.img, op)
}

func drawPatrolIcon(dst *ebiten.Image, x, y, size float64) {
	p := &vector.Path{}
	x0 := float32(x + size*0.18)
	y0 := float32(y + size*0.72)
	p.MoveTo(x0, y0)
	p.LineTo(float32(x+size*0.55), float32(y+size*0.42))
	p.LineTo(float32(x+size*0.82), float32(y+size*0.62))
	ink := &vector.DrawPathOptions{AntiAlias: true}
	ink.ColorScale.ScaleWithColor(markEnemy)
	vector.StrokePath(dst, p, &vector.StrokeOptions{Width: float32(size * 0.06)}, ink)
}

package mapedit

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"fannon-codder/internal/sim"
	"fannon-codder/internal/terrain"
)

const (
	// drawStep is the minimum distance, in world pixels, between samples.
	drawStep = 2
	// drawMinArea drops a click or a stroke that never opened into a patch.
	drawMinArea = 48
)

// editing is the regions view of a freeform phase. Terrain view and a
// tile grid still use the left button to scroll.
func (e *Editor) editing() bool {
	return e.err == nil && e.gridView && e.freeform && e.chart.W > 0 && e.chart.H > 0
}

// beginStroke starts a tree outline at a world point. The map outside
// the regions view, and a point off the map, do not start one.
func (e *Editor) beginStroke(x, y float64) bool {
	if !e.editing() || !e.onMap(x, y) {
		return false
	}
	e.drag = false
	e.stroke = [][2]float64{{x, y}}
	return true
}

// extendStroke appends a sample once the pointer has moved drawStep.
// A point past the map edge is pulled back onto the edge.
func (e *Editor) extendStroke(x, y float64) {
	if len(e.stroke) == 0 {
		return
	}
	x, y = e.clampMap(x, y)
	last := e.stroke[len(e.stroke)-1]
	dx, dy := x-last[0], y-last[1]
	if dx*dx+dy*dy < drawStep*drawStep {
		return
	}
	e.stroke = append(e.stroke, [2]float64{x, y})
}

// finishStroke records the release point, then closes the outline.
func (e *Editor) finishStroke(x, y float64) bool {
	if len(e.stroke) == 0 {
		return false
	}
	x, y = e.clampMap(x, y)
	last := e.stroke[len(e.stroke)-1]
	dx, dy := x-last[0], y-last[1]
	if dx*dx+dy*dy >= 0.25 {
		e.stroke = append(e.stroke, [2]float64{x, y})
	}
	return e.endStroke()
}

// endStroke closes the samples with the straight segment back to the
// first point and keeps the ring as a tree region. That segment is the
// shortest path between the two ends. A ring that is still a speck is
// dropped. The phase file is not written.
func (e *Editor) endStroke() bool {
	pts := e.stroke
	e.stroke = nil
	if len(pts) < 3 || ringArea(pts) < drawMinArea {
		return false
	}
	cr := terrain.ChartRegion{Tile: sim.TileTree, Name: "tree", Points: pts}
	e.chart.Regions = append(e.chart.Regions, cr)
	if e.added == nil {
		e.added = map[int][]terrain.ChartRegion{}
	}
	e.added[e.index] = append(e.added[e.index], cr)
	return true
}

func (e *Editor) onMap(x, y float64) bool {
	return x >= 0 && y >= 0 && x < float64(e.chart.W) && y < float64(e.chart.H)
}

func (e *Editor) clampMap(x, y float64) (float64, float64) {
	w, h := float64(e.chart.W), float64(e.chart.H)
	if x < 0 {
		x = 0
	} else if w > 0 && x > w {
		x = w
	}
	if y < 0 {
		y = 0
	} else if h > 0 && y > h {
		y = h
	}
	return x, y
}

func (e *Editor) worldCursor(cx, cy float64) (float64, float64) {
	s := e.scale
	if s <= 0 {
		s = 1
	}
	return e.scrollX + (cx-float64(RailWidth(e.offW)))/s, e.scrollY + cy/s
}

// ringArea is the absolute shoelace area. The closing edge is included.
func ringArea(pts [][2]float64) float64 {
	if len(pts) < 3 {
		return 0
	}
	sum := 0.0
	j := len(pts) - 1
	for i := range pts {
		sum += pts[j][0]*pts[i][1] - pts[i][0]*pts[j][1]
		j = i
	}
	return math.Abs(sum) / 2
}

func (e *Editor) drawStroke(dst *ebiten.Image) {
	pts := e.stroke
	if dst == nil || len(pts) == 0 {
		return
	}
	s := e.scale
	if s <= 0 {
		s = 1
	}
	ox := -e.scrollX * s
	oy := -e.scrollY * s
	path := regionPath(pts, ox, oy, s)
	if len(pts) >= 3 {
		fill := &vector.DrawPathOptions{AntiAlias: true}
		fill.ColorScale.ScaleWithColor(color.RGBA{R: 24, G: 72, B: 28, A: 140})
		vector.FillPath(dst, path, nil, fill)
	}
	hand := openPath(pts, ox, oy, s)
	ink := &vector.DrawPathOptions{AntiAlias: true}
	ink.ColorScale.ScaleWithColor(gridInk)
	width := float32(math.Max(1, s))
	vector.StrokePath(dst, hand, &vector.StrokeOptions{Width: width, LineJoin: vector.LineJoinRound}, ink)
	if len(pts) >= 2 {
		chord := &vector.Path{}
		last := pts[len(pts)-1]
		chord.MoveTo(float32(ox+last[0]*s), float32(oy+last[1]*s))
		chord.LineTo(float32(ox+pts[0][0]*s), float32(oy+pts[0][1]*s))
		closeInk := &vector.DrawPathOptions{AntiAlias: true}
		closeInk.ColorScale.ScaleWithColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
		vector.StrokePath(dst, chord, &vector.StrokeOptions{Width: width, LineJoin: vector.LineJoinRound}, closeInk)
	}
}

func openPath(pts [][2]float64, ox, oy, scale float64) *vector.Path {
	p := &vector.Path{}
	p.MoveTo(float32(ox+pts[0][0]*scale), float32(oy+pts[0][1]*scale))
	for _, pt := range pts[1:] {
		p.LineTo(float32(ox+pt[0]*scale), float32(oy+pt[1]*scale))
	}
	return p
}

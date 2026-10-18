package sim

import "math"

// The arrival wedge. The leader stays on the player start. The rest step
// north in pairs, left then right, so they are not stacked under the ropes.
const (
	dropFormX = 18.0
	dropFormY = 12.0
)

// FormDrop spreads the active squad into that wedge. A spot in water, a
// tree, or a cliff keeps the man's old position. The trail matches the
// new spots so the file does not walk on the first step.
func (w *World) FormDrop() {
	if w == nil {
		return
	}
	s := w.ActiveSquad()
	if s == nil {
		return
	}
	s.Kneel = true
	leader := w.Unit(s.LeaderID)
	if leader == nil {
		return
	}
	lx, ly := leader.X, leader.Y
	pts := make([]Vec2, 0, len(s.MemberIDs))
	for i, id := range s.MemberIDs {
		u := w.Unit(id)
		if u == nil || !u.Living() {
			continue
		}
		p := Vec2{X: u.X, Y: u.Y}
		if i > 0 {
			want := dropFormPoint(lx, ly, i)
			if dropFormStand(w, want) {
				p = want
			}
		}
		u.X, u.Y = p.X, p.Y
		u.Facing = -math.Pi / 2
		pts = append(pts, p)
	}
	if len(pts) > 1 {
		s.Trail = make([]Vec2, len(pts))
		for i, p := range pts {
			s.Trail[len(pts)-1-i] = p
		}
	}
}

func dropFormPoint(lx, ly float64, i int) Vec2 {
	row := (i + 1) / 2
	side := 1.0
	if i%2 == 1 {
		side = -1
	}
	return Vec2{
		X: lx + side*dropFormX*float64(row),
		Y: ly - dropFormY*float64(row),
	}
}

func dropFormStand(w *World, p Vec2) bool {
	if !w.Map.active() {
		return true
	}
	if !w.Map.Walkable(p.X, p.Y, UnitSize/2) {
		return false
	}
	switch w.Map.TileAtPixel(p.X, p.Y) {
	case TileWaterDeep, TileWaterShallow, TileQuicksand:
		return false
	default:
		return true
	}
}

package sim

import "math"

const losSample = 2.0 // px between samples; trees are 16 px so this cannot skip a tile

// HasLOS is true when no solid tile (tree/cliff) sits between the two points.
// Inactive maps never block (keeps empty-field tests working).
func (m *Map) HasLOS(x0, y0, x1, y1 float64) bool {
	_, _, hit := m.FirstSolid(x0, y0, x1, y1)
	return !hit
}

// FirstSolid walks the segment in small steps and returns the first solid
// sample. The origin cell is skipped so a muzzle on grass is not blocked.
func (m *Map) FirstSolid(x0, y0, x1, y1 float64) (hx, hy float64, hit bool) {
	if !m.active() {
		return 0, 0, false
	}
	dx := x1 - x0
	dy := y1 - y0
	dist := math.Hypot(dx, dy)
	if dist < 1e-9 {
		return 0, 0, false
	}
	nx, ny := dx/dist, dy/dist
	for d := losSample; d < dist; d += losSample {
		x := x0 + nx*d
		y := y0 + ny*d
		if m.solidAt(x, y) {
			return x, y, true
		}
	}
	if m.solidAt(x1, y1) {
		return x1, y1, true
	}
	return 0, 0, false
}

func (m *Map) solidAt(x, y float64) bool {
	tx := int(math.Floor(x / float64(TileSize)))
	ty := int(math.Floor(y / float64(TileSize)))
	return !m.WalkableTile(m.At(tx, ty))
}

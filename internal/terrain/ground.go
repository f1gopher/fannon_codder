package terrain

import "fannon-codder/internal/sim"

// Ground is what a blast throws. It follows the tile and the phase kind.
// Jungle and forest shores also count the thin sand strip.
type Ground uint8

const (
	GroundSoil Ground = iota
	GroundSand
	GroundSnow
	GroundWater
)

// GroundAt reports the material under a world point.
func GroundAt(m sim.Map, terrain string, x, y float64) Ground {
	if m.W <= 0 || m.H <= 0 || len(m.Tiles) < m.W*m.H {
		return GroundSoil
	}
	switch m.TileAtPixel(x, y) {
	case sim.TileWaterShallow, sim.TileWaterDeep:
		return GroundWater
	case sim.TileIce:
		return GroundSnow
	case sim.TileQuicksand:
		return GroundSand
	}
	switch KindOf(terrain) {
	case Snow:
		return GroundSnow
	case Desert:
		return GroundSand
	}
	if shoreSand(m, x, y) {
		return GroundSand
	}
	return GroundSoil
}

// shoreSand is the jungle and forest beach: a few pixels of sand inland of water.
func shoreSand(m sim.Map, x, y float64) bool {
	tx := int(x / sim.TileSize)
	ty := int(y / sim.TileSize)
	if tx < 0 || ty < 0 || tx >= m.W || ty >= m.H {
		return false
	}
	lx := x - float64(tx*sim.TileSize)
	ly := y - float64(ty*sim.TileSize)
	const reach = 6.0
	water := func(ox, oy int) bool {
		t := m.At(tx+ox, ty+oy)
		return t == sim.TileWaterShallow || t == sim.TileWaterDeep
	}
	if water(-1, 0) && lx < reach {
		return true
	}
	if water(1, 0) && lx > sim.TileSize-reach {
		return true
	}
	if water(0, -1) && ly < reach {
		return true
	}
	if water(0, 1) && ly > sim.TileSize-reach {
		return true
	}
	return false
}

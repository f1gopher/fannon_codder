package sim

import "math"

const TileSize = 16

// Tile is one map cell.
type Tile int

const (
	TileGrass Tile = iota
	TileTree
	TileWaterShallow
	TileWaterDeep
	TileIce
	TileQuicksand
	TileCliff
	TileBridge
	TileMine
)

// Map is a tile grid in world pixels (tile size 16).
type Map struct {
	W, H  int
	Tiles []Tile
}

func (m *Map) active() bool {
	return m != nil && m.W > 0 && m.H > 0 && len(m.Tiles) >= m.W*m.H
}

func (m *Map) PixelSize() (float64, float64) {
	if !m.active() {
		return 320, 256
	}
	return float64(m.W * TileSize), float64(m.H * TileSize)
}

func (m *Map) At(tx, ty int) Tile {
	if !m.active() || tx < 0 || ty < 0 || tx >= m.W || ty >= m.H {
		return TileTree // out of bounds is solid
	}
	return m.Tiles[ty*m.W+tx]
}

func (m *Map) WalkableTile(t Tile) bool {
	switch t {
	case TileTree, TileCliff:
		return false
	default:
		return true
	}
}

// Walkable reports whether a unit centred on (x,y) may stand there.
func (m *Map) Walkable(x, y, half float64) bool {
	if !m.active() {
		return true
	}
	x0 := int(x - half)
	y0 := int(y - half)
	x1 := int(x + half - 0.001)
	y1 := int(y + half - 0.001)
	tx0 := x0 / TileSize
	ty0 := y0 / TileSize
	tx1 := x1 / TileSize
	ty1 := y1 / TileSize
	if x0 < 0 {
		tx0 = -1
	}
	if y0 < 0 {
		ty0 = -1
	}
	for ty := ty0; ty <= ty1; ty++ {
		for tx := tx0; tx <= tx1; tx++ {
			if !m.WalkableTile(m.At(tx, ty)) {
				return false
			}
		}
	}
	return true
}

func TileCenter(tx, ty int) Vec2 {
	return Vec2{
		X: float64(tx*TileSize) + float64(TileSize)/2,
		Y: float64(ty*TileSize) + float64(TileSize)/2,
	}
}

// TileAtPixel is the tile under a world point (unit centre). Inactive maps are grass.
func (m *Map) TileAtPixel(x, y float64) Tile {
	if !m.active() {
		return TileGrass
	}
	tx := int(math.Floor(x / float64(TileSize)))
	ty := int(math.Floor(y / float64(TileSize)))
	return m.At(tx, ty)
}

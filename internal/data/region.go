package data

import (
	"fmt"

	"fannon-codder/internal/sim"
)

// Region is one freeform patch on a phase map. Points are world pixels,
// an implicit closed ring. Later regions paint over earlier ones.
// The rest of the map is the phase terrain's open ground.
type Region struct {
	Type   string      `json:"type"`
	Points [][]float64 `json:"points"`
}

// RegionTile maps a region type onto the sim material a tile centre keeps
// when it falls inside that region. "water" is shallow.
func RegionTile(kind string) (sim.Tile, error) {
	switch kind {
	case "tree":
		return sim.TileTree, nil
	case "shallow", "water":
		return sim.TileWaterShallow, nil
	case "deep":
		return sim.TileWaterDeep, nil
	case "ice":
		return sim.TileIce, nil
	case "quicksand":
		return sim.TileQuicksand, nil
	case "cliff":
		return sim.TileCliff, nil
	case "ramp":
		return sim.TileRamp, nil
	case "bridge":
		return sim.TileBridge, nil
	case "mine":
		return sim.TileMine, nil
	default:
		return 0, fmt.Errorf("unknown region %q", kind)
	}
}

func (tm Tilemap) regionsToSim() (sim.Map, error) {
	if tm.W <= 0 || tm.H <= 0 {
		return sim.Map{}, fmt.Errorf("invalid map size %dx%d", tm.W, tm.H)
	}
	m := sim.Map{W: tm.W, H: tm.H, Tiles: make([]sim.Tile, tm.W*tm.H)}
	for i, r := range tm.Regions {
		tile, err := RegionTile(r.Type)
		if err != nil {
			return sim.Map{}, fmt.Errorf("region %d: %w", i, err)
		}
		pts, err := regionPoints(r.Points)
		if err != nil {
			return sim.Map{}, fmt.Errorf("region %d: %w", i, err)
		}
		for ty := 0; ty < tm.H; ty++ {
			for tx := 0; tx < tm.W; tx++ {
				c := sim.TileCenter(tx, ty)
				if pointInPolygon(c.X, c.Y, pts) {
					m.Tiles[ty*tm.W+tx] = tile
				}
			}
		}
	}
	return m, nil
}

func regionPoints(raw [][]float64) ([][2]float64, error) {
	if len(raw) < 3 {
		return nil, fmt.Errorf("region needs at least 3 points")
	}
	pts := make([][2]float64, len(raw))
	for i, p := range raw {
		if len(p) != 2 {
			return nil, fmt.Errorf("point %d", i)
		}
		pts[i] = [2]float64{p[0], p[1]}
	}
	return pts, nil
}

// pointInPolygon is the even-odd test. The terrain painter uses the same rule.
func pointInPolygon(x, y float64, p [][2]float64) bool {
	in := false
	j := len(p) - 1
	for i := range p {
		yi, yj := p[i][1], p[j][1]
		if (yi > y) != (yj > y) {
			xin := (p[j][0]-p[i][0])*(y-yi)/(yj-yi) + p[i][0]
			if x < xin {
				in = !in
			}
		}
		j = i
	}
	return in
}

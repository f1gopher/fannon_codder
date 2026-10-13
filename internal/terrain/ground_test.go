package terrain

import "testing"

import "fannon-codder/internal/sim"

func TestGroundAtMatchesTheSurface(t *testing.T) {
	tiles := []sim.Tile{
		sim.TileGrass, sim.TileWaterShallow, sim.TileGrass,
		sim.TileGrass, sim.TileGrass, sim.TileQuicksand,
		sim.TileIce, sim.TileWaterDeep, sim.TileGrass,
	}
	m := sim.Map{W: 3, H: 3, Tiles: tiles}
	mid := func(tx, ty int) (float64, float64) {
		return float64(tx*sim.TileSize + 8), float64(ty*sim.TileSize + 8)
	}
	x, y := mid(0, 0)
	if g := GroundAt(m, "jungle", x, y); g != GroundSoil {
		t.Fatalf("jungle grass = %v", g)
	}
	x, y = mid(1, 0)
	if g := GroundAt(m, "jungle", x, y); g != GroundWater {
		t.Fatalf("shallow = %v", g)
	}
	x, y = mid(2, 0)
	if g := GroundAt(m, "desert", x, y); g != GroundSand {
		t.Fatalf("desert = %v", g)
	}
	x, y = mid(2, 1)
	if g := GroundAt(m, "jungle", x, y); g != GroundSand {
		t.Fatalf("quicksand = %v", g)
	}
	x, y = mid(0, 2)
	if g := GroundAt(m, "jungle", x, y); g != GroundSnow {
		t.Fatalf("ice = %v", g)
	}
	x, y = mid(2, 2)
	if g := GroundAt(m, "snow", x, y); g != GroundSnow {
		t.Fatalf("snow = %v", g)
	}
	// A couple of pixels inland of shallow water is beach sand.
	if g := GroundAt(m, "jungle", float64(sim.TileSize)-2, 8); g != GroundSand {
		t.Fatalf("beach = %v", g)
	}
	if g := GroundAt(m, "jungle", 4, 8); g != GroundSoil {
		t.Fatalf("inland of the beach = %v", g)
	}
}

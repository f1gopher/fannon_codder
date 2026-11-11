package data

import (
	"testing"

	"fannon-codder/internal/sim"
)

func TestCampaignPhasesAreRegions(t *testing.T) {
	c, err := LoadCampaign()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range c.Phases {
		p, err := LoadPhase(name)
		if err != nil {
			t.Fatal(err)
		}
		if !p.Map.Freeform() || len(p.Map.Tiles) != 0 || len(p.Map.Regions) == 0 {
			t.Fatalf("%s tiles=%d regions=%d", name, len(p.Map.Tiles), len(p.Map.Regions))
		}
		if name == "m01p01.json" && len(p.Map.Regions) != 26 {
			t.Fatalf("m01 regions=%d", len(p.Map.Regions))
		}
	}
}

func TestMission1HasTheJungleRoute(t *testing.T) {
	p, err := LoadPhase("m01p01.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.Terrain != "jungle" || p.Map.W != 64 || p.Map.H != 40 || len(p.Map.Tiles) != 0 {
		t.Fatalf("map %s %dx%d tiles=%d", p.Terrain, p.Map.W, p.Map.H, len(p.Map.Tiles))
	}
	m, err := p.Grid()
	if err != nil {
		t.Fatal(err)
	}
	if m.At(30, 37) != sim.TileGrass || m.At(30, 31) != sim.TileBridge {
		t.Fatalf("start %v bridge %v", m.At(30, 37), m.At(30, 31))
	}
	if m.At(10, 31) != sim.TileWaterDeep || m.At(27, 31) != sim.TileWaterShallow {
		t.Fatal("the stream should be deep water with a shallow lip by the bridge")
	}
	if m.At(3, 7) != sim.TileQuicksand || m.At(54, 5) != sim.TileMine {
		t.Fatal("quicksand or mines missing")
	}
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			switch m.At(tx, ty) {
			case sim.TileIce, sim.TileCliff, sim.TileRamp:
				t.Fatalf("tile %d,%d is %v", tx, ty, m.At(tx, ty))
			}
		}
	}
}

func TestRegionTileNames(t *testing.T) {
	if _, err := RegionTile("pond"); err == nil {
		t.Fatal("pond is not a region type")
	}
	got, err := RegionTile("water")
	if err != nil || got != sim.TileWaterShallow {
		t.Fatalf("water = %v %v", got, err)
	}
}

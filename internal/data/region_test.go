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
		if name == "m01p01.json" && len(p.Map.Regions) != 4 {
			t.Fatalf("m01 regions=%d", len(p.Map.Regions))
		}
	}
}

func TestMission1TreeRegionsCoverTheOldStand(t *testing.T) {
	p, err := LoadPhase("m01p01.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := p.Grid()
	if err != nil {
		t.Fatal(err)
	}
	trees := map[[2]int]bool{
		{12, 1}: true, {13, 1}: true, {14, 1}: true,
		{12, 2}: true, {13, 2}: true, {14, 2}: true, {15, 2}: true,
		{6, 4}: true, {7, 4}: true, {8, 4}: true,
		{6, 5}: true, {7, 5}: true, {8, 5}: true, {9, 5}: true,
		{6, 6}: true, {7, 6}: true,
		{11, 7}: true, {12, 7}: true, {13, 7}: true,
		{10, 8}: true, {11, 8}: true, {13, 8}: true, {14, 8}: true,
		{11, 11}: true, {12, 11}: true, {13, 11}: true,
	}
	if len(trees) != 26 {
		t.Fatalf("stand list %d", len(trees))
	}
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			got := m.At(tx, ty)
			if trees[[2]int{tx, ty}] {
				if got != sim.TileTree {
					t.Fatalf("tile %d,%d = %v, want tree", tx, ty, got)
				}
				continue
			}
			if got != sim.TileGrass {
				t.Fatalf("tile %d,%d = %v, want open ground", tx, ty, got)
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

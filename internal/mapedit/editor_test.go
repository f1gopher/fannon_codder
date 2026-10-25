package mapedit

import (
	"fmt"
	"testing"

	"fannon-codder/internal/sim"
)

func TestOpensTheFirstPhase(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(ed.phases) != 11 {
		t.Fatalf("phases = %d", len(ed.phases))
	}
	if ed.phases[0].name != "m01p01" || ed.phases[0].title != "It's a Jungle Out There" {
		t.Fatalf("first = %q %q", ed.phases[0].name, ed.phases[0].title)
	}
	if ed.err != nil {
		t.Fatal(ed.err)
	}
	if ed.grid.W != 24 || ed.grid.H != 16 {
		t.Fatalf("grid = %dx%d", ed.grid.W, ed.grid.H)
	}
	if ed.biome != "jungle" {
		t.Fatalf("terrain = %q", ed.biome)
	}
	terrain, size := ed.infoLines()
	if terrain != "jungle" || size != "384 × 256" {
		t.Fatalf("info = %q %q", terrain, size)
	}
	if !ed.freeform || len(ed.chart.Regions) != 4 {
		t.Fatalf("regions = %v %d", ed.freeform, len(ed.chart.Regions))
	}
	if ed.gridView || ed.viewName() != "terrain" {
		t.Fatalf("view = %v %q", ed.gridView, ed.viewName())
	}
	ed.gridView = true
	if ed.viewName() != "regions" {
		t.Fatalf("view = %q", ed.viewName())
	}
	ed.load(3)
	if !ed.freeform {
		t.Fatal("m03p01 should be regions")
	}
	if ed.phases[3].name != "m03p01" || ed.biome != "snow" {
		t.Fatalf("phase 4 = %s %q", ed.phases[3].name, ed.biome)
	}
	terrain, size = ed.infoLines()
	want := fmt.Sprintf("%d × %d", ed.grid.W*sim.TileSize, ed.grid.H*sim.TileSize)
	if terrain != "snow" || size != want {
		t.Fatalf("info = %q %q", terrain, size)
	}
	if ed.scrollX != 0 || ed.scrollY != 0 {
		t.Fatalf("scroll = %v,%v", ed.scrollX, ed.scrollY)
	}
	if !ed.gridView || ed.viewName() != "regions" {
		t.Fatalf("view = %v %q", ed.gridView, ed.viewName())
	}
}

func TestShowsHutsCratesAndMines(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(ed.buildings) != 0 || len(ed.pickups) != 1 || ed.pickups[0].Kind != "grenades" {
		t.Fatalf("m01 props buildings=%d pickups=%+v", len(ed.buildings), ed.pickups)
	}
	found := false
	for i, row := range ed.phases {
		if row.name != "m04p04" {
			continue
		}
		found = true
		ed.load(i)
	}
	if !found || ed.err != nil {
		t.Fatal(ed.err)
	}
	mines := 0
	for ty := 0; ty < ed.grid.H; ty++ {
		for tx := 0; tx < ed.grid.W; tx++ {
			if ed.grid.At(tx, ty) == sim.TileMine {
				mines++
			}
		}
	}
	if mines == 0 || len(ed.buildings) == 0 || len(ed.pickups) == 0 {
		t.Fatalf("m04p04 mines=%d buildings=%d pickups=%d", mines, len(ed.buildings), len(ed.pickups))
	}
}

func TestMarksStartAndEnemies(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if !ed.hasStart || ed.start != [2]int{12, 13} {
		t.Fatalf("start = %v %v", ed.hasStart, ed.start)
	}
	if len(ed.enemies) != 3 || ed.enemies[0].X != 19 || ed.enemies[0].Y != 3 {
		t.Fatalf("enemies = %+v", ed.enemies)
	}
	ed.scale = 2
	ed.scrollX, ed.scrollY = 16, 0
	x, y := ed.markAt(12, 13)
	// Tile 12 centre is 12.5*16 world px, then scale 2, then scroll 16.
	if x != (12.5*16-16)*2 || y != 13.5*16*2 {
		t.Fatalf("mark = %v,%v", x, y)
	}
}

func TestTileFillsDiffer(t *testing.T) {
	tiles := []sim.Tile{
		sim.TileGrass, sim.TileTree, sim.TileWaterShallow, sim.TileWaterDeep,
		sim.TileIce, sim.TileQuicksand, sim.TileCliff, sim.TileBridge,
		sim.TileMine, sim.TileRamp,
	}
	seen := map[string]sim.Tile{}
	for _, tile := range tiles {
		c := tileFill(tile)
		key := fmt.Sprintf("%d,%d,%d", c.R, c.G, c.B)
		if prev, ok := seen[key]; ok {
			t.Fatalf("%v and %v share %s", prev, tile, key)
		}
		seen[key] = tile
	}
}

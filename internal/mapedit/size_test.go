package mapedit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"fannon-codder/internal/data"
	"fannon-codder/internal/sim"
)

func TestResizeRegionMap(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ed.enemies = append(ed.enemies, data.Enemy{X: 50, Y: 30, Kind: "grunt", Line: [][2]int{{2, 2}, {60, 30}}})
	ed.buildings = append(ed.buildings, data.BuildingSpec{X: 22, Y: 14, W: 4, H: 3})
	ed.start = [2]int{60, 38}
	ed.hasStart = true
	ed.startType(typeSize)
	if ed.draft != "1024 × 640" {
		t.Fatalf("draft = %q", ed.draft)
	}
	ed.draft = "1000 × 640"
	ed.commitType()
	if !ed.typing() || ed.phase.Map.W != 64 {
		t.Fatalf("rejected size should stay open, w=%d typing=%v note=%q", ed.phase.Map.W, ed.typing(), ed.typeNote)
	}
	ed.draft = "384 × 256"
	ed.commitType()
	if ed.typing() || ed.dirty == false {
		t.Fatalf("typing=%v dirty=%v note=%q", ed.typing(), ed.dirty, ed.typeNote)
	}
	if ed.grid.W != 24 || ed.grid.H != 16 || ed.chart.W != 384 || ed.chart.H != 256 {
		t.Fatalf("grid %dx%d chart %dx%d", ed.grid.W, ed.grid.H, ed.chart.W, ed.chart.H)
	}
	if ed.phase.Map.W != 24 || ed.phase.Map.H != 16 || !ed.phase.Map.Freeform() {
		t.Fatalf("map = %+v freeform=%v", ed.phase.Map.W, ed.phase.Map.Freeform())
	}
	_, size := ed.infoLines()
	if size != "384 × 256" {
		t.Fatalf("size = %q", size)
	}
	if ed.start != [2]int{23, 15} {
		t.Fatalf("start = %v", ed.start)
	}
	for _, en := range ed.enemies {
		if en.X >= 24 || en.Y >= 16 {
			t.Fatalf("enemy outside %+v", en)
		}
		for _, p := range en.Line {
			if p[0] >= 24 || p[1] >= 16 {
				t.Fatalf("route point outside %v", p)
			}
		}
	}
	for _, b := range ed.buildings {
		if b.X+b.W > 24 || b.Y+b.H > 16 {
			t.Fatalf("building outside %+v", b)
		}
	}
	path := filepath.Join(t.TempDir(), "m01p01.json")
	if err := ed.writePhase(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var p data.Phase
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Map.W != 24 || p.Map.H != 16 || len(p.Map.Tiles) != 0 {
		t.Fatalf("saved map %dx%d tiles=%d", p.Map.W, p.Map.H, len(p.Map.Tiles))
	}
}

func TestResizeTileMapKeepsTheOverlap(t *testing.T) {
	rows := make([]string, 16)
	for y := range rows {
		rows[y] = "........................"
	}
	rows[1] = ".T......................"
	ed := &Editor{
		phase: &data.Phase{Map: data.Tilemap{W: 24, H: 16, Tiles: rows}},
		grid:  mustGrid(t, rows),
	}
	if !ed.applySize("26 × 18") {
		t.Fatal(ed.typeNote)
	}
	if ed.grid.W != 26 || ed.grid.H != 18 || ed.grid.At(1, 1) != sim.TileTree || ed.grid.At(25, 17) != sim.TileGrass {
		t.Fatalf("grown %dx%d tree=%v corner=%v", ed.grid.W, ed.grid.H, ed.grid.At(1, 1), ed.grid.At(25, 17))
	}
	if len(ed.phase.Map.Tiles) != 18 || len(ed.phase.Map.Tiles[1]) != 26 || ed.phase.Map.Tiles[1][1] != 'T' {
		t.Fatalf("rows = %d %q", len(ed.phase.Map.Tiles), ed.phase.Map.Tiles[1])
	}
	if ed.applySize("10 × 10") {
		t.Fatal("accepted a map below 24×16")
	}
	if ed.grid.W != 26 {
		t.Fatalf("grid changed to %d", ed.grid.W)
	}
}

func mustGrid(t *testing.T, rows []string) sim.Map {
	t.Helper()
	g, err := (&data.Phase{Map: data.Tilemap{W: len(rows[0]), H: len(rows), Tiles: rows}}).Grid()
	if err != nil {
		t.Fatal(err)
	}
	return g
}

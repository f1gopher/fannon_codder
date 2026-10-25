package mapedit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"fannon-codder/internal/data"
	"fannon-codder/internal/sim"
)

func TestStampTipsNameTheTool(t *testing.T) {
	if stampTip(stampPatrol) == "" || stampTip(stampDoor) == stampTip(stampHut) {
		t.Fatalf("tips %q %q", stampTip(stampDoor), stampTip(stampHut))
	}
	for _, s := range barStamps {
		if stampTip(s) == "" {
			t.Fatalf("no tip for %s", s)
		}
	}
}

func TestToolHitPicksTheMenuIcon(t *testing.T) {
	size, gap, x0, y0 := toolButtons(3)
	i, ok := ToolHit(x0+size/2, y0+size/2, 3)
	if !ok || i != 0 || barStamps[i] != stampGrunt {
		t.Fatalf("first = %d %v", i, ok)
	}
	x := x0 + 4*(size+gap) + 2
	i, ok = ToolHit(x, y0+4, 3)
	if !ok || barStamps[i] != stampBarrel {
		t.Fatalf("barrel = %d %v", i, ok)
	}
	if _, ok := ToolHit(x0+size+1, y0+4, 3); ok {
		t.Fatal("gap between icons hit a tool")
	}
	if _, ok := ToolHit(x0, y0+size+2, 3); ok {
		t.Fatal("point under the bar hit a tool")
	}
}

func TestDragMovesStartEnemyBarrelHutAndCrate(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	start := sim.TileCenter(ed.start[0], ed.start[1])
	if got := ed.pickAt(start.X, start.Y); got.kind != pickStart {
		t.Fatalf("pick start = %+v", got)
	}
	ed.beginGrab(pick{kind: pickStart}, start.X, start.Y)
	ed.dragGrab(start.X+sim.TileSize, start.Y)
	if ed.start != [2]int{13, 13} {
		t.Fatalf("start = %v", ed.start)
	}

	grunt := ed.enemies[0]
	c := sim.TileCenter(grunt.X, grunt.Y)
	ed.beginGrab(ed.pickAt(c.X, c.Y), c.X, c.Y)
	ed.dragGrab(c.X, c.Y+sim.TileSize)
	if ed.enemies[0].X != grunt.X || ed.enemies[0].Y != grunt.Y+1 {
		t.Fatalf("grunt = %+v", ed.enemies[0])
	}

	drum := ed.barrels[0]
	c = sim.TileCenter(drum.X, drum.Y)
	ed.beginGrab(ed.pickAt(c.X, c.Y), c.X, c.Y)
	ed.dragGrab(c.X-sim.TileSize, c.Y)
	if ed.barrels[0].X != drum.X-1 || ed.barrels[0].Y != drum.Y {
		t.Fatalf("barrel = %+v", ed.barrels[0])
	}

	box := ed.pickups[0]
	c = sim.TileCenter(box.X, box.Y)
	ed.beginGrab(ed.pickAt(c.X, c.Y), c.X, c.Y)
	ed.dragGrab(c.X, c.Y-sim.TileSize)
	if ed.pickups[0].X != box.X || ed.pickups[0].Y != box.Y-1 || ed.pickups[0].Kind != "grenades" {
		t.Fatalf("crate = %+v", ed.pickups[0])
	}

	for i, row := range ed.phases {
		if row.name == "m04p04" {
			ed.load(i)
		}
	}
	if len(ed.buildings) == 0 {
		t.Fatal("no hut")
	}
	hut := ed.buildings[0]
	// Footprint is the hut body, above its centre mark.
	ed.beginGrab(ed.pickAt(float64(hut.X*sim.TileSize+4), float64(hut.Y*sim.TileSize+4)), float64(hut.X*sim.TileSize+4), float64(hut.Y*sim.TileSize+4))
	if ed.grab.kind != pickHut {
		t.Fatalf("pick hut = %+v", ed.grab)
	}
	ed.dragGrab(float64((hut.X+1)*sim.TileSize+4), float64(hut.Y*sim.TileSize+4))
	if ed.buildings[0].X != hut.X+1 || ed.buildings[0].Y != hut.Y || ed.buildings[0].W != hut.W || ed.buildings[0].Door != hut.Door {
		t.Fatalf("hut = %+v", ed.buildings[0])
	}
	if !ed.dirty {
		t.Fatal("move should be dirty")
	}
}

func TestPlaceAddsEnemiesHutsBarrelsAndCrates(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	n := len(ed.enemies)
	ed.stamp = stampGrenadier
	ed.placeAt(2*sim.TileSize, 3*sim.TileSize)
	ed.stamp = stampRocketeer
	ed.placeAt(4*sim.TileSize, 3*sim.TileSize)
	ed.stamp = stampPatrol
	ed.placeAt(6*sim.TileSize, 3*sim.TileSize)
	ed.addPatrolPoint(8*sim.TileSize, 4*sim.TileSize)
	ed.stamp = stampBarrel
	ed.placeAt(9*sim.TileSize, 9*sim.TileSize)
	ed.stamp = stampDoor
	ed.placeAt(10*sim.TileSize, 10*sim.TileSize)
	ed.stamp = stampRockets
	ed.placeAt(11*sim.TileSize, 11*sim.TileSize)
	if len(ed.enemies) != n+3 {
		t.Fatalf("enemies = %d", len(ed.enemies))
	}
	if ed.enemies[n].Kind != "grenadier" || ed.enemies[n+1].Kind != "rocketeer" || ed.enemies[n+2].Kind != "patrol" {
		t.Fatalf("added = %+v", ed.enemies[n:])
	}
	if len(ed.enemies[n+2].Line) != 1 || ed.enemies[n+2].Line[0] != [2]int{8, 4} {
		t.Fatalf("line = %+v", ed.enemies[n+2].Line)
	}
	if ed.barrels[len(ed.barrels)-1] != (data.BarrelSpec{X: 9, Y: 9}) {
		t.Fatalf("barrel = %+v", ed.barrels[len(ed.barrels)-1])
	}
	hut := ed.buildings[len(ed.buildings)-1]
	if !hut.Door || hut.W != 2 || hut.H != 2 || hut.Spawn != 2.5 {
		t.Fatalf("hut = %+v", hut)
	}
	if ed.pickups[len(ed.pickups)-1].Kind != "rockets" {
		t.Fatalf("crate = %+v", ed.pickups[len(ed.pickups)-1])
	}
	ed.removeSelected()
	if ed.pickups[len(ed.pickups)-1].Kind == "rockets" {
		t.Fatal("rocket crate still there")
	}
}

func TestSaveWritesPlacementsAndKeepsTheMap(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ed.start = [2]int{4, 5}
	ed.dirty = true
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
	if p.PlayerStart != [2]int{4, 5} || p.Title != "It's a Jungle Out There" || len(p.Map.Regions) != 4 {
		t.Fatalf("saved start=%v title=%q regions=%d", p.PlayerStart, p.Title, len(p.Map.Regions))
	}
	if len(p.Enemies) != 4 || len(p.Barrels) != 3 || len(p.Pickups) != 1 {
		t.Fatalf("lists enemies=%d barrels=%d pickups=%d", len(p.Enemies), len(p.Barrels), len(p.Pickups))
	}
	if !ed.dirty {
		t.Fatal("writePhase should leave the session dirty")
	}
}

func TestEditsStayWhenThePhaseChanges(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ed.start = [2]int{2, 2}
	ed.dirty = true
	ed.load(1)
	if ed.start == [2]int{2, 2} {
		t.Fatal("phase 2 kept mission 1's start")
	}
	ed.load(0)
	if ed.start != [2]int{2, 2} || !ed.dirty {
		t.Fatalf("restored start=%v dirty=%v", ed.start, ed.dirty)
	}
}

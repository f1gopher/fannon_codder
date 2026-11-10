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
		t.Fatal("point under the icon hit a tool")
	}
	rowH := barRowHeight(3)
	i, ok = ToolHit(x0+size/2, y0+rowH+size/2, 3)
	if !ok || allStamps()[i] != stampCivilian {
		t.Fatalf("second row = %d %v", i, ok)
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

func TestCivilianPlaceDragDeleteAndSave(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range ed.phases {
		if row.name == "m04p03" {
			ed.load(i)
		}
	}
	if len(ed.civilians) != 3 {
		t.Fatalf("loaded civilians = %d", len(ed.civilians))
	}
	c := sim.TileCenter(ed.civilians[0].X, ed.civilians[0].Y)
	ed.beginGrab(ed.pickAt(c.X, c.Y), c.X, c.Y)
	if ed.grab.kind != pickCivilian {
		t.Fatalf("pick = %+v", ed.grab)
	}
	ed.dragGrab(c.X+sim.TileSize, c.Y)
	if ed.civilians[0].X != 13 || ed.civilians[0].Y != 9 {
		t.Fatalf("moved = %+v", ed.civilians[0])
	}
	ed.stamp = stampCivilian
	ed.placeAt(2*sim.TileSize, 2*sim.TileSize)
	if len(ed.civilians) != 4 || ed.sel.kind != pickCivilian {
		t.Fatalf("placed %+v n=%d", ed.sel, len(ed.civilians))
	}
	ed.removeSelected()
	if len(ed.civilians) != 3 {
		t.Fatalf("after delete = %d", len(ed.civilians))
	}
	ed.dirty = true
	path := filepath.Join(t.TempDir(), "m04p03.json")
	if err := ed.writePhase(path); err != nil {
		t.Fatal(err)
	}
	p, err := data.LoadPhase(path)
	if err != nil {
		// LoadPhase looks beside the campaign, so read the temp file.
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Fatal(err)
		}
		p = &data.Phase{}
		if jerr := json.Unmarshal(b, p); jerr != nil {
			t.Fatal(jerr)
		}
	}
	if len(p.Civilians) != 3 || p.Civilians[0].X != 13 || p.Civilians[0].Y != 9 {
		t.Fatalf("saved civilians = %+v", p.Civilians)
	}
	if p.Title == "" || len(p.Map.Regions) == 0 {
		t.Fatal("save dropped the rest of the phase")
	}
}

func TestHostageAndTentRoundTrip(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ed.stamp = stampHostage
	ed.placeAt(3*sim.TileSize, 4*sim.TileSize)
	if len(ed.hostages) != 1 || ed.sel.kind != pickHostage {
		t.Fatalf("placed %+v n=%d", ed.sel, len(ed.hostages))
	}
	c := sim.TileCenter(ed.hostages[0].X, ed.hostages[0].Y)
	ed.beginGrab(ed.pickAt(c.X, c.Y), c.X, c.Y)
	ed.dragGrab(c.X+sim.TileSize, c.Y)
	if ed.hostages[0].X != 4 || ed.hostages[0].Y != 4 {
		t.Fatalf("moved = %+v", ed.hostages[0])
	}
	ed.stamp = stampTent
	ed.placeAt(8*sim.TileSize, 4*sim.TileSize)
	if len(ed.buildings) != 1 || ed.buildings[0].Kind != "tent" {
		t.Fatalf("tent = %+v", ed.buildings)
	}
	kept := len(ed.buildings)
	ed.placeAt(10*sim.TileSize, 6*sim.TileSize)
	if len(ed.buildings) != kept+1 || ed.buildings[kept].Kind != "tent" {
		t.Fatal("second tent")
	}
	ed.sel = pick{kind: pickHut, i: 0}
	nHost := len(ed.hostages)
	ed.removeSelected()
	if len(ed.buildings) != 1 || ed.buildings[0].Kind != "tent" || len(ed.hostages) != nHost {
		t.Fatal("deleting the tent removed the hostage or the other tent")
	}
	ed.dirty = true
	path := filepath.Join(t.TempDir(), "phase.json")
	if err := ed.writePhase(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := &data.Phase{}
	if err := json.Unmarshal(b, p); err != nil {
		t.Fatal(err)
	}
	if len(p.Hostages) != 1 || p.Hostages[0].X != 4 || p.Hostages[0].Y != 4 {
		t.Fatalf("saved hostages = %+v", p.Hostages)
	}
	if len(p.Buildings) != 1 || p.Buildings[0].Kind != "tent" {
		t.Fatalf("saved tent = %+v", p.Buildings)
	}
}

func TestHomePlaceDragDeleteAndSave(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ed.stamp = stampCivilian
	ed.placeAt(3*sim.TileSize, 4*sim.TileSize)
	ed.flipHome()
	if !ed.civilians[len(ed.civilians)-1].Home {
		t.Fatal("H should mark the civilian")
	}
	ed.flipHome()
	if ed.civilians[len(ed.civilians)-1].Home {
		t.Fatal("H should clear the flag")
	}
	ed.flipHome()
	c := sim.TileCenter(ed.civilians[len(ed.civilians)-1].X, ed.civilians[len(ed.civilians)-1].Y)
	ed.beginGrab(ed.pickAt(c.X, c.Y), c.X, c.Y)
	ed.dragGrab(c.X+sim.TileSize, c.Y)
	last := len(ed.civilians) - 1
	if ed.civilians[last].X != 4 || !ed.civilians[last].Home {
		t.Fatalf("moved = %+v", ed.civilians[last])
	}
	ed.stamp = stampHome
	ed.placeAt(8*sim.TileSize, 4*sim.TileSize)
	if ed.buildings[len(ed.buildings)-1].Kind != "home" {
		t.Fatalf("home = %+v", ed.buildings)
	}
	bi := len(ed.buildings) - 1
	ed.sel = pick{kind: pickHut, i: bi}
	ed.removeSelected()
	for _, b := range ed.buildings {
		if b.Kind == "home" {
			t.Fatal("home stayed after delete")
		}
	}
	if !ed.civilians[last].Home {
		t.Fatal("deleting the home cleared the civilian")
	}
	ed.stamp = stampHome
	ed.placeAt(8*sim.TileSize, 4*sim.TileSize)
	ed.dirty = true
	path := filepath.Join(t.TempDir(), "phase.json")
	if err := ed.writePhase(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := &data.Phase{}
	if err := json.Unmarshal(b, p); err != nil {
		t.Fatal(err)
	}
	foundHome := false
	for _, spec := range p.Buildings {
		if spec.Kind == "home" {
			foundHome = true
		}
	}
	foundFlag := false
	for _, civ := range p.Civilians {
		if civ.X == 4 && civ.Y == 4 && civ.Home {
			foundFlag = true
		}
	}
	if !foundHome || !foundFlag {
		t.Fatalf("saved buildings=%+v civilians=%+v", p.Buildings, p.Civilians)
	}
}

func TestLeaderAndHandoverRoundTrip(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	before := len(ed.enemies)
	ed.stamp = stampLeader
	ed.placeAt(3*sim.TileSize, 4*sim.TileSize)
	i := len(ed.enemies) - 1
	if i < before || ed.enemies[i].Kind != "leader" || ed.sel.kind != pickEnemy {
		t.Fatalf("placed %+v enemies=%+v", ed.sel, ed.enemies)
	}
	c := sim.TileCenter(ed.enemies[i].X, ed.enemies[i].Y)
	ed.beginGrab(ed.pickAt(c.X, c.Y), c.X, c.Y)
	ed.dragGrab(c.X+sim.TileSize, c.Y)
	if ed.enemies[i].X != 4 || ed.enemies[i].Y != 4 {
		t.Fatalf("moved = %+v", ed.enemies[i])
	}
	ed.stamp = stampHandover
	ed.placeAt(8*sim.TileSize, 4*sim.TileSize)
	if len(ed.buildings) != 1 || ed.buildings[0].Kind != "handover" {
		t.Fatalf("handover = %+v", ed.buildings)
	}
	ed.sel = pick{kind: pickEnemy, i: i}
	ed.removeSelected()
	if len(ed.enemies) != before || len(ed.buildings) != 1 || ed.buildings[0].Kind != "handover" {
		t.Fatal("deleting the leader removed the handover")
	}
	ed.stamp = stampLeader
	ed.placeAt(2*sim.TileSize, 2*sim.TileSize)
	ed.dirty = true
	path := filepath.Join(t.TempDir(), "phase.json")
	if err := ed.writePhase(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := &data.Phase{}
	if err := json.Unmarshal(b, p); err != nil {
		t.Fatal(err)
	}
	var leaders int
	for _, en := range p.Enemies {
		if en.Kind == "leader" && en.X == 2 && en.Y == 2 {
			leaders++
		}
	}
	if leaders != 1 {
		t.Fatalf("saved leader = %+v", p.Enemies)
	}
	if len(p.Buildings) != 1 || p.Buildings[0].Kind != "handover" {
		t.Fatalf("saved handover = %+v", p.Buildings)
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

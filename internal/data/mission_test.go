package data

import (
	"math"
	"testing"

	"fannon-codder/internal/sim"
)

func TestLoadMission1(t *testing.T) {
	p, err := LoadPhase("m01p01.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.Deploy != 2 {
		t.Fatalf("deploy=%d, want 2", p.Deploy)
	}
	if len(p.Enemies) != 3 {
		t.Fatalf("enemies=%d, want 3", len(p.Enemies))
	}
	w, err := p.World()
	if err != nil {
		t.Fatal(err)
	}
	sx, sy := sim.TileCenter(10, 13).X, sim.TileCenter(10, 13).Y
	if !w.Map.Walkable(sx, sy, 4) {
		t.Fatal("player start should be walkable")
	}
	tx, ty := sim.TileCenter(11, 1).X, sim.TileCenter(11, 1).Y
	if w.Map.Walkable(tx, ty, 4) {
		t.Fatal("tree tile should block")
	}
	nP, nE := 0, 0
	for i := range w.Units {
		if !w.Units[i].Living() {
			continue
		}
		switch w.Units[i].Side {
		case sim.SidePlayer:
			nP++
		case sim.SideEnemy:
			nE++
		}
	}
	if nP != 2 || nE != 3 {
		t.Fatalf("units player=%d enemy=%d, want 2 and 3", nP, nE)
	}
}

func TestPhaseLoadsHutAndCrate(t *testing.T) {
	rows := make([]string, 8)
	for i := range rows {
		rows[i] = "........"
	}
	p := &Phase{
		Deploy:      1,
		PlayerStart: [2]int{1, 6},
		Map:         Tilemap{W: 8, H: 8, Tiles: rows},
		Buildings:   []BuildingSpec{{X: 4, Y: 1, Door: true, Spawn: 3}},
		Pickups:     []PickupSpec{{X: 2, Y: 5, Kind: "grenades"}},
		Objectives:  []string{"kill_all_enemy", "destroy_enemy_buildings"},
	}
	w, err := p.World()
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Buildings) != 1 || !w.Buildings[0].HasDoor {
		t.Fatal("expected a door hut")
	}
	if w.Buildings[0].SpawnEvery != 3 {
		t.Fatalf("spawn=%v", w.Buildings[0].SpawnEvery)
	}
	if len(w.Pickups) != 1 || w.Pickups[0].Amount != sim.CrateAmount {
		t.Fatal("expected a 4-grenade crate")
	}
	if len(w.Objectives) != 2 {
		t.Fatalf("objectives=%d", len(w.Objectives))
	}
}

func TestCampaignRunsMission2AfterMission1(t *testing.T) {
	c, err := LoadCampaign()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Phases) < 3 || c.Phases[0] != "m01p01.json" || c.Phases[1] != "m02p01.json" || c.Phases[2] != "m02p02.json" {
		t.Fatalf("phases=%v", c.Phases)
	}
}

func TestMission2BridgePhase(t *testing.T) {
	p := mustPhase(t, "m02p01.json")
	if p.Mission != 2 || p.Phase != 1 || p.Deploy != 3 {
		t.Fatalf("mission/phase/deploy = %d/%d/%d", p.Mission, p.Phase, p.Deploy)
	}
	if len(p.Enemies) < 15 || len(p.Enemies) > 18 {
		t.Fatalf("enemies=%d, want 15–18", len(p.Enemies))
	}
	if len(p.Objectives) != 1 || p.Objectives[0] != "kill_all_enemy" {
		t.Fatalf("objectives=%v", p.Objectives)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	bridges, deepEnemy := 0, 0
	minBX, maxBX := 999, -1
	for y := 0; y < w.Map.H; y++ {
		for x := 0; x < w.Map.W; x++ {
			if w.Map.At(x, y) == sim.TileBridge {
				bridges++
				if x < minBX {
					minBX = x
				}
				if x > maxBX {
					maxBX = x
				}
			}
		}
	}
	if bridges == 0 || maxBX-minBX > 1 {
		t.Fatalf("want one narrow bridge, tiles=%d span=%d", bridges, maxBX-minBX)
	}
	for _, e := range p.Enemies {
		tile := w.Map.At(e.X, e.Y)
		if tile == sim.TileTree {
			t.Fatalf("grunt on a tree at %d,%d", e.X, e.Y)
		}
		if tile == sim.TileWaterDeep {
			deepEnemy++
			if tileDist(e.X, e.Y, p.PlayerStart[0], p.PlayerStart[1]) > 160 {
				t.Fatalf("deep grunt at %d,%d is far from spawn", e.X, e.Y)
			}
		}
	}
	if deepEnemy != 1 {
		t.Fatalf("want one swimmer near spawn, got %d", deepEnemy)
	}
	assertSpawnSurvives(t, w)
}

func TestMission2HQPhase(t *testing.T) {
	p := mustPhase(t, "m02p02.json")
	if p.Mission != 2 || p.Phase != 2 || p.Deploy != 3 {
		t.Fatalf("mission/phase/deploy = %d/%d/%d", p.Mission, p.Phase, p.Deploy)
	}
	if len(p.Objectives) != 2 {
		t.Fatalf("objectives=%v", p.Objectives)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	water, land := 0, 0
	for y := 0; y < w.Map.H; y++ {
		for x := 0; x < w.Map.W; x++ {
			switch w.Map.At(x, y) {
			case sim.TileWaterDeep, sim.TileWaterShallow:
				water++
			case sim.TileGrass, sim.TileTree:
				land++
			}
		}
	}
	if water <= land {
		t.Fatalf("phase 2 should be mostly water, water=%d land=%d", water, land)
	}
	if len(w.Buildings) != 1 || !w.Buildings[0].HasDoor {
		t.Fatal("want one door hut")
	}
	if len(w.Pickups) != 1 || w.Pickups[0].Amount != sim.CrateAmount {
		t.Fatal("want one crate of 4 grenades")
	}
	b := w.Buildings[0]
	d := math.Hypot(w.Pickups[0].X-(b.X+b.W/2), w.Pickups[0].Y-(b.Y+b.H/2))
	if d > 80 || d <= sim.GrenadeRadius {
		t.Fatalf("crate should sit near the hut but outside the blast, dist=%v", d)
	}
	if w.Map.TileAtPixel(w.Pickups[0].X, w.Pickups[0].Y) != sim.TileGrass {
		t.Fatal("crate should be on grass so you can walk onto it")
	}
	for _, e := range p.Enemies {
		if w.Map.At(e.X, e.Y) == sim.TileTree {
			t.Fatalf("grunt on a tree at %d,%d", e.X, e.Y)
		}
	}
	assertSpawnSurvives(t, w)
}

func mustPhase(t *testing.T, name string) *Phase {
	t.Helper()
	p, err := LoadPhase(name)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func mustWorld(t *testing.T, p *Phase) *sim.World {
	t.Helper()
	w, err := p.World()
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func assertBigMap(t *testing.T, w *sim.World) {
	t.Helper()
	mw, mh := w.Map.PixelSize()
	if mw <= w.Camera.ViewW || mh <= w.Camera.ViewH {
		t.Fatalf("map %v×%v should scroll past the 320×256 view", mw, mh)
	}
}

func assertStartWalkable(t *testing.T, p *Phase, w *sim.World) {
	t.Helper()
	start := sim.TileCenter(p.PlayerStart[0], p.PlayerStart[1])
	for i := 0; i < p.Deploy; i++ {
		x := start.X - float64(i)*sim.FileSpacing
		if !w.Map.Walkable(x, start.Y, 4) {
			t.Fatalf("deploy slot %d is not walkable", i)
		}
		if tile := w.Map.TileAtPixel(x, start.Y); tile == sim.TileWaterDeep || tile == sim.TileWaterShallow {
			t.Fatalf("deploy slot %d starts in water", i)
		}
	}
}

func assertSpawnSurvives(t *testing.T, w *sim.World) {
	t.Helper()
	w.Spread = 0
	for i := 0; i < 90; i++ {
		w.Step(1.0 / 60)
	}
	if n := livingPlayers(w); n < 2 {
		t.Fatalf("spawn wiped the squad, living=%d", n)
	}
}

func livingPlayers(w *sim.World) int {
	n := 0
	for i := range w.Units {
		if w.Units[i].Side == sim.SidePlayer && w.Units[i].Living() {
			n++
		}
	}
	return n
}

func tileDist(ax, ay, bx, by int) float64 {
	return math.Hypot(float64(ax-bx)*16, float64(ay-by)*16)
}

func TestParseWaterTiles(t *testing.T) {
	cases := []struct {
		ch rune
		t  sim.Tile
	}{
		{'~', sim.TileWaterShallow},
		{'W', sim.TileWaterDeep},
		{'B', sim.TileBridge},
		{'=', sim.TileBridge},
	}
	for _, c := range cases {
		got, err := parseTile(c.ch)
		if err != nil {
			t.Fatalf("parse %q: %v", string(c.ch), err)
		}
		if got != c.t {
			t.Fatalf("parse %q = %v, want %v", string(c.ch), got, c.t)
		}
	}
}

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

func TestMission3IcePhase(t *testing.T) {
	p := mustPhase(t, "m03p01.json")
	if p.Mission != 3 || p.Phase != 1 || p.Deploy != 4 || p.Terrain != "arctic" {
		t.Fatalf("mission/phase/deploy/terrain = %d/%d/%d/%q", p.Mission, p.Phase, p.Deploy, p.Terrain)
	}
	if p.Title != "Blast It's Cold" {
		t.Fatal(p.Title)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	ice, cliffs, ramps := 0, 0, 0
	for y := 0; y < w.Map.H; y++ {
		for x := 0; x < w.Map.W; x++ {
			switch w.Map.At(x, y) {
			case sim.TileIce:
				ice++
			case sim.TileCliff:
				cliffs++
			case sim.TileRamp:
				ramps++
			}
		}
	}
	if ice == 0 || cliffs == 0 || ramps == 0 {
		t.Fatalf("ice=%d cliff=%d ramp=%d", ice, cliffs, ramps)
	}
	doors := 0
	for i := range w.Buildings {
		if w.Buildings[i].HasDoor {
			doors++
		}
	}
	if doors != 4 {
		t.Fatalf("door huts=%d, want 4", doors)
	}
	grenades := 0
	for i := range w.Pickups {
		pk := &w.Pickups[i]
		grenades += pk.Amount
		if w.Map.TileAtPixel(pk.X, pk.Y) == sim.TileCliff {
			t.Fatal("crate on a cliff")
		}
		for bi := range w.Buildings {
			b := &w.Buildings[bi]
			d := math.Hypot(pk.X-(b.X+b.W/2), pk.Y-(b.Y+b.H/2))
			if d <= sim.GrenadeRadius {
				t.Fatalf("crate %d is inside hut %d blast (dist=%v); shooting the hut would waste it", i, bi, d)
			}
		}
	}
	if len(w.Pickups) != 2 || grenades != 8 {
		t.Fatalf("want 2 crates / 8 grenades, got %d pickups / %d grenades", len(w.Pickups), grenades)
	}
	objs := map[sim.Objective]bool{}
	for _, o := range w.Objectives {
		objs[o] = true
	}
	if !objs[sim.KillAllEnemy] || !objs[sim.DestroyEnemyBuildings] {
		t.Fatalf("objectives=%v", w.Objectives)
	}
	assertSpawnSurvives(t, w)
}

func TestPhaseSpawnsCivilian(t *testing.T) {
	rows := []string{
		"........",
		"...Q....",
		"........",
		".....M..",
		"........",
		"........",
		"........",
		"........",
	}
	p := &Phase{
		Deploy:      1,
		PlayerStart: [2]int{1, 6},
		Map:         Tilemap{W: 8, H: 8, Tiles: rows},
		Civilians:   []Enemy{{X: 4, Y: 4}},
		Buildings:   []BuildingSpec{{X: 5, Y: 1, Door: false}},
		Objectives:  []string{"kill_all_enemy", "protect_civilians"},
	}
	w, err := p.World()
	if err != nil {
		t.Fatal(err)
	}
	if w.Map.At(3, 1) != sim.TileQuicksand || w.Map.At(5, 3) != sim.TileMine {
		t.Fatal("Q and M tiles did not load")
	}
	n := 0
	for i := range w.Units {
		if w.Units[i].Side == sim.SideCivilian && w.Units[i].Living() {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("civilians=%d", n)
	}
	if len(w.Buildings) != 1 || w.Buildings[0].HasDoor {
		t.Fatal("expected a doorless hut")
	}
}

func TestStartGrenadesPerTrooper(t *testing.T) {
	rows := make([]string, 8)
	for i := range rows {
		rows[i] = "........"
	}
	base := Phase{
		Deploy:      4,
		PlayerStart: [2]int{4, 4},
		Map:         Tilemap{W: 8, H: 8, Tiles: rows},
	}
	base.StartGrenadesPerTrooper = 2
	w := mustWorld(t, &base)
	if got := w.ActiveSquad().Grenades; got != 8 {
		t.Fatalf("4 troopers × 2 = %d, want 8", got)
	}
	base.Deploy = 5
	base.StartGrenadesPerTrooper = 2
	w = mustWorld(t, &base)
	if got := w.ActiveSquad().Grenades; got != 10 {
		t.Fatalf("5 troopers × 2 = %d, want 10", got)
	}
	base.StartGrenadesPerTrooper = 0
	w = mustWorld(t, &base)
	if got := w.ActiveSquad().Grenades; got != 0 {
		t.Fatalf("no free grenades, got %d", got)
	}
}

func TestUnknownEnemyKind(t *testing.T) {
	rows := make([]string, 4)
	for i := range rows {
		rows[i] = "...."
	}
	p := &Phase{
		Deploy:      1,
		PlayerStart: [2]int{1, 1},
		Map:         Tilemap{W: 4, H: 4, Tiles: rows},
		Enemies:     []Enemy{{X: 2, Y: 2, Kind: "tank"}},
	}
	if _, err := p.World(); err == nil {
		t.Fatal("expected unknown enemy kind to fail")
	}
}

func TestMission4BeachyHead(t *testing.T) {
	p := mustPhase(t, "m04p01.json")
	if p.Mission != 4 || p.Phase != 1 || p.Deploy != 4 || p.Title != "Beachy Head" {
		t.Fatalf("%d/%d deploy %d %q", p.Mission, p.Phase, p.Deploy, p.Title)
	}
	if p.StartGrenadesPerTrooper != 0 {
		t.Fatal("Beachy Head does not hand out grenades")
	}
	if len(p.Objectives) != 1 || p.Objectives[0] != "destroy_enemy_buildings" {
		t.Fatalf("objectives=%v", p.Objectives)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	assertGrenadeEconomy(t, w, 5, 8)
	if w.ActiveSquad().Grenades != 0 {
		t.Fatal("squad should start empty")
	}
	assertSpawnSurvives(t, w)
	for i := range w.Buildings {
		w.Buildings[i].Alive = false
	}
	w.Step(1.0 / 60)
	if w.Status != sim.Won {
		t.Fatal("destroying the huts should win even with grunts left")
	}
	if livingEnemies(w) == 0 {
		t.Fatal("expected surviving grunts")
	}
}

func TestMission4PierPressure(t *testing.T) {
	p := mustPhase(t, "m04p02.json")
	if p.Title != "Pier Pressure" || p.Deploy != 4 || p.StartGrenadesPerTrooper != 2 {
		t.Fatalf("%q deploy %d grenades %d", p.Title, p.Deploy, p.StartGrenadesPerTrooper)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	assertGrenadeEconomy(t, w, 4, 4)
	if w.ActiveSquad().Grenades != 8 {
		t.Fatalf("starter grenades=%d, want 8", w.ActiveSquad().Grenades)
	}
	bridge := false
	for _, e := range p.Enemies {
		if w.Map.At(e.X, e.Y) == sim.TileBridge {
			bridge = true
		}
	}
	if !bridge {
		t.Fatal("want a grunt holding the pier")
	}
	assertSpawnSurvives(t, w)
}

func TestMission4VillagePeople(t *testing.T) {
	p := mustPhase(t, "m04p03.json")
	if p.Title != "Village People" || p.Deploy != 5 || p.StartGrenadesPerTrooper != 2 {
		t.Fatalf("%q deploy %d grenades %d", p.Title, p.Deploy, p.StartGrenadesPerTrooper)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	doors, scenery := 0, 0
	for i := range w.Buildings {
		if w.Buildings[i].HasDoor {
			doors++
		} else {
			scenery++
		}
	}
	if doors != 2 || scenery < 2 {
		t.Fatalf("doors=%d scenery=%d", doors, scenery)
	}
	if livingCivilians(w) < 3 {
		t.Fatal("village should have civilians")
	}
	if !hasTile(w, sim.TileQuicksand) {
		t.Fatal("want a quicksand warning")
	}
	assertGrenadeEconomy(t, w, doors, 4)
	if w.ActiveSquad().Grenades != 10 {
		t.Fatalf("starter grenades=%d, want 10", w.ActiveSquad().Grenades)
	}
	assertSpawnSurvives(t, w)
}

func TestMission4Quicksand(t *testing.T) {
	p := mustPhase(t, "m04p04.json")
	if p.Title != "Quicksand" || p.Deploy != 5 || p.StartGrenadesPerTrooper != 2 {
		t.Fatalf("%q deploy %d grenades %d", p.Title, p.Deploy, p.StartGrenadesPerTrooper)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	if !hasTile(w, sim.TileQuicksand) || !hasTile(w, sim.TileMine) {
		t.Fatal("want quicksand pools and mines")
	}
	n := 0
	for i := range w.Units {
		u := &w.Units[i]
		if u.Kind != sim.KindGrenadier {
			continue
		}
		n++
		if u.Bombs != sim.GrenadierBombs || u.GrenadeCD != sim.GrenadierFirstDelay {
			t.Fatalf("grenadier bombs=%d cd=%v", u.Bombs, u.GrenadeCD)
		}
	}
	if n < 3 {
		t.Fatalf("grenadiers=%d, want at least 3", n)
	}
	doors := 0
	for i := range w.Buildings {
		if w.Buildings[i].HasDoor {
			doors++
		}
	}
	assertGrenadeEconomy(t, w, doors, 4)
	if w.ActiveSquad().Grenades != 10 {
		t.Fatalf("starter grenades=%d, want 10", w.ActiveSquad().Grenades)
	}
	assertSpawnSurvives(t, w)
}

func assertGrenadeEconomy(t *testing.T, w *sim.World, doors, crateGrenades int) {
	t.Helper()
	gotDoors := 0
	for i := range w.Buildings {
		if w.Buildings[i].HasDoor {
			gotDoors++
		}
	}
	if gotDoors != doors {
		t.Fatalf("door huts=%d, want %d", gotDoors, doors)
	}
	got := 0
	for i := range w.Pickups {
		pk := &w.Pickups[i]
		got += pk.Amount
		if w.Map.TileAtPixel(pk.X, pk.Y) != sim.TileGrass {
			t.Fatal("crate is not on grass")
		}
		for bi := range w.Buildings {
			b := &w.Buildings[bi]
			d := distToRect(pk.X, pk.Y, b.X, b.Y, b.W, b.H)
			if d <= sim.GrenadeRadius {
				t.Fatalf("crate %d is inside hut %d blast (dist=%v)", i, bi, d)
			}
		}
	}
	if got != crateGrenades {
		t.Fatalf("crate grenades=%d, want %d", got, crateGrenades)
	}
}

func distToRect(cx, cy, x, y, w, h float64) float64 {
	nx, ny := cx, cy
	if nx < x {
		nx = x
	} else if nx > x+w {
		nx = x + w
	}
	if ny < y {
		ny = y
	} else if ny > y+h {
		ny = y + h
	}
	return math.Hypot(cx-nx, cy-ny)
}

func hasTile(w *sim.World, want sim.Tile) bool {
	for y := 0; y < w.Map.H; y++ {
		for x := 0; x < w.Map.W; x++ {
			if w.Map.At(x, y) == want {
				return true
			}
		}
	}
	return false
}

func livingEnemies(w *sim.World) int {
	n := 0
	for i := range w.Units {
		if w.Units[i].Side == sim.SideEnemy && w.Units[i].Living() {
			n++
		}
	}
	return n
}

func livingCivilians(w *sim.World) int {
	n := 0
	for i := range w.Units {
		if w.Units[i].Side == sim.SideCivilian && w.Units[i].Living() {
			n++
		}
	}
	return n
}

func TestMission5ValleyOfIce(t *testing.T) {
	p := mustPhase(t, "m05p01.json")
	if p.Mission != 5 || p.Phase != 1 || p.Deploy != 3 || p.Title != "Valley of Ice" || p.Terrain != "arctic" {
		t.Fatalf("%d/%d %q deploy %d terrain %q", p.Mission, p.Phase, p.Title, p.Deploy, p.Terrain)
	}
	if p.StartGrenadesPerTrooper != 2 || p.StartRocketsPerTrooper != 0 {
		t.Fatalf("grenades %d rockets %d", p.StartGrenadesPerTrooper, p.StartRocketsPerTrooper)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	assertGrenadeEconomy(t, w, 6, 8)
	if !hasTile(w, sim.TileIce) {
		t.Fatal("want an ice river")
	}
	if countKind(w, sim.KindRocketeer) < 2 {
		t.Fatal("want the first rocketeers")
	}
	if !hasPickup(w, sim.PickupGrenades) || !hasPickup(w, sim.PickupRockets) {
		t.Fatal("want a grenade crate and a rocket crate")
	}
	if !rocketeersBesideTrees(w) {
		t.Fatal("rocketeers should stand beside a tree")
	}
	if w.ActiveSquad().Grenades != 6 {
		t.Fatalf("starter grenades=%d, want 6", w.ActiveSquad().Grenades)
	}
	assertSpawnSurvives(t, w)
}

func TestMission5BarmyBazookas(t *testing.T) {
	p := mustPhase(t, "m05p02.json")
	if p.Title != "Barmy Bazookas" || p.Deploy != 3 || p.StartRocketsPerTrooper != 0 {
		t.Fatalf("%q deploy %d rockets %d", p.Title, p.Deploy, p.StartRocketsPerTrooper)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	assertGrenadeEconomy(t, w, 6, 8)
	if !hasTile(w, sim.TileWaterDeep) || !hasTile(w, sim.TileBridge) {
		t.Fatal("want a river and a bridge")
	}
	if countKind(w, sim.KindRocketeer) < 4 {
		t.Fatal("want many rocketeers")
	}
	if !rocketeersBesideTrees(w) {
		t.Fatal("rocketeers should stand beside a tree")
	}
	assertSpawnSurvives(t, w)
}

func TestMission5Skidoo(t *testing.T) {
	p := mustPhase(t, "m05p03.json")
	if p.Title != "My Beautiful Skidoo" || p.Deploy != 4 || p.StartRocketsPerTrooper != 1 {
		t.Fatalf("%q deploy %d rockets %d", p.Title, p.Deploy, p.StartRocketsPerTrooper)
	}
	if len(p.Objectives) != 1 || p.Objectives[0] != "destroy_enemy_buildings" {
		t.Fatalf("objectives=%v", p.Objectives)
	}
	w := mustWorld(t, p)
	assertBigMap(t, w)
	assertStartWalkable(t, p, w)
	assertGrenadeEconomy(t, w, 3, 4)
	if !hasTile(w, sim.TileIce) {
		t.Fatal("want ice to skid on")
	}
	if w.ActiveSquad().Rockets != 4 || w.ActiveSquad().Grenades != 8 {
		t.Fatalf("rockets=%d grenades=%d, want 4 and 8", w.ActiveSquad().Rockets, w.ActiveSquad().Grenades)
	}
	player, enemy := 0, 0
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive || !v.Armed {
			t.Fatal("skidoos should start armed and intact")
		}
		if v.Side == sim.SidePlayer && !vehicleOccupied(w, v) {
			player++
		}
		if v.Side == sim.SideEnemy && vehicleOccupied(w, v) {
			enemy++
		}
	}
	if player != 1 || enemy != 1 {
		t.Fatalf("player skidoos=%d enemy skidoos=%d", player, enemy)
	}
	assertSpawnSurvives(t, w)
}

func vehicleOccupied(w *sim.World, v *sim.Vehicle) bool {
	for _, id := range v.Occupants {
		if u := w.Unit(id); u != nil && u.Living() {
			return true
		}
	}
	return false
}

func countKind(w *sim.World, kind sim.UnitKind) int {
	n := 0
	for i := range w.Units {
		if w.Units[i].Living() && w.Units[i].Kind == kind {
			n++
		}
	}
	return n
}

func hasPickup(w *sim.World, kind sim.PickupKind) bool {
	for i := range w.Pickups {
		if w.Pickups[i].Alive && w.Pickups[i].Kind == kind {
			return true
		}
	}
	return false
}

func rocketeersBesideTrees(w *sim.World) bool {
	for i := range w.Units {
		u := &w.Units[i]
		if u.Kind != sim.KindRocketeer || !u.Living() {
			continue
		}
		tx := int(u.X) / 16
		ty := int(u.Y) / 16
		near := false
		for dy := -1; dy <= 1 && !near; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				if w.Map.At(tx+dx, ty+dy) == sim.TileTree {
					near = true
					break
				}
			}
		}
		if !near {
			return false
		}
	}
	return true
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
		{'I', sim.TileIce},
		{'C', sim.TileCliff},
		{'R', sim.TileRamp},
		{'Q', sim.TileQuicksand},
		{'M', sim.TileMine},
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

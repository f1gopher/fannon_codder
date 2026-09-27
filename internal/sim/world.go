package sim

import "math/rand/v2"

const (
	UnitSize = 8
)

// World is the battle simulation. No Ebitengine types.
type World struct {
	Camera      Camera
	Units       []Unit
	Squads      []Squad
	Projectiles []Projectile
	Buildings   []Building
	Pickups     []Pickup
	Grenades    []Grenade
	Explosions  []Explosion
	Vehicles    []Vehicle
	AimX, AimY  float64
	Firing      bool
	Spread      float64
	AI          bool
	Status      Status
	Objectives  []Objective
	Map         Map
	// Selected members of the active squad will form the next split.
	Selected     []int
	GrenadeShare AmmoShare
	RocketShare  AmmoShare
	Special      Special
	BoardID      int
	Driving      bool
	DriveX       float64
	DriveY       float64
	nextID       int
	nextVID      int
	rng          *rand.Rand
}

// NewEmpty is a playable blank world (no units, no tiles).
func NewEmpty() *World {
	return &World{
		Camera: Camera{
			ViewW: 320,
			ViewH: 256,
			MapW:  320,
			MapH:  256,
		},
		Spread:     MGSpread,
		AI:         true,
		Status:     Playing,
		Objectives: []Objective{KillAllEnemy},
		nextID:     1,
		nextVID:    1,
		rng:        newRNG(),
	}
}

// NewDemoWorld is an empty-field 2v3 used by older tests.
func NewDemoWorld() *World {
	w := NewEmpty()
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		{X: 80, Y: 128},
		{X: 80 - FileSpacing, Y: 128},
	})
	w.SpawnUnit(SideEnemy, Vec2{X: 200, Y: 56})
	w.SpawnUnit(SideEnemy, Vec2{X: 248, Y: 140})
	w.SpawnUnit(SideEnemy, Vec2{X: 176, Y: 208})
	return w
}

// NewCoverWorld is the chunk 11 debug map: larger than one screen, a tree
// line that blocks MG, and grunts you only reach by panning / flanking.
func NewCoverWorld() *World {
	const W, H = 40, 30
	tiles := make([]Tile, W*H)
	for ty := 0; ty <= 18; ty++ {
		tiles[ty*W+16] = TileTree
		tiles[ty*W+17] = TileTree
	}
	w := NewEmpty()
	w.Map = Map{W: W, H: H, Tiles: tiles}
	mw, mh := w.Map.PixelSize()
	w.Camera.MapW = mw
	w.Camera.MapH = mh
	start := TileCenter(4, 8)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		start,
		{X: start.X - FileSpacing, Y: start.Y},
	})
	w.SpawnUnit(SideEnemy, TileCenter(24, 8)) // east of the wall, off the first screen
	w.SpawnUnit(SideEnemy, TileCenter(6, 26)) // south, walk off the starting view
	return w
}

// NewRiverWorld is the river sandbox: one bridge, swimmers, and three
// troopers so a squad can be left guarding the bridge.
func NewRiverWorld() *World {
	const W, H = 20, 16
	tiles := make([]Tile, W*H)
	for x := 0; x < W; x++ {
		tiles[4*W+x] = TileWaterShallow
		tiles[5*W+x] = TileWaterDeep
		tiles[6*W+x] = TileWaterDeep
		tiles[7*W+x] = TileWaterDeep
		tiles[8*W+x] = TileWaterShallow
	}
	for ty := 4; ty <= 8; ty++ {
		tiles[ty*W+10] = TileBridge
		tiles[ty*W+11] = TileBridge
	}
	w := NewEmpty()
	w.Map = Map{W: W, H: H, Tiles: tiles}
	mw, mh := w.Map.PixelSize()
	w.Camera.MapW = mw
	w.Camera.MapH = mh
	start := TileCenter(10, 12)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		start,
		{X: start.X - FileSpacing, Y: start.Y},
		{X: start.X - 2*FileSpacing, Y: start.Y},
	})
	w.SpawnUnit(SideEnemy, TileCenter(8, 6))  // swimmer, sitting duck
	w.SpawnUnit(SideEnemy, TileCenter(16, 6)) // swimmer
	w.SpawnUnit(SideEnemy, TileCenter(14, 1)) // far bank
	w.refreshTerrain()
	return w
}

// NewHutWorld is the chunk 14 sandbox: a door hut, a grenade crate, and reds.
func NewHutWorld() *World {
	const W, H = 20, 16
	tiles := make([]Tile, W*H)
	for _, p := range [][2]int{{8, 6}, {9, 6}, {8, 7}} {
		tiles[p[1]*W+p[0]] = TileTree
	}
	w := NewEmpty()
	w.Map = Map{W: W, H: H, Tiles: tiles}
	mw, mh := w.Map.PixelSize()
	w.Camera.MapW = mw
	w.Camera.MapH = mh
	w.Objectives = []Objective{KillAllEnemy, DestroyEnemyBuildings}
	start := TileCenter(6, 12)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		start,
		{X: start.X - FileSpacing, Y: start.Y},
	})
	w.AddDoorHut(13, 3)
	w.AddGrenadeCrate(TileCenter(10, 9))
	w.SpawnUnit(SideEnemy, w.Buildings[0].DoorSpawn())
	w.refreshTerrain()
	return w
}

// NewSkidooWorld is the chunk 20 sandbox: a skidoo, a rocket crate, and a hut.
func NewSkidooWorld() *World {
	const W, H = 24, 18
	tiles := make([]Tile, W*H)
	for tx := 8; tx <= 14; tx++ {
		for ty := 8; ty <= 10; ty++ {
			tiles[ty*W+tx] = TileIce
		}
	}
	w := NewEmpty()
	w.Map = Map{W: W, H: H, Tiles: tiles}
	mw, mh := w.Map.PixelSize()
	w.Camera.MapW = mw
	w.Camera.MapH = mh
	w.Objectives = []Objective{DestroyEnemyBuildings}
	start := TileCenter(4, 14)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		start,
		{X: start.X - FileSpacing, Y: start.Y},
	})
	w.AddSkidoo(TileCenter(8, 14), SidePlayer, true)
	w.AddRocketCrate(TileCenter(6, 12))
	w.AddBuilding(16, 3, 2, 2, true, SpawnEvery)
	w.AddSkidoo(TileCenter(20, 4), SideEnemy, true)
	w.refreshTerrain()
	return w
}

// NewHazardWorld is the chunk 18 sandbox: a mine, a quicksand pool,
// a wandering civilian, and a doorless hut that is not an objective.
func NewHazardWorld() *World {
	const W, H = 20, 16
	tiles := make([]Tile, W*H)
	tiles[12*W+7] = TileMine
	for ty := 9; ty <= 11; ty++ {
		for tx := 10; tx <= 13; tx++ {
			tiles[ty*W+tx] = TileQuicksand
		}
	}
	w := NewEmpty()
	w.Map = Map{W: W, H: H, Tiles: tiles}
	mw, mh := w.Map.PixelSize()
	w.Camera.MapW = mw
	w.Camera.MapH = mh
	w.Objectives = []Objective{KillAllEnemy, DestroyEnemyBuildings}
	start := TileCenter(3, 12)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		start,
		{X: start.X - FileSpacing, Y: start.Y},
	})
	w.SpawnUnit(SideCivilian, TileCenter(16, 13))
	w.AddBuilding(14, 2, 2, 2, false, 0)
	w.SpawnUnit(SideEnemy, TileCenter(3, 3))
	w.refreshTerrain()
	return w
}

// SpawnPlayerSquad adds a player squad. The first position is the leader.
func (w *World) SpawnPlayerSquad(id SquadID, positions []Vec2) *Squad {
	s := Squad{ID: id, Active: len(w.Squads) == 0}
	for _, p := range positions {
		u := Unit{
			ID:      w.nextID,
			Side:    SidePlayer,
			HP:      Alive,
			X:       p.X,
			Y:       p.Y,
			SquadID: id,
		}
		w.nextID++
		w.Units = append(w.Units, u)
		s.MemberIDs = append(s.MemberIDs, u.ID)
	}
	if len(s.MemberIDs) > 0 {
		s.LeaderID = s.MemberIDs[0]
	}
	w.Squads = append(w.Squads, s)
	return &w.Squads[len(w.Squads)-1]
}

// SpawnUnit adds a lone trooper (dummy, later enemies) and returns it.
func (w *World) SpawnUnit(side Side, p Vec2) *Unit {
	if w.rng == nil {
		w.rng = newRNG()
	}
	u := Unit{
		ID:   w.nextID,
		Side: side,
		HP:   Alive,
		X:    p.X,
		Y:    p.Y,
	}
	w.nextID++
	w.Units = append(w.Units, u)
	return &w.Units[len(w.Units)-1]
}

func (w *World) Unit(id int) *Unit {
	for i := range w.Units {
		if w.Units[i].ID == id {
			return &w.Units[i]
		}
	}
	return nil
}

func (w *World) ActiveSquad() *Squad {
	for i := range w.Squads {
		if w.Squads[i].Active {
			return &w.Squads[i]
		}
	}
	if len(w.Squads) > 0 {
		return &w.Squads[0]
	}
	return nil
}

// CommandMove orders the active squad toward a world-space point.
// newOrder resets the file trail (a fresh click). Dragging the pointer
// while held updates the destination without breaking the file.
func (w *World) CommandMove(x, y float64, newOrder bool) {
	s := w.ActiveSquad()
	if s == nil {
		return
	}
	s.HasDest = true
	s.DestX = x
	s.DestY = y
	if newOrder {
		w.BoardID = 0
		if l := w.Unit(s.LeaderID); l != nil {
			s.Trail = []Vec2{{X: l.X, Y: l.Y}}
		}
	}
}

// Step advances movement and combat by dt seconds.
func (w *World) Step(dt float64) {
	if dt <= 0 {
		return
	}
	w.trapQuicksand()
	w.pruneSquads()
	for i := range w.Squads {
		w.stepSquad(&w.Squads[i], dt)
	}
	w.tryCompleteBoard()
	w.stepMerge()
	w.stepPickups()
	w.refreshTerrain()
	w.stepCivilians(dt)
	w.stepAI(dt)
	w.stepVehicles(dt)
	w.stepMines()
	w.stepQuicksand(dt)
	w.stepFire(dt)
	w.stepInactiveFire(dt)
	w.stepProjectiles(dt)
	w.stepGrenades(dt)
	w.stepSpawners(dt)
	w.stepBlasts(dt)
	w.evaluateObjectives()
}

func (w *World) stepSquad(s *Squad, dt float64) {
	leader := w.Unit(s.LeaderID)
	if leader == nil || !leader.Living() {
		if len(s.MemberIDs) == 0 {
			return
		}
		leader = w.Unit(s.MemberIDs[0])
		if leader == nil || !leader.Living() {
			return
		}
		s.LeaderID = leader.ID
	}
	if leader.VehicleID != 0 {
		leader.VX = 0
		leader.VY = 0
		return
	}
	if s.HasDest {
		arrived := w.steerToward(leader, s.DestX, s.DestY, WalkSpeed, dt, ArrivalRadius)
		if arrived {
			s.HasDest = false
		}
	} else {
		leader.VX = 0
		leader.VY = 0
	}
	recordTrail(s, Vec2{X: leader.X, Y: leader.Y})

	for i, id := range s.MemberIDs {
		if id == s.LeaderID {
			continue
		}
		u := w.Unit(id)
		if u == nil || !u.Living() {
			continue
		}
		target, _ := trailPoint(Vec2{X: leader.X, Y: leader.Y}, s.Trail, FileSpacing*float64(i))
		w.steerToward(u, target.X, target.Y, WalkSpeed, dt, FollowerArrive)
	}
}

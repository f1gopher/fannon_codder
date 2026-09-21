package sim

const (
	// SpawnEvery is the default door-hut interval.
	SpawnEvery = 2.5
	// maxDoorSpawns caps living enemies so a hut spews reds without flooding the map.
	maxDoorSpawns = 6
	buildingHP    = 1
)

// Building is a hut. Door huts spawn grunts until a grenade destroys them.
// Doorless huts are scenery: they do not count for destroy_enemy_buildings.
type Building struct {
	X, Y, W, H float64
	HasDoor    bool
	HP         int
	Alive      bool
	SpawnEvery float64
	SpawnCD    float64
}

// DoorSpawn is just south of the hut, clear of the walls.
func (b *Building) DoorSpawn() Vec2 {
	return Vec2{X: b.X + b.W/2, Y: b.Y + b.H + float64(UnitSize)}
}

// AddDoorHut places a 2×2-tile spawner whose top-left tile is (tx, ty).
func (w *World) AddDoorHut(tx, ty int) int {
	return w.addBuilding(tx, ty, 2, 2, true, SpawnEvery)
}

// AddBuilding places a hut from mission data. w/h are tile counts (default 2).
func (w *World) AddBuilding(tx, ty, tw, th int, door bool, spawn float64) int {
	return w.addBuilding(tx, ty, tw, th, door, spawn)
}

func (w *World) addBuilding(tx, ty, tw, th int, door bool, spawn float64) int {
	if tw < 1 {
		tw = 2
	}
	if th < 1 {
		th = 2
	}
	b := Building{
		X:       float64(tx * TileSize),
		Y:       float64(ty * TileSize),
		W:       float64(tw * TileSize),
		H:       float64(th * TileSize),
		HasDoor: door,
		HP:      buildingHP,
		Alive:   true,
	}
	if door {
		if spawn <= 0 {
			spawn = SpawnEvery
		}
		b.SpawnEvery = spawn
		b.SpawnCD = spawn * 0.25
	}
	w.Buildings = append(w.Buildings, b)
	return len(w.Buildings) - 1
}

func (w *World) stepSpawners(dt float64) {
	if w.livingCount(SideEnemy) >= maxDoorSpawns {
		return
	}
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if !b.Alive || !b.HasDoor || b.SpawnEvery <= 0 {
			continue
		}
		b.SpawnCD -= dt
		if b.SpawnCD > 0 {
			continue
		}
		b.SpawnCD = b.SpawnEvery
		w.SpawnUnit(SideEnemy, b.DoorSpawn())
		if w.livingCount(SideEnemy) >= maxDoorSpawns {
			return
		}
	}
}

func (w *World) livingCount(side Side) int {
	n := 0
	for i := range w.Units {
		if w.Units[i].Side == side && w.Units[i].Living() {
			n++
		}
	}
	return n
}

func (w *World) anyDoorBuilding() bool {
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if b.Alive && b.HasDoor {
			return true
		}
	}
	return false
}

func (w *World) pointInBuilding(x, y float64) bool {
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if b.Alive && x >= b.X && y >= b.Y && x < b.X+b.W && y < b.Y+b.H {
			return true
		}
	}
	return false
}

func (w *World) unitHitsBuilding(x, y, half float64) bool {
	ax, ay := x-half, y-half
	aw, ah := half*2, half*2
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if !b.Alive {
			continue
		}
		if ax < b.X+b.W && ax+aw > b.X && ay < b.Y+b.H && ay+ah > b.Y {
			return true
		}
	}
	return false
}

// buildingHit is the first point where the segment enters a standing hut.
func (w *World) buildingHit(x0, y0, x1, y1 float64) (float64, float64, bool) {
	dx, dy := x1-x0, y1-y0
	dist := hypot(dx, dy)
	if dist < 1e-6 {
		if w.pointInBuilding(x1, y1) {
			return x1, y1, true
		}
		return 0, 0, false
	}
	nx, ny := dx/dist, dy/dist
	for d := 2.0; d < dist; d += 2 {
		x := x0 + nx*d
		y := y0 + ny*d
		if w.pointInBuilding(x, y) {
			return x, y, true
		}
	}
	if w.pointInBuilding(x1, y1) {
		return x1, y1, true
	}
	return 0, 0, false
}

// lineClear is map LOS plus standing huts. Grenades do not use this.
func (w *World) lineClear(x0, y0, x1, y1 float64) bool {
	if !w.Map.HasLOS(x0, y0, x1, y1) {
		return false
	}
	_, _, hit := w.buildingHit(x0, y0, x1, y1)
	return !hit
}

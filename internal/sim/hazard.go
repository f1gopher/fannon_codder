package sim

import "math"

const (
	// SinkTime is how long a unit flails in quicksand before dying.
	SinkTime = 2.0

	CivilianSpeed = 14.0
	civilianPause = 0.8
	// civilianFlee is about two tiles. A shot or a blast inside it
	// sends a wandering civilian the other way.
	civilianFlee     = 2 * float64(TileSize)
	civilianFleeTime = 0.7
)

// trapQuicksand sticks anyone already standing in a pool so they cannot walk out.
func (w *World) trapQuicksand() {
	for i := range w.Units {
		w.noteQuicksand(&w.Units[i])
	}
}

func (w *World) noteQuicksand(u *Unit) {
	if u == nil || !u.Living() || u.Sinking || u.VehicleID != 0 {
		return
	}
	if w.Map.TileAtPixel(u.X, u.Y) == TileQuicksand {
		u.Sinking = true
		u.VX = 0
		u.VY = 0
		w.emit(CueSink, u.X, u.Y)
	}
}

// stepQuicksand advances the sink timer. Death is at SinkTime.
func (w *World) stepQuicksand(dt float64) {
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Living() || !u.Sinking {
			continue
		}
		u.VX = 0
		u.VY = 0
		u.Sink += dt
		if u.Sink >= SinkTime {
			w.kill(u)
		}
	}
}

// stepMines detonates a mine when a living unit's centre steps on it.
// The blast is a grenade explosion, and the tile is spent.
func (w *World) stepMines() {
	if !w.Map.active() {
		return
	}
	type trig struct{ tx, ty, id int }
	var hits []trig
	seen := map[[2]int]bool{}
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Living() || u.VehicleID != 0 {
			continue
		}
		tx := int(math.Floor(u.X / float64(TileSize)))
		ty := int(math.Floor(u.Y / float64(TileSize)))
		if w.Map.At(tx, ty) != TileMine {
			continue
		}
		key := [2]int{tx, ty}
		if seen[key] {
			continue
		}
		seen[key] = true
		hits = append(hits, trig{tx, ty, u.ID})
	}
	for _, h := range hits {
		if w.Map.At(h.tx, h.ty) != TileMine {
			continue
		}
		w.Map.Set(h.tx, h.ty, TileGrass)
		c := TileCenter(h.tx, h.ty)
		w.explode(c.X, c.Y, h.id)
	}
	if !w.Map.active() {
		return
	}
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if !v.Alive {
			continue
		}
		tx := int(math.Floor(v.X / float64(TileSize)))
		ty := int(math.Floor(v.Y / float64(TileSize)))
		if w.Map.At(tx, ty) != TileMine {
			continue
		}
		w.Map.Set(tx, ty, TileGrass)
		c := TileCenter(tx, ty)
		w.explode(c.X, c.Y, 0)
	}
}

// stepCivilians strolls yellow units. They do not shoot.
func (w *World) stepCivilians(dt float64) {
	if w.rng == nil {
		w.rng = newRNG()
	}
	for i := range w.Units {
		u := &w.Units[i]
		if u.Side != SideCivilian || !u.Living() || u.Sinking {
			continue
		}
		if !u.Escorting {
			if tx, ty, ok := w.civilianThreat(u); ok {
				w.fleeCivilian(u, tx, ty)
			}
		}
		u.WanderT -= dt
		if u.WanderT <= 0 {
			w.pickWander(u)
		}
		w.steerToward(u, u.WanderX, u.WanderY, CivilianSpeed, dt, ArrivalRadius)
	}
}

// civilianThreat is the nearest live round or blast inside civilianFlee.
func (w *World) civilianThreat(u *Unit) (float64, float64, bool) {
	best := civilianFlee
	var x, y float64
	found := false
	note := func(px, py float64) {
		d := hypot(px-u.X, py-u.Y)
		if d < best {
			best = d
			x, y = px, py
			found = true
		}
	}
	for i := range w.Projectiles {
		p := &w.Projectiles[i]
		if p.Alive {
			note(p.X, p.Y)
		}
	}
	for i := range w.Explosions {
		note(w.Explosions[i].X, w.Explosions[i].Y)
	}
	return x, y, found
}

func (w *World) fleeCivilian(u *Unit, tx, ty float64) {
	dx, dy := u.X-tx, u.Y-ty
	d := hypot(dx, dy)
	if d < 1 {
		ang := w.rng.Float64() * 2 * math.Pi
		dx, dy = math.Cos(ang), math.Sin(ang)
		d = 1
	}
	x := u.X + dx/d*civilianFlee
	y := u.Y + dy/d*civilianFlee
	u.WanderX, u.WanderY = clampStroll(w, x, y)
	u.WanderT = civilianFleeTime
}

func (w *World) pickWander(u *Unit) {
	ang := w.rng.Float64() * 2 * math.Pi
	dist := 24 + w.rng.Float64()*40
	x := u.X + math.Cos(ang)*dist
	y := u.Y + math.Sin(ang)*dist
	u.WanderX, u.WanderY = clampStroll(w, x, y)
	u.WanderT = civilianPause + w.rng.Float64()*1.6
}

func clampStroll(w *World, x, y float64) (float64, float64) {
	if w == nil || !w.Map.active() {
		return x, y
	}
	mw, mh := w.Map.PixelSize()
	margin := float64(TileSize)
	if x < margin {
		x = margin
	}
	if y < margin {
		y = margin
	}
	if x > mw-margin {
		x = mw - margin
	}
	if y > mh-margin {
		y = mh - margin
	}
	return x, y
}

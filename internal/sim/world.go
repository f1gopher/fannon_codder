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
	AimX, AimY  float64
	Firing      bool
	Spread      float64
	nextID      int
	rng         *rand.Rand
}

// NewDemoWorld is the Chunk 03 sandbox: two player troopers, one screen of grass.
func NewDemoWorld() *World {
	w := &World{
		Camera: Camera{
			ViewW: 320,
			ViewH: 256,
			MapW:  320,
			MapH:  256,
		},
		Spread: MGSpread,
		nextID: 1,
		rng:    newRNG(),
	}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		{X: 80, Y: 128},
		{X: 80 - FileSpacing, Y: 128},
	})
	// Stationary dummy so MG can be verified (no AI this chunk).
	w.SpawnUnit(SideEnemy, Vec2{X: 160, Y: 100})
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
	for i := range w.Squads {
		w.stepSquad(&w.Squads[i], dt)
	}
	w.stepFire(dt)
	w.stepProjectiles(dt)
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
	if s.HasDest {
		arrived := steerToward(leader, s.DestX, s.DestY, WalkSpeed, dt, ArrivalRadius)
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
		steerToward(u, target.X, target.Y, WalkSpeed, dt, FollowerArrive)
	}
}

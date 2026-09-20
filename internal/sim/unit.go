package sim

// Side is who a unit fights for.
type Side int

const (
	SidePlayer Side = iota
	SideEnemy
	SideCivilian
)

// HP is the coarse health state. One MG hit kills a healthy infantryman later.
type HP int

const (
	Alive HP = iota
	Wounded
	Dead
)

// Unit is one trooper (or later, a corpse still drawn on the map).
type Unit struct {
	ID        int
	Name      string
	Rank      int
	Side      Side
	HP        HP
	X, Y      float64
	VX, VY    float64
	Facing    float64 // radians, 0 = east
	InWater   bool
	SquadID   SquadID
	VehicleID int
	Kills     int
	FireCD    float64 // seconds until next MG round
}

func (u *Unit) Dead() bool { return u.HP == Dead }

func (u *Unit) Living() bool {
	return u.HP == Alive
}

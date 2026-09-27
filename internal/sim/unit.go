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

// UnitKind distinguishes a grenadier from an ordinary grunt.
type UnitKind int

const (
	KindInfantry UnitKind = iota
	KindGrenadier
	KindRocketeer
)

// Unit is one trooper (or later, a corpse still drawn on the map).
type Unit struct {
	ID          int
	Name        string
	Rank        int
	Side        Side
	Kind        UnitKind
	HP          HP
	X, Y        float64
	VX, VY      float64
	Facing      float64 // radians, 0 = east
	InWater     bool
	SquadID     SquadID
	VehicleID   int
	Kills       int
	FireCD      float64 // seconds until next MG round
	Sinking     bool    // trapped in quicksand; cannot move or fire
	Sink        float64 // seconds spent sinking; SinkTime is death
	WanderX     float64 // civilian stroll target
	WanderY     float64
	WanderT     float64 // seconds until the next stroll pick
	Bombs       int     // grenades a grenadier still carries
	GrenadeWind float64 // seconds left in the throw telegraph; the bomb leaves at 0
	GrenadeCD   float64 // seconds until the next windup may start
	RocketWind  float64 // rocketeer aim telegraph
	RocketCD    float64 // seconds until the next rocket windup
}

func (u *Unit) Dead() bool { return u.HP == Dead }

func (u *Unit) Living() bool {
	return u.HP == Alive
}

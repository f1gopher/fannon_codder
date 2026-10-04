package sim

// Side is who a unit fights for.
type Side int

const (
	SidePlayer Side = iota
	SideEnemy
	SideCivilian
)

// HP is the coarse health state. An MG round kills a standing man or drops him
// wounded. A second round, a blast, or a vehicle finishes a wounded man.
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
	WindHeld    bool    // telegraph finished; release on the first frame he faces the target
	SpotT       float64 // seconds of unbroken contact (LOS or a heard shot). Zeroed when both end.
	ReactAt     float64 // SpotT must reach this before the first MG round. 0 means no live contact.
	HearID      int     // infantry shooter whose MG round woke an idle grunt. 0 means none.
	BurstN      int     // MG rounds fired in the current chatter.
	BurstGap    float64 // seconds of silence left after a full burst.
	HasPost     bool    // walking from a door to PostX, PostY. Map spawns leave this false.
	Aggressive  bool    // closes on an acquired player, then stops to shoot. Mission 1 leaves this false.
	PostX       float64
	PostY       float64
	SinceShot   float64 // seconds since this unit's last MG round. Render only.
	SinceThrow  float64 // seconds since this unit's last grenade or rocket. Render only.
	Hop         float64 // corpse height above the ground, in world pixels
	VZ          float64 // corpse vertical speed; positive is up
	sampled     bool    // terrain has been read once; a man placed in water does not splash
}

// animRest is SinceShot and SinceThrow for a trooper who has not fired.
// The picture treats the first 0.12s after a shot and the first 0.25s after
// a throw as those poses. Zero would play them on the spawn frame.
const animRest = 1

func (u *Unit) Dead() bool { return u.HP == Dead }

func (u *Unit) Wounded() bool { return u.HP == Wounded }

func (u *Unit) Living() bool {
	return u.HP == Alive
}

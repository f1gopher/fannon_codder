package sim

// ProjKind separates machine-gun rounds from bazooka rockets.
type ProjKind int

const (
	ProjMG ProjKind = iota
	ProjRocket
	ProjShell
	ProjHoming
)

// Projectile is a machine-gun round or a bazooka rocket.
type Projectile struct {
	Kind      ProjKind
	X, Y      float64
	VX, VY    float64
	OwnerID   int
	OwnerSide Side
	// TargetID is the man a homing rocket is turning toward.
	TargetID int
	Left     float64 // remaining travel in px
	Alive    bool
}

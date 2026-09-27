package sim

// ProjKind separates machine-gun rounds from bazooka rockets.
type ProjKind int

const (
	ProjMG ProjKind = iota
	ProjRocket
)

// Projectile is a machine-gun round or a bazooka rocket.
type Projectile struct {
	Kind      ProjKind
	X, Y      float64
	VX, VY    float64
	OwnerID   int
	OwnerSide Side
	Left      float64 // remaining travel in px
	Alive     bool
}

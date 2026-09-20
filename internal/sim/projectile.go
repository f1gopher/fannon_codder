package sim

// Projectile is a machine-gun round (very fast; range-limited).
type Projectile struct {
	X, Y      float64
	VX, VY    float64
	OwnerID   int
	OwnerSide Side
	Left      float64 // remaining travel in px
	Alive     bool
}

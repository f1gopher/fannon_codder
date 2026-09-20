package sim

const (
	// EdgeScrollMargin is how close the pointer must be to a playfield
	// edge before the camera pans (Amiga look-around).
	EdgeScrollMargin = 16
	// EdgeScrollSpeed is camera pan speed in world pixels per second.
	EdgeScrollSpeed = 120
)

// Camera is the world-space top-left of the visible playfield.
type Camera struct {
	X, Y             float64
	ViewW, ViewH     float64
	MapW, MapH       float64
}

// Clamp keeps the view inside the map. A map smaller than the view pins to 0,0.
func (c *Camera) Clamp() {
	maxX := c.MapW - c.ViewW
	if maxX < 0 {
		maxX = 0
	}
	maxY := c.MapH - c.ViewH
	if maxY < 0 {
		maxY = 0
	}
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.X > maxX {
		c.X = maxX
	}
	if c.Y > maxY {
		c.Y = maxY
	}
}

// ScrollToward pans when the pointer (screen pixels) is near a view edge.
// screenW/screenH are the playfield size in screen pixels (usually ViewW/ViewH).
func (c *Camera) ScrollToward(pointerX, pointerY, screenW, screenH, dt float64) {
	if dt < 0 {
		dt = 0
	}
	step := EdgeScrollSpeed * dt
	if pointerX < EdgeScrollMargin {
		c.X -= step
	} else if screenW > 0 && pointerX > screenW-EdgeScrollMargin {
		c.X += step
	}
	if pointerY < EdgeScrollMargin {
		c.Y -= step
	} else if screenH > 0 && pointerY > screenH-EdgeScrollMargin {
		c.Y += step
	}
	c.Clamp()
}

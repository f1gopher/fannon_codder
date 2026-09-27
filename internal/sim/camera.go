package sim

const (
	// EdgeScrollMargin is how close the pointer must be to a playfield
	// edge before the camera pans (Amiga look-around).
	EdgeScrollMargin = 16
	// EdgeScrollSpeed is camera pan speed in world pixels per second.
	EdgeScrollSpeed = 120
	// LeaderMargin keeps the active leader this far inside the playfield.
	// Walking into the strip pushes the camera; the pointer cannot pan the
	// squad off the view.
	LeaderMargin = 32
)

// Camera is the world-space top-left of the visible playfield.
// OriginX is the screen pixel where that top-left is drawn (right of the
// status strip). ViewW and ViewH are the playfield, not the full screen.
type Camera struct {
	X, Y         float64
	ViewW, ViewH float64
	MapW, MapH   float64
	OriginX      float64
}

// ScreenX converts a world x to a screen pixel.
func (c Camera) ScreenX(wx float64) float64 { return wx - c.X + c.OriginX }

// ScreenY converts a world y to a screen pixel.
func (c Camera) ScreenY(wy float64) float64 { return wy - c.Y }

// WorldX converts a screen pixel to a world x.
func (c Camera) WorldX(sx float64) float64 { return sx - c.OriginX + c.X }

// WorldY converts a screen pixel to a world y.
func (c Camera) WorldY(sy float64) float64 { return sy + c.Y }

// CenterOn puts (x, y) in the middle of the view, then clamps.
// A map that fits the screen stays pinned at 0,0.
func (c *Camera) CenterOn(x, y float64) {
	c.X = x - c.ViewW/2
	c.Y = y - c.ViewH/2
	c.Clamp()
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

// Contain scrolls so (x, y) stays inside the soft margin, then clamps.
func (c *Camera) Contain(x, y float64) {
	c.containAxis(&c.X, x, c.ViewW)
	c.containAxis(&c.Y, y, c.ViewH)
	c.Clamp()
}

func (c *Camera) containAxis(cam *float64, pos, view float64) {
	if view <= 0 {
		return
	}
	if view <= LeaderMargin*2 {
		*cam = pos - view/2
		return
	}
	if pos < *cam+LeaderMargin {
		*cam = pos - LeaderMargin
	} else if pos > *cam+view-LeaderMargin {
		*cam = pos + LeaderMargin - view
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

package input

import "math"

const (
	// PadDeadzone is the radial stick zone that does not move the pointer.
	PadDeadzone = 0.22
	// PadCursorSpeed is pointer travel at a full stick, in frame pixels per second.
	PadCursorSpeed = 420
	// PadRepeatDelay is how long a held d-pad waits before the menu repeats.
	PadRepeatDelay = 0.28
	// PadRepeatEvery is the gap between those repeats.
	PadRepeatEvery = 0.14
)

// Pad is one frame of a standard gamepad. Sticks are in [-1, 1].
// Vertical stick axes are positive downward, matching the playfield.
type Pad struct {
	Connected bool
	StickX    float64
	StickY    float64
	RightX    float64
	RightY    float64
	// South is the bottom face button (A / Cross). East is right (B / Circle).
	// West is left (X / Square). North is top (Y / Triangle).
	South, East, West, North bool
	LB, RB                   bool
	LT, RT                   bool
	Start, Back              bool
	Up, Down, Left, Right    bool
}

// Edges is what that pad asks the game to do this tick.
type Edges struct {
	Click, Fire bool
	Confirm     bool
	Cancel      bool
	Deny        bool
	Grenade     bool
	Special     bool
	Weapon      bool
	Map         bool
	Heroes      bool
	Up, Down    bool
	// Squad is -1, 0, or +1. The battle skips a squad that is not on the field.
	Squad int
	// DX, DY move the pointer. PanX, PanY are the right stick after the dead zone.
	DX, DY float64
	PanX   float64
	PanY   float64
	Moving bool
	// Digital is set when the d-pad is steering and the stick is still.
	// The pause menu uses the d-pad for the highlight and leaves the pointer put.
	Digital bool
}

// PadState remembers button levels so edges and menu repeat stay stable.
type PadState struct {
	prev     Pad
	upHold   float64
	downHold float64
}

// Step turns a pad sample into pointer motion and button edges. dt is seconds.
func (s *PadState) Step(p Pad, dt float64) Edges {
	if dt < 0 {
		dt = 0
	}
	var e Edges
	if !p.Connected {
		s.prev = Pad{}
		s.upHold = 0
		s.downHold = 0
		return e
	}
	sx, sy, stick := deflected(p.StickX, p.StickY)
	digital := p.Left || p.Right || p.Up || p.Down
	if p.Left {
		sx = -1
	}
	if p.Right {
		sx = 1
	}
	if p.Up {
		sy = -1
	}
	if p.Down {
		sy = 1
	}
	e.Moving = stick || digital
	e.Digital = digital && !stick
	e.DX = sx * PadCursorSpeed * dt
	e.DY = sy * PadCursorSpeed * dt
	e.PanX, e.PanY, _ = deflected(p.RightX, p.RightY)

	e.Click = p.South
	e.Fire = p.RT
	e.Confirm = p.South && !s.prev.South
	e.Cancel = (p.East && !s.prev.East) || (p.Start && !s.prev.Start)
	e.Deny = p.East && !s.prev.East
	e.Grenade = p.West && !s.prev.West
	e.Special = p.North && !s.prev.North
	e.Weapon = p.LT && !s.prev.LT
	e.Map = p.Back && !s.prev.Back
	e.Heroes = p.North && !s.prev.North
	e.Up = menuRepeat(&s.upHold, p.Up, s.prev.Up, dt)
	e.Down = menuRepeat(&s.downHold, p.Down, s.prev.Down, dt)
	if p.LB && !s.prev.LB {
		e.Squad = -1
	}
	if p.RB && !s.prev.RB {
		e.Squad = 1
	}
	s.prev = p
	return e
}

func deflected(x, y float64) (float64, float64, bool) {
	m := math.Hypot(x, y)
	if m < PadDeadzone || m == 0 || math.IsNaN(m) {
		return 0, 0, false
	}
	if m > 1 {
		m = 1
	}
	scale := (m - PadDeadzone) / (1 - PadDeadzone)
	return x / m * scale, y / m * scale, true
}

// menuRepeat is true on the press and again while the button stays down.
func menuRepeat(held *float64, down, was bool, dt float64) bool {
	if !down {
		*held = 0
		return false
	}
	if !was {
		*held = 0
		return true
	}
	*held += dt
	if *held >= PadRepeatDelay {
		*held -= PadRepeatEvery
		return true
	}
	return false
}

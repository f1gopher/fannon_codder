package input

// Pointer is the mouse state in logical screen pixels (320×256).
type Pointer struct {
	X, Y float64

	Left  bool
	Right bool

	// LeftDown / RightDown are true only on the tick the button was pressed.
	LeftDown  bool
	RightDown bool

	// ChordGrenade is right held and left just pressed (Amiga special weapon).
	ChordGrenade bool
}

// Tracker turns raw button levels into edges and the grenade chord.
type Tracker struct {
	prevLeft  bool
	prevRight bool
	muteLeft  bool
	muteRight bool
}

// Update records this tick's cursor and buttons. x,y are logical screen pixels.
// inside is false in the letterbox bars. A press that starts there stays mute
// until that button is released. A drag that started inside keeps the button
// while the cursor is outside, so coming back is not a new press.
func (t *Tracker) Update(x, y float64, left, right, inside bool) Pointer {
	if !inside {
		if left && !t.prevLeft {
			t.muteLeft = true
		}
		if right && !t.prevRight {
			t.muteRight = true
		}
	}
	if !left {
		t.muteLeft = false
	}
	if !right {
		t.muteRight = false
	}
	if t.muteLeft {
		left = false
	}
	if t.muteRight {
		right = false
	}
	p := Pointer{
		X:            x,
		Y:            y,
		Left:         left,
		Right:        right,
		LeftDown:     left && !t.prevLeft,
		RightDown:    right && !t.prevRight,
		ChordGrenade: right && left && !t.prevLeft,
	}
	t.prevLeft = left
	t.prevRight = right
	return p
}

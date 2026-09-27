package app

// quitChoice is what the quit prompt does this frame.
type quitChoice int

const (
	quitIdle quitChoice = iota
	quitHold
	quitCancel
	quitExit
)

// quitPrompt pauses the game until the player confirms leaving the process.
type quitPrompt struct {
	open bool
}

// decide opens the prompt on escape, and while it is open treats confirm as
// exit and cancel or a second escape as staying. Opening and cancelling never
// happen on the same press.
func (q *quitPrompt) decide(escape, confirm, cancel bool) quitChoice {
	if !q.open {
		if escape {
			q.open = true
			return quitHold
		}
		return quitIdle
	}
	if confirm && !escape {
		q.open = false
		return quitExit
	}
	if cancel || escape {
		q.open = false
		return quitCancel
	}
	return quitHold
}

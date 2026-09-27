package app

import "testing"

func TestQuitPromptConfirmsAndCancels(t *testing.T) {
	var q quitPrompt
	if got := q.decide(false, true, false); got != quitIdle || q.open {
		t.Fatalf("confirm before open: %v open=%v", got, q.open)
	}
	if got := q.decide(true, false, false); got != quitHold || !q.open {
		t.Fatalf("escape should open: %v open=%v", got, q.open)
	}
	if got := q.decide(true, true, false); got != quitCancel || q.open {
		t.Fatalf("escape while open should cancel even with confirm: %v open=%v", got, q.open)
	}

	q.decide(true, false, false)
	if got := q.decide(false, true, false); got != quitExit || q.open {
		t.Fatalf("confirm should exit: %v open=%v", got, q.open)
	}

	q.decide(true, false, false)
	if got := q.decide(false, false, true); got != quitCancel || q.open {
		t.Fatalf("cancel should close: %v open=%v", got, q.open)
	}
	if got := q.decide(false, false, false); got != quitIdle {
		t.Fatalf("closed prompt should idle: %v", got)
	}
}

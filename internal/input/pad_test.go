package input

import "testing"

func TestStickInsideTheDeadzoneDoesNotMove(t *testing.T) {
	var s PadState
	e := s.Step(Pad{Connected: true, StickX: 0.1, StickY: -0.1}, 1)
	if e.Moving || e.DX != 0 || e.DY != 0 {
		t.Fatalf("dead zone still moved %#v", e)
	}
}

func TestFullStickDownMovesAtCursorSpeed(t *testing.T) {
	var s PadState
	e := s.Step(Pad{Connected: true, StickY: 1}, 0.5)
	if !e.Moving || e.DY != PadCursorSpeed*0.5 || e.DX != 0 {
		t.Fatalf("stick down %#v", e)
	}
}

func TestDpadUpAimsThePointerUp(t *testing.T) {
	var s PadState
	e := s.Step(Pad{Connected: true, Up: true}, 1)
	if e.DY >= 0 || !e.Up {
		t.Fatalf("d-pad up %#v", e)
	}
}

func TestFaceButtonsAreEdges(t *testing.T) {
	var s PadState
	first := s.Step(Pad{Connected: true, South: true, West: true, RT: true}, 1.0/60)
	if !first.Click || !first.Confirm || !first.Grenade || !first.Fire {
		t.Fatalf("press %#v", first)
	}
	held := s.Step(Pad{Connected: true, South: true, West: true, RT: true}, 1.0/60)
	if !held.Click || !held.Fire || held.Confirm || held.Grenade {
		t.Fatalf("hold %#v", held)
	}
}

func TestMenuDownRepeatsWhileHeld(t *testing.T) {
	var s PadState
	if e := s.Step(Pad{Connected: true, Down: true}, 0.01); !e.Down {
		t.Fatal("first press")
	}
	if e := s.Step(Pad{Connected: true, Down: true}, 0.1); e.Down {
		t.Fatal("repeated too soon")
	}
	if e := s.Step(Pad{Connected: true, Down: true}, 0.2); !e.Down {
		t.Fatal("expected a repeat")
	}
}

func TestDisconnectClearsAHeldButton(t *testing.T) {
	var s PadState
	s.Step(Pad{Connected: true, South: true}, 0.01)
	s.Step(Pad{}, 0.01)
	e := s.Step(Pad{Connected: true, South: true}, 0.01)
	if !e.Confirm {
		t.Fatal("reconnect did not see a new press")
	}
}

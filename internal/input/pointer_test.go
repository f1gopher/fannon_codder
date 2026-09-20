package input

import "testing"

func TestChordGrenadeRightHeldLeftJustPressed(t *testing.T) {
	var tr Tracker
	// Hold right.
	p := tr.Update(10, 10, false, true)
	if p.ChordGrenade {
		t.Fatal("chord on right-only tick")
	}
	// Click left while right still down.
	p = tr.Update(10, 10, true, true)
	if !p.ChordGrenade {
		t.Fatal("expected chord when left is pressed while right is held")
	}
	// Still holding both: not a new press.
	p = tr.Update(10, 10, true, true)
	if p.ChordGrenade {
		t.Fatal("chord should fire only on the left-down tick")
	}
}

func TestLeftDownEdge(t *testing.T) {
	var tr Tracker
	p := tr.Update(0, 0, true, false)
	if !p.LeftDown || !p.Left {
		t.Fatalf("first press: LeftDown=%v Left=%v", p.LeftDown, p.Left)
	}
	p = tr.Update(0, 0, true, false)
	if p.LeftDown {
		t.Fatal("held left should not be LeftDown")
	}
	p = tr.Update(0, 0, false, false)
	if p.Left || p.LeftDown {
		t.Fatal("release should clear left")
	}
}

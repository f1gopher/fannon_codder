package app

import "testing"

func TestIntegerScaleDefaultWindow(t *testing.T) {
	scale, ox, oy := integerScale(DefaultWindowWidth, DefaultWindowHeight, ScreenWidth, ScreenHeight)
	if scale != 3 {
		t.Fatalf("scale = %d, want 3", scale)
	}
	if ox != 0 || oy != 0 {
		t.Fatalf("offset = (%d,%d), want (0,0)", ox, oy)
	}
}

func TestIntegerScaleLetterbox(t *testing.T) {
	// Wide window: 1600×768 → 3× with horizontal bars.
	scale, ox, oy := integerScale(1600, 768, ScreenWidth, ScreenHeight)
	if scale != 3 {
		t.Fatalf("scale = %d, want 3", scale)
	}
	if ox != (1600-320*3)/2 {
		t.Fatalf("offsetX = %d, want %d", ox, (1600-320*3)/2)
	}
	if oy != 0 {
		t.Fatalf("offsetY = %d, want 0", oy)
	}
}

func TestIntegerScaleNeverFractional(t *testing.T) {
	scale, _, _ := integerScale(700, 500, ScreenWidth, ScreenHeight)
	if scale != 1 {
		t.Fatalf("scale = %d, want 1 (floor of 700/320 and 500/256)", scale)
	}
}

func TestIntegerScaleMinimumOne(t *testing.T) {
	scale, _, _ := integerScale(100, 100, ScreenWidth, ScreenHeight)
	if scale != 1 {
		t.Fatalf("scale = %d, want 1", scale)
	}
}

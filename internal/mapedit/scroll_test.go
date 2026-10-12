package mapedit

import "testing"

func TestClampScroll(t *testing.T) {
	// Larger than the view. Negative and past-the-end both clamp.
	x, y := ClampScroll(-10, 500, 576, 416, 200, 150)
	if x != 0 || y != 266 {
		t.Fatalf("large map scroll = %v,%v", x, y)
	}
	// Smaller than the view stays at the origin.
	x, y = ClampScroll(10, 10, 100, 80, 200, 150)
	if x != 0 || y != 0 {
		t.Fatalf("small map scroll = %v,%v", x, y)
	}
	x, y = ClampScroll(-5, -8, 400, 400, 100, 100)
	if x != 0 || y != 0 {
		t.Fatalf("negative scroll = %v,%v", x, y)
	}
}

func TestInfoTop(t *testing.T) {
	// Three list rows at scale 1. The block is pinned to the bottom.
	if top := InfoTop(200, 1); top != 200-InfoHeight(1) {
		t.Fatalf("info top = %v", top)
	}
	if InfoHeight(1) != RowHeight(1)*InfoLines {
		t.Fatalf("info height = %v", InfoHeight(1))
	}
	if InfoTop(10, 1) != 0 {
		t.Fatal("short rail should pin the block at 0")
	}
}

func TestRailHit(t *testing.T) {
	row, ok := RailHit(10, 25, 220, 20, 11)
	if !ok || row != 1 {
		t.Fatalf("row = %d ok=%v", row, ok)
	}
	if _, ok := RailHit(220, 10, 220, 20, 11); ok {
		t.Fatal("viewport point hit the rail")
	}
	if _, ok := RailHit(10, 20*11, 220, 20, 11); ok {
		t.Fatal("point past the last row hit a row")
	}
}

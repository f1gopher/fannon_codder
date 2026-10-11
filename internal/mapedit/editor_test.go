package mapedit

import "testing"

func TestOpensTheFirstPhase(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(ed.phases) != 11 {
		t.Fatalf("phases = %d", len(ed.phases))
	}
	if ed.phases[0].name != "m01p01" || ed.phases[0].title != "It's a Jungle Out There" {
		t.Fatalf("first = %q %q", ed.phases[0].name, ed.phases[0].title)
	}
	if ed.err != nil {
		t.Fatal(ed.err)
	}
	if ed.grid.W != 24 || ed.grid.H != 16 {
		t.Fatalf("grid = %dx%d", ed.grid.W, ed.grid.H)
	}
	if ed.biome != "jungle" {
		t.Fatalf("terrain = %q", ed.biome)
	}
	ed.load(3)
	if ed.phases[3].name != "m03p01" || ed.biome != "snow" {
		t.Fatalf("phase 4 = %s %q", ed.phases[3].name, ed.biome)
	}
	if ed.scrollX != 0 || ed.scrollY != 0 {
		t.Fatalf("scroll = %v,%v", ed.scrollX, ed.scrollY)
	}
}

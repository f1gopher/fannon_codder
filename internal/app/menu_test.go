package app

import (
	"testing"

	"fannon-codder/internal/render"
)

func TestMenuMoveWraps(t *testing.T) {
	var m battleMenu
	m.open = true
	m.move(-1)
	if m.sel != menuQuit {
		t.Fatalf("up from the top landed on %d", m.sel)
	}
	m.move(1)
	if m.sel != menuRestart {
		t.Fatalf("down from quit landed on %d", m.sel)
	}
}

func TestMenuMarksTheSelectedRow(t *testing.T) {
	m := battleMenu{open: true, sel: menuSave, note: "Saved"}
	lines := m.lines()
	if lines[0] != "PAUSED" {
		t.Fatalf("headline %q", lines[0])
	}
	if lines[1+menuSave] != "> Save game" {
		t.Fatalf("selected row %q", lines[1+menuSave])
	}
	if lines[1+menuRestart] != "  Restart level" {
		t.Fatalf("other row %q", lines[1+menuRestart])
	}
	if lines[len(lines)-2] != "Saved" {
		t.Fatalf("note %q", lines[len(lines)-2])
	}
}

func TestMenuClickHitsTheRow(t *testing.T) {
	render.SetPictureScale(1)
	m := battleMenu{open: true, sel: menuRestart}
	found := false
	for y := 0.0; y < float64(ScreenHeight); y++ {
		row, ok := m.rowAt(float64(ScreenWidth), float64(ScreenHeight), float64(ScreenWidth)/2, y)
		if ok && row == menuLoad {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no point on the plaque lands on Load game")
	}
	if _, ok := m.rowAt(float64(ScreenWidth), float64(ScreenHeight), 1, 1); ok {
		t.Fatal("the corner is not a menu row")
	}
}

func TestRestartReturnsTheSquad(t *testing.T) {
	p := NewProgress()
	before := p.Pool.Remaining()
	b := NewBattle(p)
	if p.Pool.Remaining() >= before {
		t.Fatal("deploy should take men out of the queue")
	}
	if len(b.deployed) == 0 {
		t.Fatal("expected a deployed squad")
	}
	b.restoreSquad()
	if p.Pool.Remaining() != before {
		t.Fatalf("queue %d, want %d after restart restore", p.Pool.Remaining(), before)
	}
	if len(b.deployed) != 0 {
		t.Fatal("restore should clear the battle's squad so it cannot be returned twice")
	}
}

func TestSaveSnapshotKeepsTheSquad(t *testing.T) {
	p := NewProgress()
	b := NewBattle(p)
	s := p.saveSnapshot(b.deployed)
	if len(s.Recruits) != len(b.deployed)+p.Pool.Remaining() {
		t.Fatalf("saved queue %d, want deployed plus the live queue", len(s.Recruits))
	}
	for i, man := range b.deployed {
		if s.Recruits[i].Name != man.Name {
			t.Fatalf("saved recruit %d is %s, want %s", i, s.Recruits[i].Name, man.Name)
		}
	}
	if p.Pool.Remaining()+len(b.deployed) != len(s.Recruits) {
		t.Fatal("saving must not put the squad back into the live queue")
	}
	// The live pool still does not contain the deployed names.
	live := map[string]bool{}
	for _, man := range p.Pool.Recruits {
		live[man.Name] = true
	}
	for _, man := range b.deployed {
		if live[man.Name] {
			t.Fatalf("%s was written back into the live pool", man.Name)
		}
	}
}

package mapedit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"fannon-codder/internal/data"
)

func TestEditsTitleAndBriefing(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ed.startType(typeTitle)
	if !ed.typing() || ed.draft != "It's a Jungle Out There" {
		t.Fatalf("title draft = %q typing=%v", ed.draft, ed.typing())
	}
	ed.draft = "Night Raid"
	ed.commitType()
	if ed.typing() || ed.phase.Title != "Night Raid" || !ed.dirty {
		t.Fatalf("title = %q dirty=%v typing=%v", ed.phase.Title, ed.dirty, ed.typing())
	}
	if ed.phases[0].title != "Night Raid" {
		t.Fatalf("rail title = %q", ed.phases[0].title)
	}
	ed.startType(typeBrief)
	ed.draft = "  Cross the stream.  "
	ed.commitType()
	if ed.phase.Briefing != "Cross the stream." {
		t.Fatalf("briefing = %q", ed.phase.Briefing)
	}
	path := filepath.Join(t.TempDir(), "m01p01.json")
	if err := ed.writePhase(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var p data.Phase
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Title != "Night Raid" || p.Briefing != "Cross the stream." {
		t.Fatalf("saved title=%q briefing=%q", p.Title, p.Briefing)
	}
}

func TestCancelTypeKeepsTheCopy(t *testing.T) {
	ed, err := New()
	if err != nil {
		t.Fatal(err)
	}
	want := ed.phase.Title
	ed.startType(typeTitle)
	ed.draft = "changed"
	ed.cancelType()
	if ed.phase.Title != want || ed.dirty {
		t.Fatalf("title = %q dirty=%v", ed.phase.Title, ed.dirty)
	}
}

func TestInfoTextHit(t *testing.T) {
	top := 100.0
	rh := 10.0
	if kind, ok := InfoTextHit(top+55, top, rh); !ok || kind != typeTitle {
		t.Fatalf("title row kind=%d ok=%v", kind, ok)
	}
	if kind, ok := InfoTextHit(top+75, top, rh); !ok || kind != typeBrief {
		t.Fatalf("brief row kind=%d ok=%v", kind, ok)
	}
	if _, ok := InfoTextHit(top+10, top, rh); ok {
		t.Fatal("terrain row started text edit")
	}
}

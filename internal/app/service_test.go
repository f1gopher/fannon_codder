package app

import (
	"strings"
	"testing"

	"fannon-codder/internal/campaign"
	"fannon-codder/internal/input"
	"fannon-codder/internal/sim"
)

func TestServiceListsTheLivingThenTheDead(t *testing.T) {
	alive := []campaign.Soldier{{Name: "Jools", Rank: campaign.Private, Kills: 4}}
	dead := []campaign.Soldier{{Name: "Jops", Rank: campaign.Corporal, Kills: 1}}
	s := NewService(NewProgress(), alive, dead)
	if len(s.pages) != 2 {
		t.Fatalf("pages %d", len(s.pages))
	}
	if s.pages[0][0] != "ACHIEVEMENTS" || !strings.Contains(s.pages[0][1], "Jools") {
		t.Fatalf("achievements %#v", s.pages[0])
	}
	if s.pages[1][0] != "REMEMBRANCE" || !strings.Contains(s.pages[1][1], "Jops") {
		t.Fatalf("remembrance %#v", s.pages[1])
	}
}

func TestServiceSkipsAnEmptyPage(t *testing.T) {
	alive := []campaign.Soldier{{Name: "Stoo", Rank: campaign.Private}}
	s := NewService(nil, alive, nil)
	if len(s.pages) != 1 || s.pages[0][0] != "ACHIEVEMENTS" {
		t.Fatalf("pages %#v", s.pages)
	}
}

type switchHost struct {
	next Scene
}

func (h *switchHost) Pointer() input.Pointer       { return input.Pointer{LeftDown: true} }
func (h *switchHost) Switch(s Scene)               { h.next = s }
func (h *switchHost) Progress() *Progress          { return nil }
func (h *switchHost) PlayCues([]sim.Cue)           {}
func (h *switchHost) SetListener(float64, float64) {}
func (h *switchHost) SetEngine(bool, float64)      {}
func (h *switchHost) SetRotor(bool)                {}
func (h *switchHost) PlayKind(sim.CueKind)         {}
func (h *switchHost) Frame() (float64, float64)    { return 320, 256 }

func TestServiceClickOpensBootHill(t *testing.T) {
	s := NewService(NewProgress(), []campaign.Soldier{{Name: "Jools"}}, nil)
	h := &switchHost{}
	if err := s.Update(h); err != nil {
		t.Fatal(err)
	}
	if _, ok := h.next.(*BootHill); !ok {
		t.Fatalf("next %T", h.next)
	}
}

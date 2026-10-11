package render

import (
	"testing"

	"fannon-codder/internal/sim"
)

func TestHitHUDSplitSwitchAndAmmo(t *testing.T) {
	w := sim.NewEmpty()
	w.SpawnPlayerSquad(sim.SquadSnake, []sim.Vec2{{}, {}, {}})
	ids := append([]int(nil), w.ActiveSquad().MemberIDs...)

	kind, id := HitHUD(w, 12, float64(hudSquadTop+2))
	if kind != HitSplit || sim.SquadID(id) != sim.SquadSnake {
		t.Fatalf("active header hit = %v id %v", kind, id)
	}
	kind, _ = HitHUD(w, 4, 4)
	if kind != HitGrenade {
		t.Fatalf("grenade hit = %v", kind)
	}
	kind, _ = HitHUD(w, 30, 4)
	if kind != HitRocket {
		t.Fatalf("rocket hit = %v", kind)
	}

	var row hudMan
	found := false
	for _, m := range layoutHUD(w).men {
		if m.id == ids[2] {
			row = m
			found = true
		}
	}
	if !found {
		t.Fatal("no row for the third man")
	}
	kind, id = HitHUD(w, 4, float64(row.y+2))
	if kind != HitMember || id != ids[2] {
		t.Fatalf("name hit = %v id %v", kind, id)
	}

	w.ToggleSelect(ids[2])
	if !w.Split() {
		t.Fatal("split")
	}
	foundEagle, foundSnake := false, false
	for _, m := range layoutHUD(w).men {
		kind, id = HitHUD(w, 4, float64(m.y+2))
		if m.squad == sim.SquadEagle {
			if kind != HitMember || id != m.id {
				t.Fatalf("selected squad row hit = %v id %v", kind, id)
			}
			foundEagle = true
		}
		if m.squad == sim.SquadSnake {
			if kind != HitSquad || sim.SquadID(id) != sim.SquadSnake {
				t.Fatalf("snake row hit = %v id %v", kind, id)
			}
			foundSnake = true
		}
	}
	if !foundEagle || !foundSnake {
		t.Fatal("split should list both squads on the HUD")
	}
}

func TestHitHUDMapIcon(t *testing.T) {
	w := sim.NewEmpty()
	w.SpawnPlayerSquad(sim.SquadSnake, []sim.Vec2{{}})
	kind, _ := HitHUD(w, 4, 242)
	if kind != HitMap {
		t.Fatalf("map icon hit = %v", kind)
	}
	kind, _ = HitHUD(w, 24, 242)
	if kind != HitPause {
		t.Fatalf("pause button hit = %v", kind)
	}
	kind, _ = HitHUD(w, 4, 4)
	if kind != HitGrenade {
		t.Fatalf("grenade hit = %v", kind)
	}
}

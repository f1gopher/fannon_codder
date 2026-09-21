package render

import (
	"testing"

	"fannon-codder/internal/sim"
)

func TestHitHUDSplitSwitchAndAmmo(t *testing.T) {
	w := sim.NewEmpty()
	w.SpawnPlayerSquad(sim.SquadSnake, []sim.Vec2{{}, {}, {}})
	ids := append([]int(nil), w.ActiveSquad().MemberIDs...)

	kind, _ := HitHUD(w, 6, 6)
	if kind != HitSplit {
		t.Fatalf("logo hit = %v", kind)
	}
	kind, _ = HitHUD(w, 4, 20)
	if kind != HitGrenade {
		t.Fatalf("grenade hit = %v", kind)
	}
	kind, _ = HitHUD(w, 30, 20)
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
	kind, id := HitHUD(w, 4, float64(row.y+2))
	if kind != HitMember || id != ids[2] {
		t.Fatalf("name hit = %v id %v", kind, id)
	}

	w.ToggleSelect(ids[2])
	if !w.Split() {
		t.Fatal("split")
	}
	found = false
	for _, m := range layoutHUD(w).men {
		if m.squad != sim.SquadEagle {
			continue
		}
		kind, id = HitHUD(w, 4, float64(m.y+2))
		if kind != HitSquad || sim.SquadID(id) != sim.SquadEagle {
			t.Fatalf("eagle row hit = %v id %v", kind, id)
		}
		found = true
	}
	if !found {
		t.Fatal("eagle squad missing from the HUD")
	}
}

func TestHitHUDMapIcon(t *testing.T) {
	w := sim.NewEmpty()
	w.SpawnPlayerSquad(sim.SquadSnake, []sim.Vec2{{}})
	kind, _ := HitHUD(w, 4, 242)
	if kind != HitMap {
		t.Fatalf("map icon hit = %v", kind)
	}
	kind, _ = HitHUD(w, 4, 20)
	if kind != HitGrenade {
		t.Fatalf("grenade hit = %v", kind)
	}
}

package sim

import "testing"

func TestAmmoIconSelectsSpecialUntilMenAreMarked(t *testing.T) {
	w := NewEmpty()
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}})
	if w.Special != SpecialGrenade {
		t.Fatal("grenades are the default special")
	}
	w.UseAmmoIcon(SpecialRocket)
	if w.Special != SpecialRocket {
		t.Fatal("clicking the rocket icon should select it")
	}
	w.ToggleSelect(w.ActiveSquad().MemberIDs[1])
	w.UseAmmoIcon(SpecialGrenade)
	if w.GrenadeShare != ShareHalf {
		t.Fatalf("marked men should cycle grenade share, got %v", w.GrenadeShare)
	}
	if w.Special != SpecialRocket {
		t.Fatal("a split-share click must leave the selected special alone")
	}
}

func TestBazookaSelectionDoesNotThrowGrenade(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	w.ActiveSquad().Grenades = 2
	w.Special = SpecialRocket
	if w.ThrowGrenade(40, 8) {
		t.Fatal("bazooka selected: do not throw a grenade")
	}
	if w.ActiveSquad().Grenades != 2 || len(w.Grenades) != 0 {
		t.Fatal("grenade ammo should be unchanged")
	}
	w.Special = SpecialGrenade
	if !w.ThrowGrenade(40, 8) {
		t.Fatal("grenade selected: the leader should throw")
	}
}

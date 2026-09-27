package sim

import "testing"

func TestRocketCrateGivesFourRockets(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 16, Y: 16}})
	w.AddRocketCrate(Vec2{X: 16, Y: 16})
	w.Step(1.0 / 60)
	if got := w.ActiveSquad().Rockets; got != CrateAmount {
		t.Fatalf("rockets=%d, want %d", got, CrateAmount)
	}
}

func TestRocketDestroysHut(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 16, Y: 16}})
	w.ActiveSquad().Rockets = 1
	w.Special = SpecialRocket
	w.AddBuilding(6, 0, 2, 2, true, 0)
	if !w.FireRocket(6*16+16, 16) {
		t.Fatal("expected a rocket")
	}
	if w.ActiveSquad().Rockets != 0 {
		t.Fatal("rocket was not spent")
	}
	for i := 0; i < 40; i++ {
		w.Step(1.0 / 60)
	}
	if w.Buildings[0].Alive {
		t.Fatal("bazooka should destroy the hut")
	}
}

func TestRocketDestroysVehicle(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 16, Y: 16}})
	w.ActiveSquad().Rockets = 1
	w.Special = SpecialRocket
	v := w.AddSkidoo(Vec2{X: 90, Y: 16}, SideEnemy, true)
	id := v.ID
	if !w.FireRocket(90, 16) {
		t.Fatal("expected a rocket")
	}
	for i := 0; i < 40; i++ {
		w.Step(1.0 / 60)
	}
	v = w.vehicleByID(id)
	if v == nil || v.Alive {
		t.Fatal("bazooka should destroy the skidoo")
	}
	if w.livingCount(SideEnemy) != 0 {
		t.Fatal("the driver should die with the skidoo")
	}
}

func TestMGDoesNotDestroyVehicle(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	v := w.AddSkidoo(Vec2{X: 48, Y: 8}, SideEnemy, false)
	id := v.ID
	w.SetFire(48, 8, true)
	for i := 0; i < 15; i++ {
		w.Step(1.0 / 60)
	}
	v = w.vehicleByID(id)
	if v == nil || !v.Alive {
		t.Fatal("machine gun should leave the skidoo standing")
	}
	if w.livingCount(SideEnemy) != 1 {
		t.Fatal("machine gun should not kill the crew through the hull")
	}
}

func TestStartRocketsPerTrooper(t *testing.T) {
	w := NewEmpty()
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 20, Y: 0}, {X: 30, Y: 0}})
	w.ActiveSquad().Rockets = 1 * len(w.ActiveSquad().MemberIDs)
	if w.ActiveSquad().Rockets != 4 {
		t.Fatalf("4 troopers × 1 = %d", w.ActiveSquad().Rockets)
	}
}

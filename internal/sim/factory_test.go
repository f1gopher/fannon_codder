package sim

import "testing"

func TestFactoryAndComputerTakeARocketOnly(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = []Objective{DestroyFactory, DestroyComputer, DestroyEnemyBuildings}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 200}})
	w.AddFactory(0, 0, 0, 0)
	w.AddComputer(8, 0, 0, 0)
	w.AddDoorHut(12, 0)
	w.Buildings[2].SpawnEvery = 0
	fac := &w.Buildings[0]
	comp := &w.Buildings[1]
	if fac.Kind != HutFactory || fac.HasDoor || fac.W != 4*float64(TileSize) || fac.H != 3*float64(TileSize) {
		t.Fatalf("factory %+v", fac)
	}
	if comp.Kind != HutComputer || comp.HasDoor || comp.W != float64(TileSize) || comp.H != float64(TileSize) {
		t.Fatalf("computer %+v", comp)
	}
	if !w.unitHitsBuilding(fac.X+8, fac.Y+8, unitHalf) {
		t.Fatal("the factory should block a man")
	}

	w.explode(fac.X+fac.W/2, fac.Y+fac.H/2, 0)
	w.explode(comp.X+4, comp.Y+4, 0)
	if !w.Buildings[0].Alive || !w.Buildings[1].Alive {
		t.Fatal("a grenade should leave both standing")
	}
	clanks := 0
	for _, c := range w.TakeCues() {
		if c.Kind == CueClank {
			clanks++
		}
	}
	if clanks != 2 {
		t.Fatalf("clanks=%d", clanks)
	}
	if w.Status != Playing {
		t.Fatal("a clank should not finish the phase")
	}

	fireRocketAt(t, w, fac.X+fac.W/2, fac.Y+fac.H/2)
	if w.Buildings[0].Alive {
		t.Fatal("a rocket should destroy the factory")
	}
	if len(w.Roofs) != 1 || w.Roofs[0].Kind != HutFactory {
		t.Fatalf("factory roof %+v", w.Roofs)
	}
	if !w.Buildings[1].Alive || !w.Buildings[2].Alive {
		t.Fatal("the computer and the door hut should still stand")
	}
	for i := 0; i < 90; i++ {
		w.Step(1.0 / 60)
	}
	if w.Status != Playing {
		t.Fatal("one factory down should not win while the computer and the door remain")
	}

	fireRocketAt(t, w, comp.X+8, comp.Y+8)
	if w.Buildings[1].Alive {
		t.Fatal("a rocket should destroy the computer")
	}
	if len(w.Roofs) != 1 {
		t.Fatal("the computer should not throw a roof")
	}
	if w.anyKind(HutFactory) || w.anyKind(HutComputer) {
		t.Fatal("both targets should be gone")
	}
	if !w.anyDoorBuilding() {
		t.Fatal("the door hut is not a factory or a computer")
	}
	for i := 0; i < 90; i++ {
		w.Step(1.0 / 60)
	}
	if w.Status == Won {
		t.Fatal("destroy_enemy_buildings is still open")
	}
}

func fireRocketAt(t *testing.T, w *World, x, y float64) {
	t.Helper()
	w.ActiveSquad().Rockets = 1
	w.Special = SpecialRocket
	if !w.FireRocket(x, y) {
		t.Fatal("expected a rocket")
	}
	for i := 0; i < 80; i++ {
		w.Step(1.0 / 60)
	}
}

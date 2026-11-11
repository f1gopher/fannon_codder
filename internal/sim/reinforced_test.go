package sim

import "testing"

func TestReinforcedDoorTakesARocketOnly(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = []Objective{DestroyEnemyBuildings}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 16, Y: 200}})
	w.AddReinforced(2, 0, 2, 2, SpawnEvery)
	b := &w.Buildings[0]
	if !b.HasDoor || b.Kind != HutReinforced {
		t.Fatalf("kind %+v", b)
	}
	b.SpawnCD = 0
	before := w.livingCount(SideEnemy)
	w.Step(1.0 / 60)
	if w.livingCount(SideEnemy) != before+1 {
		t.Fatal("a reinforced door should spawn a grunt")
	}

	w.explode(b.X+b.W/2, b.Y+b.H/2, 0)
	if !w.Buildings[0].Alive {
		t.Fatal("a grenade should leave the door standing")
	}
	clanks := 0
	booms := 0
	for _, c := range w.TakeCues() {
		if c.Kind == CueClank {
			clanks++
		}
		if c.Kind == CueBoom {
			booms++
		}
	}
	if clanks != 1 || booms != 1 {
		t.Fatalf("grenade cues clank=%d boom=%d", clanks, booms)
	}

	w.AddBarrel(Vec2{X: b.X + 4, Y: b.Y + 4})
	w.Barrels[0].Alive = false
	w.explode(b.X+4, b.Y+4, 0)
	if !w.Buildings[0].Alive {
		t.Fatal("a barrel should leave the door standing")
	}
	clanks = 0
	for _, c := range w.TakeCues() {
		if c.Kind == CueClank {
			clanks++
		}
	}
	if clanks != 1 {
		t.Fatalf("barrel clanks=%d", clanks)
	}

	w.SetFire(b.X+8, b.Y+8, true)
	for i := 0; i < 20; i++ {
		w.Step(1.0 / 60)
	}
	w.SetFire(0, 0, false)
	if !w.Buildings[0].Alive {
		t.Fatal("machine-gun fire should leave the door standing")
	}

	w.ActiveSquad().Rockets = 1
	w.Special = SpecialRocket
	if !w.FireRocket(b.X+b.W/2, b.Y+b.H/2) {
		t.Fatal("expected a rocket")
	}
	for i := 0; i < 80; i++ {
		w.Step(1.0 / 60)
	}
	if w.Buildings[0].Alive {
		t.Fatal("a rocket should destroy the door")
	}
	if w.anyDoorBuilding() {
		t.Fatal("a destroyed reinforced door should clear destroy_enemy_buildings")
	}
	clanks = 0
	for _, c := range w.TakeCues() {
		if c.Kind == CueClank {
			clanks++
		}
	}
	if clanks != 0 {
		t.Fatal("a destroying rocket should not clank")
	}
	if len(w.Roofs) != 1 || w.Roofs[0].Kind != HutReinforced {
		t.Fatalf("roof %+v", w.Roofs)
	}
}

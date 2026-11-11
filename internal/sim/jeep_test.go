package sim

import "testing"

func TestJeepUsesTheSkidooAndItsOwnEngine(t *testing.T) {
	w := NewEmpty()
	j := w.AddJeep(Vec2{X: 40, Y: 40}, SidePlayer, true)
	s := w.AddSkidoo(Vec2{X: 200, Y: 40}, SideEnemy, true)
	if j.Kind != KindJeep || s.Kind != KindSkidoo {
		t.Fatalf("kinds %s %s", j.Kind, s.Kind)
	}
	if j.Capacity != VehicleCapacity || !j.Armed || !j.Alive {
		t.Fatal("a jeep should keep the skidoo gun, seats, and hull")
	}
	plain := w.AddJeep(Vec2{X: 40, Y: 80}, SidePlayer, false)
	if plain.Armed {
		t.Fatal("an unarmed jeep should have no gun")
	}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	if !w.CommandBoard(j.ID) {
		t.Fatal("the squad should board the jeep")
	}
	h := w.EngineHumAt(40, 40, 320)
	if !h.Run || h.Voice != KindJeep {
		t.Fatalf("boarded jeep hum %+v", h)
	}
	w.CommandExit(j.ID)
	h = w.EngineHumAt(40, 40, 320)
	if !h.Run || h.Voice != KindSkidoo {
		t.Fatalf("nearby enemy skidoo should keep its own loop: %+v", h)
	}
}

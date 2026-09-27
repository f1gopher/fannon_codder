package sim

import "testing"

func TestHearEngineSilentUnlessAliveAndOccupied(t *testing.T) {
	const radius = 320.0
	cases := []struct {
		name                                 string
		ownAlive, ownOcc, nearAlive, nearOcc bool
		nearDist                             float64
	}{
		{"empty skidoo", true, false, false, false, 0},
		{"destroyed", false, true, false, false, 0},
		{"on foot, nothing near", false, false, false, false, 0},
		{"on foot, empty neighbour", false, false, true, false, 10},
		{"on foot, wreck nearby", false, false, false, true, 10},
		{"on foot, skidoo on the silence edge", false, false, true, true, radius},
		{"on foot, skidoo past the edge", false, false, true, true, radius + 40},
	}
	for _, tc := range cases {
		h := HearEngine(tc.ownAlive, tc.ownOcc, VehicleMaxSpeed, tc.nearAlive, tc.nearOcc, VehicleMaxSpeed, tc.nearDist, radius)
		if h.Run {
			t.Errorf("%s: loop ran", tc.name)
		}
	}
}

func TestHearEnginePitchTracksSpeed(t *testing.T) {
	idle := HearEngine(true, true, 0, true, true, VehicleMaxSpeed, 10, 320)
	mid := HearEngine(true, true, VehicleMaxSpeed/2, false, false, 0, 0, 320)
	top := HearEngine(true, true, VehicleMaxSpeed, false, false, 0, 0, 320)
	over := HearEngine(true, true, VehicleMaxSpeed*3, false, false, 0, 0, 320)
	if !idle.Run || !mid.Run || !top.Run {
		t.Fatal("a living occupied skidoo should hum at every speed, including a standstill")
	}
	if idle.Pitch != EngineIdlePitch {
		t.Fatalf("idle pitch=%v, want %v", idle.Pitch, EngineIdlePitch)
	}
	wantMid := EngineIdlePitch + (EngineTopPitch-EngineIdlePitch)*0.5
	if mid.Pitch != wantMid {
		t.Fatalf("half speed pitch=%v, want %v", mid.Pitch, wantMid)
	}
	if top.Pitch != EngineTopPitch || over.Pitch != EngineTopPitch {
		t.Fatalf("top=%v over=%v, want %v", top.Pitch, over.Pitch, EngineTopPitch)
	}
	if !(idle.Pitch < mid.Pitch && mid.Pitch < top.Pitch) {
		t.Fatal("pitch should rise from idle through half to full speed")
	}
	// The listener's own vehicle wins over a faster one nearby.
	if idle.Pitch == EngineTopPitch {
		t.Fatal("a nearby skidoo should not replace the one the listener is riding")
	}
}

func TestHearEngineNearestInsideRadius(t *testing.T) {
	h := HearEngine(false, false, 0, true, true, VehicleMaxSpeed, 319, 320)
	if !h.Run || h.Pitch != EngineTopPitch {
		t.Fatalf("skidoo inside the radius: %+v", h)
	}
	slow := HearEngine(true, false, 0, true, true, 0, 40, 320)
	if !slow.Run || slow.Pitch != EngineIdlePitch {
		t.Fatalf("empty own vehicle falls through to the neighbour: %+v", slow)
	}
}

func TestEngineHumFollowsBoardExitAndTheNearestSkidoo(t *testing.T) {
	const radius = 320.0
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	v := w.AddSkidoo(Vec2{X: 40, Y: 40}, SidePlayer, true)
	if w.EngineHumAt(40, 40, radius).Run {
		t.Fatal("an empty skidoo is silent")
	}
	if !w.CommandBoard(v.ID) {
		t.Fatal("board")
	}
	h := w.EngineHumAt(40, 40, radius)
	if !h.Run || h.Pitch != EngineIdlePitch {
		t.Fatalf("just boarded: %+v, want an idle hum", h)
	}
	v.VX = VehicleMaxSpeed
	h = w.EngineHumAt(40, 40, radius)
	if h.Pitch != EngineTopPitch {
		t.Fatalf("full speed pitch=%v, want %v", h.Pitch, EngineTopPitch)
	}
	v.VX = VehicleMaxSpeed / 2
	h = w.EngineHumAt(40, 40, radius)
	want := EngineIdlePitch + (EngineTopPitch-EngineIdlePitch)*0.5
	if h.Pitch != want {
		t.Fatalf("sliding at half speed pitch=%v, want %v", h.Pitch, want)
	}

	vid := v.ID
	enemy := w.AddSkidoo(Vec2{X: 40, Y: 80}, SideEnemy, false)
	enemy.VX = VehicleMaxSpeed
	v = w.vehicleByID(vid)
	v.VX = 0
	h = w.EngineHumAt(40, 40, radius)
	if !h.Run || h.Pitch != EngineIdlePitch {
		t.Fatalf("player's skidoo should win over the enemy: %+v", h)
	}
	if !w.CommandExit(v.ID) {
		t.Fatal("exit")
	}
	h = w.EngineHumAt(40, 40, radius)
	if !h.Run || h.Pitch != EngineTopPitch {
		t.Fatalf("after dismount the enemy 40px away should hum: %+v", h)
	}

	w.destroyVehicle(enemy)
	if w.EngineHumAt(40, 40, radius).Run {
		t.Fatal("dismounted, with the other skidoo destroyed, should be silent")
	}

	far := w.AddSkidoo(Vec2{X: 40 + radius, Y: 40}, SideEnemy, false)
	far.VX = VehicleMaxSpeed
	if w.EngineHumAt(40, 40, radius).Run {
		t.Fatal("a skidoo at the silence edge is out of the loop")
	}
}

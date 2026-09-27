package sim

import "testing"

func TestPlayerShotEmitsOneGunCue(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}}, Vec2{X: 200, Y: 200})
	w.SetFire(80, 0, true)
	w.Step(1.0 / 60)
	cues := w.TakeCues()
	if len(cues) != 1 {
		t.Fatalf("cues=%d, want 1", len(cues))
	}
	if cues[0].Kind != CueGun {
		t.Fatalf("kind=%v, want gun", cues[0].Kind)
	}
	// Muzzle is half a trooper in front of the unit, along the aim.
	if cues[0].X != float64(UnitSize)/2 || cues[0].Y != 0 {
		t.Fatalf("cue at (%v,%v), want the muzzle", cues[0].X, cues[0].Y)
	}
	w.Step(1.0 / 60)
	if again := w.TakeCues(); len(again) != 0 {
		t.Fatalf("cooldown frame emitted %d cues", len(again))
	}
}

func TestEachLivingTrooperEmitsAGunCue(t *testing.T) {
	w := gunWorld([]Vec2{{X: 0, Y: 0}, {X: 0, Y: 12}}, Vec2{X: 200, Y: 200})
	w.SetFire(80, 0, true)
	w.Step(1.0 / 60)
	if len(w.TakeCues()) != 2 {
		t.Fatal("each living trooper should emit one gun cue")
	}
}

func TestVehicleGunEmitsCue(t *testing.T) {
	w := NewEmpty()
	v := w.AddSkidoo(Vec2{X: 10, Y: 20}, SidePlayer, true)
	w.vehicleShoot(v, 80, 20, 1)
	cues := w.TakeCues()
	if len(cues) != 1 || cues[0].Kind != CueGun {
		t.Fatalf("cues=%v, want one gun", cues)
	}
	if cues[0].X != 20 || cues[0].Y != 20 {
		t.Fatalf("mounted muzzle (%v,%v), want (20,20)", cues[0].X, cues[0].Y)
	}
	if len(w.TakeCues()) != 0 {
		t.Fatal("TakeCues should empty the queue")
	}
}

func TestBlastsDoNotEmitGunCues(t *testing.T) {
	w := NewEmpty()
	w.explode(8, 8, 0)
	w.launchGrenade(&Unit{X: 0, Y: 0}, 40, 0)
	w.launchRocket(0, SidePlayer, 0, 0, 40, 0)
	if len(w.Cues) != 0 {
		t.Fatalf("non-gun actions emitted %d cues", len(w.Cues))
	}
}

func TestCueQueueDropsOldest(t *testing.T) {
	w := NewEmpty()
	n := maxCues + 5
	for i := 0; i < n; i++ {
		w.emit(CueGun, float64(i), 0)
	}
	if len(w.Cues) != maxCues {
		t.Fatalf("len=%d, want %d", len(w.Cues), maxCues)
	}
	if w.Cues[0].X != 5 || w.Cues[maxCues-1].X != float64(n-1) {
		t.Fatalf("queue spans %v..%v", w.Cues[0].X, w.Cues[maxCues-1].X)
	}
}

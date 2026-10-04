package sim

import (
	"math"
	"testing"
)

func TestSettlePlaysOutShotsAndBlasts(t *testing.T) {
	w := quietWorld()
	w.Status = Won
	shooter := w.SpawnUnit(SidePlayer, Vec2{X: 10, Y: 10})
	w.addMG(shooter.X, shooter.Y, 0, 80, shooter.ID, shooter.Side)
	shot := shooter.SinceShot
	w.Explosions = append(w.Explosions, Explosion{X: 40, Y: 10, R: GrenadeRadius})
	start := len(w.Projectiles)
	const dt = 0.05
	w.Settle(dt)
	if len(w.Projectiles) != start {
		t.Fatalf("projectiles %d, want the shot already in the air to remain", len(w.Projectiles))
	}
	if w.Projectiles[0].X <= 10 {
		t.Fatal("the round should keep travelling")
	}
	if w.Explosions[0].Age != dt {
		t.Fatalf("blast age %v, want %v", w.Explosions[0].Age, dt)
	}
	if shooter.SinceShot != shot+dt {
		t.Fatalf("SinceShot %v, want %v", shooter.SinceShot, shot+dt)
	}
}

func TestSettleDoesNothingWhilePlaying(t *testing.T) {
	w := quietWorld()
	w.Explosions = append(w.Explosions, Explosion{X: 4, Y: 4, R: 8})
	w.Settle(0.1)
	if w.Explosions[0].Age != 0 {
		t.Fatalf("age %v while the phase is still playing", w.Explosions[0].Age)
	}
}

func TestSettleLandsAGrenadeWithoutNewFire(t *testing.T) {
	w := quietWorld()
	w.AI = true
	w.Status = Won
	player := w.SpawnUnit(SidePlayer, Vec2{X: 0, Y: 0})
	grunt := w.SpawnUnit(SideEnemy, Vec2{X: 30, Y: 0})
	grunt.Aggressive = true
	grunt.SpotT = 10
	grunt.ReactAt = 0.1
	grunt.Facing = 0
	w.launchGrenade(player, 70, 0)
	dur := w.Grenades[0].Dur
	w.Settle(dur - 0.02)
	if len(w.Grenades) != 1 {
		t.Fatalf("grenades %d, want the bomb still in the air", len(w.Grenades))
	}
	w.Settle(0.02)
	if len(w.Grenades) != 0 {
		t.Fatalf("grenades still in the air: %d", len(w.Grenades))
	}
	if len(w.Explosions) != 1 {
		t.Fatalf("explosions %d, want the landing blast", len(w.Explosions))
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("a finished phase should not start a new burst")
	}
	if grunt.X != 30 {
		t.Fatalf("grunt walked to %v", grunt.X)
	}
}

func TestSettleReleasesAStartedWindupOnly(t *testing.T) {
	w := quietWorld()
	w.Status = Won
	w.SpawnUnit(SidePlayer, Vec2{X: 0, Y: 0})
	ready := w.SpawnUnit(SideEnemy, Vec2{X: 60, Y: 0})
	ready.Kind = KindGrenadier
	ready.Bombs = GrenadierBombs
	ready.Facing = 0
	winding := w.SpawnUnit(SideEnemy, Vec2{X: 60, Y: 40})
	winding.Kind = KindGrenadier
	winding.Bombs = GrenadierBombs
	winding.Facing = math.Atan2(-winding.Y, -winding.X)
	winding.GrenadeWind = 0.05
	w.Settle(0.05)
	if len(w.Grenades) != 1 {
		t.Fatalf("grenades %d, want only the windup that had started", len(w.Grenades))
	}
	if winding.Bombs != GrenadierBombs-1 {
		t.Fatalf("winding bombs %d", winding.Bombs)
	}
	if ready.Bombs != GrenadierBombs {
		t.Fatalf("idle grenadier threw, bombs %d", ready.Bombs)
	}
}

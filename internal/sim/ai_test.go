package sim

import (
	"math"
	"testing"
)

func TestEnemyWaitsOutReactionBeforeFiring(t *testing.T) {
	// Player is due east, which is the spawn facing, so only the clock waits.
	w := aiWorld(Vec2{X: 40, Y: 0}, Vec2{X: 0, Y: 0})
	e := enemyOf(w)
	if e.Facing != 0 {
		t.Fatalf("spawn facing %v, want east", e.Facing)
	}
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("no round on the first frame")
	}
	w.Step(0.3 - 1.0/60)
	if len(w.Projectiles) != 0 {
		t.Fatal("no round at 0.3s while already facing")
	}
	if e.SpotT < 0.25 || e.ReactAt == 0 {
		t.Fatalf("contact should be counting, SpotT=%v ReactAt=%v", e.SpotT, e.ReactAt)
	}
	// Reach the threshold on a short frame. A long step would land the round
	// and remove it before the test can see it.
	if gap := e.ReactAt - e.SpotT - 1.0/60; gap > 0 {
		w.Step(gap)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("still inside the reaction")
	}
	w.Step(1.0 / 60)
	if len(w.Projectiles) == 0 {
		t.Fatal("a round once the reaction has passed and he is facing")
	}
	if e.X != 0 || e.VX != 0 || e.VY != 0 {
		t.Fatal("he fires from the post")
	}
}

func TestEnemyFacingAwayWaitsForTheTurn(t *testing.T) {
	// Player to the west. Spawn facing is east, so the turn outlasts the reaction.
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 40, Y: 0})
	e := enemyOf(w)
	if e.Facing != 0 {
		t.Fatalf("spawn facing %v, want east", e.Facing)
	}
	// Stop a hair before the cone. π − tol is the turn that just enters tolerance.
	partial := (math.Pi-EnemyFaceTol)/EnemyTurnRate - 0.02
	w.Step(partial)
	if len(w.Projectiles) != 0 {
		t.Fatal("facing the wrong way blocks the first round")
	}
	if e.SpotT < e.ReactAt {
		t.Fatalf("reaction should already be done, SpotT=%v ReactAt=%v", e.SpotT, e.ReactAt)
	}
	if err := facingError(e, 0, 0); err <= EnemyFaceTol {
		t.Fatalf("turn should still be outside tolerance, err=%v", err)
	}
	if e.X != 40 || e.VX != 0 {
		t.Fatal("he holds the post while turning")
	}
	w.Step(0.08)
	if len(w.Projectiles) == 0 {
		t.Fatal("once he is facing the player he fires")
	}
}

func TestUnspottedGruntHoldsPost(t *testing.T) {
	for _, x := range []float64{110, 200} {
		w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: x, Y: 0})
		e := enemyOf(w)
		w.Step(0.5)
		if e.X != x || e.VX != 0 || e.VY != 0 {
			t.Fatalf("at %vpx the grunt should hold, x=%v vx=%v", x, e.X, e.VX)
		}
		if len(w.Projectiles) != 0 || e.SpotT != 0 || e.ReactAt != 0 {
			t.Fatalf("at %vpx there is no contact", x)
		}
	}
}

func TestBrokenLOSResetsReaction(t *testing.T) {
	w := aiWorld(Vec2{X: 40, Y: 0}, Vec2{X: 0, Y: 0})
	w.Step(0.35)
	e := enemyOf(w)
	if e.SpotT <= 0 || e.ReactAt == 0 {
		t.Fatal("contact should start the clock")
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("0.35s is inside every reaction window")
	}
	w.Map = Map{W: 4, H: 1, Tiles: []Tile{
		TileGrass, TileTree, TileGrass, TileGrass,
	}}
	w.Step(1.0 / 60)
	if e.SpotT != 0 || e.ReactAt != 0 {
		t.Fatalf("LOS break should zero the clock, SpotT=%v ReactAt=%v", e.SpotT, e.ReactAt)
	}
	w.Map = Map{}
	w.Step(0.35)
	if len(w.Projectiles) != 0 {
		t.Fatal("a new contact owes a full reaction")
	}
	if e.SpotT > 0.4 {
		t.Fatalf("SpotT=%v, clock should have restarted", e.SpotT)
	}
}

func TestKillAllEnemiesWins(t *testing.T) {
	w := NewDemoWorld()
	w.Spread = 0
	if w.Status != Playing {
		t.Fatal("demo should start Playing")
	}
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy {
			w.kill(&w.Units[i])
		}
	}
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatalf("status=%v, want Won", w.Status)
	}
}

func TestAllPlayersDeadLoses(t *testing.T) {
	w := NewDemoWorld()
	for i := range w.Units {
		if w.Units[i].Side == SidePlayer {
			w.kill(&w.Units[i])
		}
	}
	w.evaluateObjectives()
	if w.Status != Lost {
		t.Fatalf("status=%v, want Lost", w.Status)
	}
}

func TestDemoHasThreeGrunts(t *testing.T) {
	w := NewDemoWorld()
	n := 0
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy && w.Units[i].Living() {
			n++
		}
	}
	if n != 3 {
		t.Fatalf("got %d grunts, want 3", n)
	}
}

func TestGrenadierTelegraphsThenThrows(t *testing.T) {
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 60, Y: 0})
	e := enemyOf(w)
	e.Kind = KindGrenadier
	e.Bombs = GrenadierBombs
	e.GrenadeCD = 0
	startX := e.X
	w.Step(1.0 / 60)
	if len(w.Grenades) != 0 {
		t.Fatal("the bomb should wait out the telegraph")
	}
	if e.GrenadeWind <= 0 {
		t.Fatal("expected a windup")
	}
	if e.X != startX || e.VX != 0 {
		t.Fatal("a winding grenadier should hold still")
	}
	w.Step(e.GrenadeWind - 0.01)
	if len(w.Grenades) != 0 {
		t.Fatal("threw before the windup finished")
	}
	w.Step(0.02)
	if len(w.Grenades) != 1 || !w.Grenades[0].Alive {
		t.Fatalf("grenades=%d", len(w.Grenades))
	}
	if e.Bombs != GrenadierBombs-1 {
		t.Fatalf("bombs=%d", e.Bombs)
	}
	if e.GrenadeCD != GrenadierCooldown {
		t.Fatalf("cooldown=%v", e.GrenadeCD)
	}
}

func TestGrenadierThrowsAtMostTwo(t *testing.T) {
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 60, Y: 0})
	e := enemyOf(w)
	e.Kind = KindGrenadier
	e.Bombs = GrenadierBombs
	e.GrenadeCD = 0
	thrown := 0
	for n := 0; n < GrenadierBombs; n++ {
		e.GrenadeCD = 0
		e.GrenadeWind = 0
		w.Grenades = nil
		w.Step(1.0 / 60)
		if e.GrenadeWind <= 0 {
			t.Fatalf("throw %d did not telegraph", n+1)
		}
		w.Step(e.GrenadeWind)
		thrown += len(w.Grenades)
	}
	if e.Bombs != 0 {
		t.Fatalf("bombs=%d, want 0", e.Bombs)
	}
	if thrown != GrenadierBombs {
		t.Fatalf("thrown=%d, want %d", thrown, GrenadierBombs)
	}
	e.GrenadeCD = 0
	e.GrenadeWind = 0
	w.Grenades = nil
	w.Step(1.0 / 60)
	if e.GrenadeWind > 0 || len(w.Grenades) != 0 {
		t.Fatal("a grenadier with no bombs should go back to the gun")
	}
}

func TestGrenadierWaitsBetweenThrows(t *testing.T) {
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 60, Y: 0})
	e := enemyOf(w)
	e.Kind = KindGrenadier
	e.Bombs = GrenadierBombs
	e.GrenadeCD = 0
	w.Step(1.0 / 60)
	w.Step(GrenadierWindup)
	if len(w.Grenades) != 1 {
		t.Fatal("expected the first bomb")
	}
	w.Grenades = nil // the bomb would land on the squad; this test is the cooldown
	w.Step(1.0)
	if e.GrenadeWind > 0 || len(w.Grenades) != 0 {
		t.Fatal("cooldown should hold the second bomb")
	}
}

func TestGrenadierTooCloseShootsInstead(t *testing.T) {
	w := aiWorld(Vec2{X: 0, Y: 0}, Vec2{X: 20, Y: 0})
	e := enemyOf(w)
	e.Kind = KindGrenadier
	e.Bombs = GrenadierBombs
	e.GrenadeCD = 0
	e.Facing = math.Pi // already looking at the player, so the wait is the reaction
	w.Step(1.0 / 60)
	if e.GrenadeWind > 0 || len(w.Grenades) != 0 {
		t.Fatal("point-blank should not start a suicide throw")
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("point-blank MG waits out the reaction")
	}
	if gap := e.ReactAt - e.SpotT - 1.0/60; gap > 0 {
		w.Step(gap)
	}
	if len(w.Projectiles) != 0 || len(w.Grenades) != 0 {
		t.Fatal("the reaction is not finished yet")
	}
	w.Step(1.0 / 60)
	if e.GrenadeWind > 0 || len(w.Grenades) != 0 {
		t.Fatal("point-blank should still not throw")
	}
	if len(w.Projectiles) == 0 {
		t.Fatal("point-blank grenadier should use the MG after the reaction")
	}
	if e.X != 20 || e.VX != 0 {
		t.Fatal("a point-blank grenadier holds still")
	}
}

func TestEnemyDoesNotShootThroughTree(t *testing.T) {
	w := aiWorld(Vec2{X: 8, Y: 8}, Vec2{X: 72, Y: 8})
	w.Map = Map{W: 6, H: 1, Tiles: []Tile{
		TileGrass, TileGrass, TileTree, TileTree, TileGrass, TileGrass,
	}}
	start := enemyOf(w).X
	w.Step(0.5)
	e := enemyOf(w)
	if len(w.Projectiles) != 0 {
		t.Fatal("grunt in range but blocked by trees must not fire")
	}
	if e.X != start || e.VX != 0 || e.SpotT != 0 {
		t.Fatal("a blocked grunt holds the post and does not bank a reaction")
	}
}

func TestGruntFiresThreeRoundBurstThenPauses(t *testing.T) {
	if got := GunStatsFor(0).Spread; got != 0.12 {
		t.Fatalf("private spread %v, want 0.12", got)
	}
	// Spread 0 is the test default. The enemy cone must still open.
	w := aiWorld(Vec2{X: 40, Y: 0}, Vec2{X: 0, Y: 0})
	if w.Spread != 0 {
		t.Fatal("fixture should zero the player cone")
	}
	e := enemyOf(w)
	var angs []float64
	for n := 0; n < 180 && len(angs) < EnemyBurst; n++ {
		keepPlayerAlive(w)
		w.Step(1.0 / 60)
		angs = append(angs, freshEnemyMGAngles(w)...)
	}
	if len(angs) != EnemyBurst {
		t.Fatalf("got %d rounds in the first chatter, want %d (%v)", len(angs), EnemyBurst, angs)
	}
	if angs[0] == angs[1] || angs[1] == angs[2] || angs[0] == angs[2] {
		t.Fatalf("seeded cone should scatter the three rounds, got %v", angs)
	}
	for _, ang := range angs {
		if math.Abs(ang) > EnemyMGSpread+1e-9 {
			t.Fatalf("round angle %v outside ±%v", ang, EnemyMGSpread)
		}
	}
	if e.X != 0 || e.VX != 0 || e.VY != 0 {
		t.Fatal("a bursting grunt holds the post")
	}
	if e.BurstGap != EnemyBurstPause {
		t.Fatalf("pause %v, want %v", e.BurstGap, EnemyBurstPause)
	}

	// Sidestep during the gap: he keeps turning and does not shoot.
	player := playerOf(w)
	facing := e.Facing
	quiet := 0.0
	var fourth []float64
	for quiet < EnemyBurstPause+0.5 {
		keepPlayerAlive(w)
		if quiet < 0.2 {
			player.X, player.Y = 40, 40
		} else {
			player.X, player.Y = 40, 0
		}
		w.Step(1.0 / 60)
		quiet += 1.0 / 60
		if quiet < 0.2 && e.Facing <= facing {
			t.Fatal("he should keep turning through the pause")
		}
		if e.X != 0 || e.VX != 0 {
			t.Fatal("the pause does not walk him off the post")
		}
		fourth = freshEnemyMGAngles(w)
		if len(fourth) > 0 {
			break
		}
	}
	if len(fourth) != 1 {
		t.Fatalf("expected one round when the pause ended, got %d", len(fourth))
	}
	if math.Abs(quiet-EnemyBurstPause) > 1.0/60 {
		t.Fatalf("silence lasted %v, want %v", quiet, EnemyBurstPause)
	}
}

func TestBurstPauseLostContactResetsReaction(t *testing.T) {
	w := aiWorld(Vec2{X: 40, Y: 0}, Vec2{X: 0, Y: 0})
	e := enemyOf(w)
	n := 0
	for i := 0; i < 180 && n < EnemyBurst; i++ {
		keepPlayerAlive(w)
		w.Step(1.0 / 60)
		n += len(freshEnemyMGAngles(w))
	}
	if n != EnemyBurst || e.BurstGap <= 0 {
		t.Fatalf("want a live pause after %d rounds, gap=%v", n, e.BurstGap)
	}
	w.Projectiles = nil
	w.Map = Map{W: 4, H: 1, Tiles: []Tile{
		TileGrass, TileTree, TileGrass, TileGrass,
	}}
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 || e.BurstGap != 0 || e.BurstN != 0 || e.SpotT != 0 || e.ReactAt != 0 {
		t.Fatalf("broken LOS should idle him, gap=%v n=%v spot=%v react=%v shots=%d",
			e.BurstGap, e.BurstN, e.SpotT, e.ReactAt, len(w.Projectiles))
	}
	w.Map = Map{}
	for i := 0; i < 18; i++ { // 0.3s, inside every reaction window
		keepPlayerAlive(w)
		w.Step(1.0 / 60)
		if len(freshEnemyMGAngles(w)) != 0 {
			t.Fatal("a new contact owes a full reaction, not the rest of the pause")
		}
	}
	if e.X != 0 || e.VX != 0 {
		t.Fatal("he holds the post after the pause is cancelled")
	}
}

// freshEnemyMGAngles is the cone of enemy MG rounds spawned on this 1/60 step.
func freshEnemyMGAngles(w *World) []float64 {
	const dt = 1.0 / 60
	floor := EnemyMGRange - MGSpeed*dt - 1
	var out []float64
	for i := range w.Projectiles {
		p := &w.Projectiles[i]
		if p.OwnerSide != SideEnemy || p.Kind != ProjMG || !p.Alive || p.Left < floor {
			continue
		}
		out = append(out, math.Atan2(p.VY, p.VX))
	}
	return out
}

func keepPlayerAlive(w *World) {
	for i := range w.Units {
		if w.Units[i].Side == SidePlayer {
			w.Units[i].HP = Alive
		}
	}
}

func playerOf(w *World) *Unit {
	for i := range w.Units {
		if w.Units[i].Side == SidePlayer {
			return &w.Units[i]
		}
	}
	return nil
}

func aiWorld(player, enemy Vec2) *World {
	w := &World{Spread: 0, nextID: 1, rng: newRNG(), AI: true}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{player})
	w.SpawnUnit(SideEnemy, enemy)
	return w
}

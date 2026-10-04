package sim

import (
	"math"
	"testing"
)

func TestSplitThreeIntoTwoPlusOne(t *testing.T) {
	w := threeMan()
	third := w.ActiveSquad().MemberIDs[2]
	w.ActiveSquad().Grenades = 5
	w.ActiveSquad().Rockets = 4
	w.GrenadeShare = ShareHalf
	w.RocketShare = ShareAll
	if !w.ToggleSelect(third) {
		t.Fatal("could not highlight the third man")
	}
	if !w.Split() {
		t.Fatal("split failed")
	}
	snake := w.SquadByID(SquadSnake)
	eagle := w.SquadByID(SquadEagle)
	if snake == nil || eagle == nil || len(w.Squads) != 2 {
		t.Fatalf("want Snake+Eagle, got %d squads", len(w.Squads))
	}
	if len(snake.MemberIDs) != 2 || len(eagle.MemberIDs) != 1 {
		t.Fatalf("want 2+1, snake=%d eagle=%d", len(snake.MemberIDs), len(eagle.MemberIDs))
	}
	if eagle.MemberIDs[0] != third || w.Unit(third).SquadID != SquadEagle {
		t.Fatal("highlighted man should be the Eagle squad")
	}
	if !snake.Active || eagle.Active {
		t.Fatal("original squad stays active")
	}
	if snake.Grenades != 3 || eagle.Grenades != 2 {
		t.Fatalf("half of 5 grenades: stay %d go %d", snake.Grenades, eagle.Grenades)
	}
	if snake.Rockets != 0 || eagle.Rockets != 4 {
		t.Fatalf("all rockets should move: stay %d go %d", snake.Rockets, eagle.Rockets)
	}
	w.Step(1.0 / 60)
	if len(w.Squads) != 2 {
		t.Fatal("file spacing must not count as a merge")
	}
}

func TestSplitRefusesEmptyOrWholeSquad(t *testing.T) {
	w := threeMan()
	if w.Split() {
		t.Fatal("split with nobody highlighted")
	}
	for _, id := range w.ActiveSquad().MemberIDs {
		w.ToggleSelect(id)
	}
	if w.Split() {
		t.Fatal("split must leave at least one man")
	}
	if len(w.Squads) != 1 {
		t.Fatalf("squad count changed: %d", len(w.Squads))
	}
}

func TestNeverMoreThanThreeSquads(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		{X: 0, Y: 0},
		{X: FileSpacing, Y: 0},
		{X: 2 * FileSpacing, Y: 0},
		{X: 3 * FileSpacing, Y: 0},
	})
	for n := 0; n < 2; n++ {
		s := w.ActiveSquad()
		last := s.MemberIDs[len(s.MemberIDs)-1]
		w.ToggleSelect(last)
		if !w.Split() {
			t.Fatalf("split %d failed", n+1)
		}
	}
	if len(w.Squads) != 3 {
		t.Fatalf("want 3 squads, got %d", len(w.Squads))
	}
	if w.SquadByID(SquadPanther) == nil {
		t.Fatal("third squad should be Panther")
	}
	last := w.ActiveSquad().MemberIDs[len(w.ActiveSquad().MemberIDs)-1]
	w.ToggleSelect(last)
	if w.Split() {
		t.Fatal("fourth squad must be refused")
	}
	if len(w.Squads) != 3 {
		t.Fatalf("squad count %d", len(w.Squads))
	}
}

func TestMergeUnderActiveSquad(t *testing.T) {
	w := threeMan()
	guard := w.ActiveSquad().MemberIDs[2]
	w.ToggleSelect(guard)
	w.Split()
	eagle := w.SquadByID(SquadEagle)
	snake := w.SquadByID(SquadSnake)
	eagle.Grenades = 2
	snake.Grenades = 3
	// Walk the active squad onto the guard: he joins Snake, Snake stays in charge.
	g := w.Unit(guard)
	leader := w.Unit(snake.LeaderID)
	g.X, g.Y = leader.X, leader.Y
	w.Step(1.0 / 60)
	if len(w.Squads) != 1 {
		t.Fatalf("want one squad after merge, got %d", len(w.Squads))
	}
	s := w.ActiveSquad()
	if s.ID != SquadSnake || len(s.MemberIDs) != 3 {
		t.Fatalf("active Snake should hold 3 men, id=%v n=%d", s.ID, len(s.MemberIDs))
	}
	if s.Grenades != 5 {
		t.Fatalf("ammo should combine, got %d", s.Grenades)
	}
	if w.Unit(guard).SquadID != SquadSnake {
		t.Fatal("guard should be back on Snake")
	}
}

func TestMergeFollowsActiveSquad(t *testing.T) {
	w := threeMan()
	guard := w.ActiveSquad().MemberIDs[2]
	w.ToggleSelect(guard)
	w.Split()
	if !w.SetActiveSquad(SquadEagle) {
		t.Fatal("switch to Eagle")
	}
	if w.SquadByID(SquadSnake).HasDest {
		t.Fatal("left-behind squad should hold")
	}
	g := w.Unit(guard)
	other := w.Unit(w.SquadByID(SquadSnake).LeaderID)
	other.X, other.Y = g.X, g.Y
	w.Step(1.0 / 60)
	s := w.ActiveSquad()
	if s.ID != SquadEagle || len(s.MemberIDs) != 3 {
		t.Fatalf("merge should be under Eagle, id=%v n=%d", s.ID, len(s.MemberIDs))
	}
}

func TestInactiveSquadHoldsAndFires(t *testing.T) {
	w, g := parkedAt(Vec2{X: 240, Y: 100})
	guard := g.ID
	enemy := w.Unit(w.Units[len(w.Units)-1].ID)
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("a parked man owes the reaction before the first round")
	}
	stepUntilParkedFires(t, w, g)
	if len(w.Projectiles) != 1 || w.Projectiles[0].OwnerID != guard {
		t.Fatalf("shot should belong to the inactive man, got %+v", w.Projectiles)
	}
	if g.X != 200 || g.Y != 100 {
		t.Fatalf("guard moved to (%v,%v)", g.X, g.Y)
	}
	if !enemy.Living() {
		t.Fatal("the first round has not resolved yet")
	}
	for _, id := range w.SquadByID(SquadSnake).MemberIDs {
		if w.Projectiles[0].OwnerID == id {
			t.Fatal("active squad fired while the button was up")
		}
	}
}

func TestParkedPrivateWaitsOutReaction(t *testing.T) {
	w, g := parkedAt(Vec2{X: 240, Y: 100})
	if g.Facing != 0 {
		t.Fatalf("spawn facing %v, want east", g.Facing)
	}
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("no round on the first frame")
	}
	w.Step(0.3 - 1.0/60)
	if len(w.Projectiles) != 0 {
		t.Fatal("no round at 0.3s while already facing")
	}
	if g.SpotT < 0.25 || g.ReactAt == 0 {
		t.Fatalf("contact should be counting, SpotT=%v ReactAt=%v", g.SpotT, g.ReactAt)
	}
	if gap := g.ReactAt - g.SpotT - 1.0/60; gap > 0 {
		w.Step(gap)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("still inside the reaction")
	}
	w.Step(1.0 / 60)
	if len(w.Projectiles) == 0 {
		t.Fatal("a round once the reaction has passed and he is facing")
	}
	if g.X != 200 || g.VX != 0 || g.VY != 0 {
		t.Fatal("he fires from where he was left")
	}
}

func TestParkedPrivateFacingAwayWaitsForTheTurn(t *testing.T) {
	w, g := parkedAt(Vec2{X: 160, Y: 100})
	w.Step(1.0 / 60)
	if g.ReactAt == 0 {
		t.Fatal("contact should start the clock")
	}
	g.SpotT = g.ReactAt
	remain := facingError(g, 160, 100) - EnemyFaceTol
	w.Step(remain/EnemyTurnRate - 0.02)
	if len(w.Projectiles) != 0 {
		t.Fatal("facing the wrong way blocks the first round")
	}
	if err := facingError(g, 160, 100); err <= EnemyFaceTol {
		t.Fatalf("turn should still be outside tolerance, err=%v", err)
	}
	if g.X != 200 || g.VX != 0 {
		t.Fatal("he holds while turning")
	}
	w.Step(0.08)
	if len(w.Projectiles) == 0 {
		t.Fatal("once he is facing the enemy he fires")
	}
}

func TestParkedPrivatePausesAfterBurst(t *testing.T) {
	w, g := parkedAt(Vec2{X: 240, Y: 100})
	enemy := &w.Units[len(w.Units)-1]
	n := 0
	for i := 0; i < 240 && n < EnemyBurst; i++ {
		enemy.HP = Alive
		w.Step(1.0 / 60)
		n += len(w.Projectiles)
		w.Projectiles = nil
	}
	if n != EnemyBurst {
		t.Fatalf("got %d rounds in the first chatter, want %d", n, EnemyBurst)
	}
	if g.BurstGap != EnemyBurstPause {
		t.Fatalf("pause %v, want %v", g.BurstGap, EnemyBurstPause)
	}
	quiet := 0.0
	fourth := 0
	for quiet < EnemyBurstPause+0.5 {
		enemy.HP = Alive
		w.Step(1.0 / 60)
		quiet += 1.0 / 60
		fourth += len(w.Projectiles)
		w.Projectiles = nil
		if fourth > 0 {
			break
		}
	}
	if fourth != 1 {
		t.Fatalf("expected one round when the pause ended, got %d", fourth)
	}
	if math.Abs(quiet-EnemyBurstPause) > 1.0/60 {
		t.Fatalf("silence lasted %v, want %v", quiet, EnemyBurstPause)
	}
	if g.X != 200 || g.VX != 0 {
		t.Fatal("the pause does not walk him")
	}
}

func TestParkedPrivateLostLOSResetsClock(t *testing.T) {
	w, g := parkedAt(Vec2{X: 240, Y: 100})
	w.Step(0.35)
	if g.SpotT <= 0 || g.ReactAt == 0 {
		t.Fatal("contact should start the clock")
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("0.35s is inside every reaction window")
	}
	// Guard at (200, 100) is tile (12, 6). A tree on the next cell blocks east.
	const mw, mh = 20, 8
	tiles := make([]Tile, mw*mh)
	for i := range tiles {
		tiles[i] = TileGrass
	}
	tiles[6*mw+13] = TileTree
	w.Map = Map{W: mw, H: mh, Tiles: tiles}
	w.Step(1.0 / 60)
	if g.SpotT != 0 || g.ReactAt != 0 || g.BurstN != 0 || g.BurstGap != 0 {
		t.Fatalf("LOS break should zero the clock, SpotT=%v ReactAt=%v", g.SpotT, g.ReactAt)
	}
	w.Map = Map{}
	w.Step(0.35)
	if len(w.Projectiles) != 0 {
		t.Fatal("a new contact owes a full reaction")
	}
	if g.SpotT > 0.4 {
		t.Fatalf("SpotT=%v, clock should have restarted", g.SpotT)
	}
}

func TestActiveSquadFiresOnTheFirstFrame(t *testing.T) {
	w := threeMan()
	leader := w.Unit(w.ActiveSquad().LeaderID)
	w.SpawnUnit(SideEnemy, Vec2{X: 40, Y: 0})
	w.SetFire(40, 0, true)
	w.Step(1.0 / 60)
	hit := false
	for i := range w.Projectiles {
		if w.Projectiles[i].OwnerID == leader.ID {
			hit = true
		}
	}
	if !hit {
		t.Fatalf("the controlled squad fires on the frame the button is held, got %+v", w.Projectiles)
	}
}

func TestParkedCorporalOutrangesPrivate(t *testing.T) {
	if GunStatsFor(1).Range <= GunStatsFor(0).Range {
		t.Fatal("a Corporal's gun is longer than a Private's")
	}
	// 82px is past a Private (80) and inside a Corporal.
	w, g := parkedAt(Vec2{X: 282, Y: 100})
	w.Step(0.2)
	if g.SpotT != 0 || g.ReactAt != 0 || len(w.Projectiles) != 0 {
		t.Fatal("a parked Private has no contact past his own range")
	}
	g.Rank = 1
	w.Step(1.0 / 60)
	if g.SpotT <= 0 || g.ReactAt == 0 {
		t.Fatalf("the same man as a Corporal starts the clock, SpotT=%v ReactAt=%v", g.SpotT, g.ReactAt)
	}
}

func TestInactiveFireBlockedByTreeAndDeepWater(t *testing.T) {
	w := threeMan()
	guard := w.ActiveSquad().MemberIDs[2]
	w.ToggleSelect(guard)
	w.Split()
	g := w.Unit(guard)
	g.X, g.Y = 8, 8
	w.SpawnUnit(SideEnemy, Vec2{X: 40, Y: 8})
	w.Map = Map{W: 4, H: 1, Tiles: []Tile{TileGrass, TileTree, TileGrass, TileGrass}}
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("tree should block the guard's MG")
	}

	w = threeMan()
	guard = w.ActiveSquad().MemberIDs[2]
	w.ToggleSelect(guard)
	w.Split()
	g = w.Unit(guard)
	g.X, g.Y = 8, 8
	w.SpawnUnit(SideEnemy, Vec2{X: 40, Y: 8})
	w.Map = strip(4, TileWaterDeep)
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("swimming guard must not fire")
	}
}

func TestActiveSquadDoesNotAutoFire(t *testing.T) {
	w := threeMan()
	w.SpawnUnit(SideEnemy, Vec2{X: 40, Y: 0})
	w.SetFire(0, 0, false)
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 0 {
		t.Fatal("the controlled squad waits for the fire button")
	}
}

func TestSplitLeaderLeavesNextInCharge(t *testing.T) {
	w := threeMan()
	leader := w.ActiveSquad().LeaderID
	second := w.ActiveSquad().MemberIDs[1]
	w.ToggleSelect(leader)
	if !w.Split() {
		t.Fatal("split")
	}
	if w.SquadByID(SquadSnake).LeaderID != second {
		t.Fatal("next man should lead Snake")
	}
	if w.SquadByID(SquadEagle).LeaderID != leader {
		t.Fatal("old leader should lead Eagle")
	}
}

func TestAmmoShareNoneIsDefault(t *testing.T) {
	w := threeMan()
	w.ActiveSquad().Grenades = 4
	w.ToggleSelect(w.ActiveSquad().MemberIDs[1])
	w.Split()
	if w.SquadByID(SquadSnake).Grenades != 4 || w.SquadByID(SquadEagle).Grenades != 0 {
		t.Fatal("default split leaves ammo with the original squad")
	}
}

// parkedAt splits the third man onto his own squad at (200, 100), facing east,
// with one enemy at the given point. The button is up.
func parkedAt(enemy Vec2) (*World, *Unit) {
	w := threeMan()
	id := w.ActiveSquad().MemberIDs[2]
	w.ToggleSelect(id)
	if !w.Split() {
		panic("split")
	}
	g := w.Unit(id)
	g.X, g.Y = 200, 100
	g.Facing = 0
	w.SpawnUnit(SideEnemy, enemy)
	w.SetFire(0, 0, false)
	return w, g
}

func stepUntilParkedFires(t *testing.T, w *World, g *Unit) {
	t.Helper()
	for i := 0; i < 90 && len(w.Projectiles) == 0; i++ {
		w.Step(1.0 / 60)
	}
	if len(w.Projectiles) == 0 {
		t.Fatalf("parked man never fired, SpotT=%v ReactAt=%v", g.SpotT, g.ReactAt)
	}
}

func threeMan() *World {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		{X: 0, Y: 0},
		{X: FileSpacing, Y: 0},
		{X: 2 * FileSpacing, Y: 0},
	})
	return w
}

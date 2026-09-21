package sim

import "testing"

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
	w := threeMan()
	guard := w.ActiveSquad().MemberIDs[2]
	w.ToggleSelect(guard)
	w.Split()
	g := w.Unit(guard)
	g.X, g.Y = 200, 100
	enemy := w.SpawnUnit(SideEnemy, Vec2{X: 240, Y: 100})
	w.SetFire(0, 0, false)
	w.Step(1.0 / 60)
	if len(w.Projectiles) != 1 {
		t.Fatalf("guard should fire once, got %d projectiles", len(w.Projectiles))
	}
	if w.Projectiles[0].OwnerID != guard {
		t.Fatal("shot should belong to the inactive man")
	}
	if g.X != 200 || g.Y != 100 {
		t.Fatalf("guard moved to (%v,%v)", g.X, g.Y)
	}
	if !enemy.Living() {
		t.Fatal("one tick is not enough to resolve the hit")
	}
	// Active squad is far (x<=20) and not firing, so it contributes nothing.
	for _, id := range w.SquadByID(SquadSnake).MemberIDs {
		if w.Projectiles[0].OwnerID == id {
			t.Fatal("active squad fired while the button was up")
		}
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

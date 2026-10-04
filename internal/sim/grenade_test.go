package sim

import "testing"

func TestMGDoesNotDentHut(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 56, Y: 80}})
	w.AddDoorHut(2, 0) // pixels 32,0 size 32 → centre 48,16
	b := &w.Buildings[0]
	hp := b.HP
	w.SetFire(b.X+b.W/2, b.Y+b.H/2, true)
	for i := 0; i < 30; i++ {
		w.Step(1.0 / 60)
	}
	if !b.Alive || b.HP != hp {
		t.Fatalf("MG dented the hut: alive=%v hp=%d", b.Alive, b.HP)
	}
}

func TestGrenadeDestroysHut(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.AddDoorHut(2, 0)
	b := &w.Buildings[0]
	cx, cy := b.X+b.W/2, b.Y+b.H/2
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: cx, Y: cy + 70}})
	w.ActiveSquad().Grenades = 1
	if !w.ThrowGrenade(cx, cy) {
		t.Fatal("throw")
	}
	for i := 0; i < 120; i++ {
		w.Step(1.0 / 60)
	}
	if w.Buildings[0].Alive || w.Buildings[0].HP != 0 {
		t.Fatal("one grenade at the hut should destroy it")
	}
	if w.ActiveSquad().Grenades != 0 {
		t.Fatal("throw should spend one grenade")
	}
}

func TestCrateGivesFourGrenades(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	w.AddGrenadeCrate(Vec2{X: 40, Y: 40})
	w.Step(1.0 / 60)
	if w.ActiveSquad().Grenades != CrateAmount {
		t.Fatalf("grenades=%d, want %d", w.ActiveSquad().Grenades, CrateAmount)
	}
	if w.Pickups[0].Alive {
		t.Fatal("crate should be gone after walking over it")
	}
}

func TestShootingCrateExplodes(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	w.AddGrenadeCrate(Vec2{X: 60, Y: 0})
	victim := w.SpawnUnit(SideEnemy, Vec2{X: 60, Y: 16})
	w.SetFire(60, 0, true)
	for i := 0; i < 20; i++ {
		w.Step(1.0 / 60)
	}
	if w.Pickups[0].Alive {
		t.Fatal("MG should detonate the crate")
	}
	if victim.Living() {
		t.Fatal("crate blast should kill the nearby grunt")
	}
	if !w.Unit(w.ActiveSquad().LeaderID).Living() {
		t.Fatal("shooter was outside the blast")
	}
	if w.ActiveSquad().Grenades != 0 {
		t.Fatal("exploding a crate must not hand out grenades")
	}
}

func TestEnemyCannotDestroyCrates(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	enemy := w.SpawnUnit(SideEnemy, Vec2{X: 0, Y: 0})
	w.AddGrenadeCrate(Vec2{X: 40, Y: 0})
	w.AddRocketCrate(Vec2{X: 40, Y: 20})
	w.Projectiles = append(w.Projectiles, Projectile{
		X: 10, Y: 0, VX: MGSpeed, VY: 0, Left: 80,
		Alive: true, OwnerID: enemy.ID, OwnerSide: SideEnemy,
	})
	w.Step(1.0 / 60)
	if !w.Pickups[0].Alive || !w.Pickups[1].Alive {
		t.Fatal("enemy MG destroyed a crate")
	}
	w.launchGrenade(enemy, 40, 0)
	for i := 0; i < 120; i++ {
		w.Step(1.0 / 60)
	}
	if !w.Pickups[0].Alive || !w.Pickups[1].Alive {
		t.Fatal("enemy grenade destroyed a crate")
	}
	w.launchRocket(enemy.ID, SideEnemy, 0, 0, 40, 0)
	for i := 0; i < 30; i++ {
		w.Step(1.0 / 60)
	}
	if !w.Pickups[0].Alive || !w.Pickups[1].Alive {
		t.Fatal("enemy rocket destroyed a crate")
	}
}

func TestGrenadeHurtsFriendlies(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}, {X: 40, Y: 0}})
	w.ActiveSquad().Grenades = 1
	follower := w.ActiveSquad().MemberIDs[1]
	if !w.ThrowGrenade(40, 0) {
		t.Fatal("throw")
	}
	for i := 0; i < 90; i++ {
		w.Step(1.0 / 60)
	}
	if w.Unit(follower).Living() {
		t.Fatal("grenade should kill a friendly in the blast")
	}
	if !w.Unit(w.ActiveSquad().LeaderID).Living() {
		t.Fatal("leader was outside the blast")
	}
}

func TestGrenadeFliesOverTrees(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	tiles := make([]Tile, 8)
	tiles[2] = TileTree
	tiles[3] = TileTree
	w.Map = Map{W: 8, H: 1, Tiles: tiles}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	enemy := w.SpawnUnit(SideEnemy, Vec2{X: 88, Y: 8})
	w.ActiveSquad().Grenades = 1
	if !w.ThrowGrenade(88, 8) {
		t.Fatal("throw")
	}
	w.Step(1.0 / 60)
	if len(w.Grenades) != 1 || !w.Grenades[0].Alive {
		t.Fatal("grenade should still be in flight over the trees")
	}
	for i := 0; i < 120; i++ {
		w.Step(1.0 / 60)
	}
	if enemy.Living() {
		t.Fatal("grenade should kill through the tree line")
	}
}

func TestDeepWaterCannotThrow(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.Map = strip(4, TileWaterDeep)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 8}})
	w.ActiveSquad().Grenades = 1
	if w.ThrowGrenade(40, 8) {
		t.Fatal("swimming leader must not throw")
	}
	if w.ActiveSquad().Grenades != 1 || len(w.Grenades) != 0 {
		t.Fatal("failed throw should keep the grenade")
	}
}

func TestSpawnerKeepsPhaseOpen(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = []Objective{KillAllEnemy, DestroyEnemyBuildings}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 200}})
	w.AddDoorHut(2, 0)
	for i := 0; i < 50; i++ {
		w.Step(1.0 / 60)
	}
	if w.livingCount(SideEnemy) < 1 {
		t.Fatal("door hut should have spawned a grunt")
	}
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy {
			w.kill(&w.Units[i])
		}
	}
	w.evaluateObjectives()
	if w.Status != Playing {
		t.Fatal("hut still standing: phase must not end")
	}
	w.Buildings[0].Alive = false
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatalf("hut down and reds dead: status=%v", w.Status)
	}
}

func TestDoorlessHutIsNotAnObjective(t *testing.T) {
	w := NewEmpty()
	w.Objectives = []Objective{DestroyEnemyBuildings}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 80}})
	w.AddBuilding(2, 0, 2, 2, false, 0)
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatal("a doorless hut should not block destroy_enemy_buildings")
	}
}

func TestHutBlocksWalking(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 64, Y: 16}})
	w.AddDoorHut(0, 0) // occupies x 0..32
	w.CommandMove(-10, 16, true)
	for i := 0; i < 90; i++ {
		w.Step(1.0 / 60)
	}
	leader := w.Unit(w.ActiveSquad().LeaderID)
	if w.unitHitsBuilding(leader.X, leader.Y, unitHalf) {
		t.Fatalf("leader walked into the hut at x=%v", leader.X)
	}
	if leader.X < 32 {
		t.Fatalf("leader passed through the hut, x=%v", leader.X)
	}
}

func TestHutWorldClears(t *testing.T) {
	w := NewHutWorld()
	w.AI = false
	w.Spread = 0
	if len(w.Buildings) != 1 || !w.Buildings[0].HasDoor {
		t.Fatal("sandbox wants one door hut")
	}
	if len(w.Pickups) != 1 || w.Pickups[0].Amount != CrateAmount {
		t.Fatal("sandbox wants one grenade crate of 4")
	}
	leader := w.Unit(w.ActiveSquad().LeaderID)
	leader.X, leader.Y = w.Pickups[0].X, w.Pickups[0].Y
	w.Step(1.0 / 60)
	if w.ActiveSquad().Grenades != CrateAmount {
		t.Fatalf("pickup grenades=%d", w.ActiveSquad().Grenades)
	}
	b := &w.Buildings[0]
	cx, cy := b.X+b.W/2, b.Y+b.H/2
	leader.X, leader.Y = cx, cy+70
	if !w.ThrowGrenade(cx, cy) {
		t.Fatal("throw")
	}
	for i := 0; i < 150; i++ {
		w.Step(1.0 / 60)
	}
	if w.Buildings[0].Alive {
		t.Fatal("grenade should destroy the sandbox hut")
	}
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy && w.Units[i].Living() {
			w.kill(&w.Units[i])
		}
	}
	w.evaluateObjectives()
	if w.Status != Won {
		t.Fatalf("status=%v, want Won", w.Status)
	}
}

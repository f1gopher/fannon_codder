package sim

import "testing"

func TestSkidooBoardAndExit(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 20, Y: 20}, {X: 10, Y: 20}})
	v := w.AddSkidoo(Vec2{X: 20, Y: 20}, SidePlayer, true)
	if !w.CommandBoard(v.ID) || !w.LeaderInVehicle() {
		t.Fatal("squad standing on the skidoo should board")
	}
	if len(v.Occupants) != 2 {
		t.Fatalf("occupants=%d, want 2", len(v.Occupants))
	}
	hover, _ := w.VehicleHover(20, 20)
	if hover != HoverExit {
		t.Fatalf("hover=%v, want exit", hover)
	}
	if !w.CommandExit(v.ID) || w.LeaderInVehicle() {
		t.Fatal("clicking the occupied skidoo should exit")
	}
	if w.vehicleOccupied(v) {
		t.Fatal("skidoo should be empty")
	}
	for _, id := range w.ActiveSquad().MemberIDs {
		u := w.Unit(id)
		if u == nil || !u.Living() || u.VehicleID != 0 {
			t.Fatal("dismounted troopers should be alive and on foot")
		}
	}
	hover, _ = w.VehicleHover(v.X, v.Y)
	if hover != HoverBoard {
		t.Fatalf("empty skidoo hover=%v, want board", hover)
	}
}

func TestSkidooGoesFasterTheLongerYouHold(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	v := w.AddSkidoo(Vec2{X: 0, Y: 0}, SidePlayer, false)
	w.CommandBoard(v.ID)
	w.SetDrive(400, 0, true)
	w.Step(1.0 / 60)
	w.SetDrive(400, 0, true)
	w.Step(1.0 / 60)
	slow := hypot(v.VX, v.VY)
	for i := 0; i < 120; i++ {
		w.SetDrive(v.X+400, 0, true)
		w.Step(1.0 / 60)
	}
	fast := hypot(v.VX, v.VY)
	if fast < slow+20 {
		t.Fatalf("speed %v then %v; holding should build speed", slow, fast)
	}
	if fast < VehicleMaxSpeed-5 {
		t.Fatalf("top speed=%v, want about %v", fast, VehicleMaxSpeed)
	}
}

func TestSkidooSkidsOnIce(t *testing.T) {
	ice := NewEmpty()
	ice.AI = false
	ice.Objectives = nil
	tiles := make([]Tile, 12*8)
	for i := range tiles {
		tiles[i] = TileIce
	}
	ice.Map = Map{W: 12, H: 8, Tiles: tiles}
	ice.SpawnPlayerSquad(SquadSnake, []Vec2{TileCenter(2, 4)})
	iv := ice.AddSkidoo(TileCenter(2, 4), SidePlayer, false)
	ice.CommandBoard(iv.ID)
	for i := 0; i < 90; i++ {
		ice.SetDrive(iv.X+300, iv.Y, true)
		ice.Step(1.0 / 60)
	}
	sliding := hypot(iv.VX, iv.VY)
	ice.SetDrive(0, 0, false)
	ice.Step(0.25)
	if hypot(iv.VX, iv.VY) < sliding*0.4 {
		t.Fatalf("ice should keep speed, had %v now %v", sliding, hypot(iv.VX, iv.VY))
	}

	grass := NewEmpty()
	grass.AI = false
	grass.Objectives = nil
	grass.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 40, Y: 40}})
	gv := grass.AddSkidoo(Vec2{X: 40, Y: 40}, SidePlayer, false)
	grass.CommandBoard(gv.ID)
	grass.SetDrive(400, 40, true)
	grass.Step(1.0 / 60)
	grass.SetDrive(0, 0, false)
	grass.Step(1.0 / 60)
	if hypot(gv.VX, gv.VY) > 1 {
		t.Fatalf("grass should stop when you release, speed=%v", hypot(gv.VX, gv.VY))
	}
}

func TestSkidooRamKillsInfantry(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	v := w.AddSkidoo(Vec2{X: 0, Y: 0}, SidePlayer, false)
	w.CommandBoard(v.ID)
	enemyID := w.SpawnUnit(SideEnemy, Vec2{X: 10, Y: 0}).ID
	friendID := w.SpawnUnit(SidePlayer, Vec2{X: 0, Y: 10}).ID
	w.SetDrive(80, 0, true)
	w.Step(1.0 / 60)
	if w.Unit(enemyID).Living() {
		t.Fatal("ram should kill infantry")
	}
	if w.Unit(friendID).Living() {
		t.Fatal("ram should kill friendlies too")
	}
	if !w.Unit(w.ActiveSquad().LeaderID).Living() {
		t.Fatal("the crew should survive their own ram")
	}
}

func TestNoGrenadesOrRocketsFromInsideSkidoo(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	w.ActiveSquad().Grenades = 2
	w.ActiveSquad().Rockets = 2
	v := w.AddSkidoo(Vec2{X: 0, Y: 0}, SidePlayer, true)
	w.CommandBoard(v.ID)
	if w.ThrowGrenade(40, 0) || w.UseSpecial(40, 0) {
		t.Fatal("no grenades from inside")
	}
	w.Special = SpecialRocket
	if w.FireRocket(40, 0) || w.UseSpecial(40, 0) {
		t.Fatal("no bazooka from inside")
	}
	if w.ActiveSquad().Grenades != 2 || w.ActiveSquad().Rockets != 2 {
		t.Fatal("ammo should be unchanged")
	}
}

func TestSkidooMountedMG(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	v := w.AddSkidoo(Vec2{X: 0, Y: 0}, SidePlayer, true)
	w.CommandBoard(v.ID)
	w.SetFire(80, 0, true)
	w.Step(1.0 / 60)
	if len(w.Projectiles) == 0 {
		t.Fatal("an armed skidoo should fire its MG")
	}

	bare := NewEmpty()
	bare.AI = false
	bare.Objectives = nil
	bare.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	bv := bare.AddSkidoo(Vec2{X: 0, Y: 0}, SidePlayer, false)
	bare.CommandBoard(bv.ID)
	bare.SetFire(80, 0, true)
	bare.Step(1.0 / 60)
	if len(bare.Projectiles) != 0 {
		t.Fatal("an unarmed skidoo should not fire")
	}
}

func TestEnemySkidooChases(t *testing.T) {
	w := NewEmpty()
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 220, Y: 0}})
	v := w.AddSkidoo(Vec2{X: 0, Y: 0}, SideEnemy, true)
	id := v.ID
	for i := 0; i < 60; i++ {
		w.Step(1.0 / 60)
	}
	v = w.vehicleByID(id)
	if v == nil || !v.Alive || v.X < 15 {
		t.Fatalf("enemy skidoo should drive toward the squad, x=%v", v.X)
	}
	if v.Blink <= 0 {
		t.Fatal("enemy skidoo should blink")
	}
}

func TestRocketeerHidesBesideATree(t *testing.T) {
	w := NewEmpty()
	w.Spread = 0
	w.Objectives = nil
	const W, H = 10, 8
	tiles := make([]Tile, W*H)
	tiles[2*W+3] = TileTree
	w.Map = Map{W: W, H: H, Tiles: tiles}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{TileCenter(4, 6)})
	e := w.SpawnUnit(SideEnemy, TileCenter(4, 2))
	e.Kind = KindRocketeer
	e.RocketCD = 0
	if !w.treeAdjacent(e) {
		t.Fatal("fixture should stand beside a tree")
	}
	x := e.X
	for i := 0; i < 20; i++ {
		w.Step(1.0 / 60)
	}
	if e.X != x || e.VX != 0 {
		t.Fatalf("rocketeer in cover should hold, x %v -> %v", x, e.X)
	}
	if e.RocketWind <= 0 && len(w.Projectiles) == 0 {
		t.Fatal("expected a telegraphed rocket")
	}
}

func TestRocketeerDoesNotChaseIntoTheOpen(t *testing.T) {
	w := NewEmpty()
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	e := w.SpawnUnit(SideEnemy, Vec2{X: 110, Y: 0})
	e.Kind = KindRocketeer
	e.RocketCD = 10
	for i := 0; i < 30; i++ {
		w.Step(1.0 / 60)
	}
	if e.X != 110 {
		t.Fatalf("a rocketeer outside his approach holds; x=%v", e.X)
	}
}

func TestRocketeerFiresSlowly(t *testing.T) {
	w := NewEmpty()
	w.Spread = 0
	w.Objectives = nil
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 0, Y: 0}})
	e := w.SpawnUnit(SideEnemy, Vec2{X: 180, Y: 0})
	e.Kind = KindRocketeer
	e.RocketCD = 0
	w.Step(1.0 / 60)
	if e.RocketWind <= 0 {
		t.Fatal("rocket should be telegraphed")
	}
	w.Step(RocketeerWindup)
	if len(w.Projectiles) != 1 || w.Projectiles[0].Kind != ProjRocket {
		t.Fatal("expected one rocket")
	}
	w.Projectiles = nil
	w.Step(1)
	if e.RocketWind > 0 || len(w.Projectiles) != 0 {
		t.Fatal("the second rocket should wait out the cooldown")
	}
}

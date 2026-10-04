package sim

import "testing"

func TestDoorGruntWalksOutWithoutFiring(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Spread = 0
	w.AddDoorHut(2, 0)
	w.Buildings[0].SpawnCD = 0
	door := w.Buildings[0].DoorSpawn()
	// Inside gun range of the step, so a posted grunt would shoot.
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: door.X + 40, Y: door.Y}})

	w.Step(1.0 / 60)
	var grunt *Unit
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy {
			grunt = &w.Units[i]
			break
		}
	}
	if grunt == nil || grunt.Kind != KindInfantry || !grunt.HasPost {
		t.Fatalf("door should spawn an infantry grunt with a post, got %+v", grunt)
	}
	id := grunt.ID
	if grunt.X != door.X || grunt.Y != door.Y {
		t.Fatalf("spawn stands on the step, got (%v,%v)", grunt.X, grunt.Y)
	}

	for i := 0; i < 120; i++ { // two seconds
		w.Step(1.0 / 60)
		grunt = w.Unit(id)
		if grunt.HasPost && len(w.Projectiles) != 0 {
			t.Fatal("a door grunt fired during the walk")
		}
		w.Projectiles = nil
	}
	if hypot(grunt.X-door.X, grunt.Y-door.Y) < float64(TileSize) {
		t.Fatalf("after two seconds he is still on the doorstep (%v,%v)", grunt.X, grunt.Y)
	}
	if grunt.HasPost {
		t.Fatal("open ground should reach the post inside two seconds")
	}
	if w.livingCount(SideEnemy) > maxDoorSpawns {
		t.Fatalf("living enemies %d over the cap", w.livingCount(SideEnemy))
	}
}

func TestMapGruntStaysOnHisTile(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.SpawnPlayerSquad(SquadSnake, []Vec2{{X: 8, Y: 200}})
	u := w.SpawnUnit(SideEnemy, Vec2{X: 80, Y: 64})
	for i := 0; i < 120; i++ {
		w.Step(1.0 / 60)
	}
	if u.X != 80 || u.Y != 64 || u.HasPost {
		t.Fatalf("a map grunt left his tile (%v,%v) post=%v", u.X, u.Y, u.HasPost)
	}
}

func TestDoorGruntSlidesAroundTreeAndPosts(t *testing.T) {
	w := NewEmpty()
	w.AI = true
	w.Spread = 0
	const mw, mh = 16, 12
	tiles := make([]Tile, mw*mh)
	for i := range tiles {
		tiles[i] = TileGrass
	}
	// Straight south of the step is a tree. The post steps east onto grass.
	tiles[3*mw+3] = TileTree
	tiles[4*mw+3] = TileTree
	tiles[5*mw+3] = TileTree
	w.Map = Map{W: mw, H: mh, Tiles: tiles}
	w.AddDoorHut(2, 0)
	w.Buildings[0].SpawnCD = 0
	door := w.Buildings[0].DoorSpawn()

	w.Step(1.0 / 60)
	var grunt *Unit
	for i := range w.Units {
		if w.Units[i].Side == SideEnemy {
			grunt = &w.Units[i]
			break
		}
	}
	if grunt == nil || !grunt.HasPost {
		t.Fatal("expected a walking door grunt")
	}
	id := grunt.ID
	if grunt.PostX == door.X && grunt.PostY == door.Y {
		t.Fatal("the post should leave the step")
	}

	for i := 0; i < 240 && grunt.HasPost; i++ {
		w.Step(1.0 / 60)
		grunt = w.Unit(id)
		if grunt.HasPost && len(w.Projectiles) != 0 {
			t.Fatal("fired while still walking to the post")
		}
	}
	if grunt.HasPost {
		t.Fatal("a blocked walk should end posted")
	}
	if grunt.X == door.X && grunt.Y == door.Y {
		t.Fatal("he never left the step")
	}
	if grunt.X == door.X {
		t.Fatalf("the tree should slide him off the door line, x=%v", grunt.X)
	}
	if grunt.VX != 0 || grunt.VY != 0 {
		t.Fatalf("posted grunt still moving (%v,%v)", grunt.VX, grunt.VY)
	}
}

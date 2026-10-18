package sim

import (
	"math"
	"testing"
)

func TestFormDropWedge(t *testing.T) {
	w := NewEmpty()
	start := Vec2{X: 160, Y: 200}
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		start,
		{X: start.X - FileSpacing, Y: start.Y},
		{X: start.X - 2*FileSpacing, Y: start.Y},
		{X: start.X - 3*FileSpacing, Y: start.Y},
	})
	w.FormDrop()
	s := w.ActiveSquad()
	got := make([]Vec2, len(s.MemberIDs))
	for i, id := range s.MemberIDs {
		u := w.Unit(id)
		got[i] = Vec2{X: u.X, Y: u.Y}
	}
	want := []Vec2{
		start,
		dropFormPoint(start.X, start.Y, 1),
		dropFormPoint(start.X, start.Y, 2),
		dropFormPoint(start.X, start.Y, 3),
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("man %d at %v, want %v", i, got[i], want[i])
		}
	}
	if s.Trail[len(s.Trail)-1] != start {
		t.Fatalf("trail should end on the leader, got %v", s.Trail)
	}
	if !s.Kneel {
		t.Fatal("the wedge should kneel until ordered")
	}
	for i := 0; i < 30; i++ {
		w.Step(1.0 / 60)
	}
	for i, id := range s.MemberIDs {
		u := w.Unit(id)
		if u.X != want[i].X || u.Y != want[i].Y {
			t.Fatalf("man %d left the wedge for %v,%v", i, u.X, u.Y)
		}
		if u.Facing != -math.Pi/2 {
			t.Fatalf("man %d facing %v, want north", i, u.Facing)
		}
	}
	w.CommandMove(start.X+80, start.Y, true)
	if s.Kneel {
		t.Fatal("a move should stand the squad up")
	}
	w.Step(1.0 / 60)
	if w.Unit(s.LeaderID).X <= start.X {
		t.Fatal("the leader should have started walking")
	}
}

func TestFormDropSkipsWater(t *testing.T) {
	w := NewEmpty()
	const n = 16
	tiles := make([]Tile, n*n)
	w.Map = Map{W: n, H: n, Tiles: tiles}
	start := TileCenter(8, 8)
	w.SpawnPlayerSquad(SquadSnake, []Vec2{
		start,
		{X: start.X - FileSpacing, Y: start.Y},
	})
	want := dropFormPoint(start.X, start.Y, 1)
	w.Map.Set(int(want.X)/TileSize, int(want.Y)/TileSize, TileWaterShallow)
	w.FormDrop()
	u := w.Unit(w.ActiveSquad().MemberIDs[1])
	if u.X != start.X-FileSpacing || u.Y != start.Y {
		t.Fatalf("water spot should keep the file, got %v,%v", u.X, u.Y)
	}
}

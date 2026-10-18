package render

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/assets/art"
	"fannon-codder/internal/sim"
)

func TestDropHeadingMatchesThePainting(t *testing.T) {
	nx, ny := dropHeading()
	if math.Abs(nx*nx+ny*ny-1) > 1e-9 {
		t.Fatalf("nose length %v", math.Hypot(nx, ny))
	}
	rx, ry := dropRight()
	if math.Abs(nx*rx+ny*ry) > 1e-9 {
		t.Fatalf("right is not perpendicular to the nose")
	}
	// The painted nose points east and south, so the flight does too.
	if nx <= 0 || ny <= 0 {
		t.Fatalf("nose (%v, %v) should run south-east", nx, ny)
	}
}

func TestSoldiersAbseilThenTheHelicopterLeaves(t *testing.T) {
	w := sim.NewDemoWorld()
	in := NewInsert(w)
	if len(in.men) != 2 {
		t.Fatalf("men %d", len(in.men))
	}
	id0 := in.men[0].id
	if in.Landed() || in.Gone() || !in.Hides(id0) {
		t.Fatal("the squad should still be aboard at the start")
	}
	gx, gy, alt := in.pose()
	hx, hy := in.hover()
	dx, dy := hx-gx, hy-gy
	dist := math.Hypot(dx, dy)
	if math.Abs(dist-dropInbound) > 0.5 {
		t.Fatalf("approach start %v, want %v", dist, dropInbound)
	}
	nx, ny := dropHeading()
	if dx/dist*nx+dy/dist*ny < 0.99 {
		t.Fatal("the helicopter does not approach along its nose")
	}
	if alt < dropAltLow {
		t.Fatal("it should still be high on the way in")
	}

	in.Step(dropApproach + dropRope + 0.05)
	_, y, out := in.feet(0)
	if !out || y >= in.men[0].y-8 {
		t.Fatalf("first man should be on the rope, y=%v ground=%v out=%v", y, in.men[0].y, out)
	}
	if _, _, out := in.feet(1); out {
		t.Fatal("the second man should still be in the cabin")
	}
	if !in.Hides(id0) || in.Landed() {
		t.Fatal("boots are not down yet")
	}

	in.t = in.landAt()
	if !in.Landed() || in.Hides(id0) || in.Gone() {
		t.Fatal("landing should put them on the marks and leave the aircraft in view")
	}
	for i := range in.men {
		x, y, out := in.feet(i)
		if !out || math.Hypot(x-in.men[i].x, y-in.men[i].y) > 0.05 {
			t.Fatalf("man %d landed at %v,%v, want %v,%v", i, x, y, in.men[i].x, in.men[i].y)
		}
	}

	in.Step(dropRopeUp + dropDepart)
	if !in.Gone() {
		t.Fatal("the helicopter should have left")
	}
	empty := NewInsert(sim.NewEmpty())
	if !empty.Landed() || !empty.Gone() {
		t.Fatal("a battle with no squad should not wait on a helicopter")
	}
}

func TestTwoStraightRopesThenAKneel(t *testing.T) {
	w := sim.NewEmpty()
	w.SpawnPlayerSquad(sim.SquadSnake, []sim.Vec2{
		{X: 80, Y: 120},
		{X: 70, Y: 120},
		{X: 60, Y: 120},
		{X: 50, Y: 120},
		{X: 40, Y: 120},
	})
	in := NewInsert(w)
	if in.ropes() != 2 {
		t.Fatalf("ropes %d, want 2", in.ropes())
	}
	seen := map[int]bool{}
	for i := range in.men {
		seen[in.ropeIndex(i)] = true
		fx, fy := in.ropeFoot(i)
		if math.Abs(fy-120) > 0.01 {
			t.Fatalf("rope %d foot y %v", i, fy)
		}
		// Every man on a rope shares that rope's x, so the line is vertical.
		for j := range in.men {
			if in.ropeIndex(j) != in.ropeIndex(i) {
				continue
			}
			ox, _ := in.ropeFoot(j)
			if math.Abs(ox-fx) > 0.01 {
				t.Fatalf("rope %d x differs: %v vs %v", in.ropeIndex(i), ox, fx)
			}
		}
	}
	if len(seen) != 2 {
		t.Fatalf("rope ids %v", seen)
	}

	in.t = in.ropeExit(0) + dropSlide
	x, y, out := in.feet(0)
	fx, fy := in.ropeFoot(0)
	if !out || math.Abs(x-fx) > 0.05 || math.Abs(y-fy) > 0.05 {
		t.Fatalf("first boot should be under the rope at %v,%v, got %v,%v", fx, fy, x, y)
	}
	if b, _ := in.beat(0); b != beatSpread {
		t.Fatalf("beat %v, want spread", b)
	}

	in.t = in.ropeExit(0) + dropSlide + dropSpread
	if b, _ := in.beat(0); b != beatKneel {
		t.Fatalf("beat %v, want kneel", b)
	}
	x, y, _ = in.feet(0)
	if math.Hypot(x-in.men[0].x, y-in.men[0].y) > 0.05 {
		t.Fatalf("kneel should be on the mark, got %v,%v", x, y)
	}

	// Halfway through the walk he has left the rope toward his mark.
	in.t = in.ropeExit(4) + dropSlide + dropSpread*0.5
	x, _, _ = in.feet(4)
	fx, _ = in.ropeFoot(4)
	if math.Abs(x-fx) < 1 || math.Abs(x-in.men[4].x) < 1 {
		t.Fatalf("man 4 should be between the rope (%v) and his mark (%v), at %v", fx, in.men[4].x, x)
	}

	alone := NewInsert(sim.NewEmpty())
	alone.men = []dropMan{{id: 1, x: 10, y: 10}}
	if alone.ropes() != 1 {
		t.Fatalf("one soldier should get one rope, got %d", alone.ropes())
	}
}

func TestChopperSheet(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	sh := lib.Get("chopper/fly")
	if sh == nil || sh.Frames != 1 || sh.FrameW != 440 || sh.FrameH != 380 {
		t.Fatalf("chopper/fly %+v", sh)
	}
	if sh.AnchorX != flyAnchorX || sh.AnchorY != flyAnchorY {
		t.Fatalf("anchor %d,%d", sh.AnchorX, sh.AnchorY)
	}
	img, _, ok := sh.image("SE", 0)
	if !ok || img == nil {
		t.Fatal("missing south-east frame")
	}
}

func TestDropPaints(t *testing.T) {
	w := sim.NewDemoWorld()
	w.Camera.ViewW = 320
	w.Camera.ViewH = 256
	w.Camera.MapW = 400
	w.Camera.MapH = 300
	w.Camera.CenterOn(80, 128)
	in := NewInsert(w)
	in.Step(dropApproach + dropRope + dropSlide*0.45)
	SetPictureScale(2)
	t.Cleanup(func() { SetPictureScale(1) })
	dst := ebiten.NewImage(640, 512)
	var chopper, soldiers int
	spriteHook = func(img *ebiten.Image, _ bool) {
		b := img.Bounds()
		if b.Dx() == 440 && b.Dy() == 380 {
			chopper++
		} else if b.Dy() > 40 && b.Dy() < 200 {
			soldiers++
		}
	}
	t.Cleanup(func() { spriteHook = nil })
	UseDrop(in)
	defer UseDrop(nil)
	Field(dst, w, "jungle")
	if chopper != 1 {
		t.Fatalf("helicopter draws %d, want 1", chopper)
	}
	if soldiers < 2 {
		t.Fatalf("men on the ropes %d, want 2", soldiers)
	}
}

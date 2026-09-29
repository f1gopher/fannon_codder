package render

import (
	"math"
	"sort"
	"testing"

	"fannon-codder/internal/sim"
)

func TestSpriteScale(t *testing.T) {
	restoreRender(t)
	SetPictureScale(3)
	if math.Abs(SpriteScale()-0.375) > 1e-9 {
		t.Fatalf("scale at S=3 is %v, want 0.375", SpriteScale())
	}
	SetPictureScale(2160.0 / 256.0)
	want := (2160.0 / 256.0) / ArtScale
	if math.Abs(SpriteScale()-want) > 1e-9 {
		t.Fatalf("scale at 4K is %v, want %v", SpriteScale(), want)
	}
}

func TestPoseFirstMatch(t *testing.T) {
	u := &sim.Unit{HP: sim.Alive, SinceShot: 1, SinceThrow: 1}
	if got := unitPose(u, 0, 0); got != poseIdle {
		t.Fatalf("standing = %s", got.name())
	}
	u.VX = walkMin
	if got := unitPose(u, 0, 0); got != poseIdle {
		t.Fatalf("speed %v = %s, want idle", walkMin, got.name())
	}
	u.VX = walkMin + 0.01
	if got := unitPose(u, 0, 0); got != poseWalk {
		t.Fatalf("walking = %s", got.name())
	}
	u.VX = 30
	u.SinceShot = 0
	if got := unitPose(u, 0, 0); got != poseShoot {
		t.Fatalf("shot while moving = %s", got.name())
	}
	u.SinceShot = shotWindow
	if got := unitPose(u, 0, 0); got != poseWalk {
		t.Fatalf("shot window closed = %s", got.name())
	}
	u.SinceThrow = 0
	if got := unitPose(u, 0, 0); got != poseThrow {
		t.Fatalf("flourish = %s", got.name())
	}
	u.SinceThrow = throwWindow
	u.GrenadeWind = 0.1
	if got := unitPose(u, 0, 0); got != poseThrow {
		t.Fatalf("windup = %s", got.name())
	}
	u.GrenadeWind = 0
	u.InWater = true
	if got := unitPose(u, 0, 0); got != poseSwim {
		t.Fatalf("swimming = %s", got.name())
	}
	u.Sinking = true
	if got := unitPose(u, 0, 0); got != poseSink {
		t.Fatalf("sinking = %s", got.name())
	}
	u.HP = sim.Dead
	if got := unitPose(u, 0, 0.4); got != poseDeath {
		t.Fatalf("dying = %s", got.name())
	}
	if got := unitPose(u, 0.4, 0.4); got != poseCorpse {
		t.Fatalf("death finished = %s", got.name())
	}
	if got := unitPose(u, 0, 0); got != poseCorpse {
		t.Fatalf("no death sheet = %s, want corpse", got.name())
	}
}

func TestThrowScrub(t *testing.T) {
	u := &sim.Unit{GrenadeWind: sim.GrenadierWindup, SinceThrow: 0}
	if throwFrac(u) != 0 {
		t.Fatalf("start of windup = %v", throwFrac(u))
	}
	u.GrenadeWind = sim.GrenadierWindup / 2
	if math.Abs(throwFrac(u)-0.5) > 1e-9 {
		t.Fatalf("mid windup = %v", throwFrac(u))
	}
	u.GrenadeWind = 0
	u.RocketWind = sim.RocketeerWindup / 2
	if math.Abs(throwFrac(u)-0.5) > 1e-9 {
		t.Fatalf("mid rocket windup = %v", throwFrac(u))
	}
	u.RocketWind = 0
	u.SinceThrow = throwWindow / 2
	if math.Abs(throwFrac(u)-0.5) > 1e-9 {
		t.Fatalf("mid flourish = %v", throwFrac(u))
	}
}

func TestFootSort(t *testing.T) {
	w := sim.NewEmpty()
	tiles := make([]sim.Tile, 16)
	tiles[0] = sim.TileTree
	w.Map = sim.Map{W: 4, H: 4, Tiles: tiles}
	w.SpawnUnit(sim.SidePlayer, sim.Vec2{X: 8, Y: 8})
	w.SpawnUnit(sim.SidePlayer, sim.Vec2{X: 8, Y: 40})
	bodies := collectBodies(w)
	sort.SliceStable(bodies, func(i, j int) bool { return bodies[i].y < bodies[j].y })
	if len(bodies) != 3 {
		t.Fatalf("bodies = %d, want a tree and two men", len(bodies))
	}
	if bodies[0].y != 8 || bodies[1].y != 16 || bodies[2].y != 40 {
		t.Fatalf("foot order %v, %v, %v; want man 8, tree 16, man 40", bodies[0].y, bodies[1].y, bodies[2].y)
	}
}

func TestDeathClockIsPerUnit(t *testing.T) {
	restoreRender(t)
	ResetAnim()
	w := sim.NewEmpty()
	w.AI = false
	w.Objectives = nil
	a := w.SpawnUnit(sim.SideEnemy, sim.Vec2{X: 1, Y: 1})
	b := w.SpawnUnit(sim.SideEnemy, sim.Vec2{X: 4, Y: 1})
	w.Unit(a.ID).HP = sim.Dead
	Advance(0.25, w)
	age, ok := deathAge[a.ID]
	if !ok || age != 0 {
		t.Fatalf("first sight of a corpse is %v ok=%v, want 0", age, ok)
	}
	if _, seen := deathAge[b.ID]; seen {
		t.Fatal("a living man has a death clock")
	}
	Advance(0.25, w)
	if math.Abs(deathAge[a.ID]-0.25) > 1e-12 {
		t.Fatalf("death clock = %v, want 0.25", deathAge[a.ID])
	}
	if animTime != 0.5 {
		t.Fatalf("anim clock = %v, want 0.5", animTime)
	}
}

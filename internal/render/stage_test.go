package render

import (
	"math"
	"sort"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
	ground "fannon-codder/internal/terrain"
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
	u.HP = sim.Wounded
	if got := unitPose(u, 0, 0.4); got != poseWounded || got.name() != "corpse" {
		t.Fatalf("wounded = %s", got.name())
	}
	u.HP = sim.Dead
	u.Side = sim.SideEnemy
	if got := unitPose(u, 0, 0.4); got != poseDeath {
		t.Fatalf("dying = %s", got.name())
	}
	if got := unitPose(u, 0.4, 0.4); got != poseCorpse {
		t.Fatalf("death finished = %s", got.name())
	}
	if got := unitPose(u, 0, 0); got != poseCorpse {
		t.Fatalf("no death sheet = %s, want corpse", got.name())
	}
	u.Side = sim.SidePlayer
	if got := unitPose(u, 0, 0.4); got != poseCorpse {
		t.Fatalf("dead player = %s, want corpse", got.name())
	}
}

func TestVictorySalutePose(t *testing.T) {
	victorySalute = true
	defer func() { victorySalute = false }()
	u := &sim.Unit{HP: sim.Alive, Side: sim.SidePlayer, SinceShot: 1, SinceThrow: 1}
	if got := unitPose(u, 0, 0); got != poseSalute || got.name() != "salute" {
		t.Fatalf("winner = %s", got.name())
	}
	u.Side = sim.SideEnemy
	if got := unitPose(u, 0, 0); got != poseIdle {
		t.Fatalf("enemy = %s, want idle", got.name())
	}
	u.Side = sim.SidePlayer
	u.InWater = true
	if got := unitPose(u, 0, 0); got != poseSwim {
		t.Fatalf("swimmer = %s, want swim", got.name())
	}
	u.InWater = false
	u.HP = sim.Dead
	if got := unitPose(u, 0, 0); got != poseCorpse {
		t.Fatalf("dead winner = %s, want corpse", got.name())
	}
	victorySalute = false
	u.HP = sim.Alive
	if got := unitPose(u, 0, 0); got != poseIdle {
		t.Fatalf("still playing = %s, want idle", got.name())
	}
}

func TestParkedSquadKneels(t *testing.T) {
	controlledSquad = sim.SquadSnake
	defer func() { controlledSquad = sim.SquadSnake }()
	u := &sim.Unit{HP: sim.Alive, Side: sim.SidePlayer, SquadID: sim.SquadEagle, SinceShot: 1, SinceThrow: 1}
	if got := unitPose(u, 0, 0); got != poseKneel || got.name() != "kneel" {
		t.Fatalf("parked eagle = %s", got.name())
	}
	u.SquadID = sim.SquadSnake
	if got := unitPose(u, 0, 0); got != poseIdle {
		t.Fatalf("controlled snake = %s", got.name())
	}
	kneelHold = map[sim.SquadID]bool{sim.SquadSnake: true}
	t.Cleanup(func() { kneelHold = nil })
	if got := unitPose(u, 0, 0); got != poseKneel {
		t.Fatalf("arrival wedge = %s, want kneel", got.name())
	}
	u.SinceShot = 0
	if got := unitPose(u, 0, 0); got != poseKneel {
		t.Fatalf("firing from the wedge = %s, want kneel", got.name())
	}
	kneelHold = nil
	u.SquadID = sim.SquadEagle
	if got := unitPose(u, 0, 0); got != poseKneel {
		t.Fatalf("parked shot = %s, want kneel", got.name())
	}
	u.SquadID = sim.SquadSnake
	if got := unitPose(u, 0, 0); got != poseShoot {
		t.Fatalf("standing shot = %s", got.name())
	}
	u.SquadID = sim.SquadEagle
	u.VX = walkMin + 1
	if got := unitPose(u, 0, 0); got != poseWalk {
		t.Fatalf("walking eagle = %s", got.name())
	}
	u.VX = 0
	u.SinceShot = 1
	u.Side = sim.SideEnemy
	if got := unitPose(u, 0, 0); got != poseIdle {
		t.Fatalf("enemy = %s", got.name())
	}
}

func TestSelectedSquadHoldsIdleFrame(t *testing.T) {
	controlledSquad = sim.SquadSnake
	defer func() { controlledSquad = sim.SquadSnake }()
	sh := activeSheets().Get("snake/idle")
	if sh == nil || sh.Frames < 2 {
		t.Fatal("snake idle sheet")
	}
	saved := animTime
	animTime = 0.3 // fps 8, frame 2 of 4
	defer func() { animTime = saved }()
	player := &sim.Unit{Side: sim.SidePlayer, SquadID: sim.SquadSnake}
	if got := poseFrame(poseIdle, player, sh); got != 0 {
		t.Fatalf("selected idle frame = %d, want 0", got)
	}
	enemy := &sim.Unit{Side: sim.SideEnemy}
	if got := poseFrame(poseIdle, enemy, sh); got != 0 {
		t.Fatalf("enemy idle frame = %d, want the standing cel", got)
	}
	civil := &sim.Unit{Side: sim.SideCivilian}
	if got := poseFrame(poseIdle, civil, sh); got == 0 {
		t.Fatal("civilian idle stayed on frame 0")
	}
	if got := poseFrame(poseWalk, player, sh); got == 0 {
		t.Fatal("selected walk stayed on frame 0")
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

func TestBlastOutlineOnlyWhenSheetMissing(t *testing.T) {
	restoreRender(t)
	var rects int
	blastRectHook = func() { rects++ }
	t.Cleanup(func() { blastRectHook = nil })

	w := sim.NewEmpty()
	w.Explosions = []sim.Explosion{{X: 40, Y: 40, R: sim.GrenadeRadius, Age: 0.25}}
	dst := ebiten.NewImage(80, 80)
	Blasts(dst, w)
	if rects != 0 {
		t.Fatal("a finished blast sheet still drew the orange outline")
	}

	rects = 0
	w.Explosions[0].Age = 0
	Blasts(dst, w)
	if rects != 0 {
		t.Fatal("the blast sheet drew the orange outline over the first frame")
	}

	SetSheets(&Library{})
	Blasts(dst, w)
	if rects != 1 {
		t.Fatalf("missing blast sheet outlines = %d, want 1", rects)
	}
}

func TestWalkingEnemyUsesWalk(t *testing.T) {
	restoreRender(t)
	SetPictureScale(1)
	ResetAnim()
	var sprites, rects int
	var got *ebiten.Image
	spriteHook = func(img *ebiten.Image, _ bool) {
		sprites++
		got = img
	}
	rectHook = func() { rects++ }
	t.Cleanup(func() { spriteHook, rectHook = nil, nil })

	w := sim.NewEmpty()
	w.AI = false
	w.Objectives = nil
	u := w.SpawnUnit(sim.SideEnemy, sim.Vec2{X: 40, Y: 40})
	u.VX = 20
	u.Facing = 0
	dst := ebiten.NewImage(64, 64)
	for _, kind := range []sim.UnitKind{sim.KindInfantry, sim.KindGrenadier, sim.KindRocketeer} {
		u.Kind = kind
		sprites, rects = 0, 0
		got = nil
		Field(dst, w, "grass")
		key := actorKey(u)
		sh := activeSheets().Get(key + "/walk")
		if sh == nil {
			t.Fatalf("missing %s/walk", key)
		}
		want, _, ok := sh.image("E", poseFrame(poseWalk, u, sh))
		if !ok || got != want || sprites != 1 || rects != 0 {
			t.Fatalf("walking %s: sprites=%d rects=%d ok=%v", key, sprites, rects, ok && got == want)
		}
	}
}

func TestPropShadowFollowsTheSprite(t *testing.T) {
	sh := activeSheets().Get("hut/door")
	rgba := sh.stillRGBA()
	if rgba == nil {
		t.Fatal("hut sheet has no CPU frame for the shadow")
	}
	sil := ground.PropSilhouette(rgba)
	if sil == nil || sil.Bounds().Dx() <= rgba.Bounds().Dx() {
		t.Fatal("hut shadow is not a padded copy of the sprite")
	}
	// The empty corner of the frame stays clear. A wall pixel stays dark.
	if _, _, _, a := sil.At(0, 0).RGBA(); a != 0 {
		t.Fatalf("hut shadow corner alpha = %v", a)
	}
	px := 128 + ground.SilhouettePad
	py := 280 + ground.SilhouettePad
	if _, _, _, a := sil.At(px, py).RGBA(); a < 40*256 {
		t.Fatalf("hut wall shadow alpha = %v", a)
	}
	if activeSheets().Get("crate/grenade").stillRGBA() == nil {
		t.Fatal("crate sheet has no CPU frame")
	}

	w := sim.NewEmpty()
	w.AddDoorHut(1, 1)
	w.AddGrenadeCrate(sim.TileCenter(4, 4))
	hut := buildingBody(w.Camera, w.Buildings[0])
	box := crateBody(w.Camera, w.Pickups[0])
	if hut.shadow == nil || box.shadow == nil {
		t.Fatal("a painted hut and crate cast a ground shadow")
	}
	dst := ebiten.NewImage(8, 8)
	hut.shadow(dst)
	box.shadow(dst)
}

func TestBarrelDrawsOnTheField(t *testing.T) {
	saved := pictureScale
	pictureScale = 2
	t.Cleanup(func() { pictureScale = saved })
	w := sim.NewEmpty()
	w.AddBarrel(sim.Vec2{X: 40, Y: 40})
	body := barrelBody(w.Camera, w.Barrels[0])
	if body.draw == nil || body.shadow == nil {
		t.Fatal("a live barrel should paint")
	}
	dst := ebiten.NewImage(200, 200)
	body.shadow(dst)
	body.draw(dst)
	Field(dst, w, "jungle")
}

func TestFootSort(t *testing.T) {
	w := sim.NewEmpty()
	tiles := make([]sim.Tile, 16)
	tiles[0] = sim.TileTree
	w.Map = sim.Map{W: 4, H: 4, Tiles: tiles}
	w.SpawnUnit(sim.SidePlayer, sim.Vec2{X: 8, Y: 8})
	w.SpawnUnit(sim.SidePlayer, sim.Vec2{X: 8, Y: 40})
	bodies := collectBodies(w, "grass")
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

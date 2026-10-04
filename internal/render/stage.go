package render

import (
	"image"
	"math"
	"sort"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

const (
	shotWindow  = 0.12
	throwWindow = 0.25
	walkMin     = 2 // px/s; at or under this, a man idles
)

// pose is the first row of the pose table that matches. Death time is kept
// here, keyed by unit id, and is not a sim field.
type pose int

const (
	poseDeath pose = iota
	poseCorpse
	poseSink
	poseSwim
	poseThrow
	poseShoot
	poseWalk
	poseIdle
)

func (p pose) name() string {
	switch p {
	case poseDeath:
		return "death"
	case poseCorpse:
		return "corpse"
	case poseSink:
		return "sink"
	case poseSwim:
		return "swim"
	case poseThrow:
		return "throw"
	case poseShoot:
		return "shoot"
	case poseWalk:
		return "walk"
	default:
		return "idle"
	}
}

func unitPose(u *sim.Unit, deathAge, deathDur float64) pose {
	if u.Dead() {
		if deathDur > 0 && deathAge < deathDur {
			return poseDeath
		}
		return poseCorpse
	}
	if u.Sinking {
		return poseSink
	}
	if u.InWater {
		return poseSwim
	}
	if u.GrenadeWind > 0 || u.RocketWind > 0 || u.SinceThrow < throwWindow {
		return poseThrow
	}
	if u.SinceShot < shotWindow {
		return poseShoot
	}
	if math.Hypot(u.VX, u.VY) > walkMin {
		return poseWalk
	}
	return poseIdle
}

// throwFrac scrubs the throw across the windup, or across the 0.25s flourish
// after the bomb or the rocket has left.
func throwFrac(u *sim.Unit) float64 {
	if u.GrenadeWind > 0 {
		if sim.GrenadierWindup <= 0 {
			return 0
		}
		return clamp01(1 - u.GrenadeWind/sim.GrenadierWindup)
	}
	if u.RocketWind > 0 {
		if sim.RocketeerWindup <= 0 {
			return 0
		}
		return clamp01(1 - u.RocketWind/sim.RocketeerWindup)
	}
	return clamp01(u.SinceThrow / throwWindow)
}

func clamp01(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

// sheetForPose is the painted cycle for this pose. Enemy ranks have no walk
// sheet, so a man leaving a door (or a rocketeer closing in) keeps his idle.
func sheetForPose(u *sim.Unit, p pose) *Sheet {
	key := actorKey(u)
	sh := activeSheets().Get(key + "/" + p.name())
	if sh == nil && p == poseWalk {
		sh = activeSheets().Get(key + "/idle")
	}
	return sh
}

func actorKey(u *sim.Unit) string {
	switch u.Side {
	case sim.SideCivilian:
		return "civilian"
	case sim.SideEnemy:
		switch u.Kind {
		case sim.KindGrenadier:
			return "grenadier"
		case sim.KindRocketeer:
			return "rocketeer"
		default:
			return "grunt"
		}
	default:
		switch u.SquadID {
		case sim.SquadEagle:
			return "eagle"
		case sim.SquadPanther:
			return "panther"
		default:
			return "snake"
		}
	}
}

func tileSheetKey(t sim.Tile) string {
	switch t {
	case sim.TileWaterShallow:
		return "ground/water-shallow"
	case sim.TileWaterDeep:
		return "ground/water-deep"
	case sim.TileIce:
		return "ground/ice"
	case sim.TileQuicksand:
		return "ground/quicksand"
	case sim.TileCliff:
		return "ground/cliff"
	case sim.TileBridge:
		return "ground/bridge"
	case sim.TileMine:
		return "ground/mine"
	case sim.TileRamp:
		return "ground/ramp"
	default:
		return ""
	}
}

func deathDuration(u *sim.Unit) float64 {
	sh := activeSheets().Get(actorKey(u) + "/death")
	if sh == nil || sh.FPS <= 0 || sh.Frames <= 0 {
		return 0
	}
	return float64(sh.Frames) / sh.FPS
}

func poseFrame(p pose, u *sim.Unit, sh *Sheet) int {
	switch p {
	case poseDeath:
		return FrameAt(deathAge[u.ID], sh.FPS, sh.Frames, false)
	case poseCorpse:
		return 0
	case poseSink:
		frac := 0.0
		if sim.SinkTime > 0 {
			frac = u.Sink / sim.SinkTime
		}
		return FrameScrub(frac, sh.Frames)
	case poseThrow:
		return FrameScrub(throwFrac(u), sh.Frames)
	case poseShoot:
		return FrameScrub(u.SinceShot/shotWindow, sh.Frames)
	default:
		return FrameAt(animTime, sh.FPS, sh.Frames, sh.Loop)
	}
}

// SpriteScale is how a source pixel maps into the offscreen. S is the
// picture scale. At the default window this is 0.375. At 4K it is about 1.05.
func SpriteScale() float64 {
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	return s / ArtScale
}

// spriteHook, when set, sees each cel DrawSprite is about to draw.
// Tests use it because a pixel read is not available before the game starts.
var spriteHook func(img *ebiten.Image, mirror bool)

// DrawSheet paints one frame of a loaded sheet. The anchor sits on x, y
// in world-frame pixels. dir is a facing such as "S" or "SE".
func DrawSheet(dst *ebiten.Image, key, dir string, frame int, x, y float64) bool {
	lib := activeSheets()
	if lib == nil {
		return false
	}
	sh := lib.Get(key)
	if sh == nil {
		return false
	}
	img, mirror, ok := sh.image(dir, frame)
	if !ok {
		return false
	}
	DrawSprite(dst, img, sh.AnchorX, sh.AnchorY, mirror, x, y)
	return true
}

// DrawSprite puts the cel's anchor on a logical screen point. mirror flips
// the cel around that anchor. The filter is linear.
func DrawSprite(dst, img *ebiten.Image, anchorX, anchorY int, mirror bool, x, y float64) {
	if dst == nil || img == nil {
		return
	}
	if spriteHook != nil {
		spriteHook(img, mirror)
	}
	sc := SpriteScale()
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	sx := sc
	if mirror {
		sx = -sc
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(-float64(anchorX), -float64(anchorY))
	op.GeoM.Scale(sx, sc)
	op.GeoM.Translate(x*ps, y*ps)
	dst.DrawImage(img, op)
}

// DrawSpriteAngle is DrawSprite with a clockwise spin around the anchor.
// angle 0 keeps the cel's right side pointing east, which matches a velocity
// of atan2(vy, vx) because GeoM.Rotate is clockwise in screen space.
func DrawSpriteAngle(dst, img *ebiten.Image, anchorX, anchorY int, x, y, angle float64) {
	if dst == nil || img == nil {
		return
	}
	if spriteHook != nil {
		spriteHook(img, false)
	}
	sc := SpriteScale()
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(-float64(anchorX), -float64(anchorY))
	op.GeoM.Rotate(angle)
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x*ps, y*ps)
	dst.DrawImage(img, op)
}

// drawTopLeft pins a ground frame to the cell. Seamless tiles keep anchor
// out of it; a wrong anchor would open a seam.
//
// Shrinking with mipmaps averages the atlas's transparent padding into the
// edge texels. The battlefield fill is a brighter green than the grass, so
// that fringe reads as a grid. The bled cel keeps a copy of the edge just
// outside the frame, mipmaps stay off, and the tile overlaps by one device
// pixel so a crack cannot open.
func drawTopLeft(dst, img *ebiten.Image, x, y float64) {
	if dst == nil || img == nil {
		return
	}
	img = bled(img)
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	sc := SpriteScale()
	fw := float64(img.Bounds().Dx())
	fh := float64(img.Bounds().Dy())
	sx, sy := sc, sc
	if fw > 0 {
		sx = (fw*sc + 1) / fw
	}
	if fh > 0 {
		sy = (fh*sc + 1) / fh
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.DisableMipmaps = true
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(x*ps-0.5, y*ps-0.5)
	dst.DrawImage(img, op)
}

// bled is the same frame with a one-pixel copy of each edge outside it.
// Draw the returned subimage; a linear sample then stays on grass.
var bledCache sync.Map

func bled(img *ebiten.Image) *ebiten.Image {
	if v, ok := bledCache.Load(img); ok {
		return v.(*ebiten.Image)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return img
	}
	pad := ebiten.NewImage(w+2, h+2)
	put := func(src *ebiten.Image, x, y float64) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		pad.DrawImage(src, op)
	}
	put(img, 1, 1)
	put(img.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Min.X+1, b.Max.Y)).(*ebiten.Image), 0, 1)
	put(img.SubImage(image.Rect(b.Max.X-1, b.Min.Y, b.Max.X, b.Max.Y)).(*ebiten.Image), float64(w+1), 1)
	put(img.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+1)).(*ebiten.Image), 1, 0)
	put(img.SubImage(image.Rect(b.Min.X, b.Max.Y-1, b.Max.X, b.Max.Y)).(*ebiten.Image), 1, float64(h+1))
	put(img.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Min.X+1, b.Min.Y+1)).(*ebiten.Image), 0, 0)
	put(img.SubImage(image.Rect(b.Max.X-1, b.Min.Y, b.Max.X, b.Min.Y+1)).(*ebiten.Image), float64(w+1), 0)
	put(img.SubImage(image.Rect(b.Min.X, b.Max.Y-1, b.Min.X+1, b.Max.Y)).(*ebiten.Image), 0, float64(h+1))
	put(img.SubImage(image.Rect(b.Max.X-1, b.Max.Y-1, b.Max.X, b.Max.Y)).(*ebiten.Image), float64(w+1), float64(h+1))
	inner := pad.SubImage(image.Rect(1, 1, w+1, h+1)).(*ebiten.Image)
	bledCache.Store(img, inner)
	return inner
}

func drawLoopSheet(dst *ebiten.Image, sh *Sheet, tx, ty int, x, y float64) bool {
	if sh == nil || sh.Frames == 0 {
		return false
	}
	img, ok := sh.still(SceneryFrame(tx, ty, sh.Frames, animTime, sh.FPS))
	if !ok {
		return false
	}
	drawTopLeft(dst, img, x, y)
	return true
}

var (
	animTime float64
	deathAge = map[int]float64{}
)

// ResetAnim zeroes the battle clock and the per-unit death timers.
func ResetAnim() {
	animTime = 0
	deathAge = map[int]float64{}
}

// Advance moves the animation clock and the death timers. The frame a unit
// is first seen dead starts his death at 0. Later calls add dt. Call it
// from the battle Update, after Step.
func Advance(dt float64, w *sim.World) {
	if dt < 0 {
		dt = 0
	}
	animTime += dt
	if w == nil {
		return
	}
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Dead() {
			continue
		}
		if _, ok := deathAge[u.ID]; !ok {
			deathAge[u.ID] = 0
			continue
		}
		deathAge[u.ID] += dt
	}
}

// Field paints the playfield: ground, shadows, then trees, huts, crates,
// flavour props, men, and vehicles by foot Y, then birds, grenades,
// tracers, and blasts.
// terrain is "snow" or "grass" and selects the base loop when that sheet exists.
func Field(dst *ebiten.Image, w *sim.World, terrain string) {
	if dst == nil || w == nil {
		return
	}
	paintTerrain(dst, w.Map, w.Camera, terrain)
	Tiles(dst, w.Map, w.Camera)
	bodies := collectBodies(w, terrain == "snow")
	sort.SliceStable(bodies, func(i, j int) bool { return bodies[i].y < bodies[j].y })
	for i := range bodies {
		if bodies[i].shadow != nil {
			bodies[i].shadow(dst)
		}
	}
	for i := range bodies {
		bodies[i].draw(dst)
	}
	drawBirds(dst, w)
	Muzzles(dst, w)
	Bombs(dst, w)
	Projectiles(dst, w)
	Blasts(dst, w)
}

type body struct {
	y      float64
	shadow func(*ebiten.Image)
	draw   func(*ebiten.Image)
}

func collectBodies(w *sim.World, arctic bool) []body {
	cam := w.Camera
	var bodies []body
	bodies = append(bodies, treeBodies(w.Map, cam)...)
	bodies = append(bodies, flavourBodies(w, arctic)...)
	for i := range w.Buildings {
		b := w.Buildings[i]
		if !b.Alive {
			continue
		}
		bodies = append(bodies, buildingBody(cam, b))
	}
	for i := range w.Pickups {
		p := w.Pickups[i]
		if !p.Alive {
			continue
		}
		bodies = append(bodies, crateBody(cam, p))
	}
	for i := range w.Units {
		u := w.Units[i]
		if u.VehicleID != 0 {
			continue
		}
		bodies = append(bodies, unitBody(cam, u))
	}
	for i := range w.Vehicles {
		v := w.Vehicles[i]
		if !v.Alive {
			continue
		}
		bodies = append(bodies, vehicleBody(cam, v))
	}
	return bodies
}

func paintTerrain(dst *ebiten.Image, m sim.Map, cam sim.Camera, terrain string) {
	if m.W == 0 {
		return
	}
	if terrain == "" {
		terrain = "grass"
	}
	sh := activeSheets().Get("ground/" + terrain)
	if sh == nil {
		return
	}
	tx0, ty0, tx1, ty1 := visibleTiles(m, cam)
	for ty := ty0; ty < ty1; ty++ {
		for tx := tx0; tx < tx1; tx++ {
			t := m.At(tx, ty)
			if t != sim.TileGrass && t != sim.TileTree {
				continue
			}
			// Frame 0 is the rest pose. The other frames only shift fine
			// strokes, and that shimmer reads as noise, so the fill stays put.
			img, ok := sh.still(0)
			if !ok {
				continue
			}
			drawTopLeft(dst, img,
				cam.ScreenX(float64(tx*sim.TileSize)),
				cam.ScreenY(float64(ty*sim.TileSize)),
			)
		}
	}
}

func treeBodies(m sim.Map, cam sim.Camera) []body {
	if m.W == 0 {
		return nil
	}
	tx0, ty0, tx1, ty1 := visibleTiles(m, cam)
	tx0 -= 2
	ty0 -= 4
	if tx0 < 0 {
		tx0 = 0
	}
	if ty0 < 0 {
		ty0 = 0
	}
	var bodies []body
	for ty := ty0; ty < ty1; ty++ {
		for tx := tx0; tx < tx1; tx++ {
			if m.At(tx, ty) != sim.TileTree {
				continue
			}
			bodies = append(bodies, treeBody(cam, tx, ty))
		}
	}
	return bodies
}

func treeBody(cam sim.Camera, tx, ty int) body {
	footX := cam.ScreenX(float64(tx*sim.TileSize) + float64(sim.TileSize)/2)
	footY := cam.ScreenY(float64((ty + 1) * sim.TileSize))
	y := float64((ty + 1) * sim.TileSize)
	sh := activeSheets().Get("tree/sway")
	var img *ebiten.Image
	if sh != nil {
		// Rows are the three silhouettes. Neighbours take different shapes and different sway frames.
		row := mod(tx+ty*2, sh.RowCount())
		img, _ = sh.Row(row, SceneryFrame(tx, ty, sh.Frames, animTime, sh.FPS))
	}
	if img == nil {
		return body{y: y, draw: func(dst *ebiten.Image) {
			blit(dst, tileImage(sim.TileTree),
				cam.ScreenX(float64(tx*sim.TileSize)),
				cam.ScreenY(float64(ty*sim.TileSize)),
				1, 1,
			)
		}}
	}
	ax, ay := sh.AnchorX, sh.AnchorY
	return spriteBody(y, footX, footY, img, ax, ay, false, nil)
}

func buildingBody(cam sim.Camera, b sim.Building) body {
	y := b.Y + b.H
	sx := cam.ScreenX(b.X + b.W/2)
	sy := cam.ScreenY(y)
	key := "hut/plain"
	if b.HasDoor {
		key = "hut/door"
	}
	sh := activeSheets().Get(key)
	var img *ebiten.Image
	if sh != nil {
		img, _ = sh.still(FrameAt(animTime, sh.FPS, sh.Frames, sh.Loop))
	}
	fallback := func(dst *ebiten.Image) { drawBuildingRect(dst, cam, b) }
	if img == nil {
		return body{y: y, draw: fallback}
	}
	return spriteBody(y, sx, sy, img, sh.AnchorX, sh.AnchorY, false, fallback)
}

func crateBody(cam sim.Camera, p sim.Pickup) body {
	sx := cam.ScreenX(p.X)
	sy := cam.ScreenY(p.Y)
	key := "crate/grenade"
	if p.Kind == sim.PickupRockets {
		key = "crate/rocket"
	}
	sh := activeSheets().Get(key)
	var img *ebiten.Image
	if sh != nil {
		img, _ = sh.still(FrameAt(animTime, sh.FPS, sh.Frames, sh.Loop))
	}
	fallback := func(dst *ebiten.Image) { drawCrateRect(dst, cam, p) }
	if img == nil {
		return body{y: p.Y, draw: fallback}
	}
	return spriteBody(p.Y, sx, sy, img, sh.AnchorX, sh.AnchorY, false, fallback)
}

func unitBody(cam sim.Camera, u sim.Unit) body {
	sx := cam.ScreenX(u.X)
	sy := cam.ScreenY(u.Y)
	fallback := func(dst *ebiten.Image) { drawUnitRect(dst, cam, &u) }
	p := unitPose(&u, deathAge[u.ID], deathDuration(&u))
	sh := sheetForPose(&u, p)
	if sh == nil {
		return body{y: u.Y, draw: fallback}
	}
	img, mirror, ok := sh.image(Direction(u.Facing), poseFrame(p, &u, sh))
	if !ok {
		return body{y: u.Y, draw: fallback}
	}
	return spriteBody(u.Y, sx, sy, img, sh.AnchorX, sh.AnchorY, mirror, fallback)
}

func vehicleBody(cam sim.Camera, v sim.Vehicle) body {
	sx := cam.ScreenX(v.X)
	sy := cam.ScreenY(v.Y)
	key := "skidoo/idle"
	if math.Hypot(v.VX, v.VY) > walkMin {
		key = "skidoo/move"
	}
	sh := activeSheets().Get(key)
	var img *ebiten.Image
	var mirror bool
	var ax, ay int
	if sh != nil {
		ax, ay = sh.AnchorX, sh.AnchorY
		img, mirror, _ = sh.image(Direction(v.Facing), FrameAt(animTime, sh.FPS, sh.Frames, sh.Loop))
	}
	return body{
		y: v.Y,
		shadow: func(dst *ebiten.Image) {
			if img != nil {
				castShadow(dst, sx, sy, img)
			}
		},
		draw: func(dst *ebiten.Image) {
			if img != nil {
				DrawSprite(dst, img, ax, ay, mirror, sx, sy)
			} else {
				drawVehicleRect(dst, cam, &v)
			}
			drawEnemyLamp(dst, cam, &v)
		},
	}
}

func spriteBody(y, x, sy float64, img *ebiten.Image, ax, ay int, mirror bool, fallback func(*ebiten.Image)) body {
	if img == nil {
		return body{y: y, draw: fallback}
	}
	return body{
		y:      y,
		shadow: func(dst *ebiten.Image) { castShadow(dst, x, sy, img) },
		draw:   func(dst *ebiten.Image) { DrawSprite(dst, img, ax, ay, mirror, x, sy) },
	}
}

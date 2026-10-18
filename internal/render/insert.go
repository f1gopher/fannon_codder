package render

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

// The opening drop. A helicopter flies in along the nose of chopper/fly.
// One rope hangs straight down, or two when more than one soldier is
// aboard. They slide down, walk out to the file they spawned on, and
// kneel. The world does not step until that kneel is done. The aircraft
// then leaves along the same nose. This is a picture, not a vehicle.

const (
	dropApproach = 1.4
	dropRope     = 0.35
	dropSlide    = 0.95
	dropStagger  = 0.18
	dropSpread   = 0.7
	dropKneel    = 0.75
	dropRopeUp   = 0.22
	dropDepart   = 1.35
	dropInbound  = 240.0
	dropOutbound = 300.0
	dropAltHigh  = 92.0
	dropAltLow   = 64.0
	// Two ropes sit side by side under the door, in world pixels.
	dropRopeGap = 7.0

	// Anchor and hubs measured on chopper/fly.png. The anchor is the
	// bottom of the open door, where the ropes leave. Source +y is down.
	flyAnchorX = 210
	flyAnchorY = 270
	flyMastX   = 270
	flyMastY   = 105
	flyTailX   = 52
	flyTailY   = 36
	flyNoseX   = 400
	flyNoseY   = 250
)

// activeDrop hides the squad and paints the aircraft. Battle.Draw sets it.
var activeDrop *Insert

// UseDrop binds the arrival for the next Field. nil clears it.
func UseDrop(in *Insert) { activeDrop = in }

func dropHides(id int) bool {
	return activeDrop != nil && activeDrop.Hides(id)
}

// Insert is one arrival. Build it from the squads as they stand at spawn.
type Insert struct {
	t   float64
	men []dropMan
}

type dropMan struct {
	id    int
	squad sim.SquadID
	x, y  float64
}

// NewInsert records every living player on foot. No squad means it is
// already over, so a test world with nobody in it does not wait.
func NewInsert(w *sim.World) *Insert {
	in := &Insert{}
	if w == nil {
		return in
	}
	for i := range w.Squads {
		s := &w.Squads[i]
		for _, id := range s.MemberIDs {
			u := w.Unit(id)
			if u == nil || !u.Living() || u.Side != sim.SidePlayer || u.VehicleID != 0 {
				continue
			}
			in.men = append(in.men, dropMan{id: id, squad: s.ID, x: u.X, y: u.Y})
		}
	}
	return in
}

// Step moves the arrival. A finished flight stays finished.
func (in *Insert) Step(dt float64) {
	if in == nil || in.Gone() || dt <= 0 {
		return
	}
	in.t += dt
}

// Landed is true once every boot is on its mark.
func (in *Insert) Landed() bool {
	if in == nil || len(in.men) == 0 {
		return true
	}
	return in.t+1e-6 >= in.landAt()
}

// Gone is true once the helicopter has left.
func (in *Insert) Gone() bool {
	if in == nil || len(in.men) == 0 {
		return true
	}
	return in.t+1e-6 >= in.landAt()+dropRopeUp+dropDepart
}

// Hides reports a soldier who is still in the aircraft or on a rope.
func (in *Insert) Hides(id int) bool {
	if in == nil || in.Gone() {
		return false
	}
	i := in.index(id)
	if i < 0 {
		return false
	}
	return in.progress(i) < 1
}

func (in *Insert) index(id int) int {
	for i := range in.men {
		if in.men[i].id == id {
			return i
		}
	}
	return -1
}

func (in *Insert) landAt() float64 {
	n := len(in.men)
	if n == 0 {
		return 0
	}
	return in.ropeExit(n-1) + dropSlide + dropSpread + dropKneel
}

// ropeExit is when man i leaves the cabin.
func (in *Insert) ropeExit(i int) float64 {
	return dropApproach + dropRope + dropStagger*float64(i)
}

// ropes is one line for a single soldier and two otherwise.
func (in *Insert) ropes() int {
	n := len(in.men)
	if n <= 1 {
		return n
	}
	return 2
}

func (in *Insert) ropeIndex(i int) int {
	r := in.ropes()
	if r <= 1 {
		return 0
	}
	return i % r
}

// ropeFoot is the ground point under man i's rope. X matches the door
// slot so the line is vertical. Y is the ground under the aircraft.
func (in *Insert) ropeFoot(i int) (x, y float64) {
	hx, hy := in.hover()
	r := in.ropes()
	if r == 0 {
		return hx, hy
	}
	mid := float64(r-1) / 2
	slot := (float64(in.ropeIndex(i)) - mid) * dropRopeGap
	return hx + slot, hy
}

type dropBeat int

const (
	beatCabin dropBeat = iota
	beatRope
	beatSpread
	beatKneel
	beatDown
)

func (in *Insert) beat(i int) (dropBeat, float64) {
	t0 := in.ropeExit(i)
	if in.t < t0 {
		return beatCabin, 0
	}
	if in.t < t0+dropSlide {
		return beatRope, (in.t - t0) / dropSlide
	}
	t0 += dropSlide
	if in.t < t0+dropSpread {
		return beatSpread, (in.t - t0) / dropSpread
	}
	t0 += dropSpread
	if in.t < t0+dropKneel {
		return beatKneel, (in.t - t0) / dropKneel
	}
	return beatDown, 1
}

func (in *Insert) progress(i int) float64 {
	b, u := in.beat(i)
	switch b {
	case beatCabin:
		return -1
	case beatDown:
		return 1
	case beatRope:
		return u * 0.5
	default:
		// Spread and kneel are both "out of the cabin" and not yet done.
		return 0.5 + u*0.49
	}
}

func (in *Insert) hover() (x, y float64) {
	for i := range in.men {
		x += in.men[i].x
		y += in.men[i].y
	}
	n := float64(len(in.men))
	return x / n, y / n
}

// dropHeading is the painted nose, in world pixels. Source +x is east
// and source +y is south, the same axes as the map, so the flight can
// use this vector directly. Right, with y down, is (-y, x).
func dropHeading() (x, y float64) {
	dx := float64(flyNoseX - flyTailX)
	dy := float64(flyNoseY - flyTailY)
	l := math.Hypot(dx, dy)
	return dx / l, dy / l
}

func dropRight() (x, y float64) {
	nx, ny := dropHeading()
	return -ny, nx
}

// pose is the ground point under the door, and how high the door sits.
func (in *Insert) pose() (x, y, alt float64) {
	hx, hy := in.hover()
	nx, ny := dropHeading()
	switch {
	case in.t < dropApproach:
		u := smooth01(in.t / dropApproach)
		x = hx - nx*dropInbound*(1-u)
		y = hy - ny*dropInbound*(1-u)
		alt = dropAltHigh + (dropAltLow-dropAltHigh)*u
	case in.t < in.landAt()+dropRopeUp:
		x, y = hx, hy
		alt = dropAltLow + math.Sin(in.t*5.5)*1.2
	default:
		u := clamp01((in.t - (in.landAt() + dropRopeUp)) / dropDepart)
		accel := u * u
		x = hx + nx*dropOutbound*accel
		y = hy + ny*dropOutbound*accel
		alt = dropAltLow + (dropAltHigh-dropAltLow)*u
	}
	return x, y, alt
}

// feet is where a man is drawn, and whether he has left the cabin.
// On the rope he moves only downward. On the ground he walks to his mark.
func (in *Insert) feet(i int) (x, y float64, out bool) {
	b, u := in.beat(i)
	if b == beatCabin {
		return 0, 0, false
	}
	fx, fy := in.ropeFoot(i)
	_, _, alt := in.pose()
	switch b {
	case beatRope:
		s := smooth01(u)
		return fx, (fy - alt) + alt*s, true
	case beatSpread:
		s := smooth01(u)
		return fx + (in.men[i].x-fx)*s, fy + (in.men[i].y-fy)*s, true
	default:
		return in.men[i].x, in.men[i].y, true
	}
}

// Draw paints the shadow, the ropes, the men, then the aircraft.
func (in *Insert) Draw(dst *ebiten.Image, cam sim.Camera) {
	if dst == nil || in == nil || len(in.men) == 0 || in.Gone() {
		return
	}
	gx, gy, alt := in.pose()
	doorX := cam.ScreenX(gx)
	doorY := cam.ScreenY(gy) - alt

	sh := activeSheets().Get("chopper/fly")
	var body *ebiten.Image
	if sh != nil {
		body, _ = sh.still(0)
	}
	if body != nil {
		castShadow(dst, cam.ScreenX(gx), cam.ScreenY(gy)+8, body)
	} else {
		fillRectF(dst, cam.ScreenX(gx)-10, cam.ScreenY(gy)-3, 22, 6, color.RGBA{A: 70})
	}

	ropeU := 1.0
	if in.t < dropApproach {
		ropeU = 0
	} else if in.t < dropApproach+dropRope {
		ropeU = smooth01((in.t - dropApproach) / dropRope)
	}
	if in.t > in.landAt() {
		up := clamp01((in.t - in.landAt()) / dropRopeUp)
		ropeU = 1 - smooth01(up)
	}
	rope := color.RGBA{R: 0x2a, G: 0x22, B: 0x18, A: 0xff}
	if ropeU > 0 {
		groundY := cam.ScreenY(gy)
		for k := 0; k < in.ropes(); k++ {
			x0, y0 := ropeTop(in, k, doorX, doorY)
			y1 := y0 + (groundY-y0)*ropeU
			strokeLine(dst, x0, y0, x0, y1, 1.1, rope)
		}
	}
	for i := range in.men {
		b, _ := in.beat(i)
		if b == beatCabin || b == beatDown {
			continue
		}
		x, y, out := in.feet(i)
		if !out {
			continue
		}
		drawDropSoldier(dst, cam, in, i, x, y)
	}
	drawChopper(dst, doorX, doorY, body)
}

// ropeTop is the door end of rope k. The offset is horizontal, so the
// rope falls straight down the screen.
func ropeTop(in *Insert, k int, doorX, doorY float64) (x, y float64) {
	r := in.ropes()
	if r <= 1 {
		return doorX, doorY
	}
	mid := float64(r-1) / 2
	return doorX + (float64(k)-mid)*dropRopeGap, doorY
}

func drawDropSoldier(dst *ebiten.Image, cam sim.Camera, in *Insert, i int, x, y float64) {
	m := &in.men[i]
	b, _ := in.beat(i)
	facing := math.Pi / 2
	p := poseIdle
	switch b {
	case beatSpread:
		p = poseWalk
		dx := m.x - x
		dy := m.y - y
		if math.Hypot(dx, dy) > 0.5 {
			facing = math.Atan2(dy, dx)
		}
	case beatKneel:
		p = poseKneel
		facing = -math.Pi / 2
	}
	u := sim.Unit{ID: m.id, Side: sim.SidePlayer, HP: sim.Alive, SquadID: m.squad, Facing: facing}
	sh := sheetForPose(&u, p)
	if sh == nil {
		fillRectF(dst, cam.ScreenX(x)-2, cam.ScreenY(y)-8, 4, 8, color.RGBA{R: 0x6a, G: 0x70, B: 0x38, A: 0xff})
		return
	}
	frame := 0
	if p == poseWalk && sh.Frames > 1 {
		frame = int(in.t*8) % sh.Frames
	}
	img, mirror, ok := sh.image(Direction(facing), frame)
	if !ok {
		return
	}
	sx := cam.ScreenX(x)
	sy := cam.ScreenY(y)
	castShadow(dst, cam.ScreenX(m.x), cam.ScreenY(m.y), img)
	DrawSprite(dst, img, sh.AnchorX, sh.AnchorY, mirror, sx, sy)
}

func drawChopper(dst *ebiten.Image, x, y float64, body *ebiten.Image) {
	if body != nil {
		DrawSprite(dst, body, flyAnchorX, flyAnchorY, false, x, y)
	} else {
		fillRectF(dst, x-16, y-8, 34, 12, color.RGBA{R: 0x6e, G: 0x74, B: 0x3c, A: 0xff})
		fillRectF(dst, x-28, y-4, 14, 4, color.RGBA{R: 0x5a, G: 0x60, B: 0x32, A: 0xff})
	}
	// Hubs are source pixels relative to the door anchor. ArtScale
	// turns that into world pixels, which is the space Draw uses.
	mx := x + float64(flyMastX-flyAnchorX)/ArtScale
	my := y + float64(flyMastY-flyAnchorY)/ArtScale
	tx := x + float64(flyTailX-flyAnchorX)/ArtScale
	ty := y + float64(flyTailY-flyAnchorY)/ArtScale
	drawRotor(dst, mx, my, 36, 14, animTime*16, 0.55)
	drawRotor(dst, tx, ty, 5, 5, animTime*28, 0.85)
}

func drawRotor(dst *ebiten.Image, x, y, rx, ry, angle, squash float64) {
	if dst == nil || rx <= 0 || ry <= 0 {
		return
	}
	disc := rotorDisc()
	b := disc.Bounds()
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(-float64(b.Dx())/2, -float64(b.Dy())/2)
	op.GeoM.Scale(rx*2/float64(b.Dx())*ps, ry*2/float64(b.Dy())*ps)
	op.GeoM.Translate(x*ps, y*ps)
	dst.DrawImage(disc, op)
	blade := color.RGBA{R: 0x3a, G: 0x40, B: 0x28, A: 0xb0}
	for i := 0; i < 3; i++ {
		a := angle + float64(i)*2*math.Pi/3
		strokeLine(dst, x, y, x+math.Cos(a)*rx, y+math.Sin(a)*ry*squash, 1.3, blade)
	}
}

var rotorOnce struct {
	img *ebiten.Image
}

func rotorDisc() *ebiten.Image {
	if rotorOnce.img != nil {
		return rotorOnce.img
	}
	const w, h = 128, 64
	rgba := image.NewNRGBA(image.Rect(0, 0, w, h))
	cx := (w - 1) / 2.0
	cy := (h - 1) / 2.0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := (float64(x) - cx) / (w / 2.0)
			dy := (float64(y) - cy) / (h / 2.0)
			d := dx*dx + dy*dy
			if d >= 1 {
				continue
			}
			fade := (1 - d) * (1 - d)
			rgba.SetNRGBA(x, y, color.NRGBA{R: 0xc8, G: 0xcc, B: 0xc0, A: uint8(fade * 70)})
		}
	}
	rotorOnce.img = ebiten.NewImageFromImage(rgba)
	return rotorOnce.img
}

func strokeLine(dst *ebiten.Image, x0, y0, x1, y1, thick float64, c color.Color) {
	if dst == nil || thick <= 0 {
		return
	}
	dx, dy := x1-x0, y1-y0
	length := math.Hypot(dx, dy)
	if length < 0.25 {
		return
	}
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(length*ps, thick*ps)
	op.GeoM.Translate(0, -thick*ps/2)
	// GeoM.Rotate is clockwise in screen space, and atan2(dy, dx) is
	// clockwise from east when y grows downward, so the line follows
	// the two points.
	op.GeoM.Rotate(math.Atan2(dy, dx))
	op.GeoM.Translate(x0*ps, y0*ps)
	op.ColorScale.ScaleWithColor(c)
	dst.DrawImage(pixel(), op)
}

func smooth01(u float64) float64 {
	u = clamp01(u)
	return u * u * (3 - 2*u)
}

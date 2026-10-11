package render

import (
	"image"
	"image/draw"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

// Idle fidgets are cosmetic. They do not change Facing, so the squad you
// control still aims at the pointer and a parked man still watches his sector.
// The leader of that squad never turns his head off the pointer. He can nod,
// look up, or hop. Everyone else glances, and now and then someone does a gag.

const idleBout = 4.0

// idleCue is one moment of an idle performance. Head offsets are source
// pixels: +headX is the character's right, +headY is down. Tilt tips the
// helmet toward his right. Hop is world pixels up. Salute swaps in the
// south-facing salute sheet for a beat.
type idleCue struct {
	headX, headY float64
	tilt         float64
	hop          float64
	salute       bool
}

func (c idleCue) movesHead() bool {
	return math.Abs(c.headX) > 0.05 || math.Abs(c.headY) > 0.05 || math.Abs(c.tilt) > 0.001
}

// idleCueAt is the performance for this soldier at animTime t.
// aimLock is the leader of the squad under the pointer.
// player is a trooper who has a salute sheet.
func idleCueAt(id int, t float64, aimLock, player bool) idleCue {
	if id < 0 {
		id = 0
	}
	if t < 0 {
		t = 0
	}
	local := t + float64(id)*0.73
	bout := int(math.Floor(local / idleBout))
	sec := local - float64(bout)*idleBout
	roll := mix32(uint32(id), uint32(bout))
	var c idleCue
	if roll%20 == 0 {
		c = idleGag(sec, int((roll/20)%4), player)
	} else {
		c = idleLook(sec, roll)
	}
	if aimLock {
		// He keeps his eyes on the pointer. A side glance becomes a nod.
		// A salute to the camera becomes a look up. Both still face forward.
		if c.salute {
			c.salute = false
			if c.headY == 0 {
				c.headY = -5
			}
		}
		if math.Abs(c.headX) > 0.05 || math.Abs(c.tilt) > 0.001 {
			e := math.Abs(c.tilt) / 0.42
			if e > 1 {
				e = 1
			}
			if e < 0.15 {
				e = 0.15
			}
			c.headY = 5 * e
		}
		c.headX = 0
		c.tilt = 0
		c.salute = false
	}
	c.headX = clampRange(c.headX, -18, 18)
	c.headY = clampRange(c.headY, -16, 10)
	c.tilt = clampRange(c.tilt, -0.55, 0.55)
	c.hop = clampRange(c.hop, 0, 6)
	return c
}

func idleLook(sec float64, roll uint32) idleCue {
	var c idleCue
	switch roll % 5 {
	case 4:
		e := pulse(sec, 0.35, 0.75, 1.55, 2.05)
		c.headY = -6 * e
	case 3:
		// Look one way, then the other.
		s := sideSign(roll >> 8)
		e1 := pulse(sec, 0.2, 0.45, 0.9, 1.15)
		e2 := pulse(sec, 1.3, 1.55, 2.05, 2.35)
		c.headX = s * (-6*e1 + 6*e2)
		c.tilt = s * (-0.42*e1 + 0.42*e2)
	default:
		e := pulse(sec, 0.3, 0.7, 1.5, 2.0)
		s := sideSign(roll >> 8)
		c.headX = s * 5 * e
		c.tilt = s * 0.42 * e
	}
	return c
}

func idleGag(sec float64, kind int, player bool) idleCue {
	var c idleCue
	switch kind {
	case 0:
		c.hop = hopArc(sec, 0.35, 0.9) * 3.5
	case 1:
		// Chin up, counting clouds.
		e := pulse(sec, 0.25, 0.6, 2.1, 2.6)
		c.headY = -5 * e
	case 2:
		// A little sneeze: duck, then a pop.
		e := pulse(sec, 0.4, 0.55, 0.8, 1.15)
		c.headY = 7 * e
		c.hop = hopArc(sec, 0.75, 1.15) * 3
	default:
		if player {
			e := pulse(sec, 0.25, 0.45, 2.0, 2.3)
			if e > 0 {
				c.salute = true
				// If he is the leader, the salute is turned into this look-up.
				c.headY = -5 * e
			}
			return c
		}
		// No salute sheet. Peek hard over each shoulder instead.
		e1 := pulse(sec, 0.2, 0.4, 0.85, 1.05)
		e2 := pulse(sec, 1.2, 1.4, 1.9, 2.15)
		c.headX = -7*e1 + 7*e2
		c.tilt = -0.48*e1 + 0.48*e2
	}
	return c
}

func sideSign(r uint32) float64 {
	if r&1 == 0 {
		return -1
	}
	return 1
}

func pulse(sec, rise0, rise1, fall0, fall1 float64) float64 {
	switch {
	case sec < rise0 || sec >= fall1 || rise1 <= rise0 || fall1 <= fall0:
		return 0
	case sec < rise1:
		return smoothstep((sec - rise0) / (rise1 - rise0))
	case sec < fall0:
		return 1
	default:
		return 1 - smoothstep((sec-fall0)/(fall1-fall0))
	}
}

func smoothstep(u float64) float64 {
	if u < 0 {
		return 0
	}
	if u > 1 {
		return 1
	}
	return u * u * (3 - 2*u)
}

func hopArc(sec, a, b float64) float64 {
	if sec < a || sec >= b || b <= a {
		return 0
	}
	return math.Sin((sec - a) / (b - a) * math.Pi)
}

func clampRange(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func mix32(a, b uint32) uint32 {
	x := a*0x9E3779B1 ^ b*0x85EBCA6B
	x ^= x >> 16
	x *= 0x7FEB352D
	x ^= x >> 15
	return x
}

// headParts is the helmet and face cut off a standing or kneeling cel, plus
// the same cel with that cap cleared. The pivot is the neck. Cached per cel.
type headParts struct {
	body, head     *ebiten.Image
	pivotX, pivotY int
}

var headCache struct {
	mu sync.Mutex
	m  map[*image.RGBA]headParts
}

func cachedHead(src *image.RGBA) (headParts, bool) {
	if src == nil {
		return headParts{}, false
	}
	headCache.mu.Lock()
	defer headCache.mu.Unlock()
	if headCache.m == nil {
		headCache.m = map[*image.RGBA]headParts{}
	}
	if p, ok := headCache.m[src]; ok {
		return p, p.head != nil
	}
	body, head, px, py, ok := splitHead(src)
	if !ok {
		headCache.m[src] = headParts{}
		return headParts{}, false
	}
	p := headParts{
		body:   ebiten.NewImageFromImage(body),
		head:   ebiten.NewImageFromImage(head),
		pivotX: px, pivotY: py,
	}
	headCache.m[src] = p
	return p, true
}

// splitHead lifts the cap off the shoulders. The rifle, the rocket tube, and
// the body stay. A cel with no narrow helmet, including a grenadier front
// view that sits on a painted box, is left whole.
func splitHead(src *image.RGBA) (body, head *image.RGBA, pivotX, pivotY int, ok bool) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 24 || h < 48 {
		return nil, nil, 0, 0, false
	}
	opaque := func(x, y int) bool {
		return src.Pix[src.PixOffset(x, y)+3] > 32
	}
	top := -1
	for y := 0; y < h && top < 0; y++ {
		for x := 0; x < w; x++ {
			if opaque(x, y) {
				top = y
				break
			}
		}
	}
	if top < 0 || top > h/3 {
		return nil, nil, 0, 0, false
	}
	lo, hi, found := connected(opaque, w, top, w/2)
	if !found || hi-lo+1 > w*45/100 {
		return nil, nil, 0, 0, false
	}
	cx := (lo + hi) / 2
	// The neck is the narrowest run under the crown, a little below the brim
	// and above the shoulders. A rifle beside the helmet is a different run
	// and does not widen this one.
	neckY, neckW := -1, w
	loY, hiY := top+8, top+22
	if hiY > h*3/5 {
		hiY = h * 3 / 5
	}
	for y := loY; y <= hiY; y++ {
		clo, chi, cok := connected(opaque, w, y, cx)
		if !cok {
			continue
		}
		conn := chi - clo + 1
		if conn > 6 && conn < neckW {
			neckW = conn
			neckY = y
		}
	}
	if neckY < 0 || neckW > w/3 {
		return nil, nil, 0, 0, false
	}
	cut := neckY + 2
	if cut < top+8 {
		return nil, nil, 0, 0, false
	}
	minX, maxX := w, -1
	for y := top; y < cut; y++ {
		clo, chi, cok := connected(opaque, w, y, cx)
		if !cok {
			continue
		}
		if clo < minX {
			minX = clo
		}
		if chi > maxX {
			maxX = chi
		}
	}
	if maxX < 0 || maxX-minX > w/2 {
		return nil, nil, 0, 0, false
	}
	minX--
	maxX++
	// Extra collar rows travel with the helmet so a tilt does not open a gap.
	overlap := 12
	bodyImg := image.NewRGBA(image.Rect(0, 0, w, h))
	headImg := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(bodyImg, bodyImg.Bounds(), src, b.Min, draw.Src)
	headN := 0
	for y := top; y < cut+overlap && y < h; y++ {
		for x := minX; x <= maxX; x++ {
			if x < 0 || x >= w || !opaque(x, y) {
				continue
			}
			si := src.PixOffset(x, y)
			di := headImg.PixOffset(x, y)
			copy(headImg.Pix[di:di+4], src.Pix[si:si+4])
			headN++
			if y < cut {
				bi := bodyImg.PixOffset(x, y)
				bodyImg.Pix[bi] = 0
				bodyImg.Pix[bi+1] = 0
				bodyImg.Pix[bi+2] = 0
				bodyImg.Pix[bi+3] = 0
			}
		}
	}
	if headN < 30 {
		return nil, nil, 0, 0, false
	}
	return bodyImg, headImg, cx, cut, true
}

// connected is the opaque run on this row nearest to seed. A run that starts
// far from the helmet belongs to a rifle, not the head.
func connected(opaque func(int, int) bool, w, y, seed int) (lo, hi int, ok bool) {
	best, dist := -1, w
	for x := 0; x < w; x++ {
		if !opaque(x, y) {
			continue
		}
		d := absInt(x - seed)
		if d < dist {
			dist = d
			best = x
		}
	}
	if best < 0 || dist > 18 {
		return 0, 0, false
	}
	lo, hi = best, best
	for x := best; x >= 0 && opaque(x, y); x-- {
		lo = x
	}
	for x := best; x < w && opaque(x, y); x++ {
		hi = x
	}
	return lo, hi, true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// drawFidget paints an idle or kneeling soldier. The shadow stays on the
// ground. A salute gag uses that sheet, facing the camera. Otherwise the
// helmet is drawn a second time, tipped on the neck.
func drawFidget(dst *ebiten.Image, u *sim.Unit, img *ebiten.Image, src *image.RGBA, ax, ay int, mirror bool, x, y float64, cue idleCue) {
	y -= cue.hop
	if cue.salute {
		if simg, sax, say, smirror, ok := saluteCel(u); ok {
			DrawSprite(dst, simg, sax, say, smirror, x, y)
			return
		}
	}
	if cue.movesHead() {
		if parts, ok := cachedHead(src); ok {
			DrawSprite(dst, parts.body, ax, ay, mirror, x, y)
			drawHead(dst, parts.head, ax, ay, parts.pivotX, parts.pivotY, mirror, x, y, cue)
			return
		}
	}
	DrawSprite(dst, img, ax, ay, mirror, x, y)
}

func saluteCel(u *sim.Unit) (img *ebiten.Image, ax, ay int, mirror, ok bool) {
	sh := activeSheets().Get(actorKey(u) + "/salute")
	if sh == nil {
		return nil, 0, 0, false, false
	}
	img, mirror, ok = sh.image("S", FrameAt(animTime, sh.FPS, sh.Frames, true))
	if !ok {
		return nil, 0, 0, false, false
	}
	return img, sh.AnchorX, sh.AnchorY, mirror, true
}

func drawHead(dst, head *ebiten.Image, ax, ay, px, py int, mirror bool, x, y float64, cue idleCue) {
	if dst == nil || head == nil {
		return
	}
	if spriteHook != nil {
		spriteHook(head, mirror)
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
	// +tilt tips the helmet toward the character's right. Rotate before the
	// mirror so a west-facing man, drawn from the east cel, tips the same way
	// on his own body.
	op.GeoM.Translate(-float64(px), -float64(py))
	op.GeoM.Rotate(-cue.tilt)
	op.GeoM.Translate(float64(px)+cue.headX, float64(py)+cue.headY)
	op.GeoM.Translate(-float64(ax), -float64(ay))
	op.GeoM.Scale(sx, sc)
	op.GeoM.Translate(x*ps, y*ps)
	dst.DrawImage(head, op)
}

func soldierFidget(u *sim.Unit, p pose, aimLock bool) idleCue {
	if p != poseIdle && p != poseKneel {
		return idleCue{}
	}
	if u.Side != sim.SidePlayer && u.Side != sim.SideEnemy {
		return idleCue{}
	}
	return idleCueAt(u.ID, animTime, aimLock, u.Side == sim.SidePlayer)
}

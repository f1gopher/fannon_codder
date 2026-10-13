package render

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"fannon-codder/internal/sim"
	"fannon-codder/internal/terrain"
)

// A grenade, rocket, mine, or crate blast keeps a particle cloud after the
// painted sheet has finished. The sim marker still lasts blastTime. Shed
// fireballs (Scale > 1) are unchanged.

const boomLife = 1.6

type moteKind int

const (
	moteFlash moteKind = iota
	moteRing
	moteSpark
	moteEmber
	moteDebris
	moteSmoke
)

type mote struct {
	x, y   float64
	vx, vy float64
	age    float64
	life   float64
	kind   moteKind
	r0, r1 float64
	h, vh  float64
	red    uint8
	green  uint8
	blue   uint8
}

// boomTerrain is the phase kind Field is painting. Blasts uses it so debris
// matches the ground under the blast.
var boomTerrain string

type boom struct {
	x, y float64
	age  float64
	mote []mote
}

var booms []boom

func stepBooms(dt float64) {
	if dt <= 0 || len(booms) == 0 {
		return
	}
	out := booms[:0]
	for i := range booms {
		b := booms[i]
		b.age += dt
		if b.age >= boomLife {
			continue
		}
		live := b.mote[:0]
		for j := range b.mote {
			m := b.mote[j]
			m.age += dt
			if m.age >= m.life {
				continue
			}
			if m.kind == moteDebris && (m.h > 0 || m.vh != 0) {
				stepDebrisHop(&m, dt)
				live = append(live, m)
				continue
			}
			m.x += m.vx * dt
			m.y += m.vy * dt
			if m.kind == moteSpark || m.kind == moteEmber {
				m.vy += 100 * dt
			}
			live = append(live, m)
		}
		b.mote = live
		out = append(out, b)
	}
	booms = out
}

func noteGrenadeBooms(w *sim.World) {
	if w == nil {
		return
	}
	for i := range w.Explosions {
		e := &w.Explosions[i]
		if e.Scale > 1 || e.Life > 0 {
			continue
		}
		if boomAlready(e.X, e.Y, e.Age) {
			continue
		}
		booms = append(booms, makeBoom(w.Map, e.X, e.Y, e.Age))
	}
}

func boomAlready(x, y, age float64) bool {
	for i := range booms {
		b := &booms[i]
		if math.Hypot(b.x-x, b.y-y) < 2 && math.Abs(b.age-age) < 0.05 {
			return true
		}
	}
	return false
}

func makeBoom(m sim.Map, x, y, age float64) boom {
	rng := rand.New(rand.NewSource(int64(x*1000) ^ int64(y*9176) ^ 0x5eed))
	b := boom{x: x, y: y, age: age}
	b.mote = append(b.mote, mote{x: x, y: y, life: 0.16, kind: moteFlash, r0: 7, r1: 2})
	b.mote = append(b.mote, mote{x: x, y: y, life: 0.28, kind: moteRing, r0: 3, r1: 16})
	b.mote = append(b.mote, mote{x: x, y: y, life: 0.52, kind: moteRing, r0: 5, r1: 28})
	for n := 0; n < 72; n++ {
		ang := rng.Float64() * math.Pi * 2
		spd := 32 + rng.Float64()*110
		b.mote = append(b.mote, mote{
			x: x, y: y,
			vx:   math.Cos(ang) * spd,
			vy:   math.Sin(ang) * spd,
			life: 0.4 + rng.Float64()*0.7,
			kind: moteSpark,
			r0:   0.35 + rng.Float64()*0.45,
			r1:   0.12,
		})
	}
	for n := 0; n < 36; n++ {
		ang := rng.Float64() * math.Pi * 2
		spd := 12 + rng.Float64()*36
		b.mote = append(b.mote, mote{
			x: x, y: y,
			vx:   math.Cos(ang) * spd,
			vy:   math.Sin(ang)*spd - 6,
			life: 0.9 + rng.Float64()*0.7,
			kind: moteEmber,
			r0:   0.45 + rng.Float64()*0.35,
			r1:   0.15,
		})
	}
	addGroundDebris(&b, rng, terrain.GroundAt(m, boomTerrain, x, y), x, y)
	for n := 0; n < 40; n++ {
		ang := rng.Float64() * math.Pi * 2
		spd := 4 + rng.Float64()*12
		b.mote = append(b.mote, mote{
			x: x, y: y,
			vx:   math.Cos(ang) * spd,
			vy:   math.Sin(ang)*spd - 10,
			life: 1.1 + rng.Float64()*0.7,
			kind: moteSmoke,
			r0:   1.1,
			r1:   2.4 + rng.Float64()*1.8,
		})
	}
	// Catch up if the marker is already a few frames old.
	for i := range b.mote {
		m := &b.mote[i]
		m.age = age
		if age <= 0 {
			continue
		}
		m.x += m.vx * age
		m.y += m.vy * age
		if m.kind == moteSpark || m.kind == moteEmber {
			m.vy += 100 * age
		}
		if m.kind == moteDebris && m.vh != 0 {
			stepDebrisHop(m, age)
		}
	}
	return b
}

func drawBoomLayer(dst *ebiten.Image, cam sim.Camera, smoke bool) {
	if dst == nil {
		return
	}
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	for i := range booms {
		for j := range booms[i].mote {
			m := &booms[i].mote[j]
			isSmoke := m.kind == moteSmoke || m.kind == moteDebris
			if isSmoke != smoke || m.age >= m.life {
				continue
			}
			u := m.age / m.life
			sx := float32(cam.ScreenX(m.x) * ps)
			sy := float32(cam.ScreenY(m.y-m.h) * ps)
			rad := float32((m.r0 + (m.r1-m.r0)*u) * ps)
			if rad < 0.2 {
				continue
			}
			vector.FillCircle(dst, sx, sy, rad, moteColor(m, u), true)
			if m.kind == moteRing {
				vector.StrokeCircle(dst, sx, sy, rad, 1.5*float32(ps), moteColor(&mote{kind: moteSpark}, u), true)
			}
		}
	}
}

// stepDebrisHop moves a piece until it hits the ground, then leaves it there.
// A long frame does not slide it past the landing point.
func stepDebrisHop(m *mote, dt float64) {
	if dt <= 0 || (m.h <= 0 && m.vh == 0) {
		return
	}
	const g = 220.0
	disc := m.vh*m.vh + 2*g*m.h
	tHit := dt
	if disc >= 0 {
		tHit = (m.vh + math.Sqrt(disc)) / g
	}
	if tHit < dt {
		m.x += m.vx * tHit
		m.y += m.vy * tHit
		m.h = 0
		m.vh = 0
		m.vx = 0
		m.vy = 0
		return
	}
	m.x += m.vx * dt
	m.y += m.vy * dt
	m.h += m.vh*dt - 0.5*g*dt*dt
	m.vh -= g * dt
}

func addGroundDebris(b *boom, rng *rand.Rand, g terrain.Ground, x, y float64) {
	n := 20
	reach0, reach1 := 4.0, 10.0
	lift0, lift1 := 50.0, 90.0
	r0 := 2.4
	var red, green, blue uint8
	switch g {
	case terrain.GroundWater:
		n = 32
		reach0, reach1 = 5, 12
		lift0, lift1 = 60, 100
		r0 = 1.8
		red, green, blue = 150, 206, 224
	case terrain.GroundSand:
		n = 36
		reach0, reach1 = 3, 8
		lift0, lift1 = 36, 70
		r0 = 1.5
		red, green, blue = 214, 184, 120
	case terrain.GroundSnow:
		n = 26
		reach0, reach1 = 4, 9
		lift0, lift1 = 42, 80
		r0 = 2.1
		red, green, blue = 236, 242, 246
	default:
		red, green, blue = 112, 74, 40
	}
	for i := 0; i < n; i++ {
		ang := rng.Float64() * math.Pi * 2
		vh := lift0 + rng.Float64()*(lift1-lift0)
		air := 2 * vh / 220
		if air < 0.05 {
			air = 0.05
		}
		// Land inside about one soldier height (14 px).
		spd := (reach0 + rng.Float64()*(reach1-reach0)) / air
		size := r0 + rng.Float64()*0.9
		b.mote = append(b.mote, mote{
			x: x, y: y,
			vx:   math.Cos(ang) * spd,
			vy:   math.Sin(ang) * spd,
			vh:   vh,
			life: 1.15 + rng.Float64()*0.35,
			kind: moteDebris,
			r0:   size,
			r1:   size * 0.75,
			red:  red, green: green, blue: blue,
		})
	}
}

func moteColor(m *mote, u float64) color.RGBA {
	fade := 1 - u
	if fade < 0 {
		fade = 0
	}
	switch m.kind {
	case moteFlash:
		a := uint8(255 * fade * fade)
		return color.RGBA{R: 255, G: 244, B: 220, A: a}
	case moteRing:
		a := uint8(180 * fade)
		return color.RGBA{R: 255, G: 186, B: 64, A: a}
	case moteSpark:
		a := uint8(230 * fade)
		g := uint8(220 - 150*u)
		return color.RGBA{R: 255, G: g, B: uint8(40 * fade), A: a}
	case moteEmber:
		a := uint8(210 * fade)
		return color.RGBA{R: 255, G: uint8(90 + 40*fade), B: 24, A: a}
	case moteDebris:
		// Stay solid, then fade in the last stretch.
		if u > 0.45 {
			fade = (1 - u) / 0.55
		} else {
			fade = 1
		}
		return color.RGBA{R: m.red, G: m.green, B: m.blue, A: uint8(230 * fade)}
	default:
		a := uint8(90 * fade * fade)
		shade := uint8(48 + 30*u)
		return color.RGBA{R: shade, G: shade - 6, B: shade - 8, A: a}
	}
}

package render

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

func TestGrenadeBoomOutlastsTheSheet(t *testing.T) {
	booms = nil
	t.Cleanup(func() { booms = nil })
	w := sim.NewEmpty()
	w.Camera = sim.Camera{ViewW: 200, ViewH: 200, MapW: 200, MapH: 200}
	w.Explosions = []sim.Explosion{{X: 40, Y: 40, R: sim.GrenadeRadius}}
	dst := ebiten.NewImage(120, 120)
	Blasts(dst, w)
	if len(booms) != 1 || len(booms[0].mote) < 150 {
		t.Fatalf("booms=%d motes=%d", len(booms), len(booms[0].mote))
	}
	w.Explosions = append(w.Explosions, sim.Explosion{X: 90, Y: 40, R: 20, Scale: 2.15, Life: 1.25})
	Blasts(dst, w)
	if len(booms) != 1 {
		t.Fatalf("shed fire should not add a particle cloud, got %d", len(booms))
	}
	Advance(0.9, w)
	if len(booms) != 1 || !debrisStillUp(booms[0]) {
		t.Fatal("debris should still be visible after the sparks and smoke")
	}
	Advance(0.8, w)
	if len(booms) != 0 {
		t.Fatalf("cloud still alive at %v", booms[0].age)
	}
}

func debrisStillUp(b boom) bool {
	for i := range b.mote {
		if b.mote[i].kind == moteDebris && b.mote[i].r0 >= 1.5 {
			return true
		}
	}
	return false
}

func TestBoomDebrisMatchesTheGround(t *testing.T) {
	booms = nil
	t.Cleanup(func() { booms = nil })
	w := sim.NewEmpty()
	w.Map = sim.Map{W: 2, H: 1, Tiles: []sim.Tile{sim.TileGrass, sim.TileWaterShallow}}
	w.Explosions = []sim.Explosion{{X: 8, Y: 8, R: sim.GrenadeRadius}}
	boomTerrain = "jungle"
	dst := ebiten.NewImage(40, 40)
	Blasts(dst, w)
	if !debrisIs(112, 74, 40) {
		t.Fatal("jungle grass should throw soil")
	}
	booms = nil
	w.Explosions = []sim.Explosion{{X: 24, Y: 8, R: sim.GrenadeRadius}}
	Blasts(dst, w)
	if !debrisIs(150, 206, 224) {
		t.Fatal("water should throw spray")
	}
	booms = nil
	boomTerrain = "snow"
	w.Explosions = []sim.Explosion{{X: 8, Y: 8, R: sim.GrenadeRadius}}
	Blasts(dst, w)
	if !debrisIs(236, 242, 246) {
		t.Fatal("snow should throw snow")
	}
}

func TestDebrisLandsNearTheBlast(t *testing.T) {
	booms = nil
	t.Cleanup(func() { booms = nil })
	w := sim.NewEmpty()
	w.Map = sim.Map{W: 1, H: 1, Tiles: []sim.Tile{sim.TileGrass}}
	w.Explosions = []sim.Explosion{{X: 40, Y: 40, R: sim.GrenadeRadius}}
	boomTerrain = "jungle"
	Blasts(ebiten.NewImage(8, 8), w)
	Advance(1.2, w)
	const soldier = 14.0
	for i := range booms[0].mote {
		m := &booms[0].mote[i]
		if m.kind != moteDebris {
			continue
		}
		if m.h != 0 || m.vx != 0 || m.vy != 0 {
			t.Fatalf("debris still moving at %+v", m)
		}
		if d := math.Hypot(m.x-40, m.y-40); d > soldier {
			t.Fatalf("debris landed %v px out, soldier height is %v", d, soldier)
		}
	}
}

func TestGrenadeLeavesAScorch(t *testing.T) {
	scorches = nil
	booms = nil
	t.Cleanup(func() {
		scorches = nil
		booms = nil
	})
	saved := pictureScale
	pictureScale = 1
	t.Cleanup(func() { pictureScale = saved })

	w := sim.NewEmpty()
	w.Map = sim.Map{W: 4, H: 1, Tiles: []sim.Tile{
		sim.TileGrass, sim.TileGrass, sim.TileWaterShallow, sim.TileWaterDeep,
	}}
	w.Camera = sim.Camera{ViewW: 80, ViewH: 40, MapW: 64, MapH: 16}
	boomTerrain = "jungle"
	w.Explosions = []sim.Explosion{{X: 8, Y: 8, R: sim.GrenadeRadius}}
	dst := ebiten.NewImage(80, 40)
	Blasts(dst, w)
	if len(scorches) != 1 || scorches[0].img == nil {
		t.Fatalf("soil scorch = %+v", scorches)
	}
	mask := bakeScorch(8, 8)
	mid := mask.RGBAAt(scorchTex/2, scorchTex/2)
	rim := mask.RGBAAt(2, 2)
	if mid.R >= rim.R-8 || mid.R > 210 {
		t.Fatalf("centre should darken the ground, mid=%v rim=%v", mid, rim)
	}
	if mid.B+6 >= mid.R {
		t.Fatalf("centre should warm the ground, got %v", mid)
	}
	half := mask.RGBAAt(scorchTex/2, scorchTex/2+scorchTex/5)
	if half.R <= mid.R || half.R >= rim.R {
		t.Fatalf("the stain should feather, mid=%v half=%v rim=%v", mid.R, half.R, rim.R)
	}
	w.Explosions = append(w.Explosions, sim.Explosion{X: 40, Y: 8, R: sim.GrenadeRadius})
	Blasts(dst, w)
	if len(scorches) != 1 {
		t.Fatalf("water should leave no mark, got %d", len(scorches))
	}
	w.Explosions = append(w.Explosions, sim.Explosion{X: 20, Y: 8, R: 20, Scale: 2.15, Life: 1.25})
	Blasts(dst, w)
	if len(scorches) != 1 {
		t.Fatalf("shed fire should leave no mark, got %d", len(scorches))
	}
	Blasts(dst, w)
	if len(scorches) != 1 {
		t.Fatalf("the same blast should not stack marks, got %d", len(scorches))
	}
	Advance(2, w)
	if len(scorches) != 1 {
		t.Fatal("the mark should outlast the cloud")
	}
	drawScorches(dst, w.Camera)
	ResetAnim()
	if len(scorches) != 0 {
		t.Fatal("ResetAnim should clear scorches")
	}
}

func debrisIs(r, g, b uint8) bool {
	for i := range booms {
		for j := range booms[i].mote {
			m := &booms[i].mote[j]
			if m.kind == moteDebris && m.red == r && m.green == g && m.blue == b && m.vh > 0 {
				return true
			}
		}
	}
	return false
}

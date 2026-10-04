package render

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

// Flavour is one painted prop. It does not block movement or fire.
// Arctic maps grow snowmen and igloos. Grass maps grow scrub.
// Birds are separate and cross every map.
type Flavour struct {
	Key   string
	X, Y  float64
	Tiles int
}

// Bird flies level across the map. Westbound birds mirror the east sheet.
type Bird struct {
	Y     float64
	Speed float64
	Phase float64
	West  bool
}

// FlavourSites picks open grass cells. An igloo takes the cell and the one
// east of it. Trees, water, cliffs, and hut footprints stay clear.
func FlavourSites(m sim.Map, buildings []sim.Building, arctic bool) []Flavour {
	if m.W < 3 || m.H < 3 || len(m.Tiles) < m.W*m.H {
		return nil
	}
	used := map[int]bool{}
	var out []Flavour
	for ty := 1; ty < m.H-1; ty++ {
		for tx := 1; tx < m.W-1; tx++ {
			if used[ty*m.W+tx] || !openFlavour(m, buildings, tx, ty) {
				continue
			}
			h := tx*13 + ty*29 + m.W*3 + m.H
			footX := float64(tx*sim.TileSize) + float64(sim.TileSize)/2
			footY := float64((ty + 1) * sim.TileSize)
			if arctic {
				if h%29 == 0 && tx+1 < m.W-1 && openFlavour(m, buildings, tx+1, ty) && !used[ty*m.W+tx+1] {
					out = append(out, Flavour{Key: "prop/igloo", X: footX + float64(sim.TileSize)/2, Y: footY, Tiles: 2})
					used[ty*m.W+tx] = true
					used[ty*m.W+tx+1] = true
					continue
				}
				if h%17 == 0 {
					out = append(out, Flavour{Key: "prop/snowman", X: footX, Y: footY, Tiles: 1})
					used[ty*m.W+tx] = true
				}
				continue
			}
			if h%19 == 0 {
				out = append(out, Flavour{Key: "prop/scrub", X: footX, Y: footY, Tiles: 1})
				used[ty*m.W+tx] = true
			}
		}
	}
	return out
}

func openFlavour(m sim.Map, buildings []sim.Building, tx, ty int) bool {
	if m.At(tx, ty) != sim.TileGrass {
		return false
	}
	x0 := float64(tx * sim.TileSize)
	y0 := float64(ty * sim.TileSize)
	x1 := x0 + float64(sim.TileSize)
	y1 := y0 + float64(sim.TileSize)
	for _, b := range buildings {
		if !b.Alive {
			continue
		}
		if b.X < x1 && b.X+b.W > x0 && b.Y < y1 && b.Y+b.H > y0 {
			return false
		}
	}
	return true
}

// Birds returns two to four flyers. Their lanes and phases come from the map size.
func Birds(m sim.Map) []Bird {
	w := m.W
	h := m.H
	if w < 1 {
		w = 20
	}
	if h < 1 {
		h = 16
	}
	n := 2 + w/12
	if n > 4 {
		n = 4
	}
	mapH := float64(h * sim.TileSize)
	out := make([]Bird, n)
	for i := range out {
		lane := 0.28 + float64(i)/float64(n)*0.5
		out[i] = Bird{
			Y:     mapH * lane,
			Speed: 22 + float64((i*7)%13),
			Phase: float64(i*90 + w*5),
			West:  i%2 == 1,
		}
	}
	return out
}

// At is the bird's world position at time t. x wraps past the map edges.
func (b Bird) At(mapW, t float64) (x, y float64) {
	if mapW < 1 {
		mapW = 320
	}
	span := mapW + 64
	x = math.Mod(t*b.Speed+b.Phase, span)
	if x < 0 {
		x += span
	}
	if b.West {
		x = span - x
	}
	return x - 32, b.Y
}

func flavourBodies(w *sim.World, arctic bool) []body {
	sites := FlavourSites(w.Map, w.Buildings, arctic)
	bodies := make([]body, 0, len(sites))
	for _, s := range sites {
		sh := activeSheets().Get(s.Key)
		if sh == nil {
			continue
		}
		img, ok := sh.still(0)
		if !ok {
			continue
		}
		sx := w.Camera.ScreenX(s.X)
		sy := w.Camera.ScreenY(s.Y)
		bodies = append(bodies, spriteBody(s.Y, sx, sy, img, sh.AnchorX, sh.AnchorY, false, nil))
	}
	return bodies
}

func drawBirds(dst *ebiten.Image, w *sim.World) {
	sh := activeSheets().Get("bird/flap")
	if sh == nil || w.Map.W == 0 {
		return
	}
	frame := FrameAt(animTime, sh.FPS, sh.Frames, sh.Loop)
	mapW, _ := w.Map.PixelSize()
	for _, b := range Birds(w.Map) {
		x, y := b.At(mapW, animTime)
		dir := "E"
		if b.West {
			dir = "W"
		}
		img, mirror, ok := sh.image(dir, frame)
		if !ok {
			continue
		}
		DrawSprite(dst, img, sh.AnchorX, sh.AnchorY, mirror, w.Camera.ScreenX(x), w.Camera.ScreenY(y))
	}
}

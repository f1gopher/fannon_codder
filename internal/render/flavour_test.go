package render

import (
	"testing"

	"fannon-codder/assets/art"
	"fannon-codder/internal/sim"
)

func grassMap(w, h int) sim.Map {
	return sim.Map{W: w, H: h, Tiles: make([]sim.Tile, w*h)}
}

func TestFlavourSheets(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	bird := lib.Get("bird/flap")
	if bird == nil || bird.Frames != 4 || bird.FPS != 8 || !bird.Loop || bird.FrameW != 128 || bird.FrameH != 96 {
		t.Fatalf("bird flap %+v", bird)
	}
	if _, mirror, ok := bird.image("W", 0); !ok || !mirror {
		t.Fatal("west flight should mirror the east row")
	}
	img := decodeArtPNG(t, "bird/flap.png")
	if img.Bounds().Dx() != 128*4 || img.Bounds().Dy() != 96 {
		t.Fatalf("bird bounds %v", img.Bounds())
	}
	if frameDelta(img, 128, 0, 1) < 0.4 || frameDelta(img, 128, 1, 2) < 0.4 {
		t.Fatal("wing frames do not move")
	}
	for _, key := range []string{"prop/scrub", "prop/snowman", "prop/igloo"} {
		sh := lib.Get(key)
		if sh == nil || sh.Frames != 1 {
			t.Fatalf("%s missing", key)
		}
	}
	if lib.Get("prop/igloo").FrameW != 256 || lib.Get("prop/snowman").FrameH != 192 {
		t.Fatal("prop cells")
	}
}

func TestGrassGrowsScrubAndSnowGrowsIgloos(t *testing.T) {
	m := grassMap(20, 16)
	m.Set(4, 4, sim.TileTree)
	hut := []sim.Building{{X: 6 * 16, Y: 6 * 16, W: 32, H: 32, Alive: true}}
	grass := FlavourSites(m, hut, false)
	if len(grass) == 0 {
		t.Fatal("grass map should grow scrub")
	}
	for _, s := range grass {
		if s.Key != "prop/scrub" {
			t.Fatalf("grass site %s", s.Key)
		}
	}
	snow := FlavourSites(m, hut, true)
	var igloo, man int
	for _, s := range snow {
		switch s.Key {
		case "prop/igloo":
			igloo++
		case "prop/snowman":
			man++
		default:
			t.Fatalf("snow site %s", s.Key)
		}
		tx := int(s.X) / sim.TileSize
		ty := int(s.Y)/sim.TileSize - 1
		if m.At(tx, ty) == sim.TileTree || (s.Tiles == 2 && m.At(tx-1, ty) == sim.TileTree) {
			t.Fatalf("flavour on the tree at %v", s)
		}
		if tx >= 6 && tx <= 7 && ty >= 6 && ty <= 7 {
			t.Fatalf("flavour on the hut at %v", s)
		}
	}
	if igloo == 0 || man == 0 {
		t.Fatalf("igloos=%d snowmen=%d", igloo, man)
	}
}

func TestBirdsCrossTheMap(t *testing.T) {
	m := grassMap(20, 16)
	birds := Birds(m)
	if len(birds) < 2 {
		t.Fatalf("birds %d", len(birds))
	}
	mapW, _ := m.PixelSize()
	x0, _ := birds[0].At(mapW, 0)
	x1, _ := birds[0].At(mapW, 1)
	if x0 == x1 {
		t.Fatal("bird did not move")
	}
	west := birds[1]
	if !west.West {
		t.Fatal("second bird should fly west")
	}
	a, _ := west.At(mapW, 0)
	b, _ := west.At(mapW, 1)
	if b >= a && a > 32 {
		t.Fatalf("west bird went %v to %v", a, b)
	}
}

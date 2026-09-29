package app

import (
	"math"
	"testing"
)

func TestPictureScale(t *testing.T) {
	cases := []struct {
		name       string
		dipW, dipH float64
		device     float64
		s          float64
		offW, offH int
	}{
		{"default window", 1024, 768, 1, 3, 960, 768},
		{"4K", 3840, 2160, 1, 2160.0 / 256.0, 2700, 2160},
		{"default window, 2× monitor", 1024, 768, 2, 6, 1920, 1536},
		{"4K, 2× monitor, clamped", 3840, 2160, 2, 2160.0 / 256.0, 2700, 2160},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, w, h := PictureSize(c.dipW, c.dipH, c.device)
			if math.Abs(s-c.s) > 1e-9 {
				t.Fatalf("S = %v, want %v", s, c.s)
			}
			if w != c.offW || h != c.offH {
				t.Fatalf("offscreen = %d×%d, want %d×%d", w, h, c.offW, c.offH)
			}
		})
	}
}

func TestTallWindowFitsTheWidth(t *testing.T) {
	s, w, h := PictureSize(1024, 2000, 1)
	want := 1024.0 / 320
	if math.Abs(s-want) > 1e-9 {
		t.Fatalf("S = %v, want %v", s, want)
	}
	if w != 1024 || h != int(math.Round(256*want)) {
		t.Fatalf("offscreen = %d×%d", w, h)
	}
}

func TestWiderWindowKeepsTheFrame(t *testing.T) {
	s1, w1, h1 := PictureSize(1024, 768, 1)
	s2, w2, h2 := PictureSize(1600, 768, 1)
	if s1 != s2 || w1 != w2 || h1 != h2 {
		t.Fatalf("wider window changed the picture: %v %d×%d vs %v %d×%d", s1, w1, h1, s2, w2, h2)
	}
}

func TestOffscreenCursorIdentityWhenBlitScaleIsOne(t *testing.T) {
	x, y := OffscreenCursor(100, 40, 1024, 768, 960, 768)
	if x != 100 || y != 40 {
		t.Fatalf("got (%v,%v), want (100,40)", x, y)
	}
}

func TestOffscreenCursorFollowsAResizedFramebuffer(t *testing.T) {
	// A window manager can resize the frame away from the requested 1024×768.
	// The picture then fills the real width and is letterboxed vertically.
	// Ebitengine's scale is 1, so the layout cursor is already an offscreen pixel.
	const fbW, fbH = 692, 836
	const offW, offH = 692, 554
	x, y := OffscreenCursor(0, 0, fbW, fbH, offW, offH)
	if math.Abs(x) > 0.01 || math.Abs(y) > 0.01 {
		t.Fatalf("top-left got (%v,%v), want (0,0)", x, y)
	}
	x, y = OffscreenCursor(691, 553, fbW, fbH, offW, offH)
	if math.Abs(x-691) > 0.01 || math.Abs(y-553) > 0.01 {
		t.Fatalf("bottom-right got (%v,%v), want (691,553)", x, y)
	}
}

func TestOffscreenCursorClampedHiDPI(t *testing.T) {
	// Framebuffer 7680×4320, picture 2700×2160. Ebitengine would scale by 2.
	// The 1:1 blit’s top-left is physical (2490, 1080), which Ebitengine
	// reports as layout (675, 540).
	x, y := OffscreenCursor(675, 540, 7680, 4320, 2700, 2160)
	if math.Abs(x) > 0.01 || math.Abs(y) > 0.01 {
		t.Fatalf("top-left got (%v,%v), want (0,0)", x, y)
	}
	x, y = OffscreenCursor(1350, 1080, 7680, 4320, 2700, 2160)
	if math.Abs(x-1350) > 0.01 || math.Abs(y-1080) > 0.01 {
		t.Fatalf("centre got (%v,%v), want (1350,1080)", x, y)
	}
}

func TestFramePointRejectsTheBar(t *testing.T) {
	_, w, h := PictureSize(1024, 768, 1)
	x, y, inside := FramePoint(-10, 20, 3, w, h)
	if inside {
		t.Fatal("a point in the left bar reached the frame")
	}
	if x != 0 || math.Abs(y-20.0/3) > 1e-9 {
		t.Fatalf("clamped (%v,%v)", x, y)
	}
	x, y, inside = FramePoint(150, 90, 3, w, h)
	if !inside {
		t.Fatal("a point on the picture was treated as a bar")
	}
	if math.Abs(x-50) > 1e-9 || math.Abs(y-30) > 1e-9 {
		t.Fatalf("logical (%v,%v), want (50,30)", x, y)
	}
}

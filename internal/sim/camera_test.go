package sim

import "testing"

func TestClampMapSmallerThanView(t *testing.T) {
	c := Camera{X: 40, Y: -10, ViewW: 320, ViewH: 256, MapW: 320, MapH: 256}
	c.Clamp()
	if c.X != 0 || c.Y != 0 {
		t.Fatalf("got (%v,%v), want (0,0)", c.X, c.Y)
	}
}

func TestClampMapLargerThanView(t *testing.T) {
	c := Camera{X: 999, Y: 999, ViewW: 320, ViewH: 256, MapW: 640, MapH: 512}
	c.Clamp()
	if c.X != 320 || c.Y != 256 {
		t.Fatalf("got (%v,%v), want (320,256)", c.X, c.Y)
	}
}

func TestScrollTowardNoOpWhenMapFits(t *testing.T) {
	c := Camera{ViewW: 320, ViewH: 256, MapW: 320, MapH: 256}
	c.ScrollToward(0, 0, 320, 256, 1)
	if c.X != 0 || c.Y != 0 {
		t.Fatalf("one-screen map should not pan, got (%v,%v)", c.X, c.Y)
	}
}

func TestCoverMapLargerThanView(t *testing.T) {
	w := NewCoverWorld()
	mw, mh := w.Map.PixelSize()
	if mw <= w.Camera.ViewW || mh <= w.Camera.ViewH {
		t.Fatalf("cover map %v×%v should exceed view %v×%v", mw, mh, w.Camera.ViewW, w.Camera.ViewH)
	}
	w.Camera.ScrollToward(319, 255, 320, 256, 1)
	if w.Camera.X <= 0 || w.Camera.Y <= 0 {
		t.Fatalf("expected pan on cover map, got (%v,%v)", w.Camera.X, w.Camera.Y)
	}
}

func TestScrollTowardPansOnLargeMap(t *testing.T) {
	c := Camera{ViewW: 320, ViewH: 256, MapW: 640, MapH: 512}
	c.ScrollToward(319, 255, 320, 256, 1)
	if c.X <= 0 || c.Y <= 0 {
		t.Fatalf("expected pan down-right, got (%v,%v)", c.X, c.Y)
	}
	if c.X > 320 || c.Y > 256 {
		t.Fatalf("pan exceeded clamp, got (%v,%v)", c.X, c.Y)
	}
}

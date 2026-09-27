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

func TestCenterOnLeaderClampsWhenMapFits(t *testing.T) {
	c := Camera{ViewW: 320, ViewH: 256, MapW: 320, MapH: 256}
	c.CenterOn(200, 200)
	if c.X != 0 || c.Y != 0 {
		t.Fatalf("one-screen map stays pinned, got (%v,%v)", c.X, c.Y)
	}
}

func TestCenterOnLargeMap(t *testing.T) {
	c := Camera{ViewW: 320, ViewH: 256, MapW: 640, MapH: 512}
	c.CenterOn(400, 300)
	if c.X != 240 || c.Y != 172 {
		t.Fatalf("got (%v,%v), want (240,172)", c.X, c.Y)
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

func TestOneScreenMapScrollsBesideStrip(t *testing.T) {
	// Mission 1 is 320×256, the full logical screen. The status strip covers
	// the left 52px, so the playfield is narrower and the west edge must pan
	// out from under the strip.
	c := Camera{ViewW: 268, ViewH: 256, MapW: 320, MapH: 256, OriginX: 52}
	c.CenterOn(168, 216)
	if c.X <= 0 {
		t.Fatalf("centred squad should leave room to pan west, x=%v", c.X)
	}
	west := 24.0 // tile (1, 5), the grunt behind the strip
	if c.ScreenX(west) >= c.OriginX {
		t.Fatalf("west grunt should start left of the playfield, screen x=%v", c.ScreenX(west))
	}
	c.ScrollToward(0, 128, 268, 256, 1)
	if c.X != 0 {
		t.Fatalf("edge scroll should reach the west map edge, x=%v", c.X)
	}
	if got := c.ScreenX(west); got < c.OriginX {
		t.Fatalf("west grunt still under the strip at screen x=%v", got)
	}
}

func TestContainPullsLeaderOffTheStrip(t *testing.T) {
	c := Camera{X: 34, ViewW: 268, ViewH: 256, MapW: 320, MapH: 256, OriginX: 52}
	c.Contain(24, 128)
	if c.X != 0 {
		t.Fatalf("camera x=%v, want 0", c.X)
	}
	if got := c.ScreenX(24); got < c.OriginX {
		t.Fatalf("leader still under the strip at screen x=%v", got)
	}
}

func TestContainBlocksPanningSquadOffscreen(t *testing.T) {
	c := Camera{ViewW: 268, ViewH: 256, MapW: 640, MapH: 512, OriginX: 52}
	c.Contain(300, 200)
	c.ScrollToward(0, 0, 268, 256, 1)
	c.Contain(300, 200)
	if 300 < c.X+LeaderMargin {
		t.Fatalf("leader left the west margin: cam %v leader 300", c.X)
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

package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/assets/art"
	"fannon-codder/internal/sim"
)

func TestFacing(t *testing.T) {
	cases := []struct {
		ang  float64
		want string
	}{
		{0, "E"},
		{math.Pi / 2, "S"},
		{math.Pi, "W"},
		{-math.Pi, "W"},
		{-math.Pi / 2, "N"},
	}
	for _, tc := range cases {
		if got := Direction(tc.ang); got != tc.want {
			t.Fatalf("facing %v → %s, want %s", tc.ang, got, tc.want)
		}
	}
}

func TestSceneryPhaseDiffersByTile(t *testing.T) {
	a := SceneryFrame(0, 0, 4, 0, 8)
	b := SceneryFrame(1, 0, 4, 0, 8)
	if a == b {
		t.Fatalf("tile (0,0) and (1,0) share frame %d", a)
	}
	if a != 0 || b != 3 {
		t.Fatalf("(0,0)=%d (1,0)=%d, want 0 and 3", a, b)
	}
}

func TestMirrorMustNameAStoredRow(t *testing.T) {
	dir := t.TempDir()
	body := []byte(`{"frameW":8,"frameH":8,"rows":["E"],"mirrors":{"W":"S"}}`)
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(os.DirFS(dir)); err == nil {
		t.Fatal("mirror onto a missing row was accepted")
	}
}

func TestSnakeIdleAndWalkSheets(t *testing.T) {
	f, err := art.Files.Open("style/contact-sheet.png")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	idle := lib.Get("snake/idle")
	walk := lib.Get("snake/walk")
	if idle == nil || walk == nil {
		t.Fatal("snake idle and walk sheets should be embedded")
	}
	if idle.Frames != 4 || idle.FPS != 8 || !idle.Loop {
		t.Fatalf("idle frames=%d fps=%v loop=%v", idle.Frames, idle.FPS, idle.Loop)
	}
	if walk.Frames != 6 || walk.FPS != 12 || !walk.Loop {
		t.Fatalf("walk frames=%d fps=%v loop=%v", walk.Frames, walk.FPS, walk.Loop)
	}
	if idle.FrameW != walk.FrameW || idle.FrameH != walk.FrameH || idle.AnchorX != walk.AnchorX || idle.AnchorY != walk.AnchorY {
		t.Fatalf("idle and walk cells differ: idle %dx%d @%d,%d walk %dx%d @%d,%d",
			idle.FrameW, idle.FrameH, idle.AnchorX, idle.AnchorY,
			walk.FrameW, walk.FrameH, walk.AnchorX, walk.AnchorY)
	}
	for _, d := range []string{"E", "SE", "S", "N", "NE"} {
		if _, mirror, ok := idle.image(d, 0); !ok || mirror {
			t.Fatalf("%s should be a stored row", d)
		}
		if _, _, ok := walk.image(d, 0); !ok {
			t.Fatalf("walk missing %s", d)
		}
	}
	for _, d := range []string{"W", "SW", "NW"} {
		if _, mirror, ok := idle.image(d, 0); !ok || !mirror {
			t.Fatalf("%s should mirror a stored row", d)
		}
	}
	if lib.Get("eagle/idle") != nil || lib.Get("grunt/idle") != nil {
		t.Fatal("only snake idle and walk are production sheets")
	}
}

func TestSheetPlaysAFrame(t *testing.T) {
	restoreRender(t)
	SetPictureScale(8)
	ResetAnim()
	lib := testIdleSheet(t)
	SetSheets(lib)
	sh := lib.Get("snake/idle")
	if sh == nil {
		t.Fatal("test sheet did not load")
	}
	if _, mirror, ok := sh.image("W", 0); !ok || !mirror {
		t.Fatalf("π faces W, which is a mirror: ok=%v mirror=%v", ok, mirror)
	}
	if _, mirror, ok := sh.image("E", 0); !ok || mirror {
		t.Fatalf("E is a stored row: ok=%v mirror=%v", ok, mirror)
	}

	frame0, mirror0, ok := sh.image("E", 0)
	frame1, _, ok1 := sh.image("E", 1)
	if !ok || !ok1 || frame0 == nil || frame1 == nil || frame0 == frame1 {
		t.Fatal("test sheet did not keep two distinct frames")
	}
	if mirror0 {
		t.Fatal("E is stored, not a mirror")
	}
	var drawn []*ebiten.Image
	var mirrored []bool
	var rects int
	spriteHook = func(img *ebiten.Image, mirror bool) {
		drawn = append(drawn, img)
		mirrored = append(mirrored, mirror)
	}
	rectHook = func() { rects++ }
	t.Cleanup(func() { spriteHook, rectHook = nil, nil })

	w := trooperWorld()
	u := w.Unit(w.ActiveSquad().LeaderID)
	u.Facing = 0
	paintTrooper(w)
	if rects != 0 || len(drawn) != 1 || drawn[0] != frame0 || mirrored[0] {
		t.Fatalf("east frame 0: rects=%d drawn=%d", rects, len(drawn))
	}

	drawn, mirrored, rects = nil, nil, 0
	u.Facing = math.Pi
	paintTrooper(w)
	if rects != 0 || len(drawn) != 1 || drawn[0] != frame0 || !mirrored[0] {
		t.Fatalf("west should mirror frame 0: rects=%d drawn=%d mirror=%v", rects, len(drawn), mirrored)
	}

	drawn, mirrored, rects = nil, nil, 0
	u.Facing = 0
	Advance(1, nil)
	paintTrooper(w)
	if rects != 0 || len(drawn) != 1 || drawn[0] != frame1 || mirrored[0] {
		t.Fatalf("second frame: rects=%d drawn=%d", rects, len(drawn))
	}

	drawn, rects = nil, 0
	u.Facing = math.Pi / 2
	paintTrooper(w)
	if len(drawn) != 0 || rects != 1 {
		t.Fatalf("south has no row: sprites=%d rects=%d", len(drawn), rects)
	}
}

func TestMissingSheetDrawsRectangle(t *testing.T) {
	restoreRender(t)
	SetPictureScale(1)
	ResetAnim()
	SetSheets(&Library{})
	var sprites, rects int
	spriteHook = func(*ebiten.Image, bool) { sprites++ }
	rectHook = func() { rects++ }
	t.Cleanup(func() { spriteHook, rectHook = nil, nil })
	paintTrooper(trooperWorld())
	if sprites != 0 || rects != 1 {
		t.Fatalf("missing sheet: sprites=%d rects=%d, want the rectangle", sprites, rects)
	}
}

func trooperWorld() *sim.World {
	w := sim.NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.SpawnPlayerSquad(sim.SquadSnake, []sim.Vec2{{X: 40, Y: 40}})
	return w
}

func paintTrooper(w *sim.World) {
	dst := ebiten.NewImage(64, 64)
	Field(dst, w, "grass")
}

func testIdleSheet(t *testing.T) *Library {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 8))
	red := color.RGBA{R: 255, A: 255}
	green := color.RGBA{G: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			c := red
			if x >= 4 {
				c = green
			}
			img.SetRGBA(x, y, c)
			img.SetRGBA(x+8, y, blue)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	sub := filepath.Join(dir, "snake")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := []byte(`{
		"frameW": 8,
		"frameH": 8,
		"anchorX": 4,
		"anchorY": 4,
		"fps": 1,
		"loop": true,
		"rows": ["E"],
		"mirrors": {"W": "E"}
	}`)
	if err := os.WriteFile(filepath.Join(sub, "idle.json"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "idle.png"), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	lib, err := Load(os.DirFS(dir))
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

func restoreRender(t *testing.T) {
	t.Helper()
	ps := pictureScale
	at := animTime
	ages := deathAge
	ov := sheetsOverride
	t.Cleanup(func() {
		pictureScale = ps
		animTime = at
		deathAge = ages
		sheetsOverride = ov
	})
}

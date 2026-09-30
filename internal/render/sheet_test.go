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
	for _, key := range []string{"snake/shoot", "snake/throw", "snake/death", "snake/corpse"} {
		if lib.Get(key) == nil {
			t.Fatalf("missing %s", key)
		}
	}
}

func TestSnakeFightSheets(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	shoot := lib.Get("snake/shoot")
	throw := lib.Get("snake/throw")
	death := lib.Get("snake/death")
	corpse := lib.Get("snake/corpse")
	if shoot == nil || throw == nil || death == nil || corpse == nil {
		t.Fatal("snake shoot, throw, death, and corpse should be embedded")
	}
	if shoot.Frames != 2 || shoot.Loop {
		t.Fatalf("shoot frames=%d loop=%v", shoot.Frames, shoot.Loop)
	}
	if throw.Frames != 4 || throw.Loop {
		t.Fatalf("throw frames=%d loop=%v", throw.Frames, throw.Loop)
	}
	if death.Frames != 6 || death.FPS != 15 || death.Loop {
		t.Fatalf("death frames=%d fps=%v loop=%v", death.Frames, death.FPS, death.Loop)
	}
	if float64(death.Frames)/death.FPS != 0.4 {
		t.Fatalf("death lasts %v, want 0.4s", float64(death.Frames)/death.FPS)
	}
	if FrameAt(0.39, death.FPS, death.Frames, false) != 5 {
		t.Fatal("the last death frame should still be showing just before 0.4s")
	}
	if corpse.Frames != 1 || corpse.Loop {
		t.Fatalf("corpse frames=%d loop=%v", corpse.Frames, corpse.Loop)
	}
	if corpse.FrameW != death.FrameW || corpse.FrameH != death.FrameH || corpse.AnchorX != death.AnchorX || corpse.AnchorY != death.AnchorY {
		t.Fatalf("corpse cell %dx%d @%d,%d, death cell %dx%d @%d,%d",
			corpse.FrameW, corpse.FrameH, corpse.AnchorX, corpse.AnchorY,
			death.FrameW, death.FrameH, death.AnchorX, death.AnchorY)
	}
	for _, sh := range []*Sheet{shoot, throw, death, corpse} {
		for _, d := range []string{"E", "SE", "S", "N", "NE"} {
			if _, mirror, ok := sh.image(d, 0); !ok || mirror {
				t.Fatalf("%s should be a stored row", d)
			}
		}
		for _, d := range []string{"W", "SW", "NW"} {
			if _, mirror, ok := sh.image(d, 0); !ok || !mirror {
				t.Fatalf("%s should mirror a stored row", d)
			}
		}
	}
	deathPNG := decodeArtPNG(t, "snake/death.png")
	corpsePNG := decodeArtPNG(t, "snake/corpse.png")
	if deathPNG.Bounds().Dx() != death.FrameW*death.Frames || deathPNG.Bounds().Dy() != death.FrameH*5 {
		t.Fatalf("death png %v", deathPNG.Bounds())
	}
	if corpsePNG.Bounds().Dx() != corpse.FrameW || corpsePNG.Bounds().Dy() != corpse.FrameH*5 {
		t.Fatalf("corpse png %v", corpsePNG.Bounds())
	}
	for row := 0; row < 5; row++ {
		first := celBytes(deathPNG, 0, row, death.FrameW, death.FrameH)
		last := celBytes(deathPNG, death.Frames-1, row, death.FrameW, death.FrameH)
		body := celBytes(corpsePNG, 0, row, corpse.FrameW, corpse.FrameH)
		if !bytes.Equal(last, body) {
			t.Fatalf("row %d corpse is not the death sheet's last frame", row)
		}
		if bytes.Equal(first, body) {
			t.Fatalf("row %d corpse is the death sheet's first frame", row)
		}
	}
}

func TestSnakeWaterSheets(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	swim := lib.Get("snake/swim")
	sink := lib.Get("snake/sink")
	if swim == nil || sink == nil {
		t.Fatal("snake swim and sink should be embedded")
	}
	if swim.Frames != 4 || swim.FPS != 8 || !swim.Loop {
		t.Fatalf("swim frames=%d fps=%v loop=%v", swim.Frames, swim.FPS, swim.Loop)
	}
	if sink.Frames != 8 || sink.Loop {
		t.Fatalf("sink frames=%d loop=%v", sink.Frames, sink.Loop)
	}
	if sink.FrameW != 85 || sink.FrameH != 113 || sink.AnchorX != 45 || sink.AnchorY != 108 {
		t.Fatalf("sink cell %dx%d @%d,%d", sink.FrameW, sink.FrameH, sink.AnchorX, sink.AnchorY)
	}
	for _, sh := range []*Sheet{swim, sink} {
		for _, d := range []string{"E", "SE", "S", "N", "NE"} {
			if _, mirror, ok := sh.image(d, 0); !ok || mirror {
				t.Fatalf("%s should be a stored row", d)
			}
		}
		for _, d := range []string{"W", "SW", "NW"} {
			if _, mirror, ok := sh.image(d, sh.Frames-1); !ok || !mirror {
				t.Fatalf("%s should mirror a stored row", d)
			}
		}
	}
	// The sink is scrubbed across SinkTime. The last frame is the one at p=1.
	if FrameScrub(0, sink.Frames) != 0 || FrameScrub(1, sink.Frames) != sink.Frames-1 {
		t.Fatal("sink scrub should span the sheet, ending on the last frame")
	}
	sinkPNG := decodeArtPNG(t, "snake/sink.png")
	for row := 0; row < 5; row++ {
		first := opaqueCount(sinkPNG, 0, row, sink.FrameW, sink.FrameH)
		last := opaqueCount(sinkPNG, sink.Frames-1, row, sink.FrameW, sink.FrameH)
		if last >= first/3 {
			t.Fatalf("row %d last sink frame still has %d opaque pixels, first has %d", row, last, first)
		}
		if last < 8 {
			t.Fatalf("row %d last sink frame is empty (%d)", row, last)
		}
	}
}

func TestEnemySheets(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	if lib.Get("grunt/walk") != nil {
		t.Fatal("grunts have no walk")
	}
	grunt := []struct {
		key    string
		frames int
		fps    float64
		loop   bool
	}{
		{"grunt/idle", 4, 8, true},
		{"grunt/shoot", 2, 16.67, false},
		{"grunt/death", 6, 15, false},
		{"grunt/corpse", 1, 1, false},
		{"grunt/swim", 4, 8, true},
		{"grunt/sink", 8, 4, false},
	}
	for _, tc := range grunt {
		sh := lib.Get(tc.key)
		if sh == nil {
			t.Fatalf("missing %s", tc.key)
		}
		if sh.Frames != tc.frames || sh.FPS != tc.fps || sh.Loop != tc.loop {
			t.Fatalf("%s frames=%d fps=%v loop=%v", tc.key, sh.Frames, sh.FPS, sh.Loop)
		}
		for _, d := range []string{"E", "SE", "S", "N", "NE"} {
			if _, mirror, ok := sh.image(d, 0); !ok || mirror {
				t.Fatalf("%s %s should be stored", tc.key, d)
			}
		}
		for _, d := range []string{"W", "SW", "NW"} {
			if _, mirror, ok := sh.image(d, 0); !ok || !mirror {
				t.Fatalf("%s %s should mirror", tc.key, d)
			}
		}
	}
	if float64(6)/15 != 0.4 {
		t.Fatal("death timing")
	}
	for _, key := range []string{"grenadier/idle", "grenadier/throw", "rocketeer/idle", "rocketeer/throw", "grenadier/death", "rocketeer/sink"} {
		if lib.Get(key) == nil {
			t.Fatalf("missing %s", key)
		}
	}
	throw := lib.Get("grenadier/throw")
	if throw.Frames != 4 || throw.FPS != 16 || throw.Loop {
		t.Fatalf("grenadier throw frames=%d fps=%v loop=%v", throw.Frames, throw.FPS, throw.Loop)
	}
	launch := lib.Get("rocketeer/throw")
	if launch.Frames != 4 || launch.Loop {
		t.Fatalf("rocketeer launch frames=%d loop=%v", launch.Frames, launch.Loop)
	}
	idle := decodeArtPNG(t, "grunt/idle.png")
	if idle.Bounds().Dx() != 104*4 || idle.Bounds().Dy() != 124*5 {
		t.Fatalf("grunt idle png %v", idle.Bounds())
	}
	if opaqueCount(idle, 0, 0, 104, 124) < 200 {
		t.Fatal("grunt east idle is empty")
	}
}

func TestCivilianSheets(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		key    string
		frames int
		fps    float64
		loop   bool
	}{
		{"civilian/idle", 4, 8, true},
		{"civilian/walk", 6, 12, true},
		{"civilian/death", 6, 15, false},
		{"civilian/corpse", 1, 1, false},
	}
	for _, tc := range want {
		sh := lib.Get(tc.key)
		if sh == nil {
			t.Fatalf("missing %s", tc.key)
		}
		if sh.Frames != tc.frames || sh.FPS != tc.fps || sh.Loop != tc.loop {
			t.Fatalf("%s frames=%d fps=%v loop=%v", tc.key, sh.Frames, sh.FPS, sh.Loop)
		}
		for _, d := range []string{"E", "SE", "S", "N", "NE"} {
			if _, mirror, ok := sh.image(d, 0); !ok || mirror {
				t.Fatalf("%s %s should be stored", tc.key, d)
			}
		}
		for _, d := range []string{"W", "SW", "NW"} {
			if _, mirror, ok := sh.image(d, 0); !ok || !mirror {
				t.Fatalf("%s %s should mirror", tc.key, d)
			}
		}
	}
	idle := decodeArtPNG(t, "civilian/idle.png")
	if idle.Bounds().Dx() != 112*4 || idle.Bounds().Dy() != 128*5 {
		t.Fatalf("civilian idle png %v", idle.Bounds())
	}
	if opaqueCount(idle, 0, 0, 112, 128) < 200 {
		t.Fatal("civilian east idle is empty")
	}
	death := decodeArtPNG(t, "civilian/death.png")
	stand := opaqueCount(death, 0, 0, 200, 128)
	down := opaqueCount(death, 5, 0, 200, 128)
	if stand < 200 || down < 200 {
		t.Fatalf("civilian death empty stand=%d down=%d", stand, down)
	}
}

func TestGroundSheets(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"ground/grass", "ground/snow"} {
		sh := lib.Get(key)
		if sh == nil {
			t.Fatalf("missing %s", key)
		}
		if sh.Frames != 4 || sh.FPS != 8 || !sh.Loop || sh.FrameW != 128 || sh.FrameH != 128 {
			t.Fatalf("%s frames=%d fps=%v loop=%v cell=%dx%d", key, sh.Frames, sh.FPS, sh.Loop, sh.FrameW, sh.FrameH)
		}
		img, ok := sh.still(0)
		if !ok || img == nil {
			t.Fatalf("%s has no still frame", key)
		}
	}
	for _, name := range []string{"ground/grass.png", "ground/snow.png"} {
		img := decodeArtPNG(t, name)
		b := img.Bounds()
		if b.Dx() != 128*4 || b.Dy() != 128 {
			t.Fatalf("%s bounds %v", name, b)
		}
		// Opposite edges of one cell should meet. A hard seam is tens of levels.
		seam := edgeDelta(img, 128)
		if seam > 18 {
			t.Fatalf("%s seam %.1f", name, seam)
		}
	}
	moving := []struct {
		key    string
		frames int
		fps    float64
	}{
		{"ground/water-shallow", 8, 8},
		{"ground/water-deep", 8, 8},
		{"ground/quicksand", 6, 4},
		{"ground/ice", 6, 4},
	}
	for _, tc := range moving {
		sh := lib.Get(tc.key)
		if sh == nil {
			t.Fatalf("missing %s", tc.key)
		}
		if sh.Frames != tc.frames || sh.FPS != tc.fps || !sh.Loop || sh.FrameW != 128 || sh.FrameH != 128 {
			t.Fatalf("%s frames=%d fps=%v loop=%v cell=%dx%d", tc.key, sh.Frames, sh.FPS, sh.Loop, sh.FrameW, sh.FrameH)
		}
		img := decodeArtPNG(t, tc.key+".png")
		b := img.Bounds()
		if b.Dx() != 128*tc.frames || b.Dy() != 128 {
			t.Fatalf("%s bounds %v", tc.key, b)
		}
		seam := edgeDelta(img, 128)
		if seam > 18 {
			t.Fatalf("%s seam %.1f", tc.key, seam)
		}
		// A slow sparkle moves few pixels. An unchanged frame moves none.
		if frameDelta(img, 128, 0, 1) < 0.4 {
			t.Fatalf("%s frames do not move", tc.key)
		}
	}
	for _, key := range []string{"ground/cliff", "ground/ramp", "ground/bridge"} {
		sh := lib.Get(key)
		if sh == nil || sh.Frames != 1 || sh.FrameW != 128 || sh.FrameH != 128 {
			t.Fatalf("%s missing or not a still 128 cell", key)
		}
		img := decodeArtPNG(t, key+".png")
		if seam := edgeDelta(img, 128); seam > 18 {
			t.Fatalf("%s seam %.1f", key, seam)
		}
	}
	mine := lib.Get("ground/mine")
	if mine == nil || mine.Frames != 4 || mine.FPS != 2 || !mine.Loop || mine.FrameW != 128 {
		t.Fatalf("mine %+v", mine)
	}
	mineImg := decodeArtPNG(t, "ground/mine.png")
	if mineImg.Bounds().Dx() != 128*4 || mineImg.Bounds().Dy() != 128 {
		t.Fatalf("mine bounds %v", mineImg.Bounds())
	}
	if frameDelta(mineImg, 128, 0, 1) > 0.05 {
		t.Fatal("mine rest frames should match")
	}
	if frameDelta(mineImg, 128, 0, 3) < 0.05 {
		t.Fatal("mine glint frame should differ")
	}
	tree := lib.Get("tree/sway")
	if tree == nil || tree.Frames != 4 || tree.FPS != 4 || !tree.Loop || tree.RowCount() != 3 {
		t.Fatalf("tree sway frames=%v", tree)
	}
	if tree.FrameW != 256 || tree.FrameH != 288 || tree.AnchorX != 128 || tree.AnchorY != 282 {
		t.Fatalf("tree cell %dx%d anchor %d,%d", tree.FrameW, tree.FrameH, tree.AnchorX, tree.AnchorY)
	}
	treeImg := decodeArtPNG(t, "tree/sway.png")
	if treeImg.Bounds().Dx() != 256*4 || treeImg.Bounds().Dy() != 288*3 {
		t.Fatalf("tree sway bounds %v", treeImg.Bounds())
	}
	ab := frameDeltaRow(treeImg, 256, 288, 0, 1)
	ac := frameDeltaRow(treeImg, 256, 288, 0, 2)
	if ab < 8 || ac < 8 {
		t.Fatalf("tree silhouettes too alike ab=%.1f ac=%.1f", ab, ac)
	}
	if sway := frameDelta(cropRow(treeImg, 256, 288, 0), 256, 0, 2); sway < 0.4 {
		t.Fatal("tree sway frames do not move")
	}
}

func TestHutCrateAndSkidooSheets(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"hut/door", "hut/plain"} {
		sh := lib.Get(key)
		if sh == nil || sh.Frames != 4 || sh.FPS != 5 || !sh.Loop || sh.FrameW != 256 || sh.FrameH != 320 {
			t.Fatalf("%s %+v", key, sh)
		}
		if sh.AnchorX != 128 || sh.AnchorY != 312 {
			t.Fatalf("%s anchor %d,%d", key, sh.AnchorX, sh.AnchorY)
		}
		img := decodeArtPNG(t, key+".png")
		if img.Bounds().Dx() != 256*4 || img.Bounds().Dy() != 320 {
			t.Fatalf("%s bounds %v", key, img.Bounds())
		}
		if frameDelta(img, 256, 0, 2) < 0.02 {
			t.Fatalf("%s smoke does not move", key)
		}
	}
	for _, key := range []string{"crate/grenade", "crate/rocket"} {
		sh := lib.Get(key)
		if sh == nil || sh.Frames != 1 || sh.FrameW != 128 || sh.FrameH != 128 || sh.AnchorX != 64 || sh.AnchorY != 64 {
			t.Fatalf("%s %+v", key, sh)
		}
	}
	idle := lib.Get("skidoo/idle")
	move := lib.Get("skidoo/move")
	if idle == nil || move == nil || idle.Frames != 4 || move.Frames != 4 || idle.RowCount() != 5 || move.RowCount() != 5 {
		t.Fatalf("skidoo idle=%v move=%v", idle, move)
	}
	if idle.FrameW != 240 || idle.FrameH != 176 || idle.AnchorX != 120 || idle.AnchorY != 88 {
		t.Fatalf("skidoo cell %dx%d anchor %d,%d", idle.FrameW, idle.FrameH, idle.AnchorX, idle.AnchorY)
	}
	if move.FPS != 8 || !move.Loop {
		t.Fatalf("skidoo move fps=%v loop=%v", move.FPS, move.Loop)
	}
	if _, _, ok := move.image("W", 0); !ok {
		t.Fatal("skidoo west should mirror east")
	}
	moveImg := decodeArtPNG(t, "skidoo/move.png")
	if moveImg.Bounds().Dx() != 240*4 || moveImg.Bounds().Dy() != 176*5 {
		t.Fatalf("skidoo move bounds %v", moveImg.Bounds())
	}
	if frameDelta(cropRow(moveImg, 240, 176, 0), 240, 0, 1) < 0.05 {
		t.Fatal("skidoo track frames do not move")
	}
}

func cropRow(img image.Image, fw, fh, row int) image.Image {
	sub, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		return img
	}
	return sub.SubImage(image.Rect(0, row*fh, fw*4, (row+1)*fh))
}

func frameDeltaRow(img image.Image, fw, fh, rowA, rowB int) float64 {
	var acc, samples float64
	for y := 0; y < fh; y += 4 {
		for x := 0; x < fw; x += 4 {
			r0, g0, b0, a0 := img.At(x, rowA*fh+y).RGBA()
			r1, g1, b1, a1 := img.At(x, rowB*fh+y).RGBA()
			acc += math.Abs(float64(r0)-float64(r1)) + math.Abs(float64(g0)-float64(g1)) + math.Abs(float64(b0)-float64(b1)) + math.Abs(float64(a0)-float64(a1))
			samples += 4
		}
	}
	return acc / samples / 256
}

func frameDelta(img image.Image, n, a, b int) float64 {
	var acc, samples float64
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			r0, g0, b0, _ := img.At(a*n+x, y).RGBA()
			r1, g1, b1, _ := img.At(b*n+x, y).RGBA()
			acc += math.Abs(float64(r0)-float64(r1)) + math.Abs(float64(g0)-float64(g1)) + math.Abs(float64(b0)-float64(b1))
			samples += 3
		}
	}
	return acc / samples / 256
}

func edgeDelta(img image.Image, n int) float64 {
	var acc float64
	samples := 0.0
	add := func(x0, y0, x1, y1 int) {
		r0, g0, b0, _ := img.At(x0, y0).RGBA()
		r1, g1, b1, _ := img.At(x1, y1).RGBA()
		acc += math.Abs(float64(r0)-float64(r1)) + math.Abs(float64(g0)-float64(g1)) + math.Abs(float64(b0)-float64(b1))
		samples += 3
	}
	for y := 0; y < n; y++ {
		add(0, y, n-1, y)
	}
	for x := 0; x < n; x++ {
		add(x, 0, x, n-1)
	}
	return acc / samples / 256
}

func TestEagleAndPantherRecolorSnake(t *testing.T) {
	lib, err := Load(art.Files)
	if err != nil {
		t.Fatal(err)
	}
	poses := []string{"idle", "walk", "shoot", "throw", "death", "corpse", "swim", "sink"}
	for _, pose := range poses {
		snake := lib.Get("snake/" + pose)
		if snake == nil {
			t.Fatalf("missing snake/%s", pose)
		}
		for _, squad := range []string{"eagle", "panther"} {
			sh := lib.Get(squad + "/" + pose)
			if sh == nil {
				t.Fatalf("missing %s/%s", squad, pose)
			}
			if sh.Frames != snake.Frames || sh.FPS != snake.FPS || sh.Loop != snake.Loop ||
				sh.FrameW != snake.FrameW || sh.FrameH != snake.FrameH ||
				sh.AnchorX != snake.AnchorX || sh.AnchorY != snake.AnchorY {
				t.Fatalf("%s/%s does not match the snake cell", squad, pose)
			}
			for _, d := range []string{"E", "SE", "S", "N", "NE"} {
				if _, mirror, ok := sh.image(d, 0); !ok || mirror {
					t.Fatalf("%s/%s %s should be stored", squad, pose, d)
				}
			}
			for _, d := range []string{"W", "SW", "NW"} {
				if _, mirror, ok := sh.image(d, 0); !ok || !mirror {
					t.Fatalf("%s/%s %s should mirror", squad, pose, d)
				}
			}
		}
		base := decodeArtPNG(t, "snake/"+pose+".png")
		eagle := decodeArtPNG(t, "eagle/"+pose+".png")
		panther := decodeArtPNG(t, "panther/"+pose+".png")
		if base.Bounds() != eagle.Bounds() || base.Bounds() != panther.Bounds() {
			t.Fatalf("%s png size differs", pose)
		}
		var same, moved int
		var eB, pR int
		b := base.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				sb := nrgba(base.At(x, y))
				eb := nrgba(eagle.At(x, y))
				pb := nrgba(panther.At(x, y))
				if sb.A != eb.A || sb.A != pb.A {
					t.Fatalf("%s alpha changed at %d,%d", pose, x, y)
				}
				if sb.A == 0 {
					continue
				}
				if !greenCloth(sb) {
					if sb != eb || sb != pb {
						t.Fatalf("%s recolored a non-uniform pixel at %d,%d snake %v eagle %v panther %v", pose, x, y, sb, eb, pb)
					}
					same++
					continue
				}
				if sb == eb || sb == pb {
					t.Fatalf("%s left uniform green at %d,%d", pose, x, y)
				}
				moved++
				if eb.B > eb.R {
					eB++
				}
				if pb.R > pb.G && pb.G > pb.B {
					pR++
				}
			}
		}
		if moved < 100 || eB < moved*8/10 || pR < moved*8/10 {
			t.Fatalf("%s moved=%d blue=%d amber=%d same=%d", pose, moved, eB, pR, same)
		}
	}
}

func nrgba(c color.Color) color.NRGBA {
	return color.NRGBAModel.Convert(c).(color.NRGBA)
}

// greenCloth matches the fatigues and helmet. Skin, boots, and the rifle are warmer or greyer.
func greenCloth(c color.NRGBA) bool {
	rf, gf, bf := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	max := rf
	if gf > max {
		max = gf
	}
	if bf > max {
		max = bf
	}
	min := rf
	if gf < min {
		min = gf
	}
	if bf < min {
		min = bf
	}
	if max < 0.06 || max-min < 0.14*max {
		return false
	}
	d := max - min
	var h float64
	switch max {
	case rf:
		h = (gf - bf) / d
		if h < 0 {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	default:
		h = (rf-gf)/d + 4
	}
	h *= 60
	return h >= 68 && h <= 188
}

func opaqueCount(img image.Image, frame, row, fw, fh int) int {
	n := 0
	x0 := frame * fw
	y0 := row * fh
	for y := 0; y < fh; y++ {
		for x := 0; x < fw; x++ {
			_, _, _, a := img.At(x0+x, y0+y).RGBA()
			if a > 0x2000 {
				n++
			}
		}
	}
	return n
}

func decodeArtPNG(t *testing.T, name string) image.Image {
	t.Helper()
	f, err := art.Files.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func celBytes(img image.Image, frame, row, fw, fh int) []byte {
	var buf []byte
	x0 := frame * fw
	y0 := row * fh
	for y := 0; y < fh; y++ {
		for x := 0; x < fw; x++ {
			r, g, b, a := img.At(x0+x, y0+y).RGBA()
			buf = append(buf, byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8))
		}
	}
	return buf
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

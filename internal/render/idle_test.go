package render

import (
	"image"
	"image/draw"
	"math"
	"testing"
)

func TestLeaderIdleStaysOnThePointer(t *testing.T) {
	var lived bool
	for id := 0; id < 8; id++ {
		for step := 0; step <= 600; step++ {
			c := idleCueAt(id, float64(step)*0.05, true, true)
			if c.headX != 0 || c.tilt != 0 || c.salute {
				t.Fatalf("id %d t %.2f looks away: %+v", id, float64(step)*0.05, c)
			}
			if c.hop > 1 || math.Abs(c.headY) > 2 {
				lived = true
			}
		}
	}
	if !lived {
		t.Fatal("the leader never nodded, looked up, or hopped")
	}
}

func TestIdleLooksAndGags(t *testing.T) {
	var looked, hopped, saluted, sky bool
	for id := 1; id <= 12; id++ {
		for step := 0; step <= 2000; step++ {
			c := idleCueAt(id, float64(step)*0.05, false, true)
			if math.Abs(c.headX) > 4 {
				looked = true
			}
			if c.hop > 2 {
				hopped = true
			}
			if c.salute {
				saluted = true
			}
			if c.headY < -4 {
				sky = true
			}
		}
	}
	if !looked || !hopped || !saluted || !sky {
		t.Fatalf("player idle looked=%v hop=%v salute=%v sky=%v", looked, hopped, saluted, sky)
	}
	var enemyHop, enemyLook, enemySalute bool
	for id := 1; id <= 12; id++ {
		for step := 0; step <= 2000; step++ {
			c := idleCueAt(id, float64(step)*0.05, false, false)
			if c.hop > 2 {
				enemyHop = true
			}
			if math.Abs(c.headX) > 4 {
				enemyLook = true
			}
			if c.salute {
				enemySalute = true
			}
		}
	}
	if !enemyHop || !enemyLook {
		t.Fatal("an enemy idle never looked or hopped")
	}
	if enemySalute {
		t.Fatal("an enemy used the player salute")
	}
}

func TestHeadSplitsOffTheHelmet(t *testing.T) {
	for _, key := range []string{
		"snake/idle", "snake/kneel", "grunt/idle", "grenadier/idle", "rocketeer/idle",
	} {
		sh := activeSheets().Get(key)
		if sh == nil {
			t.Fatalf("missing %s", key)
		}
		for _, dir := range []string{"E", "SE", "S", "N", "NE"} {
			src := sh.celRGBA(dir, 0)
			if src == nil {
				t.Fatalf("%s %s has no cel", key, dir)
			}
			body, head, _, _, ok := splitHead(src)
			// The grenadier's front cel is painted on a grey box, so the
			// helmet is not a separate cap. Leave that one whole.
			if key == "grenadier/idle" && dir == "S" {
				if ok {
					t.Fatal("grenadier south split the backdrop")
				}
				continue
			}
			if !ok {
				t.Fatalf("%s %s did not split", key, dir)
			}
			hy, hx := crownColumn(head)
			if hy < 0 || !clearAt(body, hx, hy) {
				t.Fatalf("%s %s body kept the crown at %d,%d", key, dir, hx, hy)
			}
			if clearAt(head, hx, hy) {
				t.Fatalf("%s %s head lost the crown", key, dir)
			}
			if feetClear(body) {
				t.Fatalf("%s %s body lost its feet", key, dir)
			}
			if !feetClear(head) {
				t.Fatalf("%s %s head took the feet", key, dir)
			}
		}
	}
}

func TestGlanceMovesTheHelmet(t *testing.T) {
	sh := activeSheets().Get("snake/idle")
	src := sh.celRGBA("S", 0)
	if src == nil {
		t.Fatal("south idle")
	}
	restX, restY := helmetCentroid(src)
	look := compositeCue(src, idleCue{headX: 14, tilt: 0.4})
	lx, _ := helmetCentroid(look)
	if math.Abs(lx-restX) < 4 {
		t.Fatalf("helmet moved %.1f px, want a clear look", lx-restX)
	}
	up := compositeCue(src, idleCue{headY: -12})
	_, uy := helmetCentroid(up)
	if uy >= restY-3 {
		t.Fatalf("look up helmet y %.1f, rest %.1f", uy, restY)
	}
}

func compositeCue(src *image.RGBA, cue idleCue) *image.RGBA {
	body, head, px, py, ok := splitHead(src)
	if !ok {
		return src
	}
	out := image.NewRGBA(src.Bounds())
	draw.Draw(out, out.Bounds(), body, image.Point{}, draw.Over)
	blitHead(out, head, px, py, cue)
	return out
}

// blitHead matches drawHead: +tilt tips the helmet toward +x, then the shift.
// Sampling is inverse so a tilt does not skip pixels the way a forward splat does.
func blitHead(dst, head *image.RGBA, px, py int, cue idleCue) {
	theta := -cue.tilt
	c, s := math.Cos(theta), math.Sin(theta)
	// Inverse of rotate(theta) is rotate(-theta).
	ic, is := c, -s
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			qx := float64(x) - float64(px) - cue.headX
			qy := float64(y) - float64(py) - cue.headY
			sx := qx*ic - qy*is + float64(px)
			sy := qx*is + qy*ic + float64(py)
			ix, iy := int(math.Round(sx)), int(math.Round(sy))
			if ix < 0 || iy < 0 || ix >= head.Bounds().Dx() || iy >= head.Bounds().Dy() {
				continue
			}
			i := head.PixOffset(ix, iy)
			if head.Pix[i+3] < 32 {
				continue
			}
			di := dst.PixOffset(x, y)
			copy(dst.Pix[di:di+4], head.Pix[i:i+4])
		}
	}
}

func helmetCentroid(img *image.RGBA) (x, y float64) {
	b := img.Bounds()
	var sx, sy, n int
	limit := b.Min.Y + b.Dy()/3
	for py := b.Min.Y; py < limit; py++ {
		for px := b.Min.X; px < b.Max.X; px++ {
			if img.Pix[img.PixOffset(px, py)+3] < 32 {
				continue
			}
			n++
			sx += px
			sy += py
		}
	}
	if n == 0 {
		return 0, 0
	}
	return float64(sx) / float64(n), float64(sy) / float64(n)
}

func crownColumn(img *image.RGBA) (y, x int) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		var sum, n int
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.Pix[img.PixOffset(x, y)+3] > 32 {
				sum += x
				n++
			}
		}
		if n > 2 && n < b.Dx()/3 {
			return y, sum / n
		}
	}
	return -1, 0
}

func clearAt(img *image.RGBA, x, y int) bool {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return true
	}
	return img.Pix[img.PixOffset(x, y)+3] <= 32
}

func feetClear(img *image.RGBA) bool {
	b := img.Bounds()
	for y := b.Min.Y + b.Dy()*3/4; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.Pix[img.PixOffset(x, y)+3] > 32 {
				return false
			}
		}
	}
	return true
}

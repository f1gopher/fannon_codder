package render

import (
	"bytes"
	"image"
	"image/png"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/assets/art"
)

func TestTextPixels(t *testing.T) {
	if got := TextPixels(3); got != 14 {
		t.Fatalf("size at S=3 is %v, want 14", got)
	}
	if got := TextPixels(6); math.Abs(got-28) > 1e-9 {
		t.Fatalf("size at S=6 is %v, want 28", got)
	}
	want := 14 * (2160.0 / 256.0) / 3
	if got := TextPixels(2160.0 / 256.0); math.Abs(got-want) > 1e-9 {
		t.Fatalf("size at 4K is %v, want %v", got, want)
	}
}

func TestNoticePlaque(t *testing.T) {
	raw, err := art.Files.ReadFile("ui/notice.png")
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != 1280 || b.Dy() != 720 {
		t.Fatalf("plaque %dx%d", b.Dx(), b.Dy())
	}
	if _, _, _, a := img.At(0, 0).RGBA(); a != 0 {
		t.Fatal("corner should be clear so the rim can sit on the field")
	}
	cr, cg, cb, ca := img.At(b.Dx()/2, b.Dy()/2).RGBA()
	if ca>>8 != 255 || cg>>8 < 60 || cg>>8 > 140 || cr>>8 > cg>>8+30 {
		t.Fatalf("field %d %d %d a=%d", cr>>8, cg>>8, cb>>8, ca>>8)
	}
}

func TestNoticeLineAt(t *testing.T) {
	SetPictureScale(2)
	lines := []string{"PAUSED", "Restart level", "Quit"}
	top, ok := NoticeLineAt(320, 256, 160, 0)
	if ok {
		t.Fatalf("top of the frame hit line %d", top)
	}
	var headline, body bool
	for y := 0.0; y < 256; y++ {
		i, hit := NoticeLineAt(320, 256, 160, y, lines...)
		if !hit {
			continue
		}
		if i == 0 {
			headline = true
		}
		if i == 1 {
			body = true
		}
	}
	if !headline || !body {
		t.Fatalf("headline %v body %v", headline, body)
	}
}

func TestNotice(t *testing.T) {
	SetPictureScale(3)
	dst := ebiten.NewImage(320*3, 256*3)
	Notice(dst, "PHASE FAILED", "Click or Enter")
	Notice(dst, "PAUSED")
	Notice(dst, "QUIT THE GAME?", "Y or Enter    quit", "N or Escape    stay")
	Notice(nil, "PAUSED")
	Notice(dst)
	Notice(dst, "")
	SetSheets(&Library{})
	Notice(dst, "PHASE FAILED", "Click or Enter")
	SetSheets(nil)
}

func TestTextCentered(t *testing.T) {
	SetPictureScale(1)
	dst := ebiten.NewImage(320, 256)
	TextCentered(dst, "PHASE FAILED\nClick or Enter")
	TextCentered(nil, "PHASE FAILED")
	TextCentered(dst, "")
	if dst.Bounds() != image.Rect(0, 0, 320, 256) {
		t.Fatalf("bounds %v", dst.Bounds())
	}
}

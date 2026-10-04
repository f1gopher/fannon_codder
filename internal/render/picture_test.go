package render

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
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

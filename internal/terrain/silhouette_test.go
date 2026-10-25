package terrain

import (
	"image"
	"image/color"
	"testing"
)

func TestPropSilhouetteKeepsTheShape(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 48, 64))
	// A tall post on the left and a wide foot. An oval would not have this notch.
	for y := 8; y < 56; y++ {
		for x := 8; x < 16; x++ {
			src.SetRGBA(x, y, color.RGBA{A: 255})
		}
	}
	for y := 48; y < 56; y++ {
		for x := 8; x < 40; x++ {
			src.SetRGBA(x, y, color.RGBA{A: 255})
		}
	}
	sil := PropSilhouette(src)
	if sil == nil {
		t.Fatal("no silhouette")
	}
	post := sil.RGBAAt(8+SilhouettePad+4, 20+SilhouettePad)
	foot := sil.RGBAAt(30+SilhouettePad, 52+SilhouettePad)
	notch := sil.RGBAAt(34+SilhouettePad, 20+SilhouettePad)
	if post.A < 40 || foot.A < 40 {
		t.Fatalf("shape faded post=%d foot=%d", post.A, foot.A)
	}
	if notch.A > 20 {
		t.Fatalf("empty notch filled in, alpha=%d", notch.A)
	}
}

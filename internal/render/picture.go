package render

import (
	"bytes"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

// pictureScale is offscreen pixels per world pixel. Game.Draw sets it
// before any scene paints. One means a test drew without a window.
var pictureScale = 1.0

var uiFont *text.GoTextFaceSource

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	uiFont = src
}

// SetPictureScale records how many offscreen pixels one world pixel occupies.
func SetPictureScale(s float64) {
	if s <= 0 {
		s = 1
	}
	pictureScale = s
}

// PictureScale is the scale SetPictureScale last recorded.
func PictureScale() float64 { return pictureScale }

// TextPixels is the HUD face size. It is 14 pixels when S is 3, and it
// scales with S.
func TextPixels(s float64) float64 {
	if s <= 0 {
		s = 1
	}
	return 14 * s / 3
}

// Text draws str with the Go regular face. x and y are world-frame pixels;
// the glyph top-left lands at that point times the picture scale.
func Text(dst *ebiten.Image, str string, x, y float64) {
	if dst == nil || str == "" {
		return
	}
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	size := TextPixels(s)
	face := &text.GoTextFace{Source: uiFont, Size: size}
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignStart
	op.SecondaryAlign = text.AlignStart
	op.LineSpacing = size * 1.25
	op.GeoM.Translate(x*s, y*s)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(dst, str, face, op)
}

// Rect fills a rectangle given in world-frame pixels.
func Rect(dst *ebiten.Image, x, y, w, h int, c color.Color) {
	fillRect(dst, x, y, w, h, c)
}

// blit draws src, whose pixels are world-frame pixels, at x,y in that frame.
// sx and sy are an extra scale (a sinking trooper) applied before the picture scale.
func blit(dst, src *ebiten.Image, x, y, sx, sy float64) {
	if dst == nil || src == nil || sx == 0 || sy == 0 {
		return
	}
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(sx*s, sy*s)
	op.GeoM.Translate(x*s, y*s)
	dst.DrawImage(src, op)
}

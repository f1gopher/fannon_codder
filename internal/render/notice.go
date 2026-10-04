package render

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Notice ink matches the title wordmark: cream fill, dark olive edge,
// sitting on the painted olive plaque.
var (
	noticeInk  = color.RGBA{R: 0xf3, G: 0xe7, B: 0xc4, A: 0xff}
	noticeEdge = color.RGBA{R: 0x2a, G: 0x24, B: 0x10, A: 0xff}
	noticeFill = color.RGBA{R: 0x3c, G: 0x40, B: 0x22, A: 0xff}
)

// Notice draws lines on a painted plaque in the middle of dst.
// The first line is the headline. A missing plaque falls back to a flat fill.
func Notice(dst *ebiten.Image, lines ...string) {
	if dst == nil || len(lines) == 0 {
		return
	}
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	head := TextPixels(s) * 1.6
	body := TextPixels(s)
	gap := head * 0.35
	var blockW, blockH float64
	for i, line := range lines {
		if line == "" {
			continue
		}
		size := body
		if i == 0 {
			size = head
		}
		face := &text.GoTextFace{Source: uiFont, Size: size}
		w, h := text.Measure(line, face, size)
		if w > blockW {
			blockW = w
		}
		if blockH > 0 {
			blockH += gap
		}
		blockH += h
	}
	if blockW == 0 {
		return
	}
	// Measure is in device pixels. The plaque is laid out in world pixels.
	// The rim is about 15 world pixels, so the pad sits the letters in the field.
	padX, padY := 46.0, 32.0
	pw := blockW/s + padX*2
	ph := blockH/s + padY*2
	b := dst.Bounds()
	x := (float64(b.Dx())/s - pw) / 2
	y := (float64(b.Dy())/s - ph) / 2
	if !drawNoticePlaque(dst, x, y, pw, ph) {
		Rect(dst, int(x), int(y), int(pw), int(ph), noticeFill)
	}
	cy := y + padY
	for i, line := range lines {
		if line == "" {
			continue
		}
		size := body
		if i == 0 {
			size = head
		}
		face := &text.GoTextFace{Source: uiFont, Size: size}
		_, h := text.Measure(line, face, size)
		drawNoticeLine(dst, line, face, x+pw/2, cy+h/s/2)
		cy += h/s + gap/s
	}
}

func drawNoticeLine(dst *ebiten.Image, str string, face *text.GoTextFace, x, y float64) {
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	// A filled disc of the glyph, one device pixel at a time, so the olive
	// edge is solid. Sparse world-pixel copies left gaps between the letters.
	rad := face.Size * 0.16
	if rad < 1 {
		rad = 1
	}
	n := int(math.Ceil(rad))
	rr := rad * rad
	for oy := -n; oy <= n; oy++ {
		for ox := -n; ox <= n; ox++ {
			if ox == 0 && oy == 0 {
				continue
			}
			if float64(ox*ox+oy*oy) > rr {
				continue
			}
			drawFace(dst, str, face, x+float64(ox)/s, y+float64(oy)/s, noticeEdge)
		}
	}
	drawFace(dst, str, face, x, y, noticeInk)
}

func drawFace(dst *ebiten.Image, str string, face *text.GoTextFace, x, y float64, c color.Color) {
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	op.GeoM.Translate(x*s, y*s)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, str, face, op)
}

// drawNoticePlaque stretches ui/notice as a 9-slice. The corner holds the
// rounded rim; the middle is the olive field.
func drawNoticePlaque(dst *ebiten.Image, x, y, w, h float64) bool {
	img, _, _, ok := uiCel("ui/notice")
	if !ok {
		return false
	}
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	c := sh / 6
	if c < 1 || c*2 >= sw || c*2 >= sh {
		return false
	}
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	sc := s / ArtScale
	cw := float64(c) * sc
	dw, dh := w*s, h*s
	dx, dy := x*s, y*s
	if dw < cw*2 {
		cw = dw / 2
	}
	ch := float64(c) * sc
	if dh < ch*2 {
		ch = dh / 2
	}
	midW := dw - cw*2
	midH := dh - ch*2
	xs := []float64{0, float64(c), float64(sw - c)}
	ys := []float64{0, float64(c), float64(sh - c)}
	ws := []float64{float64(c), float64(sw - 2*c), float64(c)}
	hs := []float64{float64(c), float64(sh - 2*c), float64(c)}
	dxs := []float64{dx, dx + cw, dx + cw + midW}
	dys := []float64{dy, dy + ch, dy + ch + midH}
	dws := []float64{cw, midW, cw}
	dhs := []float64{ch, midH, ch}
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			if dws[col] <= 0 || dhs[row] <= 0 {
				continue
			}
			sr := image.Rect(int(xs[col]), int(ys[row]), int(xs[col]+ws[col]), int(ys[row]+hs[row]))
			sub, ok := img.SubImage(sr).(*ebiten.Image)
			if !ok || sub == nil {
				return false
			}
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterLinear
			op.GeoM.Scale(dws[col]/ws[col], dhs[row]/hs[row])
			op.GeoM.Translate(dxs[col], dys[row])
			dst.DrawImage(sub, op)
		}
	}
	return true
}

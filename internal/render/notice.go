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

// noticePad is the olive field inside the cream rim, in world pixels.
const (
	noticePadX = 46.0
	noticePadY = 32.0
)

// noticeBox is one drawn line in world pixels. Index is the lines-slice index.
// The box spans the plaque interior, including the gap under the letters.
type noticeBox struct {
	index      int
	x, y, w, h float64
}

// noticeBoxes lays out the same lines Notice draws. frameW and frameH are the
// destination in world pixels (image bounds divided by the picture scale).
func noticeBoxes(frameW, frameH float64, lines []string) []noticeBox {
	s := pictureScale
	if s <= 0 {
		s = 1
	}
	head := TextPixels(s) * 1.6
	body := TextPixels(s)
	gap := head * 0.35
	var blockW, blockH float64
	type sized struct {
		i    int
		h    float64
		size float64
	}
	var drawn []sized
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
		drawn = append(drawn, sized{i: i, h: h, size: size})
	}
	if blockW == 0 || len(drawn) == 0 {
		return nil
	}
	pw := blockW/s + noticePadX*2
	ph := blockH/s + noticePadY*2
	x := (frameW - pw) / 2
	y := (frameH - ph) / 2
	cy := y + noticePadY
	boxes := make([]noticeBox, len(drawn))
	for n, d := range drawn {
		hh := d.h / s
		bh := hh
		if n < len(drawn)-1 {
			bh += gap / s
		}
		boxes[n] = noticeBox{index: d.i, x: x, y: cy, w: pw, h: bh}
		cy += hh + gap/s
	}
	return boxes
}

// NoticeLineAt reports which drawn line contains the world point x, y.
// The index matches the lines slice. The hit box is the full width of the plaque.
func NoticeLineAt(frameW, frameH, x, y float64, lines ...string) (int, bool) {
	for _, b := range noticeBoxes(frameW, frameH, lines) {
		if x >= b.x && y >= b.y && x < b.x+b.w && y < b.y+b.h {
			return b.index, true
		}
	}
	return 0, false
}

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
	b := dst.Bounds()
	boxes := noticeBoxes(float64(b.Dx())/s, float64(b.Dy())/s, lines)
	if len(boxes) == 0 {
		return
	}
	plaque := boxes[0]
	// The first box starts at the pad, so the plaque origin is above it.
	px := plaque.x
	py := plaque.y - noticePadY
	pw := plaque.w
	last := boxes[len(boxes)-1]
	ph := last.y + last.h + noticePadY - py
	if !drawNoticePlaque(dst, px, py, pw, ph) {
		Rect(dst, int(px), int(py), int(pw), int(ph), noticeFill)
	}
	head := TextPixels(s) * 1.6
	body := TextPixels(s)
	for _, box := range boxes {
		line := lines[box.index]
		size := body
		if box.index == 0 {
			size = head
		}
		face := &text.GoTextFace{Source: uiFont, Size: size}
		_, th := text.Measure(line, face, size)
		drawNoticeLine(dst, line, face, box.x+box.w/2, box.y+th/s/2)
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

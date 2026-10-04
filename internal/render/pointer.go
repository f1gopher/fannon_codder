package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	PointerArrow = iota
	PointerCrosshair
	PointerBoard
	PointerExit
)

var (
	// Gold reads on snow. Pure yellow is almost as bright as the ground.
	pointerInk   = color.RGBA{R: 0xff, G: 0xc4, B: 0x00, A: 0xff}
	pointerAim   = color.RGBA{R: 0xff, G: 0x20, B: 0x20, A: 0xff}
	pointerBlack = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
)

// Pointer draws the cursor at logical screen pixels.
// The arrow, the crosshair, and the board cursor are painted at master
// scale and drawn at S/8. The exit mark stays the small bitmap.
func Pointer(dst *ebiten.Image, x, y float64, kind int) {
	ix, iy := int(x), int(y)
	switch kind {
	case PointerCrosshair:
		if !drawUIAt(dst, "ui/crosshair", x, y) {
			drawCrosshair(dst, ix, iy)
		}
	case PointerBoard:
		if !drawUIAt(dst, "ui/board", x, y) {
			drawBoard(dst, ix, iy)
		}
	case PointerExit:
		drawExit(dst, ix, iy)
	default:
		if !drawUIAt(dst, "ui/pointer", x, y) {
			drawArrow(dst, ix, iy)
		}
	}
}

func drawUIAt(dst *ebiten.Image, key string, x, y float64) bool {
	img, ax, ay, ok := uiCel(key)
	if !ok {
		return false
	}
	if c, ok := cursorTint(key); ok {
		DrawSpriteTint(dst, img, ax, ay, x, y, c)
	} else {
		DrawSprite(dst, img, ax, ay, false, x, y)
	}
	return true
}

func cursorTint(key string) (color.Color, bool) {
	switch key {
	case "ui/crosshair":
		return pointerAim, true
	case "ui/pointer", "ui/board":
		return pointerInk, true
	default:
		return nil, false
	}
}

func uiCel(key string) (img *ebiten.Image, ax, ay int, ok bool) {
	lib := activeSheets()
	if lib == nil {
		return nil, 0, 0, false
	}
	sh := lib.Get(key)
	if sh == nil {
		return nil, 0, 0, false
	}
	img, ok = sh.still(0)
	if !ok {
		return nil, 0, 0, false
	}
	return img, sh.AnchorX, sh.AnchorY, true
}

func drawBoard(dst *ebiten.Image, x, y int) {
	// Hollow box: climb in.
	rows := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 1, 0, 0, 1},
		{1, 0, 0, 1, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	drawBitmap(dst, x-3, y-3, rows, pointerInk)
}

func drawExit(dst *ebiten.Image, x, y int) {
	// Arrow leaving a box.
	rows := [][]int{
		{0, 0, 0, 1, 0, 0, 0},
		{0, 0, 1, 1, 1, 0, 0},
		{0, 1, 0, 1, 0, 1, 0},
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}
	drawBitmap(dst, x-3, y-5, rows, pointerInk)
}

func drawArrow(dst *ebiten.Image, x, y int) {
	// 7×10 tip at (x,y), gold with a 1px black outline.
	rows := [][]int{
		{1, 0, 0, 0, 0, 0, 0},
		{1, 1, 0, 0, 0, 0, 0},
		{1, 1, 1, 0, 0, 0, 0},
		{1, 1, 1, 1, 0, 0, 0},
		{1, 1, 1, 1, 1, 0, 0},
		{1, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 0, 0, 0},
		{1, 0, 1, 1, 0, 0, 0},
		{0, 0, 0, 1, 1, 0, 0},
	}
	drawBitmap(dst, x, y, rows, pointerInk)
}

func drawCrosshair(dst *ebiten.Image, x, y int) {
	// 9×9 plus, hotspot at the centre.
	rows := [][]int{
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{1, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
	}
	drawBitmap(dst, x-4, y-4, rows, pointerAim)
}

func drawBitmap(dst *ebiten.Image, ox, oy int, rows [][]int, ink color.Color) {
	h := len(rows)
	for j, row := range rows {
		for i, v := range row {
			if v == 0 {
				continue
			}
			px, py := ox+i, oy+j
			// Outline
			for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				nx, ny := px+d[0], py+d[1]
				if !filled(rows, i+d[0], j+d[1], h) {
					fillRect(dst, nx, ny, 1, 1, pointerBlack)
				}
			}
			fillRect(dst, px, py, 1, 1, ink)
		}
	}
}

func filled(rows [][]int, x, y, h int) bool {
	if y < 0 || y >= h || x < 0 || x >= len(rows[y]) {
		return false
	}
	return rows[y][x] != 0
}

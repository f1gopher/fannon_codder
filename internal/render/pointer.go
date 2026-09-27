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
	pointerWhite = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	pointerBlack = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
)

// Pointer draws the Amiga-style cursor at logical screen pixels.
// Arrow is the destination pointer; crosshair is the weapon cursor.
func Pointer(dst *ebiten.Image, x, y float64, kind int) {
	ix, iy := int(x), int(y)
	switch kind {
	case PointerCrosshair:
		drawCrosshair(dst, ix, iy)
	case PointerBoard:
		drawBoard(dst, ix, iy)
	case PointerExit:
		drawExit(dst, ix, iy)
	default:
		drawArrow(dst, ix, iy)
	}
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
	drawBitmap(dst, x-3, y-3, rows)
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
	drawBitmap(dst, x-3, y-5, rows)
}

func drawArrow(dst *ebiten.Image, x, y int) {
	// 7×10 tip at (x,y), white with a 1px black outline.
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
	drawBitmap(dst, x, y, rows)
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
	drawBitmap(dst, x-4, y-4, rows)
}

func drawBitmap(dst *ebiten.Image, ox, oy int, rows [][]int) {
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
					dst.Set(nx, ny, pointerBlack)
				}
			}
			dst.Set(px, py, pointerWhite)
		}
	}
}

func filled(rows [][]int, x, y, h int) bool {
	if y < 0 || y >= h || x < 0 || x >= len(rows[y]) {
		return false
	}
	return rows[y][x] != 0
}

package mapedit

// RailWidth is the phase list in offscreen pixels.
func RailWidth(offW int) int {
	w := offW / 4
	if w < 220 {
		w = 220
	}
	if w > 420 {
		w = 420
	}
	if offW > 1 && w >= offW {
		w = offW - 1
	}
	return w
}

// RowHeight is the list row in offscreen pixels. The face is the game HUD size.
func RowHeight(scale float64) float64 {
	if scale <= 0 {
		scale = 1
	}
	return 14*scale/3 + 8
}

// RailHit reports the phase row under an offscreen point. A point in the
// viewport, or past the last row, is not a row.
func RailHit(x, y float64, rail int, rowH float64, n int) (int, bool) {
	if n <= 0 || rowH <= 0 || rail <= 0 || x < 0 || y < 0 || x >= float64(rail) {
		return 0, false
	}
	row := int(y / rowH)
	if row < 0 || row >= n {
		return 0, false
	}
	return row, true
}

// ClampScroll keeps the view inside the map. A map smaller than the view
// stays pinned at 0,0. scroll is world pixels.
func ClampScroll(scrollX, scrollY, mapW, mapH, viewW, viewH float64) (float64, float64) {
	maxX := mapW - viewW
	if maxX < 0 {
		maxX = 0
	}
	maxY := mapH - viewH
	if maxY < 0 {
		maxY = 0
	}
	if scrollX < 0 {
		scrollX = 0
	} else if scrollX > maxX {
		scrollX = maxX
	}
	if scrollY < 0 {
		scrollY = 0
	} else if scrollY > maxY {
		scrollY = maxY
	}
	return scrollX, scrollY
}

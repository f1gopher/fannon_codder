package mapedit

import (
	"fmt"
	"strconv"
	"strings"

	"fannon-codder/internal/data"
	"fannon-codder/internal/sim"
)

// A phase map covers the playfield. 24×16 tiles is 384×256 world pixels.
const (
	minMapTilesW = 24
	minMapTilesH = 16
	maxMapTiles  = 160
)

func (e *Editor) sizeDraft() string {
	if e.grid.W <= 0 || e.grid.H <= 0 {
		return ""
	}
	if e.freeform {
		return fmt.Sprintf("%d × %d", e.grid.W*sim.TileSize, e.grid.H*sim.TileSize)
	}
	return fmt.Sprintf("%d × %d", e.grid.W, e.grid.H)
}

// applySize reads a width and a height. A region map takes world pixels,
// a multiple of the tile. A tile map takes a tile count. The old ground
// stays in the top-left. Placements that no longer fit are removed, and
// the player start is pulled back onto the map.
func (e *Editor) applySize(text string) bool {
	if e.phase == nil || e.grid.W <= 0 || e.grid.H <= 0 {
		e.typeNote = "No map"
		return false
	}
	w, h, ok := parseSize(text)
	if !ok {
		e.typeNote = "Need a width and a height"
		return false
	}
	if e.freeform {
		if w%sim.TileSize != 0 || h%sim.TileSize != 0 {
			e.typeNote = "Pixels must be a multiple of 16"
			return false
		}
		w /= sim.TileSize
		h /= sim.TileSize
	}
	if w < minMapTilesW || h < minMapTilesH {
		e.typeNote = "At least 24 × 16 tiles"
		return false
	}
	if w > maxMapTiles || h > maxMapTiles {
		e.typeNote = "At most 160 × 160 tiles"
		return false
	}
	if w == e.grid.W && h == e.grid.H {
		return true
	}
	if e.freeform {
		oldW, oldH := e.phase.Map.W, e.phase.Map.H
		e.phase.Map.W = w
		e.phase.Map.H = h
		g, err := e.phase.Grid()
		if err != nil {
			e.phase.Map.W, e.phase.Map.H = oldW, oldH
			e.typeNote = err.Error()
			return false
		}
		e.grid = g
		e.chart.W = w * sim.TileSize
		e.chart.H = h * sim.TileSize
	} else {
		next := sim.Map{W: w, H: h, Tiles: make([]sim.Tile, w*h)}
		cw, ch := e.grid.W, e.grid.H
		if cw > w {
			cw = w
		}
		if ch > h {
			ch = h
		}
		for y := 0; y < ch; y++ {
			for x := 0; x < cw; x++ {
				next.Tiles[y*w+x] = e.grid.Tiles[y*e.grid.W+x]
			}
		}
		e.grid = next
		e.phase.Map.W = w
		e.phase.Map.H = h
		e.phase.Map.Tiles = tileRows(next)
	}
	e.clipToMap()
	e.sel = pick{}
	e.clampScroll()
	e.dirty = true
	e.typeNote = ""
	return true
}

func parseSize(text string) (int, int, bool) {
	s := strings.ToLower(text)
	s = strings.ReplaceAll(s, "tiles", " ")
	s = strings.ReplaceAll(s, "×", " ")
	s = strings.ReplaceAll(s, "x", " ")
	s = strings.ReplaceAll(s, ",", " ")
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return 0, 0, false
	}
	w, errW := strconv.Atoi(fields[0])
	h, errH := strconv.Atoi(fields[1])
	if errW != nil || errH != nil || w < 1 || h < 1 {
		return 0, 0, false
	}
	return w, h, true
}

func tileRows(g sim.Map) []string {
	rows := make([]string, g.H)
	for y := 0; y < g.H; y++ {
		var b strings.Builder
		for x := 0; x < g.W; x++ {
			b.WriteRune(data.TileRune(g.Tiles[y*g.W+x]))
		}
		rows[y] = b.String()
	}
	return rows
}

func onMap(x, y, w, h int) bool {
	return x >= 0 && y >= 0 && x < w && y < h
}

func (e *Editor) clipToMap() {
	w, h := e.grid.W, e.grid.H
	if e.hasStart && !onMap(e.start[0], e.start[1], w, h) {
		e.start[0], e.start[1] = e.clampTile(e.start[0], e.start[1])
	}
	e.enemies = clipEnemies(e.enemies, w, h)
	e.civilians = clipEnemies(e.civilians, w, h)
	e.hostages = clipEnemies(e.hostages, w, h)
	e.buildings = clipBuildings(e.buildings, w, h)
	e.pickups = clipPickups(e.pickups, w, h)
	e.barrels = clipBarrels(e.barrels, w, h)
	e.vehicles = clipVehicles(e.vehicles, w, h)
}

func clipEnemies(in []data.Enemy, w, h int) []data.Enemy {
	out := in[:0]
	for _, en := range in {
		if !onMap(en.X, en.Y, w, h) {
			continue
		}
		en.Line = clipPoints(en.Line, w, h)
		out = append(out, en)
	}
	return out
}

func clipPoints(in [][2]int, w, h int) [][2]int {
	if len(in) == 0 {
		return nil
	}
	out := in[:0]
	for _, p := range in {
		if onMap(p[0], p[1], w, h) {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func clipBuildings(in []data.BuildingSpec, w, h int) []data.BuildingSpec {
	out := in[:0]
	for _, b := range in {
		bw, bh := b.W, b.H
		if bw < 1 {
			bw = 1
		}
		if bh < 1 {
			bh = 1
		}
		if b.X < 0 || b.Y < 0 || b.X+bw > w || b.Y+bh > h {
			continue
		}
		out = append(out, b)
	}
	return out
}

func clipPickups(in []data.PickupSpec, w, h int) []data.PickupSpec {
	out := in[:0]
	for _, p := range in {
		if onMap(p.X, p.Y, w, h) {
			out = append(out, p)
		}
	}
	return out
}

func clipBarrels(in []data.BarrelSpec, w, h int) []data.BarrelSpec {
	out := in[:0]
	for _, b := range in {
		if onMap(b.X, b.Y, w, h) {
			out = append(out, b)
		}
	}
	return out
}

func clipVehicles(in []data.VehicleSpec, w, h int) []data.VehicleSpec {
	out := in[:0]
	for _, v := range in {
		if !onMap(v.X, v.Y, w, h) {
			continue
		}
		v.Line = clipPoints(v.Line, w, h)
		out = append(out, v)
	}
	return out
}

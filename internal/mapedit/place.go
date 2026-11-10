package mapedit

import (
	"math"

	"fannon-codder/internal/data"
	"fannon-codder/internal/sim"
)

// A pick is one draggable thing on the phase.
type pickKind int

const (
	pickNone pickKind = iota
	pickStart
	pickEnemy
	pickLine
	pickBarrel
	pickHut
	pickCrate
	pickCivilian
)

type pick struct {
	kind pickKind
	i, j int
}

func (p pick) ok() bool { return p.kind != pickNone }

const pickRadius = 10 // world pixels

// stamp is what a left click on empty ground adds.
const (
	stampGrunt     = "grunt"
	stampGrenadier = "grenadier"
	stampRocketeer = "rocketeer"
	stampPatrol    = "patrol"
	stampBarrel    = "barrel"
	stampHut       = "hut"
	stampDoor      = "door"
	stampGrenades  = "grenades"
	stampRockets   = "rockets"
	stampCivilian  = "civilian"
)

func stampName(s string) string {
	switch s {
	case stampGrenadier:
		return "grenadier"
	case stampRocketeer:
		return "rocketeer"
	case stampPatrol:
		return "patrol"
	case stampBarrel:
		return "barrel"
	case stampHut:
		return "hut"
	case stampDoor:
		return "door hut"
	case stampGrenades:
		return "grenade crate"
	case stampRockets:
		return "rocket crate"
	case stampCivilian:
		return "civilian"
	default:
		return "grunt"
	}
}

func (e *Editor) help() string {
	return "left-click places    drag moves    Del removes    W patrol point    Shift+1 civilian    S saves"
}

func (e *Editor) tileAt(wx, wy float64) (int, int) {
	tx := int(math.Floor(wx / sim.TileSize))
	ty := int(math.Floor(wy / sim.TileSize))
	return e.clampTile(tx, ty)
}

func (e *Editor) clampTile(tx, ty int) (int, int) {
	if e.grid.W < 1 || e.grid.H < 1 {
		return 0, 0
	}
	if tx < 0 {
		tx = 0
	} else if tx >= e.grid.W {
		tx = e.grid.W - 1
	}
	if ty < 0 {
		ty = 0
	} else if ty >= e.grid.H {
		ty = e.grid.H - 1
	}
	return tx, ty
}

func (e *Editor) pickAt(wx, wy float64) pick {
	best := pick{}
	bestD := float64(pickRadius) + 1
	take := func(px, py float64, k pick) {
		d := math.Hypot(wx-px, wy-py)
		if d <= pickRadius && d < bestD {
			bestD = d
			best = k
		}
	}
	if e.hasStart {
		c := sim.TileCenter(e.start[0], e.start[1])
		take(c.X, c.Y, pick{kind: pickStart})
	}
	for i, en := range e.enemies {
		c := sim.TileCenter(en.X, en.Y)
		take(c.X, c.Y, pick{kind: pickEnemy, i: i})
		for j, pt := range en.Line {
			c = sim.TileCenter(pt[0], pt[1])
			take(c.X, c.Y, pick{kind: pickLine, i: i, j: j})
		}
	}
	for i, c := range e.civilians {
		p := sim.TileCenter(c.X, c.Y)
		take(p.X, p.Y, pick{kind: pickCivilian, i: i})
	}
	for i, b := range e.barrels {
		c := sim.TileCenter(b.X, b.Y)
		take(c.X, c.Y, pick{kind: pickBarrel, i: i})
	}
	for i, pk := range e.pickups {
		c := sim.TileCenter(pk.X, pk.Y)
		take(c.X, c.Y, pick{kind: pickCrate, i: i})
	}
	if best.ok() {
		return best
	}
	for i, b := range e.buildings {
		tw, th := hutSize(b)
		x0 := float64(b.X * sim.TileSize)
		y0 := float64(b.Y * sim.TileSize)
		x1 := float64((b.X + tw) * sim.TileSize)
		y1 := float64((b.Y + th) * sim.TileSize)
		if wx >= x0 && wx < x1 && wy >= y0 && wy < y1 {
			return pick{kind: pickHut, i: i}
		}
	}
	return pick{}
}

func hutSize(b data.BuildingSpec) (int, int) {
	tw, th := b.W, b.H
	if tw < 1 {
		tw = 2
	}
	if th < 1 {
		th = 2
	}
	return tw, th
}

func (e *Editor) beginGrab(p pick, wx, wy float64) {
	e.sel = p
	e.grabbing = true
	e.grab = p
	e.grabTX, e.grabTY = e.tileAt(wx, wy)
	switch p.kind {
	case pickStart:
		e.homeStart = e.start
	case pickEnemy:
		e.homeEnemy = e.enemies[p.i]
		e.homeEnemy.Line = append([][2]int(nil), e.enemies[p.i].Line...)
	case pickLine:
		e.homeLine = e.enemies[p.i].Line[p.j]
	case pickBarrel:
		e.homeBarrel = e.barrels[p.i]
	case pickHut:
		e.homeHut = e.buildings[p.i]
	case pickCrate:
		e.homeCrate = e.pickups[p.i]
	case pickCivilian:
		e.homeCivilian = e.civilians[p.i]
	}
}

func (e *Editor) dragGrab(wx, wy float64) {
	if !e.grabbing || e.err != nil {
		return
	}
	tx, ty := e.tileAt(wx, wy)
	dx, dy := tx-e.grabTX, ty-e.grabTY
	switch e.grab.kind {
	case pickStart:
		nx, ny := e.clampTile(e.homeStart[0]+dx, e.homeStart[1]+dy)
		if e.start != [2]int{nx, ny} {
			e.start = [2]int{nx, ny}
			e.dirty = true
			e.status = ""
		}
	case pickEnemy:
		i := e.grab.i
		nx, ny := e.clampTile(e.homeEnemy.X+dx, e.homeEnemy.Y+dy)
		if e.enemies[i].X != nx || e.enemies[i].Y != ny {
			e.enemies[i].X, e.enemies[i].Y = nx, ny
			e.dirty = true
			e.status = ""
		}
	case pickLine:
		i, j := e.grab.i, e.grab.j
		nx, ny := e.clampTile(e.homeLine[0]+dx, e.homeLine[1]+dy)
		if e.enemies[i].Line[j] != [2]int{nx, ny} {
			e.enemies[i].Line[j] = [2]int{nx, ny}
			e.dirty = true
			e.status = ""
		}
	case pickBarrel:
		i := e.grab.i
		nx, ny := e.clampTile(e.homeBarrel.X+dx, e.homeBarrel.Y+dy)
		if e.barrels[i].X != nx || e.barrels[i].Y != ny {
			e.barrels[i].X, e.barrels[i].Y = nx, ny
			e.dirty = true
			e.status = ""
		}
	case pickHut:
		i := e.grab.i
		tw, th := hutSize(e.homeHut)
		nx, ny := e.homeHut.X+dx, e.homeHut.Y+dy
		if nx < 0 {
			nx = 0
		}
		if ny < 0 {
			ny = 0
		}
		if e.grid.W > 0 && nx > e.grid.W-tw {
			nx = e.grid.W - tw
		}
		if e.grid.H > 0 && ny > e.grid.H-th {
			ny = e.grid.H - th
		}
		if e.buildings[i].X != nx || e.buildings[i].Y != ny {
			e.buildings[i].X, e.buildings[i].Y = nx, ny
			e.dirty = true
			e.status = ""
		}
	case pickCrate:
		i := e.grab.i
		nx, ny := e.clampTile(e.homeCrate.X+dx, e.homeCrate.Y+dy)
		if e.pickups[i].X != nx || e.pickups[i].Y != ny {
			e.pickups[i].X, e.pickups[i].Y = nx, ny
			e.dirty = true
			e.status = ""
		}
	case pickCivilian:
		i := e.grab.i
		nx, ny := e.clampTile(e.homeCivilian.X+dx, e.homeCivilian.Y+dy)
		if e.civilians[i].X != nx || e.civilians[i].Y != ny {
			e.civilians[i].X, e.civilians[i].Y = nx, ny
			e.dirty = true
			e.status = ""
		}
	}
}

func (e *Editor) placeAt(wx, wy float64) {
	if e.err != nil || e.grid.W < 1 {
		return
	}
	tx, ty := e.tileAt(wx, wy)
	switch e.stamp {
	case stampGrenadier, stampRocketeer, stampPatrol:
		e.enemies = append(e.enemies, data.Enemy{X: tx, Y: ty, Kind: e.stamp})
		e.sel = pick{kind: pickEnemy, i: len(e.enemies) - 1}
	case stampBarrel:
		e.barrels = append(e.barrels, data.BarrelSpec{X: tx, Y: ty})
		e.sel = pick{kind: pickBarrel, i: len(e.barrels) - 1}
	case stampHut, stampDoor:
		tw, th := 2, 2
		if tx > e.grid.W-tw {
			tx = e.grid.W - tw
		}
		if ty > e.grid.H-th {
			ty = e.grid.H - th
		}
		if tx < 0 || ty < 0 {
			return
		}
		b := data.BuildingSpec{X: tx, Y: ty, W: tw, H: th}
		if e.stamp == stampDoor {
			b.Door = true
			b.Spawn = 2.5
		}
		e.buildings = append(e.buildings, b)
		e.sel = pick{kind: pickHut, i: len(e.buildings) - 1}
	case stampCivilian:
		e.civilians = append(e.civilians, data.Enemy{X: tx, Y: ty})
		e.sel = pick{kind: pickCivilian, i: len(e.civilians) - 1}
	case stampGrenades, stampRockets:
		kind := "grenades"
		if e.stamp == stampRockets {
			kind = "rockets"
		}
		e.pickups = append(e.pickups, data.PickupSpec{X: tx, Y: ty, Kind: kind})
		e.sel = pick{kind: pickCrate, i: len(e.pickups) - 1}
	default:
		e.enemies = append(e.enemies, data.Enemy{X: tx, Y: ty, Kind: "grunt"})
		e.sel = pick{kind: pickEnemy, i: len(e.enemies) - 1}
	}
	e.dirty = true
	e.status = ""
}

func (e *Editor) addPatrolPoint(wx, wy float64) {
	i := -1
	if e.sel.kind == pickEnemy || e.sel.kind == pickLine {
		i = e.sel.i
	}
	if i < 0 || i >= len(e.enemies) || e.enemies[i].Kind != "patrol" {
		e.status = "select a patrol first"
		return
	}
	tx, ty := e.tileAt(wx, wy)
	e.enemies[i].Line = append(e.enemies[i].Line, [2]int{tx, ty})
	e.sel = pick{kind: pickLine, i: i, j: len(e.enemies[i].Line) - 1}
	e.dirty = true
	e.status = ""
}

func (e *Editor) removeSelected() {
	p := e.sel
	switch p.kind {
	case pickNone, pickStart:
		if p.kind == pickStart {
			e.status = "the start stays"
		}
		return
	case pickEnemy:
		e.enemies = append(e.enemies[:p.i], e.enemies[p.i+1:]...)
	case pickLine:
		en := &e.enemies[p.i]
		en.Line = append(en.Line[:p.j], en.Line[p.j+1:]...)
	case pickBarrel:
		e.barrels = append(e.barrels[:p.i], e.barrels[p.i+1:]...)
	case pickHut:
		e.buildings = append(e.buildings[:p.i], e.buildings[p.i+1:]...)
	case pickCrate:
		e.pickups = append(e.pickups[:p.i], e.pickups[p.i+1:]...)
	case pickCivilian:
		e.civilians = append(e.civilians[:p.i], e.civilians[p.i+1:]...)
	}
	e.sel = pick{}
	e.grabbing = false
	e.dirty = true
	e.status = ""
}

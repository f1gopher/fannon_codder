package render

import (
	"image"
	"image/color"
	"math"

	"fannon-codder/internal/sim"
)

// terrainPx is source pixels per world pixel. 4 keeps a full phase small
// enough to rebuild when a mine becomes grass, and the linear filter covers
// the rest of the way to the screen.
const terrainPx = 4

// shoreWobble is how far a drawn edge may leave its cell, in world pixels.
const shoreWobble = 3.0

const (
	surfGrass surface = iota
	surfSnow
	surfTree
	surfShallow
	surfDeep
	surfIce
	surfSand
	surfCliff
	surfRamp
	surfBridge
)

type surface uint8

const (
	maskNone  = 0
	maskWater = 96
	maskSand  = 170
	maskIce   = 255
)

// terrainPaint is the generated picture for one map. Ground is straight-alpha
// RGBA. Anim repeats the water, sand, and ice pixels. Mask is the shader's
// material channel (R). Canopies are one image per connected forest.
type terrainPaint struct {
	ground *image.RGBA
	anim   *image.RGBA
	mask   *image.RGBA
	surf   []surface
	ids    []int
	trees  []treeCanopy
}

type treeCanopy struct {
	img     *image.RGBA
	originX float64
	originY float64
	runs    []treeRun
}

type treeRun struct {
	rect   image.Rectangle
	sprite *image.RGBA // this cell's pixels only; siblings stay transparent
	world  float64     // top-left world x
	north  float64     // top-left world y
	foot   float64     // south edge, used to Y-sort
	footX  float64
}

func cellSurface(t sim.Tile, arctic bool) surface {
	switch t {
	case sim.TileTree:
		return surfTree
	case sim.TileWaterShallow:
		return surfShallow
	case sim.TileWaterDeep:
		return surfDeep
	case sim.TileIce:
		return surfIce
	case sim.TileQuicksand:
		return surfSand
	case sim.TileCliff:
		return surfCliff
	case sim.TileRamp:
		return surfRamp
	case sim.TileBridge:
		return surfBridge
	default:
		if arctic {
			return surfSnow
		}
		return surfGrass
	}
}

func surfacesOf(m sim.Map, arctic bool) []surface {
	s := make([]surface, m.W*m.H)
	for i := range s {
		s[i] = cellSurface(m.Tiles[i], arctic)
	}
	return s
}

// labelRegions gives 8-connected cells of the same surface one id.
func labelRegions(surf []surface, w, h int) []int {
	ids := make([]int, w*h)
	for i := range ids {
		ids[i] = -1
	}
	next := 0
	stack := make([]int, 0, 64)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			start := y*w + x
			if ids[start] >= 0 {
				continue
			}
			want := surf[start]
			ids[start] = next
			stack = append(stack[:0], start)
			for len(stack) > 0 {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				nx, ny := n%w, n/w
				for oy := -1; oy <= 1; oy++ {
					for ox := -1; ox <= 1; ox++ {
						if ox == 0 && oy == 0 {
							continue
						}
						xx, yy := nx+ox, ny+oy
						if xx < 0 || yy < 0 || xx >= w || yy >= h {
							continue
						}
						j := yy*w + xx
						if ids[j] >= 0 || surf[j] != want {
							continue
						}
						ids[j] = next
						stack = append(stack, j)
					}
				}
			}
			next++
		}
	}
	return ids
}

func paintMap(m sim.Map, terrain string) terrainPaint {
	arctic := arcticTerrain(terrain)
	pal := paletteFor(terrain)
	surf := surfacesOf(m, arctic)
	ids := labelRegions(surf, m.W, m.H)
	pw := m.W * sim.TileSize * terrainPx
	ph := m.H * sim.TileSize * terrainPx
	ground := image.NewRGBA(image.Rect(0, 0, pw, ph))
	kind := make([]uint8, pw*ph)
	noise := buildNoise(m.W, m.H, 11)
	wob := buildNoise(m.W, m.H, 29)
	edge := edgeCells(surf, m.W, m.H)

	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			if edge[ty*m.W+tx] {
				paintEdgeCell(ground, kind, m, surf, pal, wob, arctic, tx, ty)
				continue
			}
			fillCell(ground, kind, m, surf, arctic, tx, ty)
		}
	}
	stampTreeShade(ground, m, surf, arctic)
	stampMines(ground, m, pal)
	anim, mask := animLayers(ground, kind)
	stampBridges(ground, m, surf, pal)
	clearBridgeAnim(anim, mask, m, surf)
	trees := paintCanopies(m, surf, ids, arctic, noise, ground)
	return terrainPaint{ground: ground, anim: anim, mask: mask, surf: surf, ids: ids, trees: trees}
}

func edgeCells(surf []surface, w, h int) []bool {
	edge := make([]bool, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s := surf[y*w+x]
			if s == surfCliff || s == surfRamp || s == surfBridge || s == surfTree {
				edge[y*w+x] = true
				continue
			}
			for oy := -1; oy <= 1 && !edge[y*w+x]; oy++ {
				for ox := -1; ox <= 1; ox++ {
					xx, yy := x+ox, y+oy
					if xx < 0 || yy < 0 || xx >= w || yy >= h {
						continue
					}
					n := surf[yy*w+xx]
					if n != s || n == surfShallow || n == surfDeep || n == surfCliff {
						edge[y*w+x] = true
						break
					}
				}
			}
		}
	}
	return edge
}

func fillCell(dst *image.RGBA, kind []uint8, m sim.Map, surf []surface, arctic bool, tx, ty int) {
	s := surf[ty*m.W+tx]
	base := groundSurface(surf, m.W, m.H, s, tx, ty)
	x0 := tx * sim.TileSize * terrainPx
	y0 := ty * sim.TileSize * terrainPx
	span := sim.TileSize * terrainPx
	pw := dst.Bounds().Dx()
	for y := 0; y < span; y++ {
		wy := (float64(y0+y) + 0.5) / terrainPx
		row := (y0 + y) * pw
		for x := 0; x < span; x++ {
			wx := (float64(x0+x) + 0.5) / terrainPx
			c, k := plateColor(base, wx, wy, arctic)
			setRGBA(dst, x0+x, y0+y, c)
			kind[row+x0+x] = k
		}
	}
}

func groundSurface(surf []surface, w, h int, s surface, tx, ty int) surface {
	switch s {
	case surfTree:
		return openGround(surf, w, h)
	case surfBridge:
		return bridgeUnder(surf, w, h, tx, ty, openGround(surf, w, h))
	default:
		return s
	}
}

func paintEdgeCell(dst *image.RGBA, kind []uint8, m sim.Map, surf []surface, pal Palette, wob *noiseGrid, arctic bool, tx, ty int) {
	pw := dst.Bounds().Dx()
	x0 := tx * sim.TileSize * terrainPx
	y0 := ty * sim.TileSize * terrainPx
	span := sim.TileSize * terrainPx
	for y := 0; y < span; y++ {
		for x := 0; x < span; x++ {
			px := x0 + x
			py := y0 + y
			wx := (float64(px) + 0.5) / terrainPx
			wy := (float64(py) + 0.5) / terrainPx
			ox := (sampleNoise(wob, wx, wy) - 0.5) * (shoreWobble * 2)
			oy := (sampleNoise(wob, wy, wx) - 0.5) * (shoreWobble * 2)
			sx, sy := cellAt(m, wx+ox, wy+oy)
			base := surf[sy*m.W+sx]
			trueS := surf[ty*m.W+tx]
			// A water cell keeps its own water. Letting the wobble pull grass
			// or the other depth in paints an olive or bright stripe on the bank.
			if trueS == surfShallow || trueS == surfDeep {
				base = trueS
			}
			c, k := colorOf(arctic, surf, m.W, m.H, base, tx, ty, wx, wy)
			c = shadeCell(pal, surf, m, tx, ty, trueS, wx, wy, c)
			setRGBA(dst, px, py, c)
			kind[py*pw+px] = k
		}
	}
}

func cellAt(m sim.Map, wx, wy float64) (int, int) {
	tx := int(math.Floor(wx / sim.TileSize))
	ty := int(math.Floor(wy / sim.TileSize))
	if tx < 0 {
		tx = 0
	}
	if ty < 0 {
		ty = 0
	}
	if tx >= m.W {
		tx = m.W - 1
	}
	if ty >= m.H {
		ty = m.H - 1
	}
	return tx, ty
}

func colorOf(arctic bool, surf []surface, w, h int, s surface, tx, ty int, wx, wy float64) (color.RGBA, uint8) {
	switch s {
	case surfTree:
		s = openGround(surf, w, h)
	case surfBridge:
		s = bridgeUnder(surf, w, h, tx, ty, openGround(surf, w, h))
	}
	return plateColor(s, wx, wy, arctic)
}

// plateColor reads the approved meadow painting. Arctic rock is the same
// ledge cooled down so it does not sit warm against the snow.
func plateColor(s surface, wx, wy float64, arctic bool) (color.RGBA, uint8) {
	switch s {
	case surfSnow:
		return snowPlate.at(wx, wy), maskNone
	case surfShallow:
		return shallowPlate.at(wx, wy), maskWater
	case surfDeep:
		return deepPlate.at(wx, wy), maskWater
	case surfIce:
		return icePlate.at(wx, wy), maskIce
	case surfSand:
		return sandPlate.at(wx, wy), maskSand
	case surfCliff:
		c := rockPlate.atScale(wx, wy, 5)
		if arctic {
			c = mixRGBA(c, color.RGBA{R: 168, G: 184, B: 196, A: 255}, 0.35)
		}
		return c, maskNone
	case surfRamp:
		return dirtPlate.at(wx, wy), maskNone
	default:
		return grassPlate.at(wx, wy), maskNone
	}
}

func openGround(surf []surface, _, _ int) surface {
	for _, s := range surf {
		if s == surfSnow {
			return surfSnow
		}
	}
	return surfGrass
}

func bridgeUnder(surf []surface, w, h, tx, ty int, open surface) surface {
	deep, shallow, ice := false, false, false
	for oy := -1; oy <= 1; oy++ {
		for ox := -1; ox <= 1; ox++ {
			if ox == 0 && oy == 0 {
				continue
			}
			xx, yy := tx+ox, ty+oy
			if xx < 0 || yy < 0 || xx >= w || yy >= h {
				continue
			}
			switch surf[yy*w+xx] {
			case surfDeep:
				deep = true
			case surfShallow:
				shallow = true
			case surfIce:
				ice = true
			}
		}
	}
	if deep {
		return surfDeep
	}
	if shallow {
		return surfShallow
	}
	if ice {
		return surfIce
	}
	return open
}

func shadeCell(pal Palette, surf []surface, m sim.Map, tx, ty int, trueS surface, wx, wy float64, c color.RGBA) color.RGBA {
	localX := wx - float64(tx*sim.TileSize)
	localY := wy - float64(ty*sim.TileSize)
	if trueS == surfGrass || trueS == surfSnow {
		if d, ok := nearWater(surf, m.W, m.H, tx, ty, wx, wy); ok && d < sim.TileSize {
			t := 1 - d/sim.TileSize
			if trueS == surfGrass {
				c = mixRGBA(c, sandPlate.at(wx, wy), t*0.85)
			} else {
				c = mixRGBA(c, grassPlate.at(wx, wy), t*0.35)
			}
			c = beadFoam(c, pal, wx, wy, d)
		}
	}
	if trueS == surfShallow || trueS == surfDeep {
		c = gradeWater(c, surf, m.W, m.H, tx, ty, trueS, wx, wy)
	}
	if trueS == surfCliff {
		if south, ok := neighbor(surf, m.W, m.H, tx, ty, 0, 1); ok && south != surfCliff && south != surfRamp && localY > 11 {
			c = mixRGBA(c, color.RGBA{R: 42, G: 38, B: 34, A: 255}, 0.5)
		}
		if east, ok := neighbor(surf, m.W, m.H, tx, ty, 1, 0); ok && east != surfCliff && east != surfRamp && localX > 12 {
			c = mixRGBA(c, color.RGBA{R: 42, G: 38, B: 34, A: 255}, 0.4)
		}
	}
	if trueS != surfCliff && trueS != surfRamp {
		if north, ok := neighbor(surf, m.W, m.H, tx, ty, 0, -1); ok && north == surfCliff && localY < 4 {
			c = mixRGBA(c, color.RGBA{R: 20, G: 24, B: 28, A: 255}, 0.4*(1-localY/4))
		}
	}
	return c
}

// beadFoam breaks the shoreline into soft clumps instead of one ruled stroke.
func beadFoam(c color.RGBA, pal Palette, wx, wy, d float64) color.RGBA {
	n := fineNoise(wx, wy)
	reach := 1.15 + n*1.9
	if d >= reach {
		return c
	}
	return mixRGBA(c, pal.Foam, (1-d/reach)*(0.4+0.55*n))
}

// gradeWater lets shallow and deep meet, and pulls a bank down to a deeper blue.
func gradeWater(c color.RGBA, surf []surface, w, h, tx, ty int, self surface, wx, wy float64) color.RGBA {
	other := surfDeep
	if self == surfDeep {
		other = surfShallow
	}
	if d, ok := distToMatch(surf, w, h, tx, ty, wx, wy, func(s surface) bool { return s == other }); ok && d < 14 {
		oc, _ := plateColor(other, wx, wy, false)
		c = mixRGBA(c, oc, (1-d/14)*0.5)
	}
	if d, ok := distToMatch(surf, w, h, tx, ty, wx, wy, landBesideWater); ok && d < 16 {
		bank := color.RGBA{R: 14, G: 96, B: 128, A: 255}
		c = mixRGBA(c, bank, (1-d/16)*0.42)
	}
	return c
}

func landBesideWater(s surface) bool {
	return s != surfShallow && s != surfDeep && s != surfBridge
}

func distToMatch(surf []surface, w, h, tx, ty int, wx, wy float64, match func(surface) bool) (float64, bool) {
	best := math.MaxFloat64
	hit := false
	for oy := -1; oy <= 1; oy++ {
		for ox := -1; ox <= 1; ox++ {
			xx, yy := tx+ox, ty+oy
			if xx < 0 || yy < 0 || xx >= w || yy >= h {
				continue
			}
			if !match(surf[yy*w+xx]) {
				continue
			}
			x0 := float64(xx * sim.TileSize)
			y0 := float64(yy * sim.TileSize)
			d := distToRect(wx, wy, x0, y0, x0+sim.TileSize, y0+sim.TileSize)
			if d < 0 {
				d = 0
			}
			if d < best {
				best = d
				hit = true
			}
		}
	}
	return best, hit
}

func nearWater(surf []surface, w, h, tx, ty int, wx, wy float64) (float64, bool) {
	best := math.MaxFloat64
	hit := false
	for oy := -1; oy <= 1; oy++ {
		for ox := -1; ox <= 1; ox++ {
			xx, yy := tx+ox, ty+oy
			if xx < 0 || yy < 0 || xx >= w || yy >= h {
				continue
			}
			s := surf[yy*w+xx]
			if s != surfShallow && s != surfDeep {
				continue
			}
			hit = true
			x0 := float64(xx * sim.TileSize)
			y0 := float64(yy * sim.TileSize)
			d := distToRect(wx, wy, x0, y0, x0+sim.TileSize, y0+sim.TileSize)
			if d < best {
				best = d
			}
		}
	}
	return best, hit
}

func neighbor(surf []surface, w, h, tx, ty, ox, oy int) (surface, bool) {
	xx, yy := tx+ox, ty+oy
	if xx < 0 || yy < 0 || xx >= w || yy >= h {
		return 0, false
	}
	return surf[yy*w+xx], true
}

func stampMines(dst *image.RGBA, m sim.Map, pal Palette) {
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			if m.Tiles[ty*m.W+tx] != sim.TileMine {
				continue
			}
			fillWorldRect(dst, float64(tx*sim.TileSize)+5, float64(ty*sim.TileSize)+5, 6, 6, pal.Mine)
			fillWorldRect(dst, float64(tx*sim.TileSize)+7, float64(ty*sim.TileSize)+7, 2, 2, pal.MineMark)
		}
	}
}

func stampBridges(dst *image.RGBA, m sim.Map, surf []surface, pal Palette) {
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			if surf[ty*m.W+tx] != surfBridge {
				continue
			}
			vertical := bridgeVertical(surf, m.W, m.H, tx, ty)
			x0 := float64(tx * sim.TileSize)
			y0 := float64(ty * sim.TileSize)
			if vertical {
				for y := 0.0; y < sim.TileSize; y += 4 {
					fillWorldRect(dst, x0, y0+y, sim.TileSize, 3, pal.Plank)
					fillWorldRect(dst, x0, y0+y+3, sim.TileSize, 1, pal.PlankGap)
				}
			} else {
				for x := 0.0; x < sim.TileSize; x += 4 {
					fillWorldRect(dst, x0+x, y0, 3, sim.TileSize, pal.Plank)
					fillWorldRect(dst, x0+x+3, y0, 1, sim.TileSize, pal.PlankGap)
				}
			}
		}
	}
}

func clearBridgeAnim(anim, mask *image.RGBA, m sim.Map, surf []surface) {
	pw := anim.Bounds().Dx()
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			if surf[ty*m.W+tx] != surfBridge {
				continue
			}
			x0 := tx * sim.TileSize * terrainPx
			y0 := ty * sim.TileSize * terrainPx
			x1 := x0 + sim.TileSize*terrainPx
			y1 := y0 + sim.TileSize*terrainPx
			for py := y0; py < y1; py++ {
				for px := x0; px < x1; px++ {
					i := (py*pw + px) * 4
					anim.Pix[i+3] = 0
					mask.Pix[i] = 0
					mask.Pix[i+3] = 0
				}
			}
		}
	}
}

func bridgeVertical(surf []surface, w, h, tx, ty int) bool {
	v, hz := 0, 0
	if s, ok := neighbor(surf, w, h, tx, ty, 0, -1); ok && s == surfBridge {
		v++
	}
	if s, ok := neighbor(surf, w, h, tx, ty, 0, 1); ok && s == surfBridge {
		v++
	}
	if s, ok := neighbor(surf, w, h, tx, ty, -1, 0); ok && s == surfBridge {
		hz++
	}
	if s, ok := neighbor(surf, w, h, tx, ty, 1, 0); ok && s == surfBridge {
		hz++
	}
	return v >= hz
}

func fillWorldRect(dst *image.RGBA, x, y, w, h float64, c color.RGBA) {
	pw := dst.Bounds().Dx()
	ph := dst.Bounds().Dy()
	x0 := int(math.Floor(x * terrainPx))
	y0 := int(math.Floor(y * terrainPx))
	x1 := int(math.Ceil((x + w) * terrainPx))
	y1 := int(math.Ceil((y + h) * terrainPx))
	for py := y0; py < y1; py++ {
		if py < 0 || py >= ph {
			continue
		}
		for px := x0; px < x1; px++ {
			if px < 0 || px >= pw {
				continue
			}
			setRGBA(dst, px, py, c)
		}
	}
}

func animLayers(ground *image.RGBA, kind []uint8) (*image.RGBA, *image.RGBA) {
	b := ground.Bounds()
	anim := image.NewRGBA(b)
	mask := image.NewRGBA(b)
	for i := 0; i < len(kind); i++ {
		k := kind[i]
		if k == maskNone {
			continue
		}
		o := i * 4
		anim.Pix[o] = ground.Pix[o]
		anim.Pix[o+1] = ground.Pix[o+1]
		anim.Pix[o+2] = ground.Pix[o+2]
		anim.Pix[o+3] = 255
		mask.Pix[o] = k
		mask.Pix[o+1] = k
		mask.Pix[o+2] = k
		mask.Pix[o+3] = 255
	}
	return anim, mask
}

func paintCanopies(m sim.Map, surf []surface, ids []int, arctic bool, noise *noiseGrid, ground *image.RGBA) []treeCanopy {
	seen := map[int]bool{}
	var out []treeCanopy
	for i, s := range surf {
		if s != surfTree || seen[ids[i]] {
			continue
		}
		seen[ids[i]] = true
		out = append(out, paintOneCanopy(m, surf, ids, ids[i], arctic, noise, ground))
	}
	return out
}

// canopyPad is how far a crown's fringe may hang past its cell, in world
// pixels. The fringe is foliage, so the slice has to include all of it.
const canopyPad = 36.0

// A crown is one tree cell: a round mass, smaller crowns stuck to it, and a
// few shrubs on the side that faces open ground.
type crown struct {
	tx, ty int
	cx, cy float64
	mainR  float64
	lobes  [6]lobe
	edge   bool
	sprigs [4]lobe
}

type lobe struct {
	x, y, r float64
}

func paintOneCanopy(m sim.Map, surf []surface, ids []int, id int, arctic bool, noise *noiseGrid, ground *image.RGBA) treeCanopy {
	minX, minY := m.W, m.H
	maxX, maxY := 0, 0
	var cells []image.Point
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			i := ty*m.W + tx
			if ids[i] != id || surf[i] != surfTree {
				continue
			}
			cells = append(cells, image.Point{X: tx, Y: ty})
			if tx < minX {
				minX = tx
			}
			if ty < minY {
				minY = ty
			}
			if tx > maxX {
				maxX = tx
			}
			if ty > maxY {
				maxY = ty
			}
		}
	}
	crowns := buildCrowns(cells, surf, m.W, m.H)
	originX := float64(minX*sim.TileSize) - canopyPad
	originY := float64(minY*sim.TileSize) - canopyPad
	worldW := float64((maxX-minX+1)*sim.TileSize) + canopyPad*2
	worldH := float64((maxY-minY+1)*sim.TileSize) + canopyPad*2
	pw := int(math.Ceil(worldW * terrainPx))
	ph := int(math.Ceil(worldH * terrainPx))
	img := image.NewRGBA(image.Rect(0, 0, pw, ph))
	owner := make([]int16, pw*ph)
	for i := range owner {
		owner[i] = -1
	}
	cover, tips, ccx, ccy := splatCrowns(crowns, originX, originY, pw, ph)
	gw := ground.Bounds().Dx()
	gh := ground.Bounds().Dy()
	for py := 0; py < ph; py++ {
		wy := originY + (float64(py)+0.5)/terrainPx
		row := py * pw
		for px := 0; px < pw; px++ {
			d := float64(cover[row+px])
			sprig := float64(tips[row+px])
			if d > 12 && sprig < 0.02 {
				continue
			}
			wx := originX + (float64(px)+0.5)/terrainPx
			a := canopyAlpha(d, wx, wy, noise)
			if sprig > a {
				a = sprig
			}
			if a < 0.012 {
				continue
			}
			// Scale 4 puts one painted crown on about a tile.
			c := canopyPlate.atScale(wx, wy, 4)
			c = shadeCrown(c, d, wx-float64(ccx[row+px]), wy-float64(ccy[row+px]), arctic, sampleNoise(noise, wx*1.7, wy*1.7))
			c = mergeFringe(c, ground, gw, gh, wx, wy, a, arctic)
			c.A = uint8(a * 255)
			if c.A == 0 {
				continue
			}
			setRGBA(img, px, py, c)
			owner[row+px] = int16(nearestTree(cells, wx, wy))
		}
	}
	can := treeCanopy{img: img, originX: originX, originY: originY}
	for i, cell := range cells {
		can.runs = append(can.runs, sliceRun(can, owner, pw, i, cell.X, cell.Y))
	}
	return can
}

func buildCrowns(cells []image.Point, surf []surface, w, h int) []crown {
	out := make([]crown, len(cells))
	for i, cell := range cells {
		hsh := hashU(cell.X, cell.Y, 5)
		jx := (float64(hsh&255)/255 - 0.5) * 4
		jy := (float64((hsh>>8)&255)/255-0.5)*3 - 1.2
		cr := crown{
			tx: cell.X,
			ty: cell.Y,
			cx: float64(cell.X*sim.TileSize) + sim.TileSize/2 + jx,
			cy: float64(cell.Y*sim.TileSize) + sim.TileSize/2 + jy,
			// Wide enough that two neighbouring cells stay solid where they meet.
			// The lobes, not this core, are what scallop the outer edge.
			mainR: 14.2 + float64((hsh>>16)&3)*0.25,
		}
		for n := 0; n < len(cr.lobes); n++ {
			hi := hashU(cell.X*4+n, cell.Y*3+n, 9)
			ang := float64(n)*(math.Pi*2/float64(len(cr.lobes))) + (float64(hi&255)/255-0.5)*1.05
			dist := 8.5 + float64((hi>>8)&255)/255*4.5
			cr.lobes[n] = lobe{
				x: cr.cx + math.Cos(ang)*dist,
				y: cr.cy + math.Sin(ang)*dist,
				r: 7.4 + float64((hi>>16)&255)/255*3.6,
			}
		}
		ox, oy, edge := standOutward(surf, w, h, cell.X, cell.Y)
		cr.edge = edge
		if edge {
			cr.lobes[0] = lobe{
				x: cr.cx + ox*9,
				y: cr.cy + oy*9,
				r: 11 + float64(hsh&3)*0.35,
			}
			baseX := float64(cell.X*sim.TileSize) + sim.TileSize/2
			baseY := float64(cell.Y*sim.TileSize) + sim.TileSize/2
			for n := 0; n < len(cr.sprigs); n++ {
				hi := hashU(cell.X*5+n, cell.Y*3+2, 40+n)
				along := 15 + float64(hi%9)
				side := (float64((hi>>8)%100)/50 - 1) * 9
				cr.sprigs[n] = lobe{
					x: baseX + ox*along - oy*side,
					y: baseY + oy*along + ox*side,
					r: 4.4 + float64((hi>>16)%4)*0.55,
				}
			}
		}
		out[i] = cr
	}
	return out
}

// standOutward points from a tree cell toward open ground. An interior cell
// of a stand has no outward side.
func standOutward(surf []surface, w, h, tx, ty int) (float64, float64, bool) {
	sx, sy := 0.0, 0.0
	for oy := -1; oy <= 1; oy++ {
		for ox := -1; ox <= 1; ox++ {
			if ox == 0 && oy == 0 {
				continue
			}
			xx, yy := tx+ox, ty+oy
			if xx >= 0 && yy >= 0 && xx < w && yy < h && surf[yy*w+xx] == surfTree {
				continue
			}
			sx += float64(ox)
			sy += float64(oy)
		}
	}
	l := math.Hypot(sx, sy)
	if l < 0.5 {
		return 0, 0, false
	}
	return sx / l, sy / l, true
}

// splatCrowns writes the signed distance of the crown cluster. Each crown is
// its own circle, so the outline is a bunch of bushes and not one blob.
// Tips hold the soft shrubs that sit outside that mass.
func splatCrowns(crowns []crown, originX, originY float64, pw, ph int) (cover, tips, ccx, ccy []float32) {
	n := pw * ph
	cover = make([]float32, n)
	tips = make([]float32, n)
	ccx = make([]float32, n)
	ccy = make([]float32, n)
	for i := range cover {
		cover[i] = 99
	}
	for i := range crowns {
		cr := &crowns[i]
		splatDisc(cover, ccx, ccy, originX, originY, pw, ph, cr.cx, cr.cy, cr.mainR, cr.cx, cr.cy)
		for n := range cr.lobes {
			lb := &cr.lobes[n]
			splatDisc(cover, ccx, ccy, originX, originY, pw, ph, lb.x, lb.y, lb.r, cr.cx, cr.cy)
		}
		if !cr.edge {
			continue
		}
		for n := range cr.sprigs {
			lb := &cr.sprigs[n]
			splatTip(tips, originX, originY, pw, ph, lb.x, lb.y, lb.r)
		}
	}
	return cover, tips, ccx, ccy
}

func splatDisc(cover, ccx, ccy []float32, originX, originY float64, pw, ph int, lx, ly, r, crownX, crownY float64) {
	const fringe = 12.0
	rad := r + fringe
	x0 := int(math.Floor((lx - rad - originX) * terrainPx))
	y0 := int(math.Floor((ly - rad - originY) * terrainPx))
	x1 := int(math.Ceil((lx + rad - originX) * terrainPx))
	y1 := int(math.Ceil((ly + rad - originY) * terrainPx))
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > pw {
		x1 = pw
	}
	if y1 > ph {
		y1 = ph
	}
	rad2 := rad * rad
	cx := float32(crownX)
	cy := float32(crownY)
	for py := y0; py < y1; py++ {
		dy := originY + (float64(py)+0.5)/terrainPx - ly
		row := py * pw
		for px := x0; px < x1; px++ {
			dx := originX + (float64(px)+0.5)/terrainPx - lx
			d2 := dx*dx + dy*dy
			if d2 > rad2 {
				continue
			}
			d := float32(math.Sqrt(d2) - r)
			i := row + px
			if d < cover[i] {
				cover[i] = d
				ccx[i] = cx
				ccy[i] = cy
			}
		}
	}
}

func splatTip(tips []float32, originX, originY float64, pw, ph int, lx, ly, r float64) {
	const fringe = 8.0
	rad := r + fringe
	x0 := int(math.Floor((lx - rad - originX) * terrainPx))
	y0 := int(math.Floor((ly - rad - originY) * terrainPx))
	x1 := int(math.Ceil((lx + rad - originX) * terrainPx))
	y1 := int(math.Ceil((ly + rad - originY) * terrainPx))
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > pw {
		x1 = pw
	}
	if y1 > ph {
		y1 = ph
	}
	rad2 := rad * rad
	for py := y0; py < y1; py++ {
		dy := originY + (float64(py)+0.5)/terrainPx - ly
		row := py * pw
		for px := x0; px < x1; px++ {
			dx := originX + (float64(px)+0.5)/terrainPx - lx
			d2 := dx*dx + dy*dy
			if d2 > rad2 {
				continue
			}
			ld := math.Sqrt(d2) - r
			a := float32((1 - smooth01((ld+1.6)/8.2)) * 0.58)
			i := row + px
			if a > tips[i] {
				tips[i] = a
			}
		}
	}
}

// canopyAlpha keeps the middle of a crown solid and lets the outer leaves
// thin out into the ground. The falloff starts inside the crown so the rim
// is foliage, not a line drawn around it.
func canopyAlpha(d, wx, wy float64, noise *noiseGrid) float64 {
	if d > 30 {
		return 0
	}
	e := d + foliageNoise(wx, wy)*3.2 + (sampleNoise(noise, wx, wy)-0.5)*2
	if e <= -4.2 {
		return 1
	}
	if e >= 7.5 {
		return 0
	}
	return 1 - smooth01((e+4.2)/11.7)
}

func shadeCrown(c color.RGBA, d, dx, dy float64, arctic bool, n float64) color.RGBA {
	if arctic && n > 0.58 && d < -2 {
		c = mixRGBA(c, color.RGBA{R: 232, G: 242, B: 246, A: 255}, 0.28)
	}
	// A little south-east shade only. A dark stroke around the whole crown
	// reads as a cut-out, which is the opposite of foliage meeting grass.
	se := clamp01((dx + dy) / 36)
	if arctic {
		c = mixRGBA(c, color.RGBA{R: 150, G: 168, B: 172, A: 255}, se*0.12)
		return c
	}
	c = mixRGBA(c, color.RGBA{R: 46, G: 78, B: 38, A: 255}, se*0.16)
	return c
}

// mergeFringe pulls the transparent tips toward the ground already painted
// there, and lifts the plate's dark gaps so the rim is leaves rather than
// a black line.
func mergeFringe(c color.RGBA, ground *image.RGBA, gw, gh int, wx, wy, a float64, arctic bool) color.RGBA {
	if ground == nil {
		return c
	}
	// Only the blackest gaps in the plate, and only near the rim. A broad
	// recolour here draws a halo around the whole stand.
	sum := int(c.R) + int(c.G) + int(c.B)
	if a < 0.85 && sum < 96 {
		leaf := color.RGBA{R: 70, G: 108, B: 46, A: 255}
		if arctic {
			leaf = color.RGBA{R: 120, G: 150, B: 130, A: 255}
		}
		c = mixRGBA(c, leaf, 0.45)
	}
	if a > 0.5 {
		return c
	}
	gx := int(math.Floor(wx * terrainPx))
	gy := int(math.Floor(wy * terrainPx))
	if gx < 0 || gy < 0 || gx >= gw || gy >= gh {
		return c
	}
	o := (gy*gw + gx) * 4
	g := color.RGBA{R: ground.Pix[o], G: ground.Pix[o+1], B: ground.Pix[o+2], A: 255}
	return mixRGBA(c, g, (0.5-a)*0.8)
}

func sliceRun(can treeCanopy, owner []int16, pw, idx, tx, ty int) treeRun {
	b := can.img.Bounds()
	x0 := int(math.Floor((float64(tx*sim.TileSize) - canopyPad - can.originX) * terrainPx))
	y0 := int(math.Floor((float64(ty*sim.TileSize) - canopyPad - can.originY) * terrainPx))
	x1 := int(math.Ceil((float64((tx+1)*sim.TileSize) + canopyPad - can.originX) * terrainPx))
	y1 := int(math.Ceil((float64((ty+1)*sim.TileSize) + canopyPad - can.originY) * terrainPx))
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > b.Dx() {
		x1 = b.Dx()
	}
	if y1 > b.Dy() {
		y1 = b.Dy()
	}
	minX, minY := b.Dx(), b.Dy()
	maxX, maxY := -1, -1
	for py := y0; py < y1; py++ {
		row := py * pw
		for px := x0; px < x1; px++ {
			if int(owner[row+px]) != idx {
				continue
			}
			if px < minX {
				minX = px
			}
			if py < minY {
				minY = py
			}
			if px > maxX {
				maxX = px
			}
			if py > maxY {
				maxY = py
			}
		}
	}
	if maxX < minX {
		return treeRun{foot: float64((ty + 1) * sim.TileSize), footX: float64(tx*sim.TileSize) + sim.TileSize/2}
	}
	w := maxX - minX + 1
	h := maxY - minY + 1
	sprite := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			src := can.img.PixOffset(minX+x, minY+y)
			dst := sprite.PixOffset(x, y)
			sprite.Pix[dst] = can.img.Pix[src]
			sprite.Pix[dst+1] = can.img.Pix[src+1]
			sprite.Pix[dst+2] = can.img.Pix[src+2]
			sprite.Pix[dst+3] = can.img.Pix[src+3]
		}
	}
	return treeRun{
		rect:   image.Rect(minX, minY, maxX+1, maxY+1),
		sprite: sprite,
		world:  can.originX + float64(minX)/terrainPx,
		north:  can.originY + float64(minY)/terrainPx,
		foot:   float64((ty + 1) * sim.TileSize),
		footX:  float64(tx*sim.TileSize) + sim.TileSize/2,
	}
}

func nearestTree(cells []image.Point, wx, wy float64) int {
	best := 0
	bestD := math.MaxFloat64
	gx := int(math.Floor(wx / sim.TileSize))
	gy := int(math.Floor(wy / sim.TileSize))
	for i, cell := range cells {
		if absInt(cell.X-gx) > 4 || absInt(cell.Y-gy) > 4 {
			continue
		}
		dx := float64(cell.X*sim.TileSize) + sim.TileSize/2 - wx
		dy := float64(cell.Y*sim.TileSize) + sim.TileSize/2 - wy
		d := dx*dx + dy*dy
		if d < bestD-1e-6 {
			bestD = d
			best = i
			continue
		}
		if math.Abs(d-bestD) > 1e-6 {
			continue
		}
		prev := cells[best]
		if cell.Y > prev.Y || (cell.Y == prev.Y && cell.X > prev.X) {
			best = i
		}
	}
	return best
}

// stampTreeShade lays a soft south-east shadow on the ground under each
// stand so the fringe blends into the meadow instead of sitting on it.
func stampTreeShade(dst *image.RGBA, m sim.Map, surf []surface, arctic bool) {
	var cells []image.Point
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			if surf[ty*m.W+tx] == surfTree {
				cells = append(cells, image.Point{X: tx, Y: ty})
			}
		}
	}
	if len(cells) == 0 {
		return
	}
	crowns := buildCrowns(cells, surf, m.W, m.H)
	pw := dst.Bounds().Dx()
	ph := dst.Bounds().Dy()
	strength := make([]uint8, pw*ph)
	minX, minY := cells[0].X, cells[0].Y
	maxX, maxY := cells[0].X, cells[0].Y
	for _, cell := range cells[1:] {
		if cell.X < minX {
			minX = cell.X
		}
		if cell.Y < minY {
			minY = cell.Y
		}
		if cell.X > maxX {
			maxX = cell.X
		}
		if cell.Y > maxY {
			maxY = cell.Y
		}
	}
	const reach = 28.0
	x0 := int(math.Floor((float64(minX*sim.TileSize) - reach) * terrainPx))
	y0 := int(math.Floor((float64(minY*sim.TileSize) - reach) * terrainPx))
	x1 := int(math.Ceil((float64((maxX+1)*sim.TileSize) + reach) * terrainPx))
	y1 := int(math.Ceil((float64((maxY+1)*sim.TileSize) + reach) * terrainPx))
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > pw {
		x1 = pw
	}
	if y1 > ph {
		y1 = ph
	}
	// One sample per world pixel. The ground is scaled with a linear filter,
	// and the shade itself is a wide falloff, so the extra source pixels
	// would not move the edge.
	for py := y0; py < y1; py += terrainPx {
		wy := (float64(py) + 0.5) / terrainPx
		for px := x0; px < x1; px += terrainPx {
			wx := (float64(px) + 0.5) / terrainPx
			// Sample back toward the light so the shade falls south-east.
			d := shadeDist(crowns, wx-3.2, wy-4.0)
			if d > 16 {
				continue
			}
			t := clamp01((d + 6) / 24)
			s := (1 - smooth01(t)) * 0.16
			if s <= 0 {
				continue
			}
			v := uint8(s * 255)
			xEnd := px + terrainPx
			yEnd := py + terrainPx
			if xEnd > x1 {
				xEnd = x1
			}
			if yEnd > y1 {
				yEnd = y1
			}
			for yy := py; yy < yEnd; yy++ {
				row := yy * pw
				for xx := px; xx < xEnd; xx++ {
					if v > strength[row+xx] {
						strength[row+xx] = v
					}
				}
			}
		}
	}
	shadow := color.RGBA{R: 78, G: 98, B: 52, A: 255}
	if arctic {
		shadow = color.RGBA{R: 118, G: 136, B: 146, A: 255}
	}
	for i, v := range strength {
		if v == 0 {
			continue
		}
		o := i * 4
		c := color.RGBA{R: dst.Pix[o], G: dst.Pix[o+1], B: dst.Pix[o+2], A: dst.Pix[o+3]}
		c = mixRGBA(c, shadow, float64(v)/255)
		dst.Pix[o] = c.R
		dst.Pix[o+1] = c.G
		dst.Pix[o+2] = c.B
		dst.Pix[o+3] = c.A
	}
}

// shadeDist is the canopy distance used for the ground shadow. Lobes are
// folded into a slightly larger crown so the shade stays soft and cheap.
func shadeDist(crowns []crown, wx, wy float64) float64 {
	best := 99.0
	gx := int(math.Floor(wx / sim.TileSize))
	gy := int(math.Floor(wy / sim.TileSize))
	for i := range crowns {
		cr := &crowns[i]
		if absInt(cr.tx-gx) > 3 || absInt(cr.ty-gy) > 3 {
			continue
		}
		dx := wx - cr.cx
		dy := wy - cr.cy
		d2 := dx*dx + dy*dy
		if d2 > 42*42 {
			continue
		}
		d := math.Sqrt(d2) - cr.mainR - 3.5
		if d < best {
			best = d
		}
	}
	return best
}

func foliageNoise(wx, wy float64) float64 {
	n := valueNoise(wx, wy, 4.4, 71)
	n2 := valueNoise(wx+3, wy-5, 9.5, 83)
	return (n-0.5)*0.7 + (n2-0.5)*0.3
}

func valueNoise(wx, wy, scale float64, seed int) float64 {
	x := wx / scale
	y := wy / scale
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	fx := x - float64(x0)
	fy := y - float64(y0)
	sx := fx * fx * (3 - 2*fx)
	sy := fy * fy * (3 - 2*fy)
	n00 := hash01(x0, y0, seed)
	n10 := hash01(x0+1, y0, seed)
	n01 := hash01(x0, y0+1, seed)
	n11 := hash01(x0+1, y0+1, seed)
	a := n00 + (n10-n00)*sx
	b := n01 + (n11-n01)*sx
	return a + (b-a)*sy
}

func hash01(x, y, seed int) float64 {
	return float64(hashU(x, y, seed)%10000) / 10000
}

// fineNoise is a small value noise for the foam beads, a few world pixels wide.
func fineNoise(wx, wy float64) float64 {
	x := wx / 4.5
	y := wy / 4.5
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	fx := x - float64(x0)
	fy := y - float64(y0)
	sx := fx * fx * (3 - 2*fx)
	sy := fy * fy * (3 - 2*fy)
	n00 := hash01(x0, y0, 5)
	n10 := hash01(x0+1, y0, 5)
	n01 := hash01(x0, y0+1, 5)
	n11 := hash01(x0+1, y0+1, 5)
	a := n00 + (n10-n00)*sx
	b := n01 + (n11-n01)*sx
	return a + (b-a)*sy
}

func smooth01(t float64) float64 {
	t = clamp01(t)
	return t * t * (3 - 2*t)
}

func distToRect(px, py, x0, y0, x1, y1 float64) float64 {
	dx := math.Max(x0-px, px-x1)
	dy := math.Max(y0-py, py-y1)
	if dx <= 0 && dy <= 0 {
		return math.Max(dx, dy)
	}
	if dx < 0 {
		dx = 0
	}
	if dy < 0 {
		dy = 0
	}
	return math.Hypot(dx, dy)
}

type noiseGrid struct {
	w, h int
	step float64
	v    []float32
}

func buildNoise(mapW, mapH, seed int) *noiseGrid {
	const step = 32.0
	worldW := float64(mapW * sim.TileSize)
	worldH := float64(mapH * sim.TileSize)
	w := int(worldW/step) + 3
	h := int(worldH/step) + 3
	v := make([]float32, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v[y*w+x] = float32(hashU(x, y, seed)%10000) / 10000
		}
	}
	return &noiseGrid{w: w, h: h, step: step, v: v}
}

func sampleNoise(n *noiseGrid, wx, wy float64) float64 {
	x := wx/n.step + 1
	y := wy/n.step + 1
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > float64(n.w-1) {
		x = float64(n.w - 1)
	}
	if y > float64(n.h-1) {
		y = float64(n.h - 1)
	}
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	if x0 >= n.w-1 {
		x0 = n.w - 2
	}
	if y0 >= n.h-1 {
		y0 = n.h - 2
	}
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	fx := x - float64(x0)
	fy := y - float64(y0)
	if fx < 0 {
		fx = 0
	}
	if fy < 0 {
		fy = 0
	}
	if fx > 1 {
		fx = 1
	}
	if fy > 1 {
		fy = 1
	}
	sx := fx * fx * (3 - 2*fx)
	sy := fy * fy * (3 - 2*fy)
	i := y0*n.w + x0
	v00 := float64(n.v[i])
	v10 := float64(n.v[i+1])
	v01 := float64(n.v[i+n.w])
	v11 := float64(n.v[i+n.w+1])
	a := v00 + (v10-v00)*sx
	b := v01 + (v11-v01)*sx
	return a + (b-a)*sy
}

func hashU(x, y, seed int) uint32 {
	n := uint32(x)*0x27d4eb2d + uint32(y)*0x165667b1 + uint32(seed)*0x9e3779b1
	n ^= n >> 15
	n *= 0x85ebca6b
	n ^= n >> 13
	return n
}

func mixRGBA(a, b color.RGBA, t float64) color.RGBA {
	t = clamp01(t)
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: uint8(float64(a.A) + (float64(b.A)-float64(a.A))*t),
	}
}

func setRGBA(img *image.RGBA, x, y int, c color.RGBA) {
	i := img.PixOffset(x, y)
	img.Pix[i] = c.R
	img.Pix[i+1] = c.G
	img.Pix[i+2] = c.B
	img.Pix[i+3] = c.A
}

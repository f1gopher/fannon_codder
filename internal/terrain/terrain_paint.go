package terrain

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

// A beach is paint where land meets water. Widths are world pixels.
// beachSandWander is how far the grass side of the sand moves in and out.
// The water line waves by shoreWave. Corners are rounded by the blurred
// distance in water.go, and the colour fades with the distance offshore.
const (
	beachSand       = 4.0
	beachSandFade   = 3.2
	beachSandWander = 2.8
	beachCorner     = 14.0
	// Shallow stays solid for a few pixels, then fades into deep. The fade
	// wanders so the light-to-dark join is not a ruled line.
	beachShallowHold = 3.5
	beachWaterFade   = 14.0
	beachFadeWander  = 6.0
	beachFadeReach   = beachShallowHold + beachWaterFade + beachFadeWander*0.5
	// beachJoin is how far a second shore pulls that fade. Two shallow
	// bands, such as a shore and a close island, blend into one pool
	// instead of leaving a dark seam where they meet.
	beachJoin = 32.0
)

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
	surfMine
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

func cellSurface(t sim.Tile, open surface) surface {
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
		return open
	}
}

func surfacesOf(m sim.Map, open surface) []surface {
	s := make([]surface, m.W*m.H)
	for i := range s {
		s[i] = cellSurface(m.Tiles[i], open)
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
	lk := lookOf(terrain)
	prepareGrass(m.W*sim.TileSize, m.H*sim.TileSize, lk.kind)
	prepareSnow(m.W*sim.TileSize, m.H*sim.TileSize, lk.kind)
	prepareSand(m.W*sim.TileSize, m.H*sim.TileSize, lk.kind)
	prepareForest(m.W*sim.TileSize, m.H*sim.TileSize, lk.kind)
	surf := surfacesOf(m, lk.open)
	prepareSnowCliffTiles(surf, m.W, m.H, lk.kind, lk.open)
	prepareShoreTiles(surf, m.W, m.H)
	ids := labelRegions(surf, m.W, m.H)
	pw := m.W * sim.TileSize * terrainPx
	ph := m.H * sim.TileSize * terrainPx
	ground := image.NewRGBA(image.Rect(0, 0, pw, ph))
	kind := make([]uint8, pw*ph)
	noise := buildNoise(m.W, m.H, 11)
	wob := buildNoise(m.W, m.H, 29)
	edge := edgeCells(surf, m.W, m.H)
	var corners []shoreCorner
	vw := 0
	var joins []*shoreJoin
	if lk.beach {
		corners, vw = shoreCorners(surf, m.W, m.H)
		joins = shoreJoins(surf, landMasses(surf, m.W, m.H), m.W, m.H)
	}

	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
			if edge[ty*m.W+tx] {
				paintEdgeCell(ground, kind, m, surf, lk, wob, corners, vw, joins, tx, ty)
				continue
			}
			fillCell(ground, kind, m, surf, lk, tx, ty)
		}
	}
	stampTreeShade(ground, m, surf, lk)
	stampMines(ground, m, lk.pal)
	anim, mask := animLayers(ground, kind)
	stampBridges(ground, anim, mask, m, surf, lk)
	trees := paintCanopies(m, surf, ids, lk, noise, ground)
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
	// The shore wave reaches about two cells inland. Mark that band so the
	// curve is painted, not clipped to the first grass cell.
	wide := make([]bool, w*h)
	copy(wide, edge)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if edge[y*w+x] {
				continue
			}
			for oy := -2; oy <= 2 && !wide[y*w+x]; oy++ {
				for ox := -2; ox <= 2; ox++ {
					xx, yy := x+ox, y+oy
					if xx < 0 || yy < 0 || xx >= w || yy >= h {
						continue
					}
					n := surf[yy*w+xx]
					if n == surfShallow || n == surfDeep {
						wide[y*w+x] = true
						break
					}
				}
			}
		}
	}
	return wide
}

func fillCell(dst *image.RGBA, kind []uint8, m sim.Map, surf []surface, lk look, tx, ty int) {
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
			c, k := lk.plateColor(base, wx, wy)
			c, k = paintSnowCliffPixel(lk, base, wx, wy, true, c, k)
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

func paintEdgeCell(dst *image.RGBA, kind []uint8, m sim.Map, surf []surface, lk look, wob *noiseGrid, corners []shoreCorner, vw int, joins []*shoreJoin, tx, ty int) {
	pw := dst.Bounds().Dx()
	x0 := tx * sim.TileSize * terrainPx
	y0 := ty * sim.TileSize * terrainPx
	span := sim.TileSize * terrainPx
	round := cellRounded(corners, vw, tx, ty)
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
			c, k := colorOf(lk, surf, m.W, m.H, base, tx, ty, wx, wy)
			var over bool
			var mk uint8
			c, mk, over = shadeCell(lk, surf, m, tx, ty, trueS, wx, wy, c, corners, vw, joins, round)
			if over {
				k = mk
			}
			c, k = paintSnowCliffPixel(lk, trueS, wx, wy, true, c, k)
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

func colorOf(lk look, surf []surface, w, h int, s surface, tx, ty int, wx, wy float64) (color.RGBA, uint8) {
	switch s {
	case surfTree:
		s = openGround(surf, w, h)
	case surfBridge:
		s = bridgeUnder(surf, w, h, tx, ty, openGround(surf, w, h))
	}
	return lk.plateColor(s, wx, wy)
}

// plateColor reads this terrain's ground, water, and nothing of another
// terrain's. Jungle open ground is grassColor. Snow open ground is
// snowColor. Desert open ground is sandColor. Forest open ground is
// forestColor. Ice, quicksand, and ramps stay the shared materials.
// A snow cliff is painted in cliff.go. Other cliffs stay on the rock plate.
func (lk look) plateColor(s surface, wx, wy float64) (color.RGBA, uint8) {
	switch s {
	case surfShallow, surfDeep:
		into := waterReach
		if sd, ok := shoreSigned(wx, wy); ok {
			into = -sd
			if into < 0 {
				into = 0
			}
		}
		return waterColor(lk, wx, wy, into, into), maskWater
	case surfIce:
		return icePlate.at(wx, wy), maskIce
	case surfSand:
		return sandPlate.at(wx, wy), maskSand
	case surfCliff:
		// Snow moves the face in cliff.go and paints it from the palette.
		// The rock plate would put the old straight texture back.
		if lk.kind == Snow {
			return snowCliffRGB(wx, wy, 0.45), maskNone
		}
		c := rockPlate.atScale(wx, wy, 5)
		if lk.rockMix > 0 {
			c = mixRGBA(c, lk.rockTint, lk.rockMix)
		}
		return c, maskNone
	case surfRamp:
		return dirtPlate.at(wx, wy), maskNone
	default:
		if lk.kind == Jungle {
			return grassSample(wx, wy), maskNone
		}
		if lk.kind == Snow {
			return snowSample(wx, wy), maskNone
		}
		if lk.kind == Desert {
			return sandSample(wx, wy), maskNone
		}
		if lk.kind == Forest {
			return forestSample(wx, wy), maskNone
		}
		return lk.ground.at(wx, wy), maskNone
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

func shadeCell(lk look, surf []surface, m sim.Map, tx, ty int, trueS surface, wx, wy float64, c color.RGBA, corners []shoreCorner, vw int, joins []*shoreJoin, round bool) (color.RGBA, uint8, bool) {
	localX := wx - float64(tx*sim.TileSize)
	localY := wy - float64(ty*sim.TileSize)
	onBeach := trueS == surfGrass || trueS == surfSnow || trueS == surfTree || trueS == surfShallow || trueS == surfDeep
	graded := false
	var mk uint8
	var over bool
	if lk.beach && onBeach {
		painted, mask, replace, took := paintBeach(lk, surf, m, tx, ty, trueS, wx, wy, c, corners, vw, joins, round)
		if took {
			c = painted
			mk = mask
			over = replace
			graded = true
		}
	} else if trueS == surfGrass || trueS == surfSnow {
		if d, ok := nearWater(surf, m.W, m.H, tx, ty, wx, wy); ok && d < sim.TileSize {
			t := 1 - d/sim.TileSize
			if lk.shore != nil && lk.shoreMix > 0 {
				c = mixRGBA(c, lk.shore.at(wx, wy), t*lk.shoreMix)
			}
			c = beadFoam(c, lk.pal, wx, wy, d)
		}
	}
	if !graded && (trueS == surfShallow || trueS == surfDeep) {
		c = gradeWater(lk, c, surf, m.W, m.H, tx, ty, trueS, wx, wy)
	}
	// Snow cliffs wave in cliff.go. These bands are the cell edge, so they
	// would rule a straight line back across that wave.
	if snowCliff == nil && trueS == surfCliff {
		if south, ok := neighbor(surf, m.W, m.H, tx, ty, 0, 1); ok && south != surfCliff && south != surfRamp && localY > 11 {
			c = mixRGBA(c, color.RGBA{R: 42, G: 38, B: 34, A: 255}, 0.5)
		}
		if east, ok := neighbor(surf, m.W, m.H, tx, ty, 1, 0); ok && east != surfCliff && east != surfRamp && localX > 12 {
			c = mixRGBA(c, color.RGBA{R: 42, G: 38, B: 34, A: 255}, 0.4)
		}
	}
	if snowCliff == nil && trueS != surfCliff && trueS != surfRamp {
		if north, ok := neighbor(surf, m.W, m.H, tx, ty, 0, -1); ok && north == surfCliff && localY < 4 {
			c = mixRGBA(c, color.RGBA{R: 20, G: 24, B: 28, A: 255}, 0.4*(1-localY/4))
		}
	}
	return c, mk, over
}

// paintBeach lays the shore, then water that pales at the line and deepens
// offshore. took is true when this pixel's colour is finished, so the old
// tile-edge grade does not run again.
func paintBeach(lk look, surf []surface, m sim.Map, tx, ty int, trueS surface, wx, wy float64, c color.RGBA, corners []shoreCorner, vw int, joins []*shoreJoin, round bool) (color.RGBA, uint8, bool, bool) {
	onWater := trueS == surfShallow || trueS == surfDeep
	sd, _, ok := beachSD(surf, m.W, m.H, tx, ty, wx, wy, onWater, corners, vw, round)
	if !ok {
		if onWater {
			c = blendWaterDepth(lk, c, surf, m.W, m.H, tx, ty, trueS, wx, wy, valueNoise(wx, wy, 18, 31))
			return c, 0, false, true
		}
		return c, 0, false, false
	}
	if sd > 0 {
		if onWater {
			return wetSand(lk, wx, wy, sd), maskNone, true, true
		}
		if trueS != surfGrass && trueS != surfSnow && trueS != surfTree {
			return c, 0, false, false
		}
		reach := sandReach(wx, wy)
		if sd >= reach {
			return c, 0, false, false
		}
		return featherLand(lk, c, wx, wy, sd, reach), maskNone, true, true
	}
	into := -sd
	fade := into
	if onWater {
		if j := shoreAt(joins, m.W, tx, ty); j != nil && into < beachFadeReach+beachJoin*0.25 {
			lx := int(wx) - tx*sim.TileSize
			ly := int(wy) - ty*sim.TileSize
			if lx < 0 {
				lx = 0
			}
			if ly < 0 {
				ly = 0
			}
			if lx >= sim.TileSize {
				lx = sim.TileSize - 1
			}
			if ly >= sim.TileSize {
				ly = sim.TileSize - 1
			}
			d2 := float64(j.far[ly*sim.TileSize+lx])
			if d2 < into+beachJoin {
				fade = smoothMin(into, d2, beachJoin)
			}
		}
	}
	return waterColor(lk, wx, wy, into, fade), waterMask(into), true, true
}

// sandReach is the inland width of the sand at this point. Noise keeps the
// grass line from being a ruled edge. The wet side stays put.
func sandReach(wx, wy float64) float64 {
	n := valueNoise(wx, wy, 14, 23)
	reach := beachSand + (n-0.5)*beachSandWander
	if reach < 2.2 {
		reach = 2.2
	}
	return reach
}

// landMasses numbers each body of land. Water stays -1. Tiles that only
// touch at a corner stay separate, so the water in that gap still blends.
func landMasses(surf []surface, w, h int) []int {
	ids := make([]int, w*h)
	for i := range ids {
		ids[i] = -1
	}
	next := 0
	stack := make([]int, 0, 64)
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			start := y*w + x
			if ids[start] >= 0 || !landBesideWater(surf[start]) {
				continue
			}
			ids[start] = next
			stack = append(stack[:0], start)
			for len(stack) > 0 {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				nx, ny := n%w, n/w
				for _, d := range dirs {
					xx, yy := nx+d[0], ny+d[1]
					if xx < 0 || yy < 0 || xx >= w || yy >= h {
						continue
					}
					j := yy*w + xx
					if ids[j] >= 0 || !landBesideWater(surf[j]) {
						continue
					}
					ids[j] = next
					stack = append(stack, j)
				}
			}
			next++
		}
	}
	return ids
}

// shoreCorner is one grid vertex. kind 1 rounds a water corner, kind 2 a land corner.
// dx and dy point from the vertex into the single odd cell.
type shoreCorner struct {
	kind uint8
	dx   int8
	dy   int8
}

func shoreCorners(surf []surface, w, h int) ([]shoreCorner, int) {
	vw := w + 1
	out := make([]shoreCorner, vw*(h+1))
	for j := 0; j <= h; j++ {
		for i := 0; i <= w; i++ {
			out[j*vw+i] = classifyVertex(surf, w, h, i, j)
		}
	}
	return out, vw
}

func classifyVertex(surf []surface, w, h, i, j int) shoreCorner {
	spots := [4]struct {
		cx, cy int
		dx, dy int8
	}{
		{i - 1, j - 1, -1, -1},
		{i, j - 1, 1, -1},
		{i - 1, j, -1, 1},
		{i, j, 1, 1},
	}
	nWater := 0
	var dry, wet shoreCorner
	for _, s := range spots {
		water, inside := cellWater(surf, w, h, s.cx, s.cy)
		if !inside {
			return shoreCorner{}
		}
		if water {
			nWater++
			wet = shoreCorner{kind: 1, dx: s.dx, dy: s.dy}
		} else {
			dry = shoreCorner{kind: 2, dx: s.dx, dy: s.dy}
		}
	}
	switch nWater {
	case 1:
		return wet
	case 3:
		return dry
	default:
		return shoreCorner{}
	}
}

func cellRounded(corners []shoreCorner, vw, tx, ty int) bool {
	if len(corners) == 0 {
		return false
	}
	for oy := 0; oy <= 1; oy++ {
		for ox := 0; ox <= 1; ox++ {
			if corners[(ty+oy)*vw+tx+ox].kind != 0 {
				return true
			}
		}
	}
	return false
}

// beachSD is the signed distance to the rounded shoreline. Positive is land,
// negative is water. dLand is the straight distance into the water, used for
// the bank shade past the shallow band. Corners use a quarter-circle of
// beachCorner instead of the tile's right angle.
func beachSD(surf []surface, w, h, tx, ty int, wx, wy float64, onWater bool, corners []shoreCorner, vw int, round bool) (float64, float64, bool) {
	sd, ok := shoreField(wx, wy)
	if !ok {
		// No shore field: fall back to the neighbouring cells.
		if onWater {
			d, hit := distToMatch(surf, w, h, tx, ty, wx, wy, landBesideWater)
			if !hit {
				return 0, 0, false
			}
			return -d, d, true
		}
		d, hit := nearWater(surf, w, h, tx, ty, wx, wy)
		if !hit {
			return 0, 0, false
		}
		if d < 0 {
			d = 0
		}
		return d, 0, true
	}
	// The distance is already blurred, so a tile corner is a slope.
	// A fillet here would pull that slope back onto the cell corner.
	sd += shoreFlow(wx, wy)
	dLand := 0.0
	if sd < 0 {
		dLand = -sd
	}
	return sd, dLand, true
}

func roundShore(corners []shoreCorner, vw, tx, ty int, wx, wy, sd float64) float64 {
	for oy := 0; oy <= 1; oy++ {
		for ox := 0; ox <= 1; ox++ {
			c := corners[(ty+oy)*vw+tx+ox]
			if c.kind == 0 {
				continue
			}
			vx := float64((tx + ox) * sim.TileSize)
			vy := float64((ty + oy) * sim.TileSize)
			if c.kind == 1 {
				sd = filletWater(vx, vy, int(c.dx), int(c.dy), wx, wy, sd)
			} else {
				sd = filletLand(vx, vy, int(c.dx), int(c.dy), wx, wy, sd)
			}
		}
	}
	return sd
}

func cellWater(surf []surface, w, h, x, y int) (water, inside bool) {
	if x < 0 || y < 0 || x >= w || y >= h {
		return false, false
	}
	// A bridge is water for the shoreline, so the bank runs straight under
	// the deck instead of turning a corner at each end.
	s := surf[y*w+x]
	return s == surfShallow || s == surfDeep || s == surfBridge, true
}

// filletWater rounds a corner where water sticks into the land.
// arc is positive on the new land side of that curve.
func filletWater(vx, vy float64, dx, dy int, wx, wy, sd float64) float64 {
	lx := (wx - vx) * float64(dx)
	ly := (wy - vy) * float64(dy)
	if lx < 0 || ly < 0 || lx > beachCorner || ly > beachCorner {
		return sd
	}
	ccx := vx + float64(dx)*beachCorner
	ccy := vy + float64(dy)*beachCorner
	arc := math.Hypot(wx-ccx, wy-ccy) - beachCorner
	if arc > sd {
		return arc
	}
	return sd
}

// filletLand rounds a corner where land sticks into the water.
// arc is positive deeper into the land, negative on the bitten-off tip.
func filletLand(vx, vy float64, dx, dy int, wx, wy, sd float64) float64 {
	lx := (wx - vx) * float64(dx)
	ly := (wy - vy) * float64(dy)
	if lx < -beachFadeReach || ly < -beachFadeReach || lx > beachCorner || ly > beachCorner {
		return sd
	}
	ccx := vx + float64(dx)*beachCorner
	ccy := vy + float64(dy)*beachCorner
	arc := beachCorner - math.Hypot(wx-ccx, wy-ccy)
	if arc < sd {
		// This corner's curve is only a few pixels off the sharp tile.
		// A bigger drop means another shore is closer, and that bank wins.
		if sd-arc > 6 {
			return sd
		}
		return arc
	}
	return sd
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
func gradeWater(lk look, c color.RGBA, surf []surface, w, h, tx, ty int, self surface, wx, wy float64) color.RGBA {
	c = blendWaterDepth(lk, c, surf, w, h, tx, ty, self, wx, wy, valueNoise(wx, wy, 18, 31))
	if d, ok := distToMatch(surf, w, h, tx, ty, wx, wy, landBesideWater); ok && d < 16 {
		c = mixRGBA(c, lk.bank, (1-d/16)*lk.bankMix)
	}
	return c
}

func blendWaterDepth(lk look, c color.RGBA, surf []surface, w, h, tx, ty int, self surface, wx, wy, n float64) color.RGBA {
	other := surfDeep
	if self == surfDeep {
		other = surfShallow
	}
	d, ok := distToMatch(surf, w, h, tx, ty, wx, wy, func(s surface) bool { return s == other })
	if !ok {
		return c
	}
	// n is shared by both sides of a tile edge, so the colours meet.
	d += (n - 0.5) * 6
	if d < 0 {
		d = 0
	}
	if d >= sim.TileSize {
		return c
	}
	oc, _ := lk.plateColor(other, wx, wy)
	t := (1 - smooth01(d/float64(sim.TileSize))) * 0.5
	return mixRGBA(c, oc, t)
}

// shoreJoin lists land rectangles, grouped by landmass, that can blend
// into one cell. A nil join means the cell sees a single shore.
// Rectangles are tile coordinates, with x1 and y1 exclusive.
// far is the blended distance to every landmass except the nearest,
// at each world pixel of the cell, so the per-pixel paint is a lookup.
type shoreJoin struct {
	n              int
	off            [17]int
	x0, y0, x1, y1 [49]int16
	far            [sim.TileSize * sim.TileSize]float32
}

// shoreJoins records, once per cell, every landmass inside the blend
// reach. Cells with one shore stay nil so a straight coast pays nothing
// per pixel. The window is wide enough that a landmass just outside it
// cannot move the colour, so the join stays continuous across a tile edge.
func shoreJoins(surf []surface, landID []int, w, h int) []*shoreJoin {
	out := make([]*shoreJoin, w*h)
	const rad = 3
	for ty := 0; ty < h; ty++ {
		for tx := 0; tx < w; tx++ {
			s := surf[ty*w+tx]
			if s != surfShallow && s != surfDeep {
				continue
			}
			var id [16]int
			var which [49]int
			var tiles [49][2]int16
			nMass, nTile := 0, 0
			for oy := -rad; oy <= rad; oy++ {
				for ox := -rad; ox <= rad; ox++ {
					xx, yy := tx+ox, ty+oy
					if xx < 0 || yy < 0 || xx >= w || yy >= h {
						continue
					}
					mass := landID[yy*w+xx]
					if mass < 0 {
						continue
					}
					slot := -1
					for i := 0; i < nMass; i++ {
						if id[i] == mass {
							slot = i
							break
						}
					}
					if slot < 0 {
						if nMass == len(id) {
							continue
						}
						slot = nMass
						id[nMass] = mass
						nMass++
					}
					tiles[nTile] = [2]int16{int16(xx), int16(yy)}
					which[nTile] = slot
					nTile++
				}
			}
			if nMass < 2 {
				continue
			}
			j := &shoreJoin{n: nMass}
			off := 0
			for m := 0; m < nMass; m++ {
				j.off[m] = off
				off = packShoreMass(&tiles, &which, nTile, m, j, off)
			}
			j.off[nMass] = off
			fillShoreFar(j, tx, ty)
			out[ty*w+tx] = j
		}
	}
	return out
}

// packShoreMass appends the solid rectangles of one landmass. Adjacent
// tiles become one rectangle so a straight bank is a single distance.
func packShoreMass(tiles *[49][2]int16, which *[49]int, nTile, slot int, j *shoreJoin, off int) int {
	var xs, ys [49]int16
	n := 0
	for k := 0; k < nTile; k++ {
		if which[k] != slot {
			continue
		}
		xs[n] = tiles[k][0]
		ys[n] = tiles[k][1]
		n++
	}
	for i := 1; i < n; i++ {
		x, y := xs[i], ys[i]
		k := i
		for k > 0 && (ys[k-1] > y || (ys[k-1] == y && xs[k-1] > x)) {
			xs[k] = xs[k-1]
			ys[k] = ys[k-1]
			k--
		}
		xs[k], ys[k] = x, y
	}
	type run struct{ x0, x1, y int16 }
	var runs [49]run
	nr := 0
	for i := 0; i < n; {
		x0, y := xs[i], ys[i]
		x1 := x0 + 1
		i++
		for i < n && ys[i] == y && xs[i] == x1 {
			x1++
			i++
		}
		runs[nr] = run{x0, x1, y}
		nr++
	}
	var used [49]bool
	for a := 0; a < nr; a++ {
		if used[a] {
			continue
		}
		y0 := runs[a].y
		y1 := y0 + 1
		for {
			found := -1
			for b := a + 1; b < nr; b++ {
				if used[b] || runs[b].y != y1 || runs[b].x0 != runs[a].x0 || runs[b].x1 != runs[a].x1 {
					continue
				}
				found = b
				break
			}
			if found < 0 {
				break
			}
			used[found] = true
			y1++
		}
		j.x0[off] = runs[a].x0
		j.x1[off] = runs[a].x1
		j.y0[off] = y0
		j.y1[off] = y1
		off++
	}
	return off
}

// fillShoreFar stores, for each world pixel, how close the other shores are.
func fillShoreFar(j *shoreJoin, tx, ty int) {
	var dist [16]float64
	for ly := 0; ly < sim.TileSize; ly++ {
		wy := float64(ty*sim.TileSize+ly) + 0.5
		for lx := 0; lx < sim.TileSize; lx++ {
			wx := float64(tx*sim.TileSize+lx) + 0.5
			for i := 0; i < j.n; i++ {
				dist[i] = math.MaxFloat64
				for k := j.off[i]; k < j.off[i+1]; k++ {
					d := distToRect(wx, wy,
						float64(int(j.x0[k])*sim.TileSize),
						float64(int(j.y0[k])*sim.TileSize),
						float64(int(j.x1[k])*sim.TileSize),
						float64(int(j.y1[k])*sim.TileSize))
					if d < 0 {
						d = 0
					}
					if d < dist[i] {
						dist[i] = d
					}
				}
			}
			first := 0
			for i := 1; i < j.n; i++ {
				if dist[i] < dist[first] {
					first = i
				}
			}
			acc := math.MaxFloat64
			for i := 0; i < j.n; i++ {
				if i == first {
					continue
				}
				acc = smoothMin(acc, dist[i], beachJoin)
			}
			j.far[ly*sim.TileSize+lx] = float32(acc)
		}
	}
}

func shoreAt(joins []*shoreJoin, w, tx, ty int) *shoreJoin {
	if ty*w+tx >= len(joins) {
		return nil
	}
	return joins[ty*w+tx]
}

func smoothMin(a, b, k float64) float64 {
	h := k - math.Abs(a-b)
	if h <= 0 {
		if a < b {
			return a
		}
		return b
	}
	h /= k
	if b < a {
		a = b
	}
	return a - h*h*k*0.25
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
			if s != surfShallow && s != surfDeep && s != surfBridge {
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

// A bridge deck sits a little past the walkable cells: a lip over the water
// on the long sides, and a short landing on the bank at each end.
const (
	bridgeLip   = 1.35
	bridgeEnd   = 2.4
	bridgeRail  = 1.75
	bridgePlank = 4.6
	bridgeShade = 6.4
)

// stampBridges paints one timber span per bridge cell. Boards run across the
// direction of travel, the outer edges are rails, and a soft shadow falls on
// the water. The shimmer is cleared only where the deck covers the picture.
func stampBridges(dst, anim, mask *image.RGBA, m sim.Map, surf []surface, lk look) {
	w, h := m.W, m.H
	cover := make([]bool, w*h)
	anyBridge := false
	for ty := 0; ty < h; ty++ {
		for tx := 0; tx < w; tx++ {
			if surf[ty*w+tx] != surfBridge {
				continue
			}
			anyBridge = true
			for oy := -1; oy <= 1; oy++ {
				for ox := -1; ox <= 1; ox++ {
					xx, yy := tx+ox, ty+oy
					if xx < 0 || yy < 0 || xx >= w || yy >= h {
						continue
					}
					cover[yy*w+xx] = true
				}
			}
		}
	}
	if !anyBridge {
		return
	}
	axes := bridgeAxes(surf, w, h)
	pw := dst.Bounds().Dx()
	ph := dst.Bounds().Dy()
	span := sim.TileSize * terrainPx
	for ty := 0; ty < h; ty++ {
		for tx := 0; tx < w; tx++ {
			if !cover[ty*w+tx] {
				continue
			}
			x0 := tx * span
			y0 := ty * span
			for y := 0; y < span; y++ {
				py := y0 + y
				if py < 0 || py >= ph {
					continue
				}
				wy := (float64(py) + 0.5) / terrainPx
				for x := 0; x < span; x++ {
					px := x0 + x
					if px < 0 || px >= pw {
						continue
					}
					wx := (float64(px) + 0.5) / terrainPx
					applyBridgePixel(dst, anim, mask, surf, axes, w, h, lk, px, py, wx, wy)
				}
			}
		}
	}
}

func applyBridgePixel(dst, anim, mask *image.RGBA, surf []surface, axes []bool, w, h int, lk look, px, py int, wx, wy float64) {
	deck, rail, sill, post, along, across := bridgeDeck(surf, axes, w, h, wx, wy)
	i := dst.PixOffset(px, py)
	base := color.RGBA{R: dst.Pix[i], G: dst.Pix[i+1], B: dst.Pix[i+2], A: 255}
	ai := anim.PixOffset(px, py)
	if deck > 0.004 {
		wood := bridgeWood(lk, along, across, rail, sill, post)
		cover := smooth01(clamp01(deck))
		// Still ground (the bank) has no shimmer to fade, so the timber is
		// painted onto it. Water keeps its picture and lets the deck cover it.
		if anim.Pix[ai+3] == 0 {
			setRGBA(dst, px, py, mixRGBA(base, wood, cover))
			return
		}
		setRGBA(dst, px, py, wood)
		a := uint8((1 - cover) * 255)
		if a < 12 {
			a = 0
		}
		anim.Pix[ai+3] = a
		if a == 0 {
			mask.Pix[ai] = 0
			mask.Pix[ai+1] = 0
			mask.Pix[ai+2] = 0
			mask.Pix[ai+3] = 0
		}
		return
	}
	sh := bridgeShadow(surf, axes, w, h, wx, wy)
	if sh < 0.02 {
		return
	}
	shade := color.RGBA{R: 16, G: 28, B: 26, A: 255}
	setRGBA(dst, px, py, mixRGBA(base, shade, sh))
	if anim.Pix[ai+3] == 0 {
		return
	}
	ac := color.RGBA{R: anim.Pix[ai], G: anim.Pix[ai+1], B: anim.Pix[ai+2], A: anim.Pix[ai+3]}
	ac = mixRGBA(ac, shade, sh)
	anim.Pix[ai] = ac.R
	anim.Pix[ai+1] = ac.G
	anim.Pix[ai+2] = ac.B
}

func bridgeAxes(surf []surface, w, h int) []bool {
	vert := make([]bool, w*h)
	for ty := 0; ty < h; ty++ {
		for tx := 0; tx < w; tx++ {
			if surf[ty*w+tx] == surfBridge {
				vert[ty*w+tx] = bridgeVertical(surf, w, h, tx, ty)
			}
		}
	}
	return vert
}

func bridgeDeck(surf []surface, axes []bool, w, h int, wx, wy float64) (deck, rail, sill, post, along, across float64) {
	tx := int(math.Floor(wx / sim.TileSize))
	ty := int(math.Floor(wy / sim.TileSize))
	best := 0.0
	for oy := -1; oy <= 1; oy++ {
		for ox := -1; ox <= 1; ox++ {
			cx, cy := tx+ox, ty+oy
			if cx < 0 || cy < 0 || cx >= w || cy >= h || surf[cy*w+cx] != surfBridge {
				continue
			}
			d, r, s, p, al, ac := deckFromCell(surf, axes, w, h, cx, cy, wx, wy)
			if d > best {
				best = d
				deck, rail, sill, post, along, across = d, r, s, p, al, ac
			}
		}
	}
	return deck, rail, sill, post, along, across
}

func deckFromCell(surf []surface, axes []bool, w, h, cx, cy int, wx, wy float64) (deck, rail, sill, post, along, across float64) {
	vertical := axes[cy*w+cx]
	left := float64(cx * sim.TileSize)
	top := float64(cy * sim.TileSize)
	right := left + sim.TileSize
	bot := top + sim.TileSize
	westB := bridgeCell(surf, w, h, cx-1, cy)
	eastB := bridgeCell(surf, w, h, cx+1, cy)
	northB := bridgeCell(surf, w, h, cx, cy-1)
	southB := bridgeCell(surf, w, h, cx, cy+1)

	x0, x1 := left, right
	y0, y1 := top, bot
	if vertical {
		if !westB {
			x0 -= bridgeLip
		}
		if !eastB {
			x1 += bridgeLip
		}
		if !northB {
			y0 -= bridgeEnd
		}
		if !southB {
			y1 += bridgeEnd
		}
	} else {
		if !northB {
			y0 -= bridgeLip
		}
		if !southB {
			y1 += bridgeLip
		}
		if !westB {
			x0 -= bridgeEnd
		}
		if !eastB {
			x1 += bridgeEnd
		}
	}
	if wx < x0 || wx > x1 || wy < y0 || wy > y1 {
		return 0, 0, 0, 0, 0, 0
	}

	inward := float64(sim.TileSize)
	endIn := float64(sim.TileSize)
	if vertical {
		if !westB {
			inward = math.Min(inward, wx-left)
		}
		if !eastB {
			inward = math.Min(inward, right-wx)
		}
		if !northB {
			endIn = math.Min(endIn, wy-top)
		}
		if !southB {
			endIn = math.Min(endIn, bot-wy)
		}
		along, across = wy, wx
	} else {
		if !northB {
			inward = math.Min(inward, wy-top)
		}
		if !southB {
			inward = math.Min(inward, bot-wy)
		}
		if !westB {
			endIn = math.Min(endIn, wx-left)
		}
		if !eastB {
			endIn = math.Min(endIn, right-wx)
		}
		along, across = wx, wy
	}
	deck = bridgeCover(inward, endIn)
	if deck <= 0 {
		return 0, 0, 0, 0, along, across
	}
	if inward < bridgeRail {
		rail = smooth01(clamp01(1 - inward/bridgeRail))
	}
	if endIn < 1.9 {
		sill = smooth01(clamp01(1 - endIn/1.9))
	}
	if inward < 1.2 && inward > -0.45 && endIn > 3 {
		var mid float64
		if vertical {
			mid = top + sim.TileSize*0.5
			post = bridgePost(inward, wy-mid)
		} else {
			mid = left + sim.TileSize*0.5
			post = bridgePost(inward, wx-mid)
		}
	}
	return deck, rail, sill, post, along, across
}

// bridgeCover is solid on the walkable cell. Only the lip past the cell fades.
func bridgeCover(inward, endIn float64) float64 {
	side := 1.0
	end := 1.0
	if inward < 0 {
		side = clamp01(1 + inward/bridgeLip)
	}
	if endIn < 0 {
		end = clamp01(1 + endIn/bridgeEnd)
	}
	if side < end {
		return side
	}
	return end
}

func bridgePost(inward, alongDelta float64) float64 {
	pd := math.Hypot((inward-0.3)/1.05, alongDelta/1.35)
	if pd >= 1 {
		return 0
	}
	return smooth01(1 - pd)
}

func bridgeCell(surf []surface, w, h, tx, ty int) bool {
	if tx < 0 || ty < 0 || tx >= w || ty >= h {
		return false
	}
	return surf[ty*w+tx] == surfBridge
}

func bridgeShadow(surf []surface, axes []bool, w, h int, wx, wy float64) float64 {
	tx := int(math.Floor(wx / sim.TileSize))
	ty := int(math.Floor(wy / sim.TileSize))
	best := 0.0
	for oy := -1; oy <= 1; oy++ {
		for ox := -1; ox <= 1; ox++ {
			cx, cy := tx+ox, ty+oy
			if cx < 0 || cy < 0 || cx >= w || cy >= h || surf[cy*w+cx] != surfBridge {
				continue
			}
			x0, y0, x1, y1 := bridgeRect(surf, axes, w, h, cx, cy)
			d, dx, dy := distOutside(wx, wy, x0, y0, x1, y1)
			if d < 0 || d > bridgeShade {
				continue
			}
			fall := smooth01(1 - d/bridgeShade)
			light := dx*0.35 + dy*0.95
			side := 0.14
			if light > 0 {
				side = 0.32 + 0.4*clamp01(light/3)
			}
			if sh := fall * side; sh > best {
				best = sh
			}
		}
	}
	return best
}

func bridgeRect(surf []surface, axes []bool, w, h, cx, cy int) (x0, y0, x1, y1 float64) {
	vertical := axes[cy*w+cx]
	left := float64(cx * sim.TileSize)
	top := float64(cy * sim.TileSize)
	right := left + sim.TileSize
	bot := top + sim.TileSize
	x0, x1, y0, y1 = left, right, top, bot
	if vertical {
		if !bridgeCell(surf, w, h, cx-1, cy) {
			x0 -= bridgeLip
		}
		if !bridgeCell(surf, w, h, cx+1, cy) {
			x1 += bridgeLip
		}
		if !bridgeCell(surf, w, h, cx, cy-1) {
			y0 -= bridgeEnd
		}
		if !bridgeCell(surf, w, h, cx, cy+1) {
			y1 += bridgeEnd
		}
		return x0, y0, x1, y1
	}
	if !bridgeCell(surf, w, h, cx, cy-1) {
		y0 -= bridgeLip
	}
	if !bridgeCell(surf, w, h, cx, cy+1) {
		y1 += bridgeLip
	}
	if !bridgeCell(surf, w, h, cx-1, cy) {
		x0 -= bridgeEnd
	}
	if !bridgeCell(surf, w, h, cx+1, cy) {
		x1 += bridgeEnd
	}
	return x0, y0, x1, y1
}

// distOutside is the distance past a rect and the direction away from it.
// Inside the rect the distance is negative.
func distOutside(px, py, x0, y0, x1, y1 float64) (d, dx, dy float64) {
	cx := math.Min(math.Max(px, x0), x1)
	cy := math.Min(math.Max(py, y0), y1)
	dx, dy = px-cx, py-cy
	if px >= x0 && px <= x1 && py >= y0 && py <= y1 {
		return -1, 0, 0
	}
	return math.Hypot(dx, dy), dx, dy
}

func bridgeWood(lk look, along, across, rail, sill, post float64) color.RGBA {
	seamAlong := along + (valueNoise(across, along, 7, 11)-0.5)*1.15
	span := bridgePlank
	plank := int(math.Floor(seamAlong / span))
	u := seamAlong/span - float64(plank)
	grain := valueNoise(across*0.85, float64(plank)*span, 5.5, 19)
	streak := valueNoise(across*0.25+float64(plank)*3, along*1.4, 2.2, 29)
	tone := hash01(plank, 7, 13)
	t := clamp01(0.34 + 0.4*grain + 0.16*(streak-0.5) + 0.22*(tone-0.5))
	gap := lk.pal.PlankGap
	face := lk.pal.Plank
	hi := color.RGBA{
		R: clampByte(int(face.R) + 38),
		G: clampByte(int(face.G) + 28),
		B: clampByte(int(face.B) + 14),
		A: 255,
	}
	c := mixRGBA(gap, face, t)
	crown := smooth01(clamp01((u-0.08)/0.22)) * (1 - smooth01(clamp01((u-0.62)/0.28)))
	c = mixRGBA(c, hi, crown*0.62)
	seamU := math.Min(u/0.13, (1-u)/0.07)
	seam := 1 - smooth01(clamp01(seamU))
	dark := color.RGBA{R: gap.R * 2 / 3, G: gap.G * 2 / 3, B: gap.B * 2 / 3, A: 255}
	c = mixRGBA(c, dark, seam*0.9)
	if lk.snowCaps {
		dust := crown * (0.58 + 0.22*grain) * (1 - seam*0.8) * (1 - rail*0.65)
		c = mixRGBA(c, color.RGBA{R: 234, G: 240, B: 244, A: 255}, dust)
	}
	if sill > 0 {
		c = mixRGBA(c, gap, sill*0.7)
	}
	if rail > 0 {
		curb := mixRGBA(gap, color.RGBA{R: 46, G: 30, B: 16, A: 255}, 0.5)
		if lk.snowCaps {
			curb = mixRGBA(curb, color.RGBA{R: 206, G: 214, B: 220, A: 255}, 0.4)
		}
		c = mixRGBA(c, curb, smooth01(rail))
		if rail > 0.22 && rail < 0.72 {
			catch := 1 - math.Abs(rail-0.45)/0.28
			c = mixRGBA(c, hi, 0.32*clamp01(catch))
		}
	}
	if post > 0 {
		bolt := color.RGBA{R: 40, G: 28, B: 16, A: 255}
		if lk.snowCaps {
			bolt = mixRGBA(bolt, color.RGBA{R: 238, G: 244, B: 248, A: 255}, 0.62)
		}
		c = mixRGBA(c, bolt, smooth01(post))
	}
	return c
}

// bridgeVertical is true when the span runs north-south, so the boards run
// east-west. A tie follows the longer run. A lone cell stands north-south.
func bridgeVertical(surf []surface, w, h, tx, ty int) bool {
	v := bridgeReach(surf, w, h, tx, ty, 0, -1) + bridgeReach(surf, w, h, tx, ty, 0, 1)
	hz := bridgeReach(surf, w, h, tx, ty, -1, 0) + bridgeReach(surf, w, h, tx, ty, 1, 0)
	if v == 0 && hz == 0 {
		return true
	}
	return v >= hz
}

func bridgeReach(surf []surface, w, h, tx, ty, ox, oy int) int {
	n := 0
	for n < 64 {
		s, ok := neighbor(surf, w, h, tx, ty, ox*(n+1), oy*(n+1))
		if !ok || s != surfBridge {
			return n
		}
		n++
	}
	return n
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

func paintCanopies(m sim.Map, surf []surface, ids []int, lk look, noise *noiseGrid, ground *image.RGBA) []treeCanopy {
	seen := map[int]bool{}
	var out []treeCanopy
	for i, s := range surf {
		if s != surfTree || seen[ids[i]] {
			continue
		}
		seen[ids[i]] = true
		out = append(out, paintOneCanopy(m, surf, ids, ids[i], lk, noise, ground))
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

func paintOneCanopy(m sim.Map, surf []surface, ids []int, id int, lk look, noise *noiseGrid, ground *image.RGBA) treeCanopy {
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
			c := lk.canopy.atScale(wx, wy, 4)
			c = shadeCrown(c, d, wx-float64(ccx[row+px]), wy-float64(ccy[row+px]), lk, sampleNoise(noise, wx*1.7, wy*1.7))
			c = mergeFringe(c, ground, gw, gh, wx, wy, a, lk)
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

func shadeCrown(c color.RGBA, d, dx, dy float64, lk look, n float64) color.RGBA {
	if lk.snowCaps && n > 0.58 && d < -2 {
		c = mixRGBA(c, color.RGBA{R: 232, G: 242, B: 246, A: 255}, 0.28)
	}
	// A little south-east shade only. A dark stroke around the whole crown
	// reads as a cut-out, which is the opposite of foliage meeting the ground.
	se := clamp01((dx + dy) / 36)
	return mixRGBA(c, lk.crown, se*lk.crownMix)
}

// mergeFringe pulls the transparent tips toward the ground already painted
// there, and lifts the plate's dark gaps so the rim is leaves rather than
// a black line.
func mergeFringe(c color.RGBA, ground *image.RGBA, gw, gh int, wx, wy, a float64, lk look) color.RGBA {
	if ground == nil {
		return c
	}
	// Only the blackest gaps in the plate, and only near the rim. A broad
	// recolour here draws a halo around the whole stand.
	sum := int(c.R) + int(c.G) + int(c.B)
	if a < 0.85 && sum < 96 {
		c = mixRGBA(c, lk.leaf, 0.45)
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
func stampTreeShade(dst *image.RGBA, m sim.Map, surf []surface, lk look) {
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
	for i, v := range strength {
		if v == 0 {
			continue
		}
		o := i * 4
		c := color.RGBA{R: dst.Pix[o], G: dst.Pix[o+1], B: dst.Pix[o+2], A: dst.Pix[o+3]}
		c = mixRGBA(c, lk.shadow, float64(v)/255)
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
	if dx == 0 {
		return dy
	}
	if dy == 0 {
		return dx
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

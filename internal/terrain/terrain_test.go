package terrain

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

func mapOf(w, h int, tiles []sim.Tile) sim.Map {
	if tiles == nil {
		tiles = make([]sim.Tile, w*h)
	}
	return sim.Map{W: w, H: h, Tiles: tiles}
}

func regionIDs(t *testing.T, paint terrainPaint, want surface) map[int]bool {
	t.Helper()
	seen := map[int]bool{}
	for i, s := range paint.surf {
		if s == want {
			seen[paint.ids[i]] = true
		}
	}
	return seen
}

func TestPaintIsStable(t *testing.T) {
	m := mapOf(4, 4, nil)
	m.Tiles[5] = sim.TileWaterShallow
	a := paintMap(m, "grass")
	b := paintMap(m, "grass")
	if !bytes.Equal(a.ground.Pix, b.ground.Pix) {
		t.Fatal("the same map painted two different grounds")
	}
}

func TestGrassBlockIsOneRegion(t *testing.T) {
	paint := paintMap(mapOf(4, 4, nil), "grass")
	if len(regionIDs(t, paint, surfGrass)) != 1 {
		t.Fatalf("grass regions = %d, want 1", len(regionIDs(t, paint, surfGrass)))
	}
}

func TestCornerTouchJoins(t *testing.T) {
	tiles := []sim.Tile{
		sim.TileGrass, sim.TileWaterShallow,
		sim.TileWaterShallow, sim.TileGrass,
	}
	paint := paintMap(mapOf(2, 2, tiles), "grass")
	if len(regionIDs(t, paint, surfGrass)) != 1 {
		t.Fatalf("diagonal grass regions = %d, want 1", len(regionIDs(t, paint, surfGrass)))
	}
	if len(regionIDs(t, paint, surfShallow)) != 1 {
		t.Fatalf("diagonal water regions = %d, want 1", len(regionIDs(t, paint, surfShallow)))
	}
}

func TestWaterSplitsGrass(t *testing.T) {
	tiles := make([]sim.Tile, 12)
	for y := 0; y < 4; y++ {
		tiles[y*3+1] = sim.TileWaterShallow
	}
	paint := paintMap(mapOf(3, 4, tiles), "grass")
	if len(regionIDs(t, paint, surfGrass)) != 2 {
		t.Fatalf("grass regions = %d, want 2", len(regionIDs(t, paint, surfGrass)))
	}
	if len(regionIDs(t, paint, surfShallow)) != 1 {
		t.Fatalf("water regions = %d, want 1", len(regionIDs(t, paint, surfShallow)))
	}
}

func TestShoreDiffersFromInterior(t *testing.T) {
	tiles := make([]sim.Tile, 8*8)
	for x := 0; x < 8; x++ {
		tiles[x] = sim.TileWaterShallow
	}
	paint := paintMap(mapOf(8, 8, tiles), "grass")
	// The shore waves, so walk south from the water until the sand shows.
	x := 4*sim.TileSize*terrainPx + 8*terrainPx
	edge := paint.ground.RGBAAt(x, sim.TileSize*terrainPx)
	found := false
	for d := 1; d <= 24; d++ {
		c := paint.ground.RGBAAt(x, sim.TileSize*terrainPx+d*terrainPx)
		if c.R > c.G+8 {
			edge = c
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no sand south of the water, last %+v", edge)
	}
	// Far enough inland that the shore wave has given the ground back.
	past := paint.ground.RGBAAt(x, sim.TileSize*terrainPx+34*terrainPx)
	mr, mg, _ := meanCell(paint.ground, 4, 6)
	if edge.R <= edge.G {
		t.Fatalf("shore pixel = %+v, want the sand strip", edge)
	}
	if past.G <= past.R {
		t.Fatalf("pixel past the sand = %+v, want grass", past)
	}
	if mg <= mr {
		t.Fatalf("interior mean r=%.0f g=%.0f, want grass", mr, mg)
	}
}

func TestBeachIsSandThenShallow(t *testing.T) {
	const n = 10
	tiles := make([]sim.Tile, n*n)
	for y := 3; y <= 6; y++ {
		for x := 2; x <= 6; x++ {
			tiles[y*n+x] = sim.TileWaterDeep
		}
	}
	tiles[3*n+7] = sim.TileTree
	for _, name := range []string{"jungle", "grass", "forest"} {
		paint := paintMap(mapOf(n, n, tiles), name)
		lk := lookOf(name)
		sandAt := func(px, py int) color.RGBA { return paint.ground.RGBAAt(px, py) }
		// The shore waves. Walk inland from the water for the sand.
		spy := shorePx(4*sim.TileSize+8, 0)
		spx := shorePx(2*sim.TileSize, -2)
		sand := sandAt(spx, spy)
		foundSand := false
		for d := 1; d <= 22; d++ {
			px := shorePx(2*sim.TileSize, -d)
			c := sandAt(px, spy)
			if c.R > c.G+12 {
				spx = px
				sand = c
				foundSand = true
				break
			}
		}
		if !foundSand {
			t.Fatalf("%s found no sand on the west bank, %+v", name, sand)
		}
		gpx, gpy := shorePx(2*sim.TileSize, -24), spy
		grass := sandAt(gpx, gpy)
		gx, gy := shoreWorld(gpx, gpy)
		hpx, hpy := shorePx(2*sim.TileSize, 2), spy
		shallow := sandAt(hpx, hpy)
		hx, hy := shoreWorld(hpx, hpy)
		dpx, dpy := shorePx(4*sim.TileSize+8, 0), shorePx(5*sim.TileSize+8, 0)
		deep := sandAt(dpx, dpy)
		dx, dy := shoreWorld(dpx, dpy)
		tpy := shorePx(3*sim.TileSize+8, 0)
		tpx := shorePx(7*sim.TileSize, 2)
		treeSand := sandAt(tpx, tpy)
		foundTree := false
		for d := -16; d <= 18; d++ {
			px := shorePx(7*sim.TileSize, d)
			c := sandAt(px, tpy)
			if c.R > c.G {
				tpx = px
				treeSand = c
				foundTree = true
				break
			}
		}
		if !foundTree {
			t.Fatalf("%s tree bank has no sand", name)
		}
		tx, ty := shoreWorld(tpx, tpy)
		wantShallow, _ := lk.plateColor(surfShallow, hx, hy)
		wantDeep, _ := lk.plateColor(surfDeep, dx, dy)
		wantGrass, _ := lk.plateColor(surfGrass, gx, gy)
		sx, sy := shoreWorld(spx, spy)
		wantGround, _ := lk.plateColor(surfGrass, sx, sy)
		// The shore waves, so this pixel is only sometimes the dry sand.
		// It stays warm, and it stays off the meadow colour.
		if sand.R <= sand.G+8 || colorDist(sand, wantGround) < 20 {
			t.Fatalf("%s sand %+v, want sand rather than ground %+v", name, sand, wantGround)
		}
		if colorDist(grass, wantGrass) > 2 {
			t.Fatalf("%s inland %+v, want grass past the sand", name, grass)
		}
		if colorDist(shallow, wantShallow) > 8 {
			t.Fatalf("%s water edge %+v, want shallow %+v", name, shallow, wantShallow)
		}
		if colorDist(deep, wantDeep) > 8 {
			t.Fatalf("%s pond centre %+v, want deep %+v", name, deep, wantDeep)
		}
		if lum(shallow)+12 < lum(deep) {
			t.Fatalf("%s edge %+v is darker than the centre %+v", name, shallow, deep)
		}
		// The canopy shadow falls on the bank. The pixel stays warm sand,
		// not the meadow under the tree.
		treeGround, _ := lk.plateColor(surfGrass, tx, ty)
		if treeSand.R <= treeSand.G || colorDist(treeSand, treeGround) < 16 {
			t.Fatalf("%s tree bank %+v, want sand rather than %+v", name, treeSand, treeGround)
		}
		if paint.mask.RGBAAt(gpx, gpy).R != maskNone {
			t.Fatalf("%s inland mask = %d, want still ground", name, paint.mask.RGBAAt(gpx, gpy).R)
		}
		if paint.mask.RGBAAt(dpx, dpy).R != maskWater {
			t.Fatalf("%s pond mask = %d, want water", name, paint.mask.RGBAAt(dpx, dpy).R)
		}
	}
	snow := paintMap(mapOf(n, n, tiles), "snow")
	snowEdge := snow.ground.RGBAAt(shorePx(2*sim.TileSize, -2), shorePx(4*sim.TileSize+8, 0))
	sx, sy := shoreWorld(shorePx(2*sim.TileSize, -2), shorePx(4*sim.TileSize+8, 0))
	if colorDist(snowEdge, sandPlate.at(sx, sy)) < 40 {
		t.Fatalf("snow shore %+v matches the jungle sand", snowEdge)
	}
}

func TestBeachGrassLineWanders(t *testing.T) {
	const w, h = 16, 6
	tiles := make([]sim.Tile, w*h)
	for x := 0; x < w; x++ {
		tiles[2*w+x] = sim.TileWaterDeep
		tiles[3*w+x] = sim.TileWaterDeep
	}
	paint := paintMap(mapOf(w, h, tiles), "jungle")
	// North shore of the water sits at world y = 2 tiles.
	edge := 2 * sim.TileSize
	var minIn, maxIn float64
	minIn, maxIn = 99, 0
	for x := 2 * sim.TileSize; x < 14*sim.TileSize; x++ {
		in := 0.0
		for d := 0.5; d < 32; d += 0.25 {
			px := int(math.Round((float64(x)+0.5)*terrainPx - 0.5))
			py := int(math.Round((float64(edge)-d)*terrainPx - 0.5))
			c := paint.ground.RGBAAt(px, py)
			if c.R > c.G+10 {
				in = d
			}
		}
		if in < minIn {
			minIn = in
		}
		if in > maxIn {
			maxIn = in
		}
	}
	if maxIn-minIn < 3 {
		t.Fatalf("grass line spans %.2f to %.2f, want a flowing shore", minIn, maxIn)
	}
}

func TestBeachCornersRound(t *testing.T) {
	const n = 12
	tiles := make([]sim.Tile, n*n)
	for i := range tiles {
		tiles[i] = sim.TileWaterDeep
	}
	for y := 4; y <= 7; y++ {
		for x := 4; x <= 7; x++ {
			tiles[y*n+x] = sim.TileGrass
		}
	}
	paint := paintMap(mapOf(n, n, tiles), "jungle")
	// NW corner of the island.
	vx := 4 * sim.TileSize
	vy := 4 * sim.TileSize
	// Along the north edge, grass is only a sand-width in.
	edgeIn := firstGrassIn(paint.ground, vx+24, vy, 0, 1)
	// Down the diagonal the rounded corner pushes grass further from the vertex.
	diagIn := firstGrassIn(paint.ground, vx, vy, 1, 1)
	if diagIn*math.Sqrt2 < edgeIn+2 {
		t.Fatalf("diagonal grass at %.1f, edge grass at %.1f, want a rounded corner", diagIn, edgeIn)
	}
	tip := paint.ground.RGBAAt(shorePx(vx, 2), shorePx(vy, 2))
	if tip.G > tip.R+30 && tip.B < 80 {
		t.Fatalf("corner tip %+v still reads as grass", tip)
	}
}

func TestShallowFadesIntoDeep(t *testing.T) {
	const w, h = 20, 8
	tiles := make([]sim.Tile, w*h)
	for y := 3; y < h; y++ {
		for x := 0; x < w; x++ {
			tiles[y*w+x] = sim.TileWaterDeep
		}
	}
	paint := paintMap(mapOf(w, h, tiles), "jungle")
	edge := 3 * sim.TileSize
	at := func(x int, d float64) color.RGBA {
		px := int(math.Round(float64(x)*terrainPx - 0.5))
		py := int(math.Round((float64(edge)+d)*terrainPx - 0.5))
		return paint.ground.RGBAAt(px, py)
	}
	x := 10 * sim.TileSize
	// Past the sand blend the water itself lightens, then darkens offshore.
	near := at(x, 10)
	mid := at(x, 18)
	far := at(x, 26)
	deep := at(x, 40)
	if lum(near) <= lum(mid) || lum(mid) <= lum(far) || lum(far) <= lum(deep) {
		t.Fatalf("water lum = %d, %d, %d, %d, want a steady drop", lum(near), lum(mid), lum(far), lum(deep))
	}
	if lum(near)-lum(deep) < 24 {
		t.Fatalf("fade only dropped lum %d to %d", lum(near), lum(deep))
	}
	// No single world pixel along the fade may jump the whole way.
	var maxStep int
	prev := at(x, 8)
	for d := 8.5; d <= 40; d += 0.5 {
		c := at(x, d)
		step := lum(prev) - lum(c)
		if step < 0 {
			step = -step
		}
		if step > maxStep {
			maxStep = step
		}
		prev = c
	}
	if maxStep > 10 {
		t.Fatalf("shallow/deep step of %d, want a fade", maxStep)
	}
	// The halfway point moves along the shore.
	half := (lum(near) + lum(deep)) / 2
	minD, maxD := 99.0, 0.0
	for x := 4 * sim.TileSize; x < 16*sim.TileSize; x += 3 {
		hit := 40.0
		for d := 8.0; d <= 40; d += 0.5 {
			if lum(at(x, d)) <= half {
				hit = d
				break
			}
		}
		if hit < minD {
			minD = hit
		}
		if hit > maxD {
			maxD = hit
		}
	}
	if maxD-minD < 3 {
		t.Fatalf("fade sits at %.1f to %.1f, want it to wander", minD, maxD)
	}
}

func TestShallowBandsJoin(t *testing.T) {
	const w, h = 24, 12
	tiles := make([]sim.Tile, w*h)
	for y := 4; y < h; y++ {
		for x := 0; x < w; x++ {
			tiles[y*w+x] = sim.TileWaterDeep
		}
	}
	// One tile of water between the shore and this island.
	for y := 5; y <= 6; y++ {
		for x := 4; x <= 6; x++ {
			tiles[y*w+x] = sim.TileGrass
		}
	}
	// Two tiles of water between the shore and this island.
	for y := 6; y <= 7; y++ {
		for x := 16; x <= 18; x++ {
			tiles[y*w+x] = sim.TileGrass
		}
	}
	paint := paintMap(mapOf(w, h, tiles), "jungle")
	at := func(wx, wy float64) color.RGBA {
		px := int(math.Round(wx*terrainPx - 0.5))
		py := int(math.Round(wy*terrainPx - 0.5))
		return paint.ground.RGBAAt(px, py)
	}
	// A channel is closer to two shores, so it stays lighter than open
	// water at the same distance from a single coast.
	gap1 := lum(at(5*sim.TileSize+8, 4*sim.TileSize+8))
	gap2 := lum(at(17*sim.TileSize+8, 5*sim.TileSize))
	open := lum(at(12*sim.TileSize+8, 5*sim.TileSize))
	if gap1 < open+18 {
		t.Fatalf("one-tile gap lum %d, open %d, want a lighter channel", gap1, open)
	}
	if gap2 < open+10 {
		t.Fatalf("two-tile gap lum %d, open %d, want a lighter channel", gap2, open)
	}
	// Across the two-tile gap the colour steps, it does not stripe.
	x := float64(17*sim.TileSize + 8)
	prev := at(x, 66)
	maxStep := 0
	for y := 66.5; y <= 94; y += 0.5 {
		c := at(x, y)
		step := int(prev.G) - int(c.G)
		if step < 0 {
			step = -step
		}
		if step > maxStep {
			maxStep = step
		}
		prev = c
	}
	if maxStep > 8 {
		t.Fatalf("joined shallow step of %d, want a smooth blend", maxStep)
	}
}

// firstGrassIn walks from a shore point along (stepX, stepY) and returns how
// far, in world pixels, the sand gives way to grass.
func firstGrassIn(img *image.RGBA, x0, y0, stepX, stepY int) float64 {
	for d := 0.5; d < 16; d += 0.25 {
		wx := float64(x0) + float64(stepX)*d
		wy := float64(y0) + float64(stepY)*d
		px := int(math.Round(wx*terrainPx - 0.5))
		py := int(math.Round(wy*terrainPx - 0.5))
		c := img.RGBAAt(px, py)
		if c.G > c.R && c.B < 90 {
			return d
		}
	}
	return 16
}

// shorePx is a source pixel at a world offset from a tile edge.
func shorePx(edge, worldDelta int) int {
	return edge*terrainPx + worldDelta*terrainPx
}

func shoreWorld(px, py int) (float64, float64) {
	return (float64(px) + 0.5) / terrainPx, (float64(py) + 0.5) / terrainPx
}

func meanCell(img *image.RGBA, tx, ty int) (r, g, b float64) {
	x0 := tx * sim.TileSize * terrainPx
	y0 := ty * sim.TileSize * terrainPx
	span := sim.TileSize * terrainPx
	var rs, gs, bs, n float64
	for y := 0; y < span; y += 4 {
		for x := 0; x < span; x += 4 {
			c := img.RGBAAt(x0+x, y0+y)
			rs += float64(c.R)
			gs += float64(c.G)
			bs += float64(c.B)
			n++
		}
	}
	return rs / n, gs / n, bs / n
}

func TestTreeRunsShareOneCanopy(t *testing.T) {
	tiles := make([]sim.Tile, 12)
	tiles[4] = sim.TileTree
	tiles[5] = sim.TileTree
	tiles[6] = sim.TileTree
	paint := paintMap(mapOf(4, 3, tiles), "grass")
	if len(paint.trees) != 1 {
		t.Fatalf("canopies = %d, want 1", len(paint.trees))
	}
	can := paint.trees[0]
	if len(can.runs) != 3 {
		t.Fatalf("runs = %d, want 3", len(can.runs))
	}
	b := can.img.Bounds()
	for i, run := range can.runs {
		if run.rect.Empty() || !run.rect.In(b) {
			t.Fatalf("run %d rect %v outside canopy %v", i, run.rect, b)
		}
	}
}

func TestCanopyEdgeFeathers(t *testing.T) {
	tiles := make([]sim.Tile, 8*8)
	tiles[3*8+3] = sim.TileTree
	paint := paintMap(mapOf(8, 8, tiles), "grass")
	if len(paint.trees) != 1 || len(paint.trees[0].runs) != 1 {
		t.Fatalf("lone tree canopies=%d runs=%d", len(paint.trees), len(paint.trees[0].runs))
	}
	can := paint.trees[0]
	img := can.img
	cx := int((float64(3*sim.TileSize) + sim.TileSize/2 - can.originX) * terrainPx)
	cy := int((float64(3*sim.TileSize) + sim.TileSize/2 - can.originY) * terrainPx)
	if img.RGBAAt(cx, cy).A < 250 {
		t.Fatalf("crown centre alpha = %d, want solid", img.RGBAAt(cx, cy).A)
	}
	lastSolid, lastAny := cx, cx
	for x := cx; x < img.Bounds().Dx(); x++ {
		a := img.RGBAAt(x, cy).A
		if a > 240 {
			lastSolid = x
		}
		if a > 12 {
			lastAny = x
		}
	}
	if lastAny-lastSolid < 12 {
		t.Fatalf("east fringe is %d px, want a soft edge", lastAny-lastSolid)
	}
	minX, maxX := img.Bounds().Dx(), -1
	for y := cy - 28; y <= cy+28; y++ {
		if y < 0 || y >= img.Bounds().Dy() {
			continue
		}
		right := -1
		for x := img.Bounds().Dx() - 1; x >= 0; x-- {
			if img.RGBAAt(x, y).A > 40 {
				right = x
				break
			}
		}
		if right < 0 {
			continue
		}
		if right < minX {
			minX = right
		}
		if right > maxX {
			maxX = right
		}
	}
	if maxX-minX < 12 {
		t.Fatalf("east silhouette span %d px, want an uneven crown", maxX-minX)
	}
	run := can.runs[0]
	if run.sprite == nil || run.rect.Empty() || !run.rect.In(img.Bounds()) {
		t.Fatalf("run rect %v sprite %v", run.rect, run.sprite)
	}
}

func TestCanopySlicesDoNotStack(t *testing.T) {
	tiles := make([]sim.Tile, 12)
	tiles[4] = sim.TileTree
	tiles[5] = sim.TileTree
	tiles[6] = sim.TileTree
	paint := paintMap(mapOf(4, 3, tiles), "grass")
	can := paint.trees[0]
	stacked, missing := 0, 0
	b := can.img.Bounds()
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if can.img.RGBAAt(x, y).A == 0 {
				continue
			}
			wx := can.originX + (float64(x)+0.5)/terrainPx
			wy := can.originY + (float64(y)+0.5)/terrainPx
			n := 0
			for _, run := range can.runs {
				if run.sprite == nil {
					continue
				}
				sx := int(math.Floor((wx - run.world) * terrainPx))
				sy := int(math.Floor((wy - run.north) * terrainPx))
				sb := run.sprite.Bounds()
				if sx < 0 || sy < 0 || sx >= sb.Dx() || sy >= sb.Dy() {
					continue
				}
				if run.sprite.RGBAAt(sx, sy).A > 0 {
					n++
				}
			}
			if n == 0 {
				missing++
			}
			if n > 1 {
				stacked++
			}
		}
	}
	if missing != 0 || stacked != 0 {
		t.Fatalf("canopy pixels missing=%d stacked=%d", missing, stacked)
	}
}

func TestKindsOwnWaterAndTrees(t *testing.T) {
	kinds := []look{jungleLook, forestLook, desertLook, snowLook}
	for i := range kinds {
		for j := i + 1; j < len(kinds); j++ {
			a, b := kinds[i], kinds[j]
			if a.shallow == b.shallow || a.deep == b.deep || a.canopy == b.canopy || a.ground == b.ground {
				t.Fatalf("%s and %s share ground, water, or trees", a.kind, b.kind)
			}
		}
	}
	const n = 8
	tiles := make([]sim.Tile, n*n)
	for y := 1; y <= 4; y++ {
		for x := 1; x <= 4; x++ {
			tiles[y*n+x] = sim.TileWaterShallow
		}
	}
	tiles[6*n+6] = sim.TileTree
	names := []string{"jungle", "forest", "desert", "snow"}
	water := make([]color.RGBA, len(names))
	crown := make([]color.RGBA, len(names))
	for i, name := range names {
		paint := paintMap(mapOf(n, n, tiles), name)
		water[i] = paint.ground.RGBAAt(2*sim.TileSize*terrainPx+8*terrainPx, 2*sim.TileSize*terrainPx+8*terrainPx)
		can := paint.trees[0]
		cx := int((float64(6*sim.TileSize) + sim.TileSize/2 - can.originX) * terrainPx)
		cy := int((float64(6*sim.TileSize) + sim.TileSize/2 - can.originY) * terrainPx)
		crown[i] = can.img.RGBAAt(cx, cy)
	}
	for i := range names {
		for j := i + 1; j < len(names); j++ {
			if colorDist(water[i], water[j]) < 40 {
				t.Fatalf("%s water %v and %s water %v are too close", names[i], water[i], names[j], water[j])
			}
			if colorDist(crown[i], crown[j]) < 24 {
				t.Fatalf("%s trees %v and %s trees %v are too close", names[i], crown[i], names[j], crown[j])
			}
		}
	}
	open := paintMap(mapOf(4, 4, nil), "desert")
	r, g, _ := meanCell(open.ground, 1, 1)
	if r <= g {
		t.Fatalf("desert ground r=%.0f g=%.0f, want sand", r, g)
	}
	forest := paintMap(mapOf(4, 4, nil), "forest")
	fr, fg, _ := meanCell(forest.ground, 1, 1)
	jr, jg, _ := meanCell(paintMap(mapOf(4, 4, nil), "jungle").ground, 1, 1)
	if fg <= fr {
		t.Fatalf("forest ground r=%.0f g=%.0f, want green", fr, fg)
	}
	if math.Abs(fr-jr)+math.Abs(fg-jg) < 8 {
		t.Fatalf("forest ground matches jungle (%.0f,%.0f)", fr, fg)
	}
}

func colorDist(a, b color.RGBA) int {
	return absInt(int(a.R)-int(b.R)) + absInt(int(a.G)-int(b.G)) + absInt(int(a.B)-int(b.B))
}

func TestArcticSnowAndIce(t *testing.T) {
	tiles := make([]sim.Tile, 16)
	tiles[0] = sim.TileIce
	paint := paintMap(mapOf(4, 4, tiles), "snow")
	sr, _, sb := meanCell(paint.ground, 2, 2)
	ir, _, ib := meanCell(paint.ground, 0, 0)
	if sr < 140 {
		t.Fatalf("open ground r=%.0f, want snow", sr)
	}
	if ib-ir <= sb-sr {
		t.Fatalf("ice blue %.0f, snow blue %.0f, want ice colder than snow", ib-ir, sb-sr)
	}
}

func TestMineRebuildKeepsInterior(t *testing.T) {
	tiles := make([]sim.Tile, 36)
	tiles[0] = sim.TileMine
	before := paintMap(mapOf(6, 6, tiles), "grass")
	tiles[0] = sim.TileGrass
	after := paintMap(mapOf(6, 6, tiles), "grass")
	x := 5*sim.TileSize*terrainPx + 4*terrainPx
	y := 5*sim.TileSize*terrainPx + 4*terrainPx
	if before.ground.RGBAAt(x, y) != after.ground.RGBAAt(x, y) {
		t.Fatalf("far grass changed from %+v to %+v", before.ground.RGBAAt(x, y), after.ground.RGBAAt(x, y))
	}
	mx := 8 * terrainPx
	my := 8 * terrainPx
	if before.ground.RGBAAt(mx, my) == after.ground.RGBAAt(mx, my) {
		t.Fatal("clearing the mine left the marker in place")
	}
	if after.surf[0] != surfGrass {
		t.Fatalf("cleared cell surface = %d, want grass", after.surf[0])
	}
}

func TestPhasePaintFitsALoad(t *testing.T) {
	tiles := make([]sim.Tile, 36*26)
	for x := 0; x < 36; x++ {
		tiles[8*36+x] = sim.TileCliff
		tiles[20*36+x] = sim.TileWaterShallow
		tiles[21*36+x] = sim.TileWaterDeep
	}
	for i := 0; i < 8; i++ {
		tiles[4*36+i] = sim.TileTree
	}
	start := time.Now()
	paintMap(mapOf(36, 26, tiles), "grass")
	if d := time.Since(start); d > time.Second {
		t.Fatalf("phase paint took %s", d)
	}
}

func riverBridge(vertical bool) sim.Map {
	const w, h = 10, 8
	tiles := make([]sim.Tile, w*h)
	if vertical {
		for y := 2; y <= 5; y++ {
			for x := 0; x < w; x++ {
				tiles[y*w+x] = sim.TileWaterDeep
			}
			tiles[y*w+4] = sim.TileBridge
			tiles[y*w+5] = sim.TileBridge
		}
		return mapOf(w, h, tiles)
	}
	for x := 2; x <= 7; x++ {
		for y := 0; y < h; y++ {
			tiles[y*w+x] = sim.TileWaterDeep
		}
		tiles[3*w+x] = sim.TileBridge
		tiles[4*w+x] = sim.TileBridge
	}
	return mapOf(w, h, tiles)
}

func worldPx(wx, wy float64) (int, int) {
	return int(wx * terrainPx), int(wy * terrainPx)
}

func lum(c color.RGBA) int {
	return int(c.R) + int(c.G) + int(c.B)
}

func TestBridgeDeckIsTimber(t *testing.T) {
	paint := paintMap(riverBridge(true), "jungle")
	// Middle of the span, clear of the rail and of a board seam.
	cx, cy := worldPx(4*sim.TileSize+8, 3*sim.TileSize+2.2)
	deck := paint.ground.RGBAAt(cx, cy)
	if deck.R < deck.B+18 || deck.R < deck.G+8 {
		t.Fatalf("deck %+v, want warm timber", deck)
	}
	if paint.anim.RGBAAt(cx, cy).A != 0 {
		t.Fatal("deck still shimmers")
	}
	// The two bridge columns share one deck, so the join is timber too.
	jx, jy := worldPx(5*sim.TileSize+0.4, 3*sim.TileSize+2.2)
	join := paint.ground.RGBAAt(jx, jy)
	if join.R < join.B+12 {
		t.Fatalf("span join %+v, want timber", join)
	}
	// Outer rail is darker than the boards.
	rx, ry := worldPx(4*sim.TileSize+0.7, 3*sim.TileSize+2.2)
	rail := paint.ground.RGBAAt(rx, ry)
	if lum(rail)+20 > lum(deck) {
		t.Fatalf("rail %+v is not darker than deck %+v", rail, deck)
	}
	// Water beside the span keeps its shimmer. The south-east side is in shadow.
	ox, oy := worldPx(7*sim.TileSize+4, 3*sim.TileSize+8)
	if paint.mask.RGBAAt(ox, oy).R != maskWater || paint.anim.RGBAAt(ox, oy).A == 0 {
		t.Fatal("open water beside the bridge lost its shimmer")
	}
	sx, sy := worldPx(6*sim.TileSize+3.4, 3*sim.TileSize+11)
	fx, fy := worldPx(6*sim.TileSize+12, 3*sim.TileSize+4)
	if lum(paint.ground.RGBAAt(sx, sy))+12 > lum(paint.ground.RGBAAt(fx, fy)) {
		t.Fatalf("shadow %+v is not darker than open water %+v", paint.ground.RGBAAt(sx, sy), paint.ground.RGBAAt(fx, fy))
	}

	side := paintMap(riverBridge(false), "jungle")
	topX, topY := worldPx(4*sim.TileSize+8, 3*sim.TileSize+0.7)
	midX, midY := worldPx(4*sim.TileSize+8, 3*sim.TileSize+8)
	if lum(side.ground.RGBAAt(topX, topY))+12 > lum(side.ground.RGBAAt(midX, midY)) {
		t.Fatalf("side rail %+v, deck %+v", side.ground.RGBAAt(topX, topY), side.ground.RGBAAt(midX, midY))
	}

	snow := paintMap(riverBridge(true), "snow")
	if lum(snow.ground.RGBAAt(cx, cy)) < lum(deck)+25 {
		t.Fatalf("snow deck %+v, jungle %+v, want a dusting", snow.ground.RGBAAt(cx, cy), deck)
	}
}

func TestWaterShaderCompiles(t *testing.T) {
	if _, err := ebiten.NewShader(waterKage); err != nil {
		t.Fatal(err)
	}
}

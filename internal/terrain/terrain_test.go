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
	// Two world pixels south of the water: inside the sand, past the foam.
	edge := paint.ground.RGBAAt(4*sim.TileSize*terrainPx+8*terrainPx, sim.TileSize*terrainPx+2*terrainPx)
	// Eight world pixels south: past the thin sand, back on grass.
	past := paint.ground.RGBAAt(4*sim.TileSize*terrainPx+8*terrainPx, sim.TileSize*terrainPx+8*terrainPx)
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
		// Open grass on the west bank, two world pixels inland.
		spx, spy := shorePx(2*sim.TileSize, -2), shorePx(4*sim.TileSize+8, 0)
		sand := sandAt(spx, spy)
		sx, sy := shoreWorld(spx, spy)
		gpx, gpy := shorePx(2*sim.TileSize, -10), spy
		grass := sandAt(gpx, gpy)
		gx, gy := shoreWorld(gpx, gpy)
		hpx, hpy := shorePx(2*sim.TileSize, 2), spy
		shallow := sandAt(hpx, hpy)
		hx, hy := shoreWorld(hpx, hpy)
		dpx, dpy := shorePx(4*sim.TileSize+8, 0), shorePx(5*sim.TileSize+8, 0)
		deep := sandAt(dpx, dpy)
		dx, dy := shoreWorld(dpx, dpy)
		tpx, tpy := shorePx(7*sim.TileSize, 2), shorePx(3*sim.TileSize+8, 0)
		treeSand := sandAt(tpx, tpy)
		tx, ty := shoreWorld(tpx, tpy)
		wantShallow, _ := lk.plateColor(surfShallow, hx, hy)
		wantDeep, _ := lk.plateColor(surfDeep, dx, dy)
		// The grass line wanders, so a pixel in the strip is sand-coloured,
		// not a perfect plate sample.
		if sand.R <= sand.G+15 || colorDist(sand, sandPlate.at(sx, sy)) >= colorDist(sand, lk.ground.at(sx, sy)) {
			t.Fatalf("%s sand %+v, want sand rather than ground", name, sand)
		}
		if colorDist(grass, lk.ground.at(gx, gy)) > 2 {
			t.Fatalf("%s inland %+v, want grass past the sand", name, grass)
		}
		if colorDist(shallow, wantShallow) > 2 {
			t.Fatalf("%s water edge %+v, want shallow %+v", name, shallow, wantShallow)
		}
		if colorDist(deep, wantDeep) > 2 {
			t.Fatalf("%s pond centre %+v, want deep %+v", name, deep, wantDeep)
		}
		// The canopy shadow falls on the bank, so the sand is darker than the plate.
		treeGround := lk.ground.at(tx, ty)
		if treeSand.R <= treeSand.G+15 || colorDist(treeSand, sandPlate.at(tx, ty)) >= colorDist(treeSand, treeGround) {
			t.Fatalf("%s tree bank %+v, want sand rather than %+v", name, treeSand, treeGround)
		}
		if paint.mask.RGBAAt(spx, spy).R != maskNone {
			t.Fatalf("%s sand mask = %d, want still ground", name, paint.mask.RGBAAt(spx, spy).R)
		}
		if paint.mask.RGBAAt(hpx, hpy).R != maskWater {
			t.Fatalf("%s shallow mask = %d, want water", name, paint.mask.RGBAAt(hpx, hpy).R)
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
		for d := 0.5; d < 8; d += 0.25 {
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
	if maxIn-minIn < 0.75 {
		t.Fatalf("grass line spans %.2f to %.2f, want a visible wander", minIn, maxIn)
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
	near := at(x, 2)
	mid := at(x, 9)
	far := at(x, 16)
	deep := at(x, 28)
	if near.G <= mid.G || mid.G <= far.G || far.G <= deep.G {
		t.Fatalf("shallow fade g = %d, %d, %d, %d, want a steady drop", near.G, mid.G, far.G, deep.G)
	}
	if int(near.G)-int(deep.G) < 20 {
		t.Fatalf("fade only dropped g %d to %d", near.G, deep.G)
	}
	// No single world pixel along the fade may jump the whole way.
	var maxStep int
	prev := at(x, 2)
	for d := 2.5; d <= 20; d += 0.5 {
		c := at(x, d)
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
		t.Fatalf("shallow/deep step of %d, want a fade", maxStep)
	}
	// The halfway point moves along the shore.
	half := (int(near.G) + int(deep.G)) / 2
	minD, maxD := 99.0, 0.0
	for x := 4 * sim.TileSize; x < 16*sim.TileSize; x += 3 {
		hit := 20.0
		for d := 3.0; d <= 20; d += 0.5 {
			if int(at(x, d).G) <= half {
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
	if maxD-minD < 1 {
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
	lk := lookOf("jungle")
	at := func(wx, wy float64) color.RGBA {
		px := int(math.Round(wx*terrainPx - 0.5))
		py := int(math.Round(wy*terrainPx - 0.5))
		return paint.ground.RGBAAt(px, py)
	}
	// Closer to the shallow plate than the deep one means the bands joined.
	joined := func(wx, wy float64) bool {
		c := at(wx, wy)
		sh, _ := lk.plateColor(surfShallow, wx, wy)
		dp, _ := lk.plateColor(surfDeep, wx, wy)
		return colorDist(c, sh) < colorDist(c, dp)
	}
	open := func(wx, wy float64) bool {
		c := at(wx, wy)
		sh, _ := lk.plateColor(surfShallow, wx, wy)
		dp, _ := lk.plateColor(surfDeep, wx, wy)
		return colorDist(c, dp) < colorDist(c, sh)
	}
	// Mid-channel of the one-tile gap, and of the two-tile gap.
	if !joined(5*sim.TileSize+8, 4*sim.TileSize+8) {
		t.Fatal("one-tile gap between island and shore still reads as deep")
	}
	if !joined(17*sim.TileSize+8, 5*sim.TileSize) {
		t.Fatal("two-tile gap between island and shore still reads as deep")
	}
	// The same distance from a lone shore stays on the deep side of the fade.
	if !open(12*sim.TileSize+8, 5*sim.TileSize) {
		t.Fatal("open water at the two-tile distance joined with nothing")
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

func TestWaterShaderCompiles(t *testing.T) {
	if _, err := ebiten.NewShader(waterKage); err != nil {
		t.Fatal(err)
	}
}

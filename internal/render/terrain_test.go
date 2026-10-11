package render

import (
	"bytes"
	"image"
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
	// Just south of the water, past the foam band, still inside the sand lip.
	edge := paint.ground.RGBAAt(4*sim.TileSize*terrainPx+8*terrainPx, sim.TileSize*terrainPx+4*terrainPx)
	mr, mg, _ := meanCell(paint.ground, 4, 6)
	if edge.R <= edge.G {
		t.Fatalf("shore pixel = %+v, want the sand lip", edge)
	}
	if mg <= mr {
		t.Fatalf("interior mean r=%.0f g=%.0f, want grass", mr, mg)
	}
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

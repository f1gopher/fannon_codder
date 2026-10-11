package render

import (
	_ "embed"
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

//go:embed water.kage
var waterKage []byte

// terrainView is the uploaded picture for one tile hash. Field rebuilds it
// when a mine becomes grass or the phase changes. The camera does not.
type terrainView struct {
	hash   uint64
	ground *ebiten.Image
	anim   *ebiten.Image
	mask   *ebiten.Image
	runs   []terrainRun
}

type terrainRun struct {
	img   *ebiten.Image
	world float64
	north float64
	foot  float64
	footX float64
}

var terrainNow terrainView

var terrainShader struct {
	s   *ebiten.Shader
	err error
	try bool
}

func terrainHash(m sim.Map, terrain string) uint64 {
	var h uint64 = 14695981039346656037
	h = (h ^ uint64(m.W)) * 1099511628211
	h = (h ^ uint64(m.H)) * 1099511628211
	if arcticTerrain(terrain) {
		h ^= 0xff
	}
	for _, t := range m.Tiles {
		h = (h ^ uint64(t)) * 1099511628211
	}
	return h
}

func ensureTerrain(m sim.Map, terrain string) {
	if m.W == 0 || m.H == 0 || len(m.Tiles) < m.W*m.H {
		terrainNow = terrainView{}
		return
	}
	h := terrainHash(m, terrain)
	if terrainNow.ground != nil && terrainNow.hash == h {
		return
	}
	painted := paintMap(m, terrain)
	view := terrainView{
		hash:   h,
		ground: uploadStraight(painted.ground),
		anim:   uploadStraight(painted.anim),
		mask:   uploadStraight(painted.mask),
	}
	for i := range painted.trees {
		can := painted.trees[i]
		for _, run := range can.runs {
			if run.sprite == nil || run.sprite.Bounds().Empty() {
				continue
			}
			view.runs = append(view.runs, terrainRun{
				img:   uploadStraight(run.sprite),
				world: run.world,
				north: run.north,
				foot:  run.foot,
				footX: run.footX,
			})
		}
	}
	terrainNow = view
}

func uploadStraight(src *image.RGBA) *ebiten.Image {
	pix := make([]byte, len(src.Pix))
	copy(pix, src.Pix)
	for i := 0; i < len(pix); i += 4 {
		a := uint32(pix[i+3])
		pix[i] = byte(uint32(pix[i]) * a / 255)
		pix[i+1] = byte(uint32(pix[i+1]) * a / 255)
		pix[i+2] = byte(uint32(pix[i+2]) * a / 255)
	}
	img := ebiten.NewImage(src.Bounds().Dx(), src.Bounds().Dy())
	img.WritePixels(pix)
	return img
}

func drawTerrain(dst *ebiten.Image, m sim.Map, cam sim.Camera, terrain string) {
	if dst == nil || m.W == 0 {
		return
	}
	ensureTerrain(m, terrain)
	if terrainNow.ground == nil {
		return
	}
	x := cam.ScreenX(0)
	y := cam.ScreenY(0)
	drawTerrainImage(dst, terrainNow.ground, x, y)
	drawAnim(dst, terrainNow.anim, terrainNow.mask, x, y)
}

func drawTerrainImage(dst, img *ebiten.Image, x, y float64) {
	if dst == nil || img == nil {
		return
	}
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	sc := ps / float64(terrainPx)
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.DisableMipmaps = true
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x*ps, y*ps)
	dst.DrawImage(img, op)
}

func drawAnim(dst, anim, mask *ebiten.Image, x, y float64) {
	if anim == nil || mask == nil {
		return
	}
	sh := waterShader()
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	sc := ps / float64(terrainPx)
	if sh == nil {
		drawTerrainImage(dst, anim, x, y)
		return
	}
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = anim
	op.Images[1] = mask
	op.Uniforms = map[string]any{"Time": float32(animTime)}
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x*ps, y*ps)
	b := anim.Bounds()
	dst.DrawRectShader(b.Dx(), b.Dy(), sh, op)
}

func waterShader() *ebiten.Shader {
	if terrainShader.try {
		return terrainShader.s
	}
	terrainShader.try = true
	terrainShader.s, terrainShader.err = ebiten.NewShader(waterKage)
	if terrainShader.err != nil {
		terrainShader.s = nil
	}
	return terrainShader.s
}

func treeShadowSrc() *ebiten.Image {
	if treeShadowImg != nil {
		return treeShadowImg
	}
	// castShadow reads the width in ArtScale units. 14 world pixels is a
	// mound's contact shadow, not the whole canopy crop.
	treeShadowImg = ebiten.NewImage(14*ArtScale, 8)
	return treeShadowImg
}

var treeShadowImg *ebiten.Image

func terrainTreeBodies(m sim.Map, cam sim.Camera, terrain string) []body {
	if m.W == 0 {
		return nil
	}
	ensureTerrain(m, terrain)
	bodies := make([]body, 0, len(terrainNow.runs))
	for i := range terrainNow.runs {
		run := terrainNow.runs[i]
		sx := cam.ScreenX(run.world)
		sy := cam.ScreenY(run.north)
		fx := cam.ScreenX(run.footX)
		fy := cam.ScreenY(run.foot)
		img := run.img
		bodies = append(bodies, body{
			y: run.foot,
			shadow: func(dst *ebiten.Image) {
				castShadow(dst, fx, fy, treeShadowSrc())
			},
			draw: func(dst *ebiten.Image) {
				drawTerrainImage(dst, img, sx, sy)
			},
		})
	}
	return bodies
}

package terrain

import (
	_ "embed"
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
)

//go:embed water.kage
var waterKage []byte

// View is the uploaded picture for one tile hash. Prepare rebuilds it when
// a mine becomes grass or the phase changes. Scrolling does not.
type View struct {
	ground   *ebiten.Image
	anim     *ebiten.Image
	mask     *ebiten.Image
	canopies []Canopy
}

// Canopy is one tree cell's sprite. FootY is the south edge the battle
// Y-sorts against huts and soldiers.
type Canopy struct {
	img            *ebiten.Image
	WorldX, NorthY float64
	FootX, FootY   float64
}

var terrainNow struct {
	hash uint64
	view *View
}

var terrainShader struct {
	s   *ebiten.Shader
	err error
	try bool
}

func terrainHash(m sim.Map, terrain string) uint64 {
	var h uint64 = 14695981039346656037
	h = (h ^ uint64(m.W)) * 1099511628211
	h = (h ^ uint64(m.H)) * 1099511628211
	h = (h ^ uint64(KindOf(terrain))) * 1099511628211
	for _, t := range m.Tiles {
		h = (h ^ uint64(t)) * 1099511628211
	}
	return h
}

// Prepare caches one painted map per process. An inactive map draws nothing.
func Prepare(m sim.Map, terrain string) *View {
	if m.W == 0 || m.H == 0 || len(m.Tiles) < m.W*m.H {
		terrainNow = struct {
			hash uint64
			view *View
		}{}
		return &View{}
	}
	h := terrainHash(m, terrain)
	if terrainNow.view != nil && terrainNow.view.ground != nil && terrainNow.hash == h {
		return terrainNow.view
	}
	painted := paintMap(m, terrain)
	view := &View{
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
			view.canopies = append(view.canopies, Canopy{
				img:    uploadStraight(run.sprite),
				WorldX: run.world,
				NorthY: run.north,
				FootX:  run.footX,
				FootY:  run.foot,
			})
		}
	}
	terrainNow.hash = h
	terrainNow.view = view
	return view
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

// DrawGround paints ground and shimmer. originX and originY are where world
// (0,0) sits, in world pixels, before scale. scale is offscreen pixels per
// world pixel. time is seconds for the water shader.
func (v *View) DrawGround(dst *ebiten.Image, originX, originY, scale, time float64) {
	if v == nil || dst == nil || v.ground == nil {
		return
	}
	drawScaled(dst, v.ground, originX, originY, scale)
	drawAnim(dst, v.anim, v.mask, originX, originY, scale, time)
}

// Canopies are the tree slices. The battle Y-sorts them with actors.
func (v *View) Canopies() []Canopy {
	if v == nil {
		return nil
	}
	return v.canopies
}

// Draw paints this canopy sprite. origin and scale match DrawGround.
func (c Canopy) Draw(dst *ebiten.Image, originX, originY, scale float64) {
	// Each cell is its own image. A linear sample past that frame is clear
	// black, which reads as a line where two crowns meet. The bled frame
	// keeps the leaf colour in range, and one extra device pixel closes
	// the crack the vertex snap can open between the slices.
	img := bledTerrain(c.img)
	if scale <= 0 {
		scale = 1
	}
	sc := scale / float64(terrainPx)
	fw := float64(img.Bounds().Dx())
	fh := float64(img.Bounds().Dy())
	sx, sy := sc, sc
	if fw > 0 {
		sx = (fw*sc + 1) / fw
	}
	if fh > 0 {
		sy = (fh*sc + 1) / fh
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.DisableMipmaps = true
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate((originX+c.WorldX)*scale, (originY+c.NorthY)*scale)
	dst.DrawImage(img, op)
}

// bledTerrain is the frame plus a one-pixel copy of each edge. Draw the
// returned subimage so a linear sample stays on the crown.
var bledTerrainCache sync.Map

func bledTerrain(img *ebiten.Image) *ebiten.Image {
	if img == nil {
		return nil
	}
	if v, ok := bledTerrainCache.Load(img); ok {
		return v.(*ebiten.Image)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return img
	}
	pad := ebiten.NewImage(w+2, h+2)
	put := func(src *ebiten.Image, x, y float64) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		pad.DrawImage(src, op)
	}
	put(img, 1, 1)
	put(img.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Min.X+1, b.Max.Y)).(*ebiten.Image), 0, 1)
	put(img.SubImage(image.Rect(b.Max.X-1, b.Min.Y, b.Max.X, b.Max.Y)).(*ebiten.Image), float64(w+1), 1)
	put(img.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+1)).(*ebiten.Image), 1, 0)
	put(img.SubImage(image.Rect(b.Min.X, b.Max.Y-1, b.Max.X, b.Max.Y)).(*ebiten.Image), 1, float64(h+1))
	put(img.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Min.X+1, b.Min.Y+1)).(*ebiten.Image), 0, 0)
	put(img.SubImage(image.Rect(b.Max.X-1, b.Min.Y, b.Max.X, b.Min.Y+1)).(*ebiten.Image), float64(w+1), 0)
	put(img.SubImage(image.Rect(b.Min.X, b.Max.Y-1, b.Min.X+1, b.Max.Y)).(*ebiten.Image), 0, float64(h+1))
	put(img.SubImage(image.Rect(b.Max.X-1, b.Max.Y-1, b.Max.X, b.Max.Y)).(*ebiten.Image), float64(w+1), float64(h+1))
	inner := pad.SubImage(image.Rect(1, 1, w+1, h+1)).(*ebiten.Image)
	bledTerrainCache.Store(img, inner)
	return inner
}

func drawScaled(dst, img *ebiten.Image, x, y, scale float64) {
	if dst == nil || img == nil {
		return
	}
	if scale <= 0 {
		scale = 1
	}
	sc := scale / float64(terrainPx)
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.DisableMipmaps = true
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x*scale, y*scale)
	dst.DrawImage(img, op)
}

func drawAnim(dst, anim, mask *ebiten.Image, x, y, scale, time float64) {
	if anim == nil || mask == nil {
		return
	}
	sh := waterShader()
	if scale <= 0 {
		scale = 1
	}
	sc := scale / float64(terrainPx)
	if sh == nil {
		drawScaled(dst, anim, x, y, scale)
		return
	}
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = anim
	op.Images[1] = mask
	op.Uniforms = map[string]any{"Time": float32(time)}
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(x*scale, y*scale)
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

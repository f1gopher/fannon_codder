package render

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/sim"
	"fannon-codder/internal/terrain"
)

// A grenade, rocket, mine, or crate blast leaves a stain on the ground.
// The stain is a soft multiply, so it darkens the paint already there
// instead of covering it with a flat colour. Water takes no mark. A shed
// fireball does not either. ResetAnim clears the list when the next battle
// starts.

const (
	scorchTex  = 96
	scorchSpan = 27.0 // world pixels across the texture
)

// scorchBlend multiplies the ground by the stain. White leaves a pixel
// alone. A darker texel burns it, and the ground's own colour stays.
var scorchBlend = ebiten.Blend{
	BlendFactorSourceRGB:        ebiten.BlendFactorDestinationColor,
	BlendFactorSourceAlpha:      ebiten.BlendFactorZero,
	BlendFactorDestinationRGB:   ebiten.BlendFactorZero,
	BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
	BlendOperationRGB:           ebiten.BlendOperationAdd,
	BlendOperationAlpha:         ebiten.BlendOperationAdd,
}

type scorch struct {
	x, y float64
	img  *ebiten.Image
}

var scorches []scorch

func clearScorches() {
	for i := range scorches {
		if scorches[i].img != nil {
			scorches[i].img.Dispose()
		}
	}
	scorches = nil
}

func noteScorch(m sim.Map, x, y float64) {
	if scorchAlready(x, y) {
		return
	}
	if terrain.GroundAt(m, boomTerrain, x, y) == terrain.GroundWater {
		return
	}
	src := bakeScorch(x, y)
	img := ebiten.NewImageFromImage(src)
	scorches = append(scorches, scorch{x: x, y: y, img: img})
}

func scorchAlready(x, y float64) bool {
	for i := range scorches {
		if math.Hypot(scorches[i].x-x, scorches[i].y-y) < 2 {
			return true
		}
	}
	return false
}

// bakeScorch builds a straight-alpha multiply mask. The centre is a warm
// darkening. The rim feathers out to white, which multiplies as no change.
// Low-frequency wobble keeps the patch from reading as a disc.
func bakeScorch(x, y float64) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, scorchTex, scorchTex))
	cx := float64(scorchTex-1) / 2
	cy := cx
	s1 := x*0.17 + y*0.03
	s2 := y*0.13 - x*0.05
	for py := 0; py < scorchTex; py++ {
		for px := 0; px < scorchTex; px++ {
			dx := (float64(px) - cx) / cx
			dy := (float64(py) - cy) / cy
			ang := math.Atan2(dy, dx)
			nx := float64(px)/scorchTex + s1
			ny := float64(py)/scorchTex + s2
			blob := 0.62*scorchWave(nx*2.1, ny*2.1) + 0.38*scorchWave(nx*4.7+3, ny*4.3)
			rim := 0.74 + 0.16*math.Sin(ang*3+s1) + 0.08*math.Sin(ang*5-s2) + 0.10*(blob-0.5)
			if rim < 0.55 {
				rim = 0.55
			}
			u := math.Hypot(dx*1.04, dy*0.92) / rim
			fr, fg, fb := 1.0, 1.0, 1.0
			if u < 1 {
				t := u * u * (3 - 2*u)
				t = t * t * (3 - 2*t)
				// A little extra char where the noise dips, still the same hue family.
				char := (0.5 - blob) * 0.12 * (1 - u)
				if char < 0 {
					char = 0
				}
				dark := (1 - t) * (0.44 + char)
				fr = 1 - dark
				fg = 1 - dark*1.08
				fb = 1 - dark*1.28
			}
			dst.SetRGBA(px, py, color.RGBA{
				R: scorchByte(fr),
				G: scorchByte(fg),
				B: scorchByte(fb),
				A: 255,
			})
		}
	}
	return dst
}

func scorchByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

func scorchWave(x, y float64) float64 {
	n := math.Sin(x*12.9898+y*78.233) * 43758.5453
	return n - math.Floor(n)
}

func drawScorches(dst *ebiten.Image, cam sim.Camera) {
	if dst == nil || len(scorches) == 0 {
		return
	}
	ps := pictureScale
	if ps <= 0 {
		ps = 1
	}
	for i := range scorches {
		s := &scorches[i]
		if s.img == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.Blend = scorchBlend
		op.Filter = ebiten.FilterLinear
		op.DisableMipmaps = true
		drawn := scorchSpan * ps
		op.GeoM.Scale(drawn/scorchTex, drawn/scorchTex)
		op.GeoM.Translate(cam.ScreenX(s.x)*ps-drawn/2, cam.ScreenY(s.y)*ps-drawn/2)
		dst.DrawImage(s.img, op)
	}
}

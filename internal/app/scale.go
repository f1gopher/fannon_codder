package app

import "math"

// Picture limits. The world frame stays 320×256. S is how many offscreen
// pixels one world pixel occupies. The offscreen is never larger than 4K.
const (
	MinWindowWidth  = 1024
	MinWindowHeight = 768
	MaxWindowWidth  = 3840
	MaxWindowHeight = 2160

	// maxPictureScale is 2160/256. Height reaches 4K before width does.
	maxPictureScale = float64(MaxWindowHeight) / float64(ScreenHeight)
)

// PictureSize is the uniform fit of the 320×256 frame into a window.
// dipW and dipH are device-independent pixels. deviceScale is the monitor
// scale. The returned offscreen is 320×S by 256×S, clamped so neither side
// exceeds 3840×2160. A wider window does not grow S once height is the fit.
func PictureSize(dipW, dipH, deviceScale float64) (s float64, offW, offH int) {
	if dipW < 1 {
		dipW = DefaultWindowWidth
	}
	if dipH < 1 {
		dipH = DefaultWindowHeight
	}
	if deviceScale <= 0 {
		deviceScale = 1
	}
	fit := dipH / float64(ScreenHeight)
	if wfit := dipW / float64(ScreenWidth); wfit < fit {
		fit = wfit
	}
	s = fit * deviceScale
	if s > maxPictureScale {
		s = maxPictureScale
	}
	if s <= 0 {
		s = 1
	}
	offW = int(math.Round(float64(ScreenWidth) * s))
	offH = int(math.Round(float64(ScreenHeight) * s))
	return s, offW, offH
}

// OffscreenCursor converts an Ebitengine layout-space cursor into pixels of
// the offscreen that DrawFinalScreen blits at 1:1 in the centre.
// Ebitengine's own letterbox scales the offscreen up to the framebuffer;
// once S is clamped that scale is not 1, so the layout point has to be
// mapped back onto the unscaled picture.
func OffscreenCursor(layoutX, layoutY float64, fbW, fbH, offW, offH int) (float64, float64) {
	if fbW <= 0 || fbH <= 0 || offW <= 0 || offH <= 0 {
		return layoutX, layoutY
	}
	scale := float64(fbW) / float64(offW)
	if sy := float64(fbH) / float64(offH); sy < scale {
		scale = sy
	}
	if scale <= 0 {
		return layoutX, layoutY
	}
	ebitenOX := (float64(fbW) - float64(offW)*scale) / 2
	ebitenOY := (float64(fbH) - float64(offH)*scale) / 2
	physX := layoutX*scale + ebitenOX
	physY := layoutY*scale + ebitenOY
	ourOX := (float64(fbW) - float64(offW)) / 2
	ourOY := (float64(fbH) - float64(offH)) / 2
	return physX - ourOX, physY - ourOY
}

// FramePoint maps an offscreen-pixel cursor into the 320×256 frame.
// inside is false when the point lies in the letterbox bar. The returned
// point is still clamped, so edge scroll sees the frame edge.
func FramePoint(offX, offY, s float64, offW, offH int) (x, y float64, inside bool) {
	if s <= 0 {
		s = 1
	}
	if offW > 0 && offH > 0 {
		inside = offX >= 0 && offY >= 0 && offX < float64(offW) && offY < float64(offH)
	} else {
		inside = offX >= 0 && offY >= 0 && offX < float64(ScreenWidth)*s && offY < float64(ScreenHeight)*s
	}
	x = offX / s
	y = offY / s
	maxX := float64(ScreenWidth - 1)
	maxY := float64(ScreenHeight - 1)
	if x < 0 {
		x = 0
	} else if x > maxX {
		x = maxX
	}
	if y < 0 {
		y = 0
	} else if y > maxY {
		y = maxY
	}
	return x, y, inside
}

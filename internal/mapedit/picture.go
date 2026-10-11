package mapedit

import "math"

// Window and picture limits match the game. Picture is app.PictureSizeFilled.
// TestPictureMatchesTheGame keeps the two copies equal.
const (
	FrameW = 320
	FrameH = 256

	DefaultWindowWidth  = 1024
	DefaultWindowHeight = 768
	MinWindowWidth      = 1024
	MinWindowHeight     = 768
	MaxWindowWidth      = 3840
	MaxWindowHeight     = 2160

	// maxPictureScale is 2160/256. Height reaches 4K before width does.
	maxPictureScale = float64(MaxWindowHeight) / float64(FrameH)
)

func pictureSize(dipW, dipH, deviceScale float64) (s float64, offW, offH int) {
	if dipW < 1 {
		dipW = DefaultWindowWidth
	}
	if dipH < 1 {
		dipH = DefaultWindowHeight
	}
	if deviceScale <= 0 {
		deviceScale = 1
	}
	fit := dipH / float64(FrameH)
	if wfit := dipW / float64(FrameW); wfit < fit {
		fit = wfit
	}
	s = fit * deviceScale
	if s > maxPictureScale {
		s = maxPictureScale
	}
	if s <= 0 {
		s = 1
	}
	offW = int(math.Round(float64(FrameW) * s))
	offH = int(math.Round(float64(FrameH) * s))
	return s, offW, offH
}

// Picture is the game's filled picture: height fit of the 256-tall frame,
// then the offscreen grows sideways to the window. Scale includes deviceScale.
func Picture(dipW, dipH, deviceScale float64) (s float64, offW, offH int) {
	s, offW, offH = pictureSize(dipW, dipH, deviceScale)
	if deviceScale <= 0 {
		deviceScale = 1
	}
	if dipW < 1 {
		dipW = DefaultWindowWidth
	}
	physW := dipW * deviceScale
	if physW > float64(MaxWindowWidth) {
		physW = float64(MaxWindowWidth)
	}
	wide := int(math.Round(physW))
	if wide > offW {
		offW = wide
	}
	return s, offW, offH
}

// OffscreenPoint converts a layout-space cursor into offscreen pixels.
// It matches app.OffscreenCursor. The editor pans with those pixels divided
// by Picture's scale.
func OffscreenPoint(layoutX, layoutY float64, fbW, fbH, offW, offH int) (float64, float64) {
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

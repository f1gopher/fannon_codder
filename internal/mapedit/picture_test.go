package mapedit_test

import (
	"testing"

	"fannon-codder/internal/app"
	"fannon-codder/internal/mapedit"
)

func TestPictureMatchesTheGame(t *testing.T) {
	sizes := [][2]float64{{1024, 768}, {1440, 900}, {1920, 1080}}
	for _, device := range []float64{1, 2} {
		for _, sz := range sizes {
			wantS, wantW, wantH := app.PictureSizeFilled(sz[0], sz[1], device)
			gotS, gotW, gotH := mapedit.Picture(sz[0], sz[1], device)
			if gotS != wantS || gotW != wantW || gotH != wantH {
				t.Fatalf("picture %v scale %v = %v %d %d, game %v %d %d", sz, device, gotS, gotW, gotH, wantS, wantW, wantH)
			}
		}
	}
}

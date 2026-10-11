package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/mapedit"
)

func main() {
	ed, err := mapedit.New()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowTitle("Fannon Codder")
	ebiten.SetWindowSize(mapedit.DefaultWindowWidth, mapedit.DefaultWindowHeight)
	ebiten.SetWindowSizeLimits(mapedit.MinWindowWidth, mapedit.MinWindowHeight, mapedit.MaxWindowWidth, mapedit.MaxWindowHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)

	if err := ebiten.RunGame(ed); err != nil {
		log.Fatal(err)
	}
}

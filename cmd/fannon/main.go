package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/app"
)

func main() {
	ebiten.SetWindowTitle("Fannon Codder")
	ebiten.SetWindowSize(app.DefaultWindowWidth, app.DefaultWindowHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(app.TPS)

	game := app.New()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

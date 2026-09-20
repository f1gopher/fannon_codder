package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/app"
)

func main() {
	skipTitle := flag.Bool("skip-title", false, "jump straight into mission 1")
	flag.Parse()

	ebiten.SetWindowTitle("Fannon Codder")
	ebiten.SetWindowSize(app.DefaultWindowWidth, app.DefaultWindowHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(app.TPS)

	game := app.New(*skipTitle)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

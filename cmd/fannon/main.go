package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/internal/app"
)

func main() {
	skipTitle := flag.Bool("skip-title", false, "jump straight into mission 1")
	cover := flag.Bool("cover", false, "debug oversized map: scrolling + tree LOS")
	river := flag.Bool("river", false, "debug river + bridge (water, swimming)")
	hut := flag.Bool("hut", false, "debug spawner hut + grenade crate")
	hazards := flag.Bool("hazards", false, "debug civilians, quicksand, and mines")
	skidoo := flag.Bool("skidoo", false, "debug skidoo, bazooka, and an enemy skidoo")
	flag.Parse()

	ebiten.SetWindowTitle("Fannon Codder")
	ebiten.SetWindowSize(app.DefaultWindowWidth, app.DefaultWindowHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(app.TPS)

	var game *app.Game
	switch {
	case *skidoo:
		game = app.NewSkidoo()
	case *hazards:
		game = app.NewHazard()
	case *hut:
		game = app.NewHut()
	case *river:
		game = app.NewRiver()
	case *cover:
		game = app.NewCover()
	default:
		game = app.New(*skipTitle)
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

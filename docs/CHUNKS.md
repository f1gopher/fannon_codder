# Chunk checklist

One chunk per Grok session. Preamble is in `PLAN.md`.

- [x] 01 Bootstrap
- [x] 02 Scenes + pointer + camera shell
- [ ] 03 Troopers walk in file
- [ ] 04 Machine guns, death, friendly-fire rules
- [ ] 05 Enemy grunts + kill-all
- [ ] 06 Mission 1 playable (data + tiny jungle)
- [ ] 07 Names, ranks, recruit pool
- [ ] 08 Boot Hill, briefing, fail/retry
- [ ] 09 Mission complete: promotions, +15, save/load
- [ ] 10 Rank gun stats + HUD ammo icons
- [ ] 11 Tilemaps, scrolling, cover
- [ ] 12 Water, swimming, bridges
- [ ] 13 Split squads (Snake / Eagle / Panther)
- [ ] 14 Buildings, spawners, grenades, crates
- [ ] 15 Mission 2 content
- [ ] 16 HUD finish + overview map
- [ ] 17 Mission 3 (ice, cliffs, grenade economy)
- [ ] 18 Civilians, quicksand, mines
- [ ] 19 Mission 4 content + free grenades
- [ ] 20 Bazookas, rocket-grunts, Skidoo, Mission 5

## Log

### Chunk 01 done
Files: `go.mod`, `go.sum`, `.gitignore`, `cmd/fannon/main.go`, `internal/app/game.go`, `internal/app/scale_test.go`, `docs/ARCHITECTURE.md`, `docs/CHUNKS.md`.
Verify: `go test ./...` then `go run ./cmd/fannon`. Grey 320×256 playfield, “Fannon Codder” in the corner, window starts at 960×768. Resize should stay chunky (integer scale + letterbox). No scenes, units, or maps yet.

### Chunk 02 done
Files: `internal/app/game.go`, `internal/app/scenes.go`, `internal/input/pointer.go`, `internal/input/pointer_test.go`, `internal/sim/camera.go`, `internal/sim/camera_test.go`, `internal/render/pointer.go`.
Verify: title screen → click or Enter → green field. OS cursor hidden. Move the white arrow; hold right mouse for a crosshair. No units or shooting.

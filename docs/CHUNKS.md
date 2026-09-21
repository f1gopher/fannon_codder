# Chunk checklist

One chunk per Grok session. Preamble is in `PLAN.md`.

- [x] 01 Bootstrap
- [x] 02 Scenes + pointer + camera shell
- [x] 03 Troopers walk in file
- [x] 04 Machine guns, death, friendly-fire rules
- [x] 05 Enemy grunts + kill-all
- [x] 06 Mission 1 playable (data + tiny jungle)
- [x] 07 Names, ranks, recruit pool
- [x] 08 Boot Hill, briefing, fail/retry
- [x] 09 Mission complete: promotions, +15, save/load
- [x] 10 Rank gun stats + HUD ammo icons
- [x] 11 Tilemaps, scrolling, cover
- [x] 12 Water, swimming, bridges
- [x] 13 Split squads (Snake / Eagle / Panther)
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

### Chunk 03 done
Files: `internal/sim/world.go`, `internal/sim/unit.go`, `internal/sim/squad.go`, `internal/sim/move.go`, `internal/sim/move_test.go`, `internal/render/units.go`, `internal/app/scenes.go`.
Verify: after the title, two green squares. Left-click: leader walks there, second man follows in file and they stop without stacking. `go test ./internal/sim`. No guns yet.

### Chunk 04 done
Files: `internal/sim/combat.go`, `internal/sim/projectile.go`, `internal/sim/combat_test.go`, `internal/sim/world.go`, `internal/sim/unit.go`, `internal/render/units.go`, `internal/app/scenes.go`.
Verify: hold right (or Ctrl) to hose MG fire. The red dummy dies in one hit and stays as a dark corpse. Shooting through your own green man does not kill him. Dead troopers drop out of the file. No enemy AI yet.

### Chunk 05 done
Files: `internal/sim/ai.go`, `internal/sim/ai_test.go`, `internal/sim/objectives.go`, `internal/sim/world.go`, `internal/sim/combat.go`, `internal/app/scenes.go`.
Verify: 2 greens vs 3 reds. Reds shoot if you enter range, otherwise wander in if close. Kill all three → PHASE COMPLETE. Lose both men → PHASE FAILED. Click/Enter returns to title. No campaign or maps yet.

### Chunk 06 done
Files: `internal/sim/map.go`, `internal/sim/map_test.go`, `internal/sim/move.go`, `internal/sim/world.go`, `internal/sim/ai.go`, `internal/sim/objectives.go`, `internal/data/mission.go`, `internal/data/mission_test.go`, `internal/render/tiles.go`, `internal/app/scenes.go`, `internal/app/game.go`, `cmd/fannon/main.go`, `data/missions/campaign.json`, `data/missions/m01p01.json`.
Verify: `go run ./cmd/fannon` (or `-skip-title`). Mission 1 jungle, 2 men, 3 isolated grunts, trees block walking (slide around). Kill all → PHASE COMPLETE. Stand in the open and you can lose. No Boot Hill/names yet.

### Chunk 07 done
Files: `internal/campaign/pool.go`, `internal/campaign/ranks.go`, `internal/campaign/names.go`, `internal/campaign/pool_test.go`, `internal/render/hud.go`, `internal/app/scenes.go`, `data/names.txt`.
Verify: left HUD shows Snake logo, Pte Jools and Pte Jops, R13 remaining. Restarting a new game deploys the same two names. `go test ./internal/campaign`. No Boot Hill, promotions, or save yet.

### Chunk 08 done
Files: `internal/app/progress.go`, `internal/app/boothill.go`, `internal/app/briefing.go`, `internal/app/scenes.go`, `internal/app/game.go`, `internal/campaign/pool.go`, `internal/campaign/pool_test.go`.
Verify: Title → Boot Hill (queue of 15, 0 graves) → briefing → battle. Wipe: graves go up, queue shrinks, next attempt is Stoo & Jon. ESC returns living men to the pool. Win → Boot Hill “MISSION COMPLETE”. No save or +15 yet.

### Chunk 09 done
Files: `internal/campaign/save.go`, `internal/campaign/save_test.go`, `internal/campaign/pool.go`, `internal/campaign/ranks.go`, `internal/app/progress.go`, `internal/app/boothill.go`, `internal/app/stub.go`, `internal/app/scenes.go`.
Verify: finish M1 with no deaths → Boot Hill queue 30, Cpl Jools and Cpl Jops. SAVE/LOAD icons (save in `~/.config/fannon-codder/save.json`). Click continues to a Mission 2 stub, not a crash. No M2 map yet.

### Chunk 10 done
Files: `internal/sim/combat.go`, `internal/sim/combat_test.go`, `internal/sim/ai.go`, `internal/render/hud.go`.
Verify: HUD shows G0 and R0. A Corporal’s MG range/RoF beat a Private (`go test ./internal/sim`). Enemies stay on the grunt table (70 px, 4/s). No grenades yet.

### Chunk 11 done
Files: `internal/sim/los.go`, `internal/sim/los_test.go`, `internal/sim/combat.go`, `internal/sim/combat_test.go`, `internal/sim/ai.go`, `internal/sim/ai_test.go`, `internal/sim/world.go`, `internal/sim/camera_test.go`, `internal/app/scenes.go`, `internal/app/game.go`, `internal/render/tiles.go`, `cmd/fannon/main.go`.
Verify: `go test ./internal/sim`. `go run ./cmd/fannon -cover` — 40×30 jungle, tree wall on the first screen. Pointer at the edge pans. Hose the wall: the east grunt lives until you walk south around the trees. Water still not in.

### Chunk 12 done
Files: `internal/sim/water.go`, `internal/sim/water_test.go`, `internal/sim/map.go`, `internal/sim/move.go`, `internal/sim/combat.go`, `internal/sim/ai.go`, `internal/sim/world.go`, `internal/data/mission.go`, `internal/data/mission_test.go`, `internal/render/tiles.go`, `internal/render/units.go`, `internal/app/scenes.go`, `internal/app/game.go`, `cmd/fannon/main.go`.
Verify: `go test ./internal/sim`. `go run ./cmd/fannon -river` — river with a brown bridge. Swimmers in deep water cannot fire (hose them from the bank). Wading shallow is slow but you can still shoot. Crossing the bridge is full speed. No grenades.

### Chunk 13 done
Files: `internal/sim/squad.go`, `internal/sim/split.go`, `internal/sim/squad_test.go`, `internal/sim/combat.go`, `internal/sim/world.go`, `internal/sim/water_test.go`, `internal/render/hud.go`, `internal/render/hud_test.go`, `internal/render/units.go`, `internal/app/scenes.go`.
Verify: `go test ./...`. `go run ./cmd/fannon -river` — three troopers (green Snake). Click a name (olive bar), click the logo or the `>S` row to peel him off as blue Eagle. He holds and shoots; walk the others across the bridge. Keys `1`/`2`/`3` or click a squad's block to switch. Walk onto the other squad to merge under whoever is active. G/R icons cycle the split share: no outline = none, open outline = half, full outline = all (counts stay 0). No grenades.

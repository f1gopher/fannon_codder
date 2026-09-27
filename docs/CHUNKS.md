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
- [x] 14 Buildings, spawners, grenades, crates
- [x] 15 Mission 2 content
- [x] 16 HUD finish + overview map
- [x] 17 Mission 3 (ice, cliffs, grenade economy)
- [x] 18 Civilians, quicksand, mines
- [x] 19 Mission 4 content + free grenades
- [x] 20 Bazookas, rocket-grunts, Skidoo, Mission 5
- [x] 21 Grunts spot, turn, and hold the post
- [x] 22 Grunt bursts and a wide cone
- 23 dropped — gunfire does not wake the next post
- [x] 24 Sound bus and gunshot
- [x] 25 Blasts
- [x] 26 Death
- [x] 27 Throw and launch
- [x] 28 Terrain and vehicles (one-shots)
- [x] 29 Distance
- [x] 30 Engine loop
- [x] 31 Scene stings

Chunks 01–22 and 24–31 are in. Chunk 23 will not be built. Sound specs are in `PLAN.md` under “Chunks 24–31”.

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

### Chunk 14 done
Files: `internal/sim/building.go`, `internal/sim/grenade.go`, `internal/sim/grenade_test.go`, `internal/sim/combat.go`, `internal/sim/ai.go`, `internal/sim/move.go`, `internal/sim/objectives.go`, `internal/sim/world.go`, `internal/data/mission.go`, `internal/data/mission_test.go`, `internal/render/props.go`, `internal/app/scenes.go`, `internal/app/game.go`, `cmd/fannon/main.go`.
Verify: `go test ./...`. `go run ./cmd/fannon -hut` — brown door hut spewing reds, grey G crate off to the west. Walk onto the crate (G4). Right-hold + left click, or Space, throws; the leader only, and it arcs over trees. One grenade on the hut destroys it (MG does not). Shoot the crate and it explodes, killing whoever is in the blast, including friendlies. Kill the remaining reds → PHASE COMPLETE. Phase stays open while the door hut stands. Doorless huts are not an objective. No bazookas.

### Chunk 15 done
Files: `data/missions/campaign.json`, `data/missions/m02p01.json`, `data/missions/m02p02.json`, `internal/data/mission_test.go`, `internal/app/progress.go`, `internal/app/progress_test.go`, `internal/app/scenes.go`, `internal/app/boothill.go`, `internal/app/stub.go`, `internal/sim/camera.go`, `internal/sim/camera_test.go`.
Verify: `go test ./...`. Fresh game: M1 (2 men) → Boot Hill queue 30, Jools and Jops corporals → M2 deploys 3, HUD remaining 27. Phase 1 “Bridge Over the River Pie”: scrolling river, one bridge, trees, a swimmer near spawn, 16 grunts, kill-all. Phase 2 “Trash Enemy HQ”: mostly water, one door hut, grenade crate on the grass beside it (outside the blast), kill-all + destroy the hut. Both phases survived → those men +2 ranks, +15 recruits, then the Mission 3 stub. An old save that was stuck on the Mission 2 stub loads phase 1 instead. Camera starts centred on the squad.

### Chunk 16 done
Files: `internal/render/hud.go`, `internal/render/hud_test.go`, `internal/render/overview.go`, `internal/sim/grenade.go`, `internal/sim/special_test.go`, `internal/sim/world.go`, `internal/app/scenes.go`.
Verify: `go test ./...`. In battle the left strip shows the troop colour, G and R counts (white border on the selected special; grenades start selected), a foot icon, ranks and names, and a green bar on the active squad. Click G or R to select that special; with names highlighted, the same click still cycles split share. C toggles the special. Bazooka selected does not spend grenades. M at the bottom of the panel toggles a schematic of the whole map (tiles, hut, crate, unit dots, view box). Click the map to close it. Split still works.

### Chunk 17 done
Files: `internal/sim/map.go`, `internal/sim/move.go`, `internal/sim/cliff_test.go`, `internal/data/mission.go`, `internal/data/mission_test.go`, `internal/render/tiles.go`, `internal/render/overview.go`, `internal/app/scenes.go`, `internal/app/progress_test.go`, `data/missions/campaign.json`, `data/missions/m03p01.json`.
Verify: `go test ./...`. After Mission 2, Boot Hill opens Mission 3 “Blast It's Cold” (deploy 4, snow field). Ice walks at grass speed. A cliff row is one-way: walk south to drop, the grey ramp on the west is the way back up. Four door huts, two crates (8 grenades). Pick the crates up before you shoot; hosing them explodes the grenades and the phase cannot be finished. One grenade per hut is enough if you do not waste the boxes. Mission 4 is still the stub.

### Chunk 18 done
Files: `internal/sim/hazard.go`, `internal/sim/hazard_test.go`, `internal/sim/unit.go`, `internal/sim/move.go`, `internal/sim/world.go`, `internal/sim/objectives.go`, `internal/sim/ai.go`, `internal/sim/water.go`, `internal/sim/map.go`, `internal/data/mission.go`, `internal/data/mission_test.go`, `internal/render/tiles.go`, `internal/render/units.go`, `internal/render/overview.go`, `internal/app/game.go`, `internal/app/scenes.go`, `cmd/fannon/main.go`.
Verify: `go test ./...`. `go run ./cmd/fannon -hazards` — grass field, brown mine with a dark pip west of the squad, tan quicksand pool further east, a yellow civilian, a doorless hut to the north. Walk onto the mine: grenade-sized blast, the tile is gone. Walk into the tan pool: the trooper stops, shrinks, and dies after about two seconds (no shooting while sinking). Hose the yellow man: he dies, and the phase stays open until the red grunt up the west road is dead. The doorless hut is not a win condition. Spears are not in. `protect_civilians` parses and does not fail the phase. Mission 4 maps are still the stub.

### Chunk 19 done
Files: `data/missions/campaign.json`, `data/missions/m04p01.json`, `data/missions/m04p02.json`, `data/missions/m04p03.json`, `data/missions/m04p04.json`, `internal/data/mission.go`, `internal/data/mission_test.go`, `internal/sim/unit.go`, `internal/sim/ai.go`, `internal/sim/ai_test.go`, `internal/sim/grenade.go`, `internal/render/units.go`, `internal/render/overview.go`, `internal/app/progress_test.go`.
Verify: `go test ./...`. New game through Mission 4. Beachy Head (4 men, 5 huts, two crates, no free grenades): blow both crates and the phase cannot be finished. Pier Pressure onward, each trooper starts with 2 grenades (G8 with 4 men, G10 with 5). Village People has yellow civilians, doorless huts, and a quicksand pool; only the two door huts count. Quicksand has pools, mines, and orange grenadiers. A grenadier stops and shows a yellow bar, then throws; each carries two bombs and will not throw again for about five seconds. Mission 5 is still the stub.

### Chunk 20 done
Files: `internal/sim/rocket.go`, `internal/sim/rocket_test.go`, `internal/sim/vehicle.go`, `internal/sim/vehicle_test.go`, `internal/sim/projectile.go`, `internal/sim/combat.go`, `internal/sim/ai.go`, `internal/sim/grenade.go`, `internal/sim/world.go`, `internal/sim/hazard.go`, `internal/sim/split.go`, `internal/sim/water.go`, `internal/data/mission.go`, `internal/data/mission_test.go`, `internal/render/units.go`, `internal/render/props.go`, `internal/render/overview.go`, `internal/render/pointer.go`, `internal/app/scenes.go`, `internal/app/game.go`, `internal/app/stub.go`, `internal/app/boothill.go`, `internal/app/progress_test.go`, `cmd/fannon/main.go`, `data/missions/campaign.json`, `data/missions/m05p01.json`, `data/missions/m05p02.json`, `data/missions/m05p03.json`.
Verify: `go test ./...`. New game through Mission 5, then save. Valley of Ice: 3 men, 6 huts, an ice river, rocketeers beside trees, a grenade crate and a rocket crate. Barmy Bazookas: 3 men, 6 huts, a bridge, many rocketeers. My Beautiful Skidoo: 4 men, 3 huts, a player skidoo, each man starts with 1 rocket (R4). Select R (or press C), then right-hold and left-click or Space to fire; a rocket destroys a hut or a skidoo, and the MG does not. Pointer over the empty skidoo is a board box; left click sends the squad in (or boards immediately if they are already on it). Hold left to drive — longer hold, higher speed. On ice the skidoo keeps sliding after you release. Right fires the mounted gun when the skidoo is armed. Grenades and rockets do not fire from inside. Running someone over kills them. The enemy skidoo has a red blinker. After Mission 5, Boot Hill says the campaign continues another day. `go run ./cmd/fannon -skidoo` is a small field with a skidoo, a rocket crate, a hut, and an enemy skidoo.

### Chunk 21 done
Files: `internal/sim/ai.go`, `internal/sim/ai_test.go`, `internal/sim/unit.go`, `internal/sim/world.go`, `internal/sim/vehicle_test.go`, `internal/render/units.go`, `internal/data/mission_test.go`.
Verify: `go test ./...`. Mission 1 spawn: the three grunts stay on their tiles and do not shoot (the south man is 80px away, gun range is 70). Each has a white nose showing facing; they spawn facing east and turn toward you once you are in range with a clear line. The first round waits out about half a second, longer if they have to turn around. Step out already aiming and you can kill the south grunt before he fires. Wait in the open and he turns, then shoots, and keeps shooting — bursts are Chunk 22. Grenadiers still telegraph a throw. Rocketeers are unchanged. `go run ./cmd/fannon`.

### Chunk 22 done
Files: `internal/sim/ai.go`, `internal/sim/ai_test.go`, `internal/sim/unit.go`.
Verify: `go test ./...`. Mission 1 again. Winning the opening aim is clean. Missing it means three rounds, a visible gap of about three quarters of a second while he keeps turning, then another burst. Shots go wide of a moving man; standing still in the open still gets him killed. Grenadiers still telegraph throws. Rocketeers are unchanged. `go run ./cmd/fannon`.

### Chunk 23 dropped
Gunfire does not wake a posted man. Mission 1 stays three separate duels.

### Chunk 24 done
Files: `internal/sim/cue.go`, `internal/sim/cue_test.go`, `internal/sim/world.go`, `internal/sim/combat.go`, `internal/sim/vehicle.go`, `internal/audio/audio.go`, `internal/audio/gun.go`, `internal/audio/gun_test.go`, `internal/app/game.go`, `internal/app/scenes.go`, `docs/ARCHITECTURE.md`.
Verify: `go test ./...`. Hold right mouse: each trooper cracks, including the mounted skidoo gun (`go run ./cmd/fannon -skidoo`). Enemy bursts crack three times, then the gap is quiet. Grenades, rockets, and blasts are still silent.

### Chunk 25 done
Files: `internal/sim/cue.go`, `internal/sim/cue_test.go`, `internal/sim/grenade.go`, `internal/audio/audio.go`, `internal/audio/boom.go`, `internal/audio/boom_test.go`, `docs/ARCHITECTURE.md`.
Verify: `go test ./...`. A grenade, a rocket, and a mine each boom once when they go off. Hosing a crate booms, and a crate it sets off booms again. The throw and the launch are still silent; death is still silent. The gun crack is unchanged.

### Chunk 26 done
Files: `internal/sim/cue.go`, `internal/sim/cue_test.go`, `internal/sim/combat.go`, `internal/audio/audio.go`, `internal/audio/yell.go`, `internal/audio/yell_test.go`, `docs/ARCHITECTURE.md`.
Verify: `go test ./...`. Shooting a grunt yells when he drops. A grenade that kills two men booms once and yells twice, at different pitches when their ids differ. A man who sinks in quicksand yells once. Calling kill again is silent. The throw and the launch are still silent.

### Chunk 27 done
Files: `internal/sim/cue.go`, `internal/sim/cue_test.go`, `internal/sim/grenade.go`, `internal/sim/rocket.go`, `internal/audio/audio.go`, `internal/audio/whoosh.go`, `internal/audio/whoosh_test.go`, `docs/ARCHITECTURE.md`.
Verify: `go test ./...`. The yellow telegraph is silent. The whoosh is the moment the bomb leaves, then the boom is the landing. A bazooka whooshes at the tube and booms on impact. Player and enemy share each clip.

### Chunk 28 done
Files: `internal/sim/cue.go`, `internal/sim/cue_test.go`, `internal/sim/unit.go`, `internal/sim/water.go`, `internal/sim/hazard.go`, `internal/sim/grenade.go`, `internal/sim/vehicle.go`, `internal/audio/audio.go`, `internal/audio/terrain.go`, `internal/audio/terrain_test.go`, `docs/ARCHITECTURE.md`.
Verify: `go test ./...`. Walking into the river splashes once; standing in it, and wading into the deep channel, stays quiet. Leaving the water is silent. The tan pool gulps once when it sticks, and the death at the end is still the yell. A grenade crate and a rocket crate each click once when picked up. Boarding the skidoo and getting off each answer once.

### Chunk 29 done
Files: `internal/audio/distance.go`, `internal/audio/distance_test.go`, `internal/audio/audio.go`, `internal/app/game.go`, `internal/app/scenes.go`, `internal/app/hear_test.go`, `internal/sim/cue.go`, `docs/ARCHITECTURE.md`.
Verify: `go test ./...`. A shot at the leader's feet is full loudness. The same crack from the far edge of a scrolling map is faint, and past 320 px it is silent. Title and briefing stings are not in yet (Chunk 31); when they arrive they stay full volume. `go run ./cmd/fannon`.

### Chunk 30 done
Files: `internal/sim/engine.go`, `internal/sim/engine_test.go`, `internal/audio/engine.go`, `internal/audio/engine_test.go`, `internal/audio/audio.go`, `internal/app/game.go`, `internal/app/scenes.go`, `docs/ARCHITECTURE.md`.
Verify: `go test ./...`. Board the skidoo (`go run ./cmd/fannon -skidoo`) and the hum starts at idle. Holding left builds the pitch as the skidoo speeds up. Letting go on grass drops it back to that idle hum. Letting go on ice keeps a lower hum while it slides. Getting off cuts it, unless another occupied skidoo is within 320 px. An empty or destroyed skidoo is silent.

### Chunk 31 done
Files: `internal/sim/cue.go`, `internal/sim/cue_test.go`, `internal/audio/sting.go`, `internal/audio/sting_test.go`, `internal/audio/audio.go`, `internal/app/game.go`, `internal/app/scenes.go`, `internal/app/briefing.go`, `docs/ARCHITECTURE.md`.
Verify: `go test ./...`. Title → click or Enter clicks, then Boot Hill is quiet. The briefing clicks when it deploys. Clearing a phase plays a short rising sting as Boot Hill opens. Escape, or clicking through a wipe, plays a short falling sting. HUD icons do not click. `go run ./cmd/fannon`.

## Backlog

Grunt behaviour is Chunks 21–22. Chunk 23 (hearing) is dropped. Do not add chase or pathfinding.

- Jeeps (skin of the skidoo), tanks, static turrets, choppers
- Hostages, kidnap, factories, protect-civilians fail
- Desert, moors, and underground tiles
- Missions 6–24 as data
- Wounded troopers and finishing them; corpse juggling
- Pixel art
- Title tune (new, not ripped). Battle sound effects, Chunks 24–31, are in.
- Birds, snowmen, igloos
- High Scoring Heroes
- Fullscreen and integer-scale options
- Headless sim replay for each phase

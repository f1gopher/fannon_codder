# Fannon Codder — Architecture

Gameplay clone of Cannon Fodder (Amiga, 1993). Product name: **Fannon Codder**.

Read this file at the start of every implementation chunk. Do not invent a new architecture.

## Display

The world pixel space did not change. The frame is still **320×256**. A tile is 16 world pixels, a trooper is an 8-pixel body, and the playfield is the 268×256 view beside the 52-pixel status strip. The camera does not zoom. Audio distances stay in world pixels. A larger window is a sharper picture of that same frame, not a wider view of the map.

The window opens at **1024×768** device-independent pixels. `SetWindowSizeLimits` keeps it inside **1024×768 … 3840×2160**. The 320×256 frame is fitted uniformly (it is not stretched to the window). On the windows this range produces, height is the fit: 1024×768 shows a 960×768 picture with 32 pixels of bar on each side (scale 3), and 3840×2160 shows a 2700×2160 picture (scale 8.4375) with bars on the sides.

`Layout` / `LayoutF` receive the window in device-independent pixels and return the offscreen, `320×S` by `256×S`. `S` is that fit times `Monitor().DeviceScaleFactor()`, clamped so the offscreen never exceeds 3840×2160. `DrawFinalScreen` clears the bars and blits the offscreen at 1:1 in the centre. The filter on that blit is nearest, because the scale is 1. Placeholders are drawn into this offscreen — logical coordinate times `S` — and not painted into a 320×256 buffer and scaled up.

HUD text is `ebiten/v2/text/v2` with the Go regular face. The size is 14 pixels when `S` is 3, and it scales with `S`. The OS cursor stays hidden. The pointer, the sim, the HUD hit tests, and edge scroll see the cursor divided by `S` and clamped to the frame. A click in the bar does not reach the game.

A painted sheet is drawn at `S/8` with a linear filter. The blit of the offscreen stays nearest. `8` is source pixels per world pixel. A body with no sheet draws its coloured rectangle.

## Loop

- `ebiten.TPS = 60`.
- `Update` samples input, steps `sim.World` once, then advances the picture clock.
- `Draw` paints the picture from sim state. **No gameplay in Draw.**

## Packages

```
cmd/fannon/          # main, window, RunGame
internal/app/        # ebiten.Game, scene stack, scaling
internal/input/      # pointer, buttons, both-buttons chord
internal/sim/        # world, units, combat, AI, objectives — NO ebiten import
internal/campaign/   # recruits, names, ranks, save JSON — NO ebiten import
internal/render/     # picture: sheets, or the coloured rectangle when a sheet is missing
internal/audio/      # cue playback — the only package that opens the audio device
internal/data/       # load mission JSON
assets/placeholder/  # later: real PNGs with the same filenames
data/missions/       # mission JSON
data/names.txt
docs/ARCHITECTURE.md
docs/CHUNKS.md
```

`internal/sim` must not import Ebitengine or `internal/audio`. Simulation is unit-tested with `go test`.

## Audio

The sim records `Cue` values (`Kind`, world `X`, `Y`, and `ID` when a sound depends on who made it) and does not play them. The battle scene takes the queue after `Step` and passes it to `internal/audio`. One mixer owns the process-wide Ebitengine context (44100 Hz, 16-bit stereo). A kind with no clip loaded is silent.

The mixer loads the gunshot (8 voices), the blast (4 voices), the death yell (4 voices at each of three pitches, chosen from the unit id), the grenade whoosh (4 voices), the rocket whoosh (3 voices), one voice each for the splash, the quicksand gulp, the crate pickup, boarding, and dismount, and one voice each for the menu click, the win sting, and the fail sting. Another effect is a `CueKind`, an emit at the cause, and `Mixer.Load(kind, pcm, voices, volume)`. Each clip keeps its own full-loudness level. The battle calls `SetListener` at the active leader (the driven vehicle counts), or at the camera centre when that squad has no leader. A world cue is full inside 48 px and falls linearly to silence at 320 px; a cue beyond that does not take a voice. There is no pan. Stings that have no world cause use `PlayKind` and stay at the clip's full volume. The title plays a click when it advances, and the briefing plays one when it starts the phase. Leaving a cleared phase plays the win sting. Leaving after a wipe plays the fail sting. Escape opens a quit prompt on every screen; Y or Enter ends the process, and N or Escape closes the prompt. The world stays frozen while the prompt is up. Boot Hill plays nothing. The world never emits those three kinds. The skidoo hum is not a cue. One looping voice on the mixer starts, stops, and changes pitch. It runs while the listener is aboard a living skidoo, or while the nearest other occupied skidoo is inside 320 px, and it is silent when that vehicle is empty or destroyed. Pitch is idle at a standstill and top at `VehicleMaxSpeed`; the listener's own vehicle wins. Ebitengine players have no playback-rate control, so the loop reader resamples a synthesised cycle. Samples are original synthesis or original recordings, never Amiga rips. There is one mixer per process.

## Core types (sim)

- `World`: map, units, projectiles, buildings, pickups, vehicles, camera, objectives, rng.
- `Unit`: id, name, rank, side (player/enemy/civilian), hp (alive / wounded / dead), pos, vel, facing, inWater, squadID, vehicleID, kills.
- `Squad`: id (Snake / Eagle / Panther), leader, member IDs in rank order, grenades, rockets, active.
- `Map`: tile grid, tile size 16×16 world pixels. Tiles: grass, tree, waterShallow, waterDeep, ice, quicksand, cliff, bridge, mine.
- `Building`: pos, hasDoor (spawner vs flavour hut), hp, spawnInterval.
- `Pickup`: grenadeCrate(4) | rocketCrate(4). Destroying it explodes.
- `Vehicle`: skidoo (later jeep as the same type with a skin).
- `Objective`: KillAllEnemy, DestroyEnemyBuildings (both can be required). ProtectCivilians parses and is ignored until a later chunk.

World space is float64 pixels. Units are ~8×8. One MG bullet kills a healthy infantry unit.

## Controls (Amiga mouse)

| Input | Effect |
|---|---|
| Move mouse | Move pointer. Near playfield edge, scroll camera. |
| Left click | Active squad leader walks toward pointer tip. Others follow in file. |
| Right held | Crosshair. Whole active squad fires MG at the crosshair. Move and fire are independent. |
| Right held + left click | Leader throws grenade or fires bazooka at the crosshair. |
| Click grenade/bazooka icon | Toggle special weapon. |
| Click names in HUD | Highlight members. Click troop logo → split into Eagle/Panther. |
| Walk two player squads together | Merge. |
| ESC | Surrender phase. Survivors return to the pool. |
| P | Pause. |

Player MG does **not** harm living friendlies. Explosives and vehicles kill everyone. Wounded friendlies on the ground can be finished by MG.

## Enemy infantry

Grunts hold the tile they spawned on (facing east). They notice a player only inside gun range with clear LOS. They turn at a limited rate and wait out a short reaction before the first round, then fire a short burst, pause, and repeat. The enemy cone is wider than a Private’s and is not the player’s `World.Spread`. Grenadiers and rocketeers keep their own windups; those windups are not stacked on the grunt reaction. A grenadier’s point-blank MG uses the grunt rules. No pathfinding and no chase. Numbers and the hearing follow-up are Chunks 21–23 in `PLAN.md`.

Inactive player squads still fire as soon as they have a target. Enemy vehicles are a separate rule.

## Camera

Pointer-driven edge scroll, clamped to the map. The view is the playfield beside the status strip, so a map the width of the full screen still scrolls by the strip's width. The active leader stays inside a soft inner margin.

## Campaign

- Start: 15 conscripts. Mission 1 deploys 2 → 13 remaining.
- After each **mission**: +15 recruits; survivors promoted **one rank per phase survived**.
- After every three missions, new recruits arrive already trained.
- Max on-field squad: briefing number, never more than six.
- Save only on Boot Hill after a full mission (`userConfig/fannon-codder/save.json`).
- Run out of recruits → game over.

Ranks (low → high): Private, Corporal, Sergeant, Staff Sergeant, Sergeant First Class, Master Sergeant, Sergeant Major, Specialist 4, Specialist 6, Warrant Officer, Chief Warrant Officer, Captain, Major, Colonel, Brigadier General, General.

Rank affects MG range, accuracy (spread), and rate of fire (Chunk 10). Until then all ranks shoot the same.

## Sprites

Sheets live in `assets/art` as one PNG and one JSON per animation. `internal/render` embeds that tree. The style board has no JSON, so it is not drawn. The key is the path without the extension.

JSON fields are `frameW`, `frameH`, `anchorX`, `anchorY`, `fps`, `loop`, `rows`, and `mirrors`. `rows` is the order of rows in the PNG. A mirror names a stored row and is not stored itself. Directions are `E, SE, S, SW, W, NW, N, NE`. Facing `0` is east, `π/2` is south, `π` is west, and `−π/2` is north. Paint E, SE, S, N, and NE. Mirror E→W, SE→SW, and NE→NW.

The anchor sits on the sim point the rectangle uses: a trooper, crate, or skidoo position; the bottom centre of a hut; the bottom centre of a tree's cell. A ground frame pins its top-left to the cell. Draw order is ground, a soft oval under each painted body, then trees, huts, crates, men, and vehicles by foot Y, then grenades, tracers, and blasts, then the HUD and the pointer. Scenery frames are offset by `tx*3+ty*5`.

Pose, first match: death (once, then the corpse frame), sink (scrubbed by `Sink/SinkTime`), swim, throw (scrubbed across the windup or the 0.25s after launch), shoot (while `SinceShot < 0.12`), walk (speed above 2 px/s), idle. `SinceShot` resets in `addMG`. `SinceThrow` resets in `launchGrenade` and `launchRocket`. Both count up each step and do not change combat. Death time is kept in the renderer by unit id. The battle `Update` advances the clock.

Keys: `snake|eagle|panther|grunt|grenadier|rocketeer|civilian` with `idle|walk|shoot|throw|death|corpse|swim|sink`. Ground: `ground/grass`, `ground/snow`, `ground/water-shallow`, `ground/water-deep`, `ground/quicksand`, `ground/ice`, `ground/bridge`, `ground/cliff`, `ground/ramp`, `ground/mine`. Props: `tree/sway`, `hut/door`, `hut/plain`, `crate/grenade`, `crate/rocket`, `skidoo/idle`, `skidoo/move`.

`ground/grass` and `ground/snow` are in: four frames, 8 fps, looping, 128×128. The battle uses snow when the phase terrain is `arctic`, and grass otherwise. Grass and tree cells take that loop. `ground/water-shallow` and `ground/water-deep` are eight-frame loops at 8 fps. `ground/quicksand` and `ground/ice` are six-frame loops at 4 fps. All four are 128×128 and pinned at the cell’s top-left. Neighbouring cells of the same kind are not on the same frame.

`ground/cliff`, `ground/ramp`, and `ground/bridge` are one frame each, 128×128, pinned at the cell’s top-left. `ground/mine` is four frames at 2 fps: three still frames and one glint, on the grass field. `tree/sway` is three silhouettes (rows E, SE, S — labels only, the tree has no facing), each a four-frame loop at 4 fps. The cell is 256×288. The anchor is the trunk foot, at (128, 282), on the bottom centre of the blocked cell. The battle picks the row with `(tx+ty*2) mod 3` and the column with the usual scenery phase, so a forest does not share one shape or one sway.

`hut/door` and `hut/plain` are four-frame chimney-smoke loops at 5 fps. The cell is 256×320, and the anchor (128, 312) is the bottom centre of a 2×2 hut. `crate/grenade` and `crate/rocket` are one frame each, 128×128, anchor at the centre (64, 64), and neither painting carries a letter. `skidoo/idle` and `skidoo/move` are four frames, cell 240×176, anchor (120, 88) on the vehicle point. Stored rows are E, SE, S, N, NE; W, SW, and NW are mirrors. Move runs at 8 fps and idle at 4. The enemy lamp is still drawn on top of that body.

`fx/flash` is two frames over the 0.12s shot window, drawn at a man’s muzzle and at a skidoo’s gun while `SinceShot` is inside that window. `fx/tracer` and `fx/rocket` are one frame each and rotate with velocity. `fx/grenade` is a two-frame tumble at 8 fps, drawn on the arc. `fx/blast` is four frames at 20 fps (0.2s) and is the only explosion. Grenades, rockets, mines, and crates all use it. The marker in the sim still lasts 0.35s; the painting stops at the end of the sheet.

The battle cursor is `ui/pointer` (the style-board arrow, hotspot on the tip), `ui/crosshair`, and `ui/board` (the skidoo, hotspot in the middle). They are drawn at `S/8`. The exit mark on a skidoo you already occupy stays the small bitmap. The status strip stays the flat dark panel. Its icons are `ui/grenade`, `ui/rocket`, `ui/foot`, `ui/vehicle`, `ui/map`, and `ui/mark-snake`, `ui/mark-eagle`, `ui/mark-panther`, also at `S/8`. Names, ranks, and the G and R counts stay text. The selected special is still the white stroke. The overview keeps its diagram and colours each tile from the accepted ground painting. Arctic grass and tree cells use the snow colour.

## Placeholders

| Thing | Draw |
|---|---|
| Player trooper | Small green rectangle + rank letter + name on HUD |
| Enemy grunt | Red rectangle |
| Civilian | Yellow rectangle |
| Corpse | Darker rectangle |
| Tree | Dark green square (blocks bullets/LOS) |
| Deep water | Dark blue |
| Shallow water | Light blue |
| Building | Brown rect; black square = door |
| Grenade crate | Grey box labelled G |
| Rocket crate | Grey box labelled R |
| Skidoo | White/grey rounded rect |
| Pointer | White arrow; crosshair when firing |
| HUD | Black panel, debug text |

Same filenames under `assets/placeholder/` so art is a later drop-in.

## Mission data

JSON, one file per phase, listed by `data/missions/campaign.json`. Maps are original layouts inspired by original mission beats, not ripped Amiga data.

## Legal

Do not extract or embed original Amiga graphics, samples, or map binaries. Do not ship the original theme tune.

## Rules for implementers

- One chunk only. No extra systems.
- `go test ./...` and `go build ./cmd/fannon` must pass.
- Update `docs/CHUNKS.md` when a chunk is done.

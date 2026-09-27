# Fannon Codder — Architecture

Gameplay clone of Cannon Fodder (Amiga, 1993). Product name: **Fannon Codder**.

Read this file at the start of every implementation chunk. Do not invent a new architecture.

## Display

- Logical screen **320×256**. Integer scale (2×/3×/4×) into the window with letterboxing. Nearest-neighbour. No rotation, no camera zoom.
- Default window 960×768 (3×). Resizable.
- Playfield is most of the screen. Status panel is a **left strip** (~48–56 px). Not drawn until a later chunk.
- Hide the OS cursor and draw our own pointer (Chunk 02).

## Loop

- `ebiten.TPS = 60`.
- `Update` samples input, steps `sim.World` once.
- `Draw` paints placeholders from sim state. **No gameplay in Draw.**

## Packages

```
cmd/fannon/          # main, window, RunGame
internal/app/        # ebiten.Game, scene stack, scaling
internal/input/      # pointer, buttons, both-buttons chord
internal/sim/        # world, units, combat, AI, objectives — NO ebiten import
internal/campaign/   # recruits, names, ranks, save JSON — NO ebiten import
internal/render/     # placeholder draw of sim + HUD
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

The mixer loads the gunshot (8 voices), the blast (4 voices), the death yell (4 voices at each of three pitches, chosen from the unit id), the grenade whoosh (4 voices), and the rocket whoosh (3 voices). Another effect is a `CueKind`, an emit at the cause, and `Mixer.Load(kind, pcm, voices, volume)`. Position is stored now; volume does not use it yet. Each clip keeps its own full-loudness level. Loops (an engine) are not cues. Stings that have no world cause use `PlayKind`. Samples are original synthesis or original recordings, never Amiga rips. There is one mixer per process.

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

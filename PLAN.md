# Fannon Codder — Implementation Plan

A Cannon Fodder (Amiga, 1993, Sensible Software) **gameplay clone** in Go + [Ebitengine v2](https://ebitengine.org/). Working title: **Fannon Codder**. Repo: `/home/andy/code/fannon_codder`.

This plan is a sequence of **small subsystems**. You paste **one chunk per Grok session**. Do not ask Grok to “build the game” or to do two chunks at once.

## Goal for this plan

A playable clone of the **engine plus missions 1–5** of the Amiga original, close in controls, structure, and feel.

| Mission | Original name | Phases | New systems this mission forces |
|---|---|---|---|
| 1 | The Sensible Initiation | 1 | Mouse pointer, file-follow, MG, 3 grunts, kill-all, tiny jungle |
| 2 | Onward Virgin Soldiers | 2 | Scrolling map, water/swim, bridges, trees, spawner hut, grenades |
| 3 | Antarctic Adventure | 1 | Ice tiles, cliffs, grenade crates as scarce resource, 4 huts |
| 4 | Super Smashing Namtastic | 4 | Start-with-grenades, civilians, door-vs-hut, quicksand, mines |
| 5 | Those Vicious Vikings | 3 | Bazookas, rocket-grunts, Skidoo (board/drive/ram) |

Later missions (6–24: jeeps, tanks, choppers, Biggunz, hostages, desert/moors/underground) are **out of scope** for this plan. The engine is built so they can be added as data + a few extra entity types.

## Constraints you already chose

- **Finish line:** engine + missions 1–5, not the full 72-phase campaign.
- **Art/audio:** coloured placeholders first. Swap real Sensible-style pixel art later without rewriting simulation.
- **Reference:** public longplays and the Amiga manual, not an emulator in-hand. Prefer Amiga footage when it disagrees with Mega Drive/SNES.
- **Chunk size:** one small subsystem per Grok session.

## Legal (non-negotiable)

Cannon Fodder, its graphics, music (“War Has Never Been So Much Fun”), map binaries, and names-as-product are copyrighted.

- **Do** recreate mechanics, control scheme, mission *structure*, and anti-war tone.
- **Do** design **original** maps that teach the same beats (tiny opener, river+bridge, spawner hut, grenade economy, civilians, quicksand, Skidoo).
- **Do not** extract or embed Amiga ADF/IPF graphics, samples, or map data.
- **Do not** ship the original theme tune or sprite sheets.
- Keep the product name **Fannon Codder**. Mission titles may homage the originals inside the game (parody/clone convention).

## How to run a chunk

Every session, paste **Preamble + the one chunk block**. After it lands, you play it, then start a **new** session for the next chunk.

### Preamble (paste every time)

```
You are implementing Fannon Codder, a Cannon Fodder (Amiga 1993) gameplay clone.

Stack: Go 1.22+, Ebitengine v2 (`github.com/hajimehoshi/ebiten/v2`).
Repo root: the current workspace.

Before writing code:
1. Read docs/ARCHITECTURE.md and docs/CHUNKS.md (if they exist).
2. Read only the files this chunk names, plus their direct neighbours.
3. Implement THIS CHUNK ONLY. Do not start the next chunk. Do not “while we’re here” extra systems.
4. Simulation lives in `internal/sim` and must be testable with `go test` (no Ebitengine in sim tests).
5. Rendering is placeholders: rectangles, circles, debug labels. No pixel-art generation this chunk unless the chunk says so.
6. Match Amiga behaviour described in the chunk. When unsure, pick the documented Amiga manual behaviour.
7. When done: run `go test ./...` and `go build ./cmd/fannon`. Fix failures.
8. Append a short “Chunk N done” note to docs/CHUNKS.md (checkbox + files touched + how to verify in-game).

Do not add README fluff, extra dependencies, or a new architecture. Follow docs/ARCHITECTURE.md.
```

If a session is going long, tell Grok: **stop at a compiling checkpoint and list what is unfinished**. Resume that list as the next session rather than widening scope.

### Verify each chunk

1. `go test ./...`
2. `go build -o /tmp/fannon ./cmd/fannon` and run it.
3. Play the acceptance steps in the chunk.
4. Only then start the next session.

Reference longplays (Amiga preferred): search YouTube for “Cannon Fodder Amiga longplay”. Mega Drive longplays (e.g. DarkEvil87) are acceptable for mission flow when Amiga footage is missing.

---

## Architecture (lock this in Chunk 01)

Keep this small. Future Grok sessions will re-read `docs/ARCHITECTURE.md`, not this whole plan.

### Display

- Logical screen **320×256** (Amiga PAL-ish). Integer scale 2×/3×/4× to the window. Nearest-neighbour. No rotation, no camera zoom.
- Playfield is most of the screen. Status panel is a **left strip** (~48–56 px), matching the Amiga HUD.
- Hide the OS cursor. Draw our own pointer.

### Loop

- `ebiten.TPS = 60`.
- `Update` samples input, steps `sim.World` once.
- `Draw` paints placeholders from sim state. **No gameplay in Draw.**

### Packages

```
cmd/fannon/          # main, window, RunGame
internal/app/        # ebiten.Game, scene stack, scaling
internal/input/      # pointer, buttons, both-buttons chord
internal/sim/        # world, units, combat, AI, objectives — NO ebiten import
internal/campaign/   # recruits, names, ranks, save JSON — NO ebiten import
internal/render/     # placeholder draw of sim + HUD
internal/data/       # load mission JSON
assets/placeholder/  # later: real PNGs with the same filenames
data/missions/       # mission_01.json …
data/names.txt       # ~400 first names
docs/ARCHITECTURE.md
docs/CHUNKS.md
```

`internal/sim` must not import Ebitengine. That is the testability rule. Break it and later chunks become expensive.

### Core types (sim)

- `World`: map, units, projectiles, buildings, pickups, vehicles, camera, objectives, rng.
- `Unit`: id, name, rank, side (player/enemy/civilian), hp (alive / wounded / dead), pos, vel, facing, inWater, squadID, vehicleID, kills.
- `Squad`: id (Snake / Eagle / Panther), leader, member IDs in rank order, grenades, rockets, active.
- `Map`: tile grid, tile size 16×16 world pixels. Tiles: grass, tree, waterShallow, waterDeep, ice, quicksand, cliff, bridge, mine.
- `Building`: pos, hasDoor (spawner vs flavour hut), hp, spawnInterval.
- `Pickup`: grenadeCrate(4) | rocketCrate(4). Destroying it explodes.
- `Vehicle`: skidoo (and later jeep as same type with a skin).
- `Objective`: KillAllEnemy, DestroyEnemyBuildings (both can be required).

World space is float64 pixels. Units are ~8×8. One MG bullet kills a healthy infantry unit.

### Controls (Amiga mouse)

| Input | Effect |
|---|---|
| Move mouse | Move pointer. Near playfield edge, **scroll camera** so you can look around the active squad. |
| Left click | Active squad leader walks toward pointer tip. Others follow **in file** (a line, not a blob). |
| Right held | Pointer becomes a crosshair. Whole active squad fires MG at the crosshair. Movement and fire direction are independent. |
| Right held + left click | Leader throws grenade **or** fires bazooka at the crosshair (whichever special is selected). |
| Click grenade/bazooka icon | Toggle special weapon. |
| Click names in HUD | Highlight members. Click troop logo → split into Eagle/Panther. |
| Walk two player squads together | Merge. |
| ESC | Surrender phase. Survivors return to the pool. |
| P | Pause. |

Player MG does **not** harm living friendlies. Explosives and vehicles kill everyone. Wounded friendlies on the ground **can** be finished by MG (Amiga manual).

### Camera

Pointer-driven edge scroll, clamped to the map. The view is the playfield beside the status strip, so a map the width of the full screen still scrolls by the strip's width. Keep the active leader inside a soft inner margin.

### Campaign numbers (Amiga)

- Start: 15 conscripts. Mission 1 deploys 2, so HUD “recruits remaining” is 13.
- After each **mission** (not each phase): +15 new recruits; survivors are **promoted one rank per phase they lived through**, applied at mission end.
- After every **three missions**, new recruits arrive already trained (higher starting rank).
- Max squad on a phase is the briefing number (2–6 in this plan). Wikipedia/Amiga: never more than six on the field.
- Save only on Boot Hill after a **full mission**. JSON file in the user config dir.
- Run out of recruits → game over.

Ranks (low → high): Private, Corporal, Sergeant, Staff Sergeant, Sergeant First Class, Master Sergeant, Sergeant Major, Specialist 4, Specialist 6, Warrant Officer, Chief Warrant Officer, Captain, Major, Colonel, Brigadier General, General.

Rank affects MG **range, accuracy (spread), and rate of fire**. Implement the stat curve in Chunk 10; until then all ranks shoot the same.

### Placeholders

| Thing | Draw |
|---|---|
| Player trooper | Small green rectangle + 1-letter rank + name on HUD |
| Enemy grunt | Red rectangle |
| Civilian | Yellow rectangle |
| Corpse | Darker rectangle, optional “blood” pixels |
| Tree | Dark green square, blocks bullets/LOS |
| Deep water | Dark blue |
| Shallow water | Light blue |
| Building | Brown rect; black square = door |
| Grenade crate | Grey box labelled G |
| Rocket crate | Grey box labelled R |
| Skidoo | White/grey rounded rect |
| Pointer | White arrow; crosshair when firing |
| HUD | Black panel, coloured icons, text via `ebiten/v2/text/v2` or `ebitenutil.DebugPrint` |

Same filenames under `assets/placeholder/` so a later art chunk is a drop-in.

### Mission data

JSON, one file per phase, listed by a campaign index:

```json
{
  "mission": 1,
  "phase": 1,
  "title": "It's a Jungle Out There",
  "briefing": "With 2 soldiers you must kill all enemy",
  "deploy": 2,
  "terrain": "jungle",
  "map": { "w": 20, "h": 16, "tiles": "....." },
  "playerStart": [10, 12],
  "enemies": [{"x": 14, "y": 4, "kind": "grunt"}],
  "buildings": [],
  "pickups": [],
  "vehicles": [],
  "objectives": ["kill_all_enemy"]
}
```

Maps are **original layouts** inspired by walkthrough beats, not ripped.

---

## Chunks

Do them in order. A chunk that says “play Mission N” assumes all previous chunks exist.

### Chunk 01 — Bootstrap

**Goal.** Window opens at integer-scaled 320×256, shows a grey playfield + “Fannon Codder” text, `go test` and `go build` work.

**Create**

- `go.mod` module `fannon-codder`
- `cmd/fannon/main.go`
- `internal/app/game.go` — `ebiten.Game`: `Update`/`Draw`/`Layout` (Layout returns 320, 256)
- `docs/ARCHITECTURE.md` — copy the Architecture section above, trimmed
- `docs/CHUNKS.md` — checklist of chunks 01–20 with empty boxes
- `.gitignore` — binaries, `.idea`, `*.exe`

**Rules.** Ebitengine v2 only. Window title `Fannon Codder`. Default window 960×768 (3×). `SetWindowResizingModeEnabled`. Nearest-neighbour (do not use linear filter).

**Done when.** `go run ./cmd/fannon` shows a 320×256 logical screen scaled up. Resize keeps integer feel (letterbox if needed, or snap scale).

**Do not.** Scenes, units, maps.

---

### Chunk 02 — Scenes + pointer + camera shell

**Goal.** Scene stack: Title → (placeholder) Battle. Mouse pointer drawn. Camera struct exists.

**Create**

- `internal/app/scenes.go` — `Scene` interface (`Update`, `Draw`, `Enter`, `Leave`)
- `internal/input/pointer.go` — pos in logical pixels, `Left`, `Right`, `LeftDown`, `RightDown`, `ChordGrenade` (right held, left just pressed)
- `internal/sim/camera.go` — position, clamp to map, `ScrollToward(pointer, screen, dt)`
- `internal/render/pointer.go`

Title scene: click or press Enter → Battle scene (empty green field, movable pointer). Battle: pointer changes to a crosshair while right is held.

**Done when.** You can move a custom cursor around a green field. Right-click shows a crosshair. No OS cursor.

**Do not.** Units, shooting, HUD.

---

### Chunk 03 — Troopers walk in file

**Goal.** One player squad of N green rectangles. Left click sets a destination. Leader steers toward it. Followers keep a short spacing in a line behind the leader’s recent path (rank and file).

**Create**

- `internal/sim/world.go`, `unit.go`, `squad.go`, `move.go`
- `internal/sim/move_test.go` — leader reaches destination; followers stay behind, not stacked on the leader
- `internal/render/units.go`

Hardcode 2 units in the battle scene for now. Speed ~30 px/s (tune to “Sensible stick-man brisk walk”). Arrival radius ~4 px. No pathfinding yet: steer directly; later chunks add collision.

**Done when.** Left click: both men walk, second trails the first. They stop at the point. `go test ./internal/sim` covers follow distance.

**Do not.** Enemies, guns, tiles.

---

### Chunk 04 — Machine guns, death, friendly-fire rules

**Goal.** Right-hold fires. Bullets are hitscan or very fast projectiles toward the crosshair. One hit kills. Player MG does not damage living player units. Corpses stay.

**Create**

- `internal/sim/combat.go`, `projectile.go`, `combat_test.go`
- Muzzle: all **living members of the active squad** fire, not only the leader.
- Spread and rate of fire: constants for now (`RoF = 8 shots/sec/unit`, tiny spread).
- Facing: when firing, unit faces the crosshair; when only walking, faces movement.
- Death: `Unit.Dead`, draw as a darker rect. Dead units are removed from the file but left on the map.

**Done when.** Two green men can hose a direction. A temporary red dummy (hardcoded) dies in one hit and stays down. Shooting through your own man does not kill him.

**Do not.** Enemy AI, grenades, objectives.

---

### Chunk 05 — Enemy grunts + kill-all

**Goal.** Red grunts with the same body rules. Simple AI: if a player unit is in range and has line of sight (always true until trees), face and shoot; else idle or slowly approach if close. Phase wins when no enemies are alive.

**Create**

- `internal/sim/ai.go`, `ai_test.go`, `objectives.go`
- Enemy MG: slower RoF and slightly worse range than a Private, so Mission 1 is easy.
- Win: `World.Status = Won`. Battle scene shows `PHASE COMPLETE` and waits for click.

**Done when.** 2 greens vs 3 reds on an empty field. Killing all three shows PHASE COMPLETE. If both greens die, `Status = Lost`.

**Do not.** Campaign, maps, HUD names.

---

### Chunk 06 — Mission 1 is playable (data + tiny jungle)

**Goal.** Load `data/missions/m01p01.json`. Original-feeling one-screen jungle: a few tree tiles (collision + later LOS), 2 deploy, 3 grunts, kill-all. This is the first time the game is “a level”.

**Create**

- `internal/data/mission.go` — JSON loader
- `internal/sim/map.go` — tile grid, walkable test (trees solid)
- `data/missions/campaign.json` — list of phases
- `data/missions/m01p01.json`
- Movement: steer around solid tiles with a cheap trick (slide along, or 8-dir nudging). **Not** full A*. Amiga troopers are stupid-direct; sliding is closer.

Map beats to hit (do not copy the original pixel map): small clearing, trees as cover blobs, 3 isolated grunts, everything fits in 320×256 minus HUD.

**Done when.** `go run ./cmd/fannon` skips title via a debug flag or title click, deploys 2 men, you kill 3 enemies among trees, PHASE COMPLETE. Losing is possible if you stand in the open.

**Do not.** Boot Hill, names, grenades, scrolling.

---

### Chunk 07 — Names, ranks, recruit pool

**Goal.** Troopers are people. Mission 1 deploys the first 2 names from a 400-name list. HUD lists them. A pool of 15 exists; remaining = 13 after deploy.

**Create**

- `internal/campaign/pool.go`, `ranks.go`, `names.go`
- `data/names.txt` — ~400 short first names. Seed with Sensible in-jokes **plus** original-feeling English names (Jools, Jops, Stoo as the first three is a nice homage; rest original).
- Deploy: take the **highest rank** available, then FIFO of remaining conscripts.
- HUD (left strip): squad logo placeholder, names top-to-bottom (leader first) with rank abbreviation.

**Done when.** HUD shows two names. Restarting the process deploys the same first two from a fresh pool. `go test` for pool math (15 start, deploy 2 → 13 remaining).

**Do not.** Boot Hill drawing, promotions, save.

---

### Chunk 08 — Boot Hill, briefing, fail/retry

**Goal.** The Amiga loop: Boot Hill → briefing → battle → back.

**Create**

- Scenes: `Title`, `BootHill`, `Briefing`, `Battle`
- Boot Hill: green hill placeholder, a queue of little men, grave markers (count = cumulative player deaths), “Mission N” label, click to start.
- Briefing: mission/phase title, “With N soldiers you must …”, click to deploy.
- Battle lose or ESC: survivors (if ESC) return to pool; dead are gone forever; briefing for the **same phase** again with the next N men.
- Battle win of a phase: if more phases in the mission, next briefing; else mission complete → Boot Hill (promotions come in Chunk 09).

**Done when.** You can fail Mission 1, see the grave count tick up, and retry with the next two names. Queue on the hill shrinks.

**Do not.** Save game, +15 recruits, rank-up.

---

### Chunk 09 — Mission complete: promotions, +15, save/load

**Goal.** Finish Mission 1 for real.

**Rules (Amiga)**

- Survivors gain **one rank per phase survived** in that mission (Mission 1 = +1).
- Then 15 new Privates (later: trained) join the pool.
- Save icon on Boot Hill writes JSON (`mission index`, `pool`, `graves`, `rng seed` optional).
- Load icon reads it. One slot is enough for now (`userConfig/fannon-codder/save.json`).

**Create** `internal/campaign/save.go` + tests for promotion and pool size 13+15=28 after M1 with 0 deaths (2 survivors stay, 13 unused + 15 new + 2 survivors = 30 total men; HUD “remaining” on M2 deploy-3 → 27). **Match the FAQ numbers:** after M1 with no deaths, M2 shows 27 remaining.

**Done when.** Complete M1, Boot Hill shows promoted names, save, quit, load, briefing for Mission 2 exists as a stub (“not implemented”) **or** simply refuses until Chunk 11. Prefer a stub screen over a crash.

**Do not.** Implement Mission 2 map yet.

---

### Chunk 10 — Rank gun stats + HUD ammo icons

**Goal.** Rank matters. HUD has empty grenade and bazooka icons (counts 0).

**Curve (tune to footage, start here)**

| Rank-ish | Range | Spread | RoF |
|---|---|---|---|
| Private | 80 px | wide | 8/s |
| mid | 110 | medium | 11/s |
| General | 150 | tight | 14/s |

Interpolate by rank index. Enemies stay on a fixed “grunt” table.

**Done when.** A promoted Corporal from a saved M1 game noticeably outranges a Private in a tiny sandbox (or a `sim` test). Icons render at 0.

**Do not.** Actual grenades.

---

### Chunk 11 — Tilemaps, scrolling, cover

**Goal.** Maps larger than one screen. Trees block **walking and bullets** (LOS). Camera edge-scrolls with the pointer.

**Create**

- Map bigger than 320×256 (e.g. 40×30 tiles).
- Bullet LOS: grid DDA or stepped samples; tree tiles stop MG.
- Camera: when pointer is within 16 px of a playfield edge, scroll that way, clamped.
- Minimap stub: clickable map icon later (Chunk 16); skip if time is short — prefer LOS+scroll working.

**Done when.** A debug oversized map: men walk off the starting screen, pointer at the edge pans, you can shoot someone on the far side of a tree line only by flanking.

**Do not.** Water.

---

### Chunk 12 — Water, swimming, bridges

**Goal.** Mission 2’s teacher.

**Rules (Amiga manual)**

- Shallow: walk slow, **can** fire.
- Deep: swim slower, **cannot** fire (and should not throw).
- Bridge tiles: treat as land.
- Enemies in deep water are sitting ducks.

**Tests.** Unit in deep water refuses to shoot. Speed multipliers are constants.

**Done when.** A sandbox strip of river + a bridge: swimming across is scary; shooting swimmers is easy; crossing the bridge is the right play.

**Do not.** Grenades, buildings.

---

### Chunk 13 — Split squads (Snake / Eagle / Panther)

**Goal.** Up to 3 player squads. Required for later missions; useful on Mission 2 already.

**Amiga**

- Highlight names, click logo, split. New squad is Eagle, then Panther.
- Grenade/rocket split: all / half / none via outlined icons (implement the three-way toggle even if ammo is 0).
- Inactive squads: **hold position and auto-fire at enemies in range** (simple). Switching active squad: click that squad’s HUD block or a hotkey (`1`/`2`/`3`).
- Merge: player squads that touch combine under the active one.

**Tests.** Split 3 men into 2+1; merge; never more than 3 squads.

**Done when.** You can leave one man guarding a bridge (he shoots) while you walk the others across.

**Do not.** Grenades yet (toggle can be dummy).

---

### Chunk 14 — Buildings, spawners, grenades, crates

**Goal.** Mission 2 phase 2 and all later “destroy buildings” phases.

**Rules**

- Door buildings spawn grunts on a timer until destroyed.
- MG cannot destroy buildings. Grenades can.
- Grenade: leader only, arc (simple ballistic or timed land-at-aim), explosion radius, hurts **everyone**, can go **over** trees/walls (manual: throw long by aiming far).
- Crate of 4 grenades. Pick up by walking over. **Shooting a crate explodes it.**
- From this chunk, objectives can be `kill_all_enemy` and/or `destroy_enemy_buildings`. Spawned living enemies count for kill-all; you must also kill the hut or it never ends.

**Tests.** MG does not dent a hut. One grenade at the door destroys it. Crate pickup increments ammo by 4.

**Done when.** Sandbox: 1 hut spewing reds, 1 crate, you blow the hut, remaining reds die, win.

---

### Chunk 15 — Mission 2 content

**Goal.** Two original-feeling phases that play the same **lesson** as the Amiga ones.

**m02p01 “Bridge Over the River Pie”**

- Deploy 3, kill-all, ~15–18 grunts.
- River, one bridge, trees to hide in, a swimmer near spawn.

**m02p02 “Trash Enemy HQ”**

- Deploy 3, kill-all + destroy buildings, 1 spawner hut, lots of water, one grenade crate near the hut (so you can blow it; do not require a lucky crate-shot).

Wire campaign.json so Boot Hill after M1 goes to M2. Mission complete → +15, promotions (2 phases = +2 ranks for full survivors).

**Done when.** You can play M1 then M2 from a fresh boot, save, and the pool numbers roughly match the FAQs (M2 remaining 27 with no M1 deaths).

Tune enemy count/placement until it is **easy-medium**, not a meat grinder. Watch a longplay for pacing, not layout tracing.

---

### Chunk 16 — HUD finish + overview map

**Goal.** Status panel matches the manual’s information, not its pixels.

- Troop logo (Snake/Eagle/Panther colours)
- Grenade count, bazooka count, selected special highlighted
- Foot vs vehicle icon
- Active squad highlight
- Names + ranks
- Map icon (bottom-left of panel): hold or toggle a top-down schematic of tiles + unit dots, no fog required

**Done when.** You can split, read who is who, see ammo, and peek the whole M2 map.

---

### Chunk 17 — Mission 3 (ice, cliffs, grenade economy)

**Goal.** `m03p01 “Blast It's Cold”`: deploy 4, kill-all + destroy 4 door-huts, arctic palette (white/grey placeholders), a **cliff drop** (one-way or slow climb — Amiga lets you jump down; going up is via a ramp tile).

**Design constraint from FAQs:** grenade crates sit next to a hut. Shooting them can soft-lock the phase. Place **just enough** crates (e.g. 2 crates = 8 grenades, 4 huts) so wasting them fails. That *is* the lesson.

**Done when.** M3 is winnable if you pick up crates and spend 1 grenade per hut; unwinnable if you panick-shoot the boxes. Ice uses the ice tile (same walk speed as grass for now; skid comes with the Skidoo).

---

### Chunk 18 — Civilians, quicksand, mines (Mission 4 systems)

**Goal.** The three new hazards/actors Mission 4 introduces.

- **Civilians:** yellow, wander, harmless unless shot; spears are optional flavour (slow, low damage). Killing them does **not** fail M4 Village People in the original (they’re a joke), so do not add a fail condition unless the phase JSON says `protect_civilians` (keep the objective type in the enum, unused).
- **Huts without doors:** scenery; not an objective.
- **Quicksand:** enter → trapped, die after a short sink animation (2 seconds).
- **Mines:** tile; trigger on walk; explosion like a grenade.

**Done when.** A sandbox/unit tests cover: sink death, mine death, shooting a civilian, doorless hut ignored by `destroy_enemy_buildings`.

---

### Chunk 19 — Mission 4 content + free grenades

**Goal.** Four phases, original-feeling.

| Phase | Title | Deploy | Objectives | Teach |
|---|---|---|---|---|
| 1 | Beachy Head | 4 | destroy buildings (5) | find crates, don’t waste them |
| 2 | Pier Pressure | 4 | kill-all + destroy (4) | **each trooper starts with 2 grenades** (manual: from this phase onward) |
| 3 | Village People | 5 | kill-all + destroy (2 door huts) | civilians, door vs hut, a quicksand warning |
| 4 | Quicksand | 5 | kill-all + destroy (1) | quicksand pools, mines, enemy grenadiers |

**Engine flag** in campaign data: `startGrenadesPerTrooper: 2` from Pier Pressure through the rest of this plan.

Enemy **grenadiers** (grunt that throws 1–2 grenades): add a `kind: "grenadier"` in JSON. Throw rarely, telegraphed.

**Done when.** Full run M1–M4 from new game is possible. Soft-lock still possible if you blow every crate — that is authentic.

---

### Chunk 20 — Bazookas, rocket-grunts, Skidoo, Mission 5

**Goal.** Finish the plan’s campaign.

**Bazooka**

- Crate of 4. Hitscan-or-fast projectile, long range, explodes on impact.
- Damages buildings, infantry, **and vehicles**.
- HUD toggle grenade ↔ bazooka (keys: click icon, or `C` as on console ports).
- From phase “My Beautiful Skidoo” onward: `startRocketsPerTrooper: 1` (manual).

**Rocket-grunt / sniper**

- `kind: "rocketeer"`: slow fire, projectile like a rocket, hides if a tree is adjacent (doesn’t walk into the open as eagerly).

**Skidoo (generic land vehicle)**

- Pointer over empty vehicle → “board” cursor; left click → squad enters (capacity 8, we deploy ≤6).
- Hold left to drive toward pointer; **hold longer → faster** (manual). Light skid on ice tiles (velocity persists).
- Right: fire mounted MG if `armed: true`.
- No grenades/bazookas from inside.
- Pointer over occupied own vehicle → “exit”.
- Ram: overlap kills infantry.
- Enemy vehicles: red blinker; same physics.
- Destroyed by rockets/explosions, not by MG.

**Mission 5 phases** (original-feeling)

1. Valley of Ice — 3 men, 6 huts, ice river, first rocketeers, grenade+rocket crates.
2. Barmy Bazookas — 3 men, 6 huts, many rocketeers, water.
3. My Beautiful Skidoo — 4 men, 3 huts, **player Skidoo**, destroy buildings; free starter rocket.

**Done when.** New game through Mission 5 is playable. Skidoo ram + rocket hut-kills work. Save after M5. Title screen can say the campaign continues another day.

---

## Chunks 21–23 — Grunts you can shoot first

Chunk 05’s rule is still what runs: a grunt who can see a player inside 70px snaps his facing and fires on that frame, at 4 rounds/s, with no pause. His cone is `World.Spread` (0.05 rad), tighter than a Private’s 0.12, and the round is aimed at the body. One hit kills. Bullets are fast (500 px/s), so the cone is the whole miss chance.

He also walks in from 140px before he has spotted anyone. On Mission 1 the south grunt spawns 80px from the squad (tiles (14,10) vs (10,13)), so he closes and opens fire without the player moving. The other two sit near 193px and only join once you step into the open. The Amiga mission is the opposite lesson: the manual’s “shoot them before they shoot you.” Posted men, a visible turn, a late first shot, then a sloppy burst you can sidestep.

Do **21, then 22, then play Mission 1**. Chunk 23 is written so it is ready, and it stays unstarted if that play already feels like the original. These chunks do not retune maps, gun range, or “one hit kills.”

Grenade windups and rocket windups stay as they are. Do not stack the grunt reaction on top of them. A grenadier who is too close to throw uses the grunt MG rules. Enemy vehicles stay on their current “drive at the player and shoot in range” rule. Inactive player squads stay snappy. No A*, no flanking, no chase.

### Chunk 21 — Spot, turn, hold the post

**Goal.** A grunt is a sentry. He stands on his spawn tile. He only notices a player inside gun range with clear LOS. He turns at a limited rate, and the first round waits until both the reaction time has elapsed and he is actually facing you. Walking into the open is a duel you can win by aiming first. Standing there is still fatal.

**Create / change**

- `internal/sim/ai.go`, `internal/sim/ai_test.go`, `internal/sim/unit.go` (a couple of timers on `Unit`).
- New constants, used by grunts only:

  | Name | Value | Meaning |
  |---|---|---|
  | `EnemyReact` | 0.55 s | Contact time before the first round is allowed |
  | `EnemyReactJitter` | 0.15 s | Added as `rng` in [−j, +j], once per contact |
  | `EnemyTurnRate` | 4 rad/s | 180° takes about 0.8 s |
  | `EnemyFaceTol` | 0.35 rad | No round until facing error is inside this |

- Contact is: living player within `EnemyMGRange` (70) and `lineClear`. Anything else is not contact.
- `SpotT` counts up only while contact holds. Losing LOS or range zeroes `SpotT` and the rolled jitter. A peek does not store a half-finished reaction.
- While waiting, turn toward the player at `EnemyTurnRate`. Do not snap `Facing`.
- The first round is allowed only when `SpotT` has passed `EnemyReact + jitter` **and** the facing error is inside `EnemyFaceTol`. Already-facing targets give you the 0.55 s window. A grunt looking the wrong way gives you the turn as well, in parallel with the clock, not after it.
- Until that moment, `VX` and `VY` stay 0. Delete the unspotted use of `EnemyApproach` (140). He does not walk toward a player he has not finished acquiring, and he does not leave his tile once he has.
- Enemy spawn facing is east (`0`). Set it where enemies are created. Player facing is unchanged.
- After the first round is allowed, this chunk may still fire continuously at `EnemyMGRoF`. Burst gaps are Chunk 22.
- Grenadier throw telegraph, bomb count, and cooldown: unchanged. Point-blank MG goes through this reaction.
- Rocketeer windup, cooldown, and tree-hiding: unchanged. Do not run them through `SpotT`.

**Tests to replace.** `TestEnemyShootsWhenPlayerInRange` (expects a round after 1/60 s) and `TestEnemyApproachesWhenCloseButOutOfShot` (expects a walk-in from 110 px). `TestGrenadierTooCloseShootsInstead` must wait out the reaction.

**Tests to add.** No round at 1 frame. No round at 0.3 s while already facing. A round once the reaction and the facing tolerance are both satisfied. Facing east with the player to the west: still no round at 0.7 s, because the turn is not finished. LOS broken at 0.4 s resets the clock. A grunt at 110 px and a grunt at 200 px both stay on their tile. Trees still block the shot. Grenadier windup and rocketeer windup timings stay on their old numbers.

**Done when.** `go test ./...`. Mission 1: at spawn, nobody walks and nobody shoots. Step out toward the south grunt already aiming and a short aim kills him before his first round. Wait in the open and he turns, then fires. The west and north grunts stay on their tiles until you reach them.

**Do not.** Bursts, spread changes, hearing, map edits, vehicle AI, pathfinding.

### Chunk 22 — Bursts and a wide cone

**Goal.** Once a grunt is allowed to shoot, he chatters and then stops to re-aim. His cone is wider than a Private’s, so a sidestep survives and standing still does not. This is the sustain half of the Amiga duel. Chunk 21 only made the first shot late.

**Create / change**

- `internal/sim/ai.go`, `internal/sim/ai_test.go`.
- Constants:

  | Name | Value | Meaning |
  |---|---|---|
  | `EnemyBurst` | 3 | Rounds, then silence |
  | `EnemyBurstPause` | 0.75 s | No MG round during the gap. He keeps turning. |
  | `EnemyMGSpread` | 0.20 rad | Wider than a Private’s 0.12. At 50 px that is about ±10 px |

- The cone is `EnemyMGSpread`, not `World.Spread`. A test that sets `World.Spread = 0` still gets the enemy cone. Player `GunStatsFor` is unchanged.
- During the pause he holds still and keeps turning. Contact lost during the pause drops him back to the Chunk 21 idle: next time, he owes a full reaction again.
- Grenadiers use this only on the MG path. A thrown bomb is not a burst round. Rocketeers and vehicles are unchanged.

**Tests.** Three rounds, then no new enemy MG round for the pause, then another round if contact held. The three angles differ under the seeded `newRNG`. A Private’s spread constant is still 0.12. `World.Spread = 0` does not collapse the enemy cone.

**Done when.** `go test ./...`. Mission 1 again. Winning the opening aim is clean. Missing it means a short chatter, a visible gap, and shots that go wide if you are moving. Waiting in the open still gets a man killed.

**Do not.** Hearing, chase, map edits, changing `EnemyMGRoF` or `EnemyMGRange`.

### Chunk 23 — Gunfire turns the next man (playtest gate)

**Play Mission 1 after Chunk 22 before starting this.** If stepping out, aiming, and taking the three grunts one at a time already feels like the original, leave this chunk unchecked. It exists because a man behind a tree who never reacts to a magazine dumped beside his bush is the one Amiga habit still missing. It is also the chunk most likely to make Mission 1 harder, which is why it is gated.

**Goal.** A shot near a posted grunt starts the same Chunk 21 reaction, even without LOS. He still does not leave his tile, and he still cannot shoot through the tree. He turns toward the noise and fires only once contact (range + LOS) is real.

**Create / change**

- `internal/sim/ai.go`, `internal/sim/combat.go` (the MG spawn is where the noise happens), `internal/sim/ai_test.go`.
- `EnemyHear` = 96 px (6 tiles). When any living infantry unit fires an MG round, every idle enemy grunt inside that distance starts a Chunk 21 contact as if he had just spotted someone: `SpotT` begins, he turns toward the shooter. Grenade blasts and rockets do not count in this chunk.
- No LOS from the noise: he turns, and he does not fire. He still needs range and `lineClear` for a round.
- He does not walk toward the noise. Forget the noise on the same  rules as losing contact (clock resets when the shooter is gone and he has no LOS target).
- Do not wake the whole map. 96 px from the Mission 1 south fight does not reach the other two posts (~193 px).

**Tests.** A grunt behind trees, 60 px from a player shot, begins turning and does not fire while the trees hold. A grunt 150 px away does not start `SpotT`. A grunt who then gets LOS still owes any remaining reaction and facing tolerance.

**Done when.** `go test ./...`. Mission 1 still plays as three separate duels. Spraying into a bush makes the man on the far side turn. He shoots only after you clear the trees and his reaction finishes.

**Do not.** Chase, last-known-position walking, alerting from explosions, vehicle guns, map edits.

---

## After Chunk 23 (not this plan)

Keep these in `docs/CHUNKS.md` as a backlog so a future plan is easy:

- Jeeps (skin of Skidoo), tanks (shell, armour), Biggunz (static turret), choppers (altitude, land-on-head, heatseekers)
- Hostages, kidnap, factories, protect-civilians fail
- Desert / moors / underground tiles
- Missions 6–24 as data
- Wounded-squirm + “finish them” (manual); corpse-juggle (manual easter egg)
- Real pixel art (Sensible-sized ~8–12 px troopers, 16-colour palettes per terrain)
- Original-feeling title tune and SFX (new audio, not ripped)
- Birds, snowmen, igloos as flavour
- High Scoring Heroes table
- Fullscreen, integer-scale options
- Headless sim replay for regression of each phase

---

## Suggested session order vs. “I want to see a game faster”

If you want a **playable toy on day one**, do **01 → 06** in order (Mission 1, nameless greens vs reds). Then 07–09 make it feel like Cannon Fodder (Boot Hill). Then 11–15 (Mission 2) is the first time it is actually the game.

Do not skip 03–05; Mission 1 is the control tutor.

Chunks 01–22 are in. Play Mission 1 before deciding on **23**.

---

## Testing policy (saves tokens)

Prefer tests in `internal/sim` and `internal/campaign` over screenshots.

Every chunk that adds a rule adds a test named after the rule, e.g. `TestDeepWaterCannotFire`, `TestMGDoesNotHurtLivingFriendly`, `TestCrateGivesFourGrenades`, `TestPoolAfterMission1NoDeaths`.

Grok must not skip tests to “save time”; they are how the next session knows the last one worked.

---

## Key decisions

| Decision | Why |
|---|---|
| Sim isolated from Ebitengine | Headless tests; Grok can reason about rules without GPU. |
| Placeholders until Chunk 20 | Tokens go to feel and rules, not sprites. |
| Data-driven phases | Missions 2–5 are JSON, not new code. Later campaign is content. |
| Original maps, original mission names | Close structure without ripping assets. |
| Integer 320×256 | Amiga silhouette; cheap to render. |
| Small subsystems, not vertical-slice dumps | Matches limited tokens; each session has a kill-condition. |
| Stop after Mission 5 | Vehicles/hazards needed for “it feels like CF” are in; tanks/choppers are a second season. |
| Grunts hold, turn, then burst (Chunks 21–22) | Shoot-on-sight made Mission 1 a meat grinder. The Amiga window is “aim first.” No chase, no A*. |

## Risks

- **Feel will be wrong until you watch a longplay while playtesting.** Budget time to compare pointer-scroll, walk speed, and MG chatter after Chunks 06, 15, 20.
- **Right-click** may be eaten by the window manager. If so, add a fallback (`Ctrl` = fire) in Chunk 04 without removing right-click.
- **Both-buttons grenade** is fiddly on some mice; keep it and also accept `Space` as “special at pointer” from Chunk 14.
- **Soft-locks** (exploding all crates) are authentic; still make Mission 2 phase 2 have a crate you do not have to shoot-walk through.
- **Scope creep** (A*, chase AI, pixel art mid-stream) will blow the budget. Grunt feel is Chunks 21–22 only: hold the post, turn, burst. Architecture.md is the brake.
- **Do not rebalance Mission 1 by deleting grunts or shortening the gun.** The south man walks in because approach is 140 px and he spawns at 80. Chunk 21 stops the walk; Chunk 22 stops the laser.

## First message to Grok after you accept this plan

Paste the **Preamble** plus **Chunk 01** only. When the window opens, come back and paste Chunk 02 in a new session.

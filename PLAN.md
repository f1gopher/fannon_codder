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
- **Art/audio:** coloured placeholders first. Battle sound is in (Chunks 24–31). The picture is Chunks 32–47: original painted art, a 1024×768 window, and a cap at 4K. Simulation stays in the current world pixel space.
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

Chunks 21 and 22 are in. Chunk 23 is dropped. These chunks do not retune maps, gun range, or “one hit kills.”

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

### Chunk 23 — dropped

**Do not implement this.** Hearing would wake the next post and make Mission 1 harder. The opener stays three separate duels. The old spec is kept below so the idea is not reinvented later.

### Chunk 23 (record only) — Gunfire turns the next man

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

## Chunks 24–31 — Sound effects

One bus, many clips. Chunk 24 is the bus and the gun. Later cue chunks add a `CueKind`, one emit at the cause, and one clip. The skidoo hum is the one exception: a loop voice, not a cue.

`internal/sim` appends `Cue{Kind, X, Y}` and never imports Ebitengine. The battle takes the queue every update (`TakeCues`) and `internal/audio.Mixer.Play` starts a voice. `Mixer.Load(kind, pcm, voices, volume)` registers 16-bit little-endian stereo PCM at 44100. A kind that is not loaded is silent. `PlayKind` is the same pool for a sting that has no world cause, and it stays at the clip's full volume. The title and the briefing click that way. A cleared phase plays the win sting on the way to Boot Hill, and an escape or a wipe plays the fail sting. Boot Hill stays quiet. The world never emits those three kinds. A world cue is full inside 48 px of the listener and silent at 320 px; the clip's own volume is the loud end of that line. The listener is the active leader, or the camera centre when there is no leader. The skidoo loop is separate from that queue. It runs while the listener is aboard a living skidoo, or while the nearest occupied one is inside 320 px, and its pitch rises from idle to `VehicleMaxSpeed`. The listener's own vehicle wins when both qualify. Eight gun voices overlap; the oldest restarts when they are all busy. The queue keeps the newest 64 cues if a frame forgets to drain.

Samples are synthesised or recorded for this game. No Amiga samples, and no theme tune in these chunks. The title tune stays on the backlog.

Chunks 25–31 are in. Do not start 23.

### Chunk 24 — Sound bus and the gun (done)

**Goal.** Every machine-gun round cracks. Player squads, posted grunts, and a mounted skidoo gun share one sample. Grenades, rockets, and blasts stay silent.

**Create / change**

- `internal/sim/cue.go`: `CueKind` (`CueNone`, `CueGun`), `Cue`, `emit`, `TakeCues`.
- `addMG` is the only spawn for an MG round (on foot and mounted). It emits `CueGun` at the muzzle.
- `internal/audio`: one context, `Load` / `Play` / `PlayKind`, synthesised `Gunshot` (~40 ms).
- The battle defers `PlayCues(TakeCues())` so a phase that ends mid-frame still plays.

**Tests.** One round, one cue at the muzzle; the cooldown frame is silent. Two living troopers, two cues. The skidoo gun emits one cue 10 px along the aim. The queue drops the oldest past 64. `Gunshot` is stable, decays, starts and ends near zero, and is the same in both channels. The throw and the launch are Chunk 27.

**Do not.** Other kinds, music, distance, wav files.

### Chunk 25 — Blasts (done)

**Goal.** One boom for every explosion that already exists: grenade, rocket impact, mine, and a crate that cooks off.

**Create / change**

- `CueBoom`. Emit it once at the top of `explode` (grenades, rockets, mines, and crate chains all go through there). A crate that sets off another crate booms again. That is the chain you already see.
- Synthesise a lower, longer clip (~180 ms) and `Load` it with 4 voices and its own volume.
- The throw and the launch stay silent. The boom is the impact, not the leaving of the hand. Death is Chunk 26.

**Tests.** `explode` appends one `CueBoom` at the blast point. A grenade that lands does too. An MG round still emits only `CueGun`.

**Done when.** `go test ./...`. A grenade, a rocket, and a mine each boom once. Hosing a crate booms, and the crates it sets off boom too.

**Do not.** A separate building-collapse sample, death yells, distance.

### Chunk 26 — Death (done)

**Goal.** A man who just died makes a short yell. The same clip, a few pitches, so a wiped squad is not one sample retriggered.

**Create / change**

- `CueDeath` from `kill`, only on the transition to dead (the early return already covers a second call).
- Store the unit id on the cue (add a field when this chunk needs it) and pick one of three pitch variants from it. Civilians use the same yell.
- A kill inside `explode` still booms and yells. Both are real.
- Quicksand death also yells. A distinct gurgle is not this chunk.

**Tests.** `kill` on a living unit emits one death cue. `kill` again emits nothing. An MG kill emits the gun cue and, once the round lands, the death cue.

**Done when.** Shooting a grunt yells when he drops. A grenade that kills two men booms once and yells twice.

**Do not.** Voice acting pulled from the original, a victory sting.

### Chunk 27 — Throw and launch (done)

**Goal.** The bomb and the rocket make a sound when they leave, distinct from the boom when they arrive.

**Create / change**

- `CueThrow` at the end of `launchGrenade` (player and grenadier share it).
- `CueRocket` at the end of `launchRocket` (player and rocketeer share it).
- Two short clips. Impacts stay `CueBoom` from Chunk 25.

**Tests.** A throw emits `CueThrow` and no boom yet. After the fuse, the boom is the only new cue. A rocket emits `CueRocket` at the tube and `CueBoom` on impact.

**Done when.** The yellow telegraph is still silent; the whoosh is the moment the bomb leaves. The bazooka whooshes, then booms.

**Do not.** A reload foley, a click on the G/R icons (that is Chunk 31).

### Chunk 28 — Terrain and vehicles, one-shots (done)

**Goal.** The world answers when you step in something. These are rare, so one voice each is enough.

**Create / change**

- `CueSplash` when `InWater` goes from false to true (a wade and a swim share it). Leaving the water is silent.
- `CueSink` when quicksand first sticks a unit. The death at the end of the sink is still Chunk 26's yell.
- `CuePickup` when a trooper takes a grenade or rocket crate.
- `CueBoard` in `enterVehicle`, `CueExit` in `dismount`.

**Tests.** Crossing onto water emits one splash, and standing in it does not emit another. Stepping into quicksand emits one sink. Walking onto a crate emits one pickup. Board and exit each emit once.

**Done when.** The river, the tan pool, a crate, and the skidoo each answer once per action.

**Do not.** Footsteps on grass, an engine loop (Chunk 30).

### Chunk 29 — Distance (done)

**Goal.** A shot across the map is quieter than a shot at your feet. The picture stays centred; Ebitengine players have volume and no pan, and the shared clip is what lets rounds overlap.

**Create / change**

- `Mixer.SetListener` from the battle, at the active leader, or the camera centre when there is no leader.
- At play time, volume is full inside 48 px and falls linearly to 0 at 320 px. `PlayKind` stays full volume (no position).
- Use the `X, Y` already stored on the cue. Do not change emit sites.

**Tests.** A pure function of distance: 0 px → full, 48 px → full, 320 px → 0, halfway between 48 and 320 → half. The gun PCM itself is unchanged.

**Done when.** A grunt you can see cracks loudly. A fight near the far edge of a scrolling map is faint.

**Do not.** Stereo pan, a low-pass, occlusion by trees.

### Chunk 30 — Engine loop (done)

**Goal.** An occupied skidoo hums, and the hum rises with speed. This is not a cue. A one-shot bus cannot hold a loop.

**Create / change**

- A loop voice on the mixer: start, stop, and a playback rate or pitch. One loop is enough (the skidoo you can hear).
- While any living occupied skidoo is the listener's vehicle, or the nearest occupied one inside the Chunk 29 radius, keep the loop running. Pitch tracks speed from idle to `VehicleMaxSpeed`.
- Silence on foot, and silence when the vehicle is empty or destroyed.

**Tests.** The loop decision is a pure function of occupied, alive, and speed. Playback itself stays on the mixer.

**Done when.** Boarding and holding left builds the hum. Letting go on ice keeps a lower hum while it slides. Dismount cuts it.

**Do not.** A second loop for the enemy skidoo if one voice cannot follow both. The player's vehicle wins.

### Chunk 31 — Scene stings (done)

**Goal.** The menus answer. No new sim emits.

**Create / change**

- `CueClick`, `CueWin`, `CueFail` as kinds with no `emit` in the world. Scenes call `PlayKind`.
- A click when the title advances and when a briefing starts the phase. A short win sting when the phase is cleared, a short fail sting on surrender or wipe. Boot Hill itself stays quiet.

**Tests.** Kinds are distinct from the gun. There is no world path that emits them.

**Done when.** Title → battle clicks. Clearing a phase stings before Boot Hill. Escaping out stings the fail.

**Do not.** The title tune, High Scoring Heroes, icon clicks on the HUD.

---

## Chunks 32–47 — Graphics

The world stays where it is. A tile is still 16 world pixels, a man is still an 8-pixel body, the playfield is still the 268×256 view beside the 52-pixel strip, and audio distances are untouched. Missions are not redrawn and the camera does not zoom. A larger window is a sharper picture of that same frame.

The window opens at **1024×768** and cannot be dragged outside **1024×768 … 3840×2160** (device-independent pixels, `SetWindowSizeLimits`). The picture inside it is the 320×256 frame fitted uniformly, so a 1024×768 window shows a 960×768 image with 32 pixels of bar on each side (scale 3). A 3840×2160 window shows a 2700×2160 image (scale 8.4375) with bars on the sides. Height is the fit. The 5:4 frame is not stretched to 16:9.

That picture is rendered at real pixels, including a high-DPI monitor. `Layout` is given device-independent pixels; the offscreen is `320×S` by `256×S` where `S` is the fit scale times `Monitor().DeviceScaleFactor()`, clamped so the offscreen never exceeds 3840×2160. `DrawFinalScreen` letterboxes that offscreen onto the framebuffer at scale 1. The old path — paint 320×256, then nearest-neighbour integer-scale it — cannot look sharp at both sizes, and 4K is not an integer multiple of 320×256.

Art is painted once, at **8 source pixels per world pixel** (`ArtScale`). A ground tile is 128×128. A trooper is painted about 96 pixels tall. Sprites are drawn at `S / 8` with a linear filter. At 4K that ratio is about 1.05. At the default window it is 0.375. On a 2× monitor the default window is scale 6, ratio 0.75. One linear downscale of a detailed painting stays readable. A one-pixel outline turns to mud, and nearest-neighbour pixels look coarse at 4K.

### The style

**Painted miniatures.** The camera sits south of the field and a little above it, so the ground is visible and a man shows his front. Shapes are soft colour masses with light from the upper left. No ink outline, no dither, no photographic texture. Jungle is deep green, water is blue, snow is blue-white, quicksand is tan, huts are warm brown, rock is grey. Snake wears green, Eagle blue, Panther amber. Grunts are a dull red-brown, grenadiers add an orange satchel, rocketeers add a tube, civilians wear a pale shirt. These are original designs. Nothing is traced from the Amiga sprites.

Chunk 33 shows this style at the 4K master size and at the 1024×768 size. That chunk stops. Chunks 35–47 wait until you accept the board. A rejection redoes 33 only.

### What every graphics session pastes

```
You are implementing Fannon Codder, a Cannon Fodder (Amiga 1993) gameplay clone.

Stack: Go 1.22+, Ebitengine v2 (`github.com/hajimehoshi/ebiten/v2`).
Repo root: the current workspace.

Before writing code:
1. Read docs/ARCHITECTURE.md, docs/CHUNKS.md, and PLAN.md section “Chunks 32–47”.
2. Read only the files this chunk names, plus their direct neighbours.
3. Implement THIS CHUNK ONLY. Do not start the next chunk.
4. Simulation lives in internal/sim and must not import Ebitengine or internal/render.
5. Do not copy Amiga graphics. Art made in an art chunk is original, and it follows the accepted Chunk 33 board.
6. A missing sheet still draws the old coloured rectangle. The field is never blank.
7. When done: go test ./... and go build ./cmd/fannon. Append the chunk note to docs/CHUNKS.md.

Chunk 33 is a stop. Do not start Chunk 34 in that session.
Chunks 35–47 do not start until the user has accepted the Chunk 33 board.
```

An art chunk loads the `game-assets` skill and produces PNG sheets plus the JSON manifest from Chunk 34. The engine never plays a video file. Animation time advances in `Update` at 1/60, not in `Draw`.

Facing matches the sim. `atan2(dy, dx)` with Y down the screen: 0 is east, π/2 is south, π is west, −π/2 is north. Sheet rows are `E, SE, S, SW, W, NW, N, NE`. Paint E, SE, S, N, and NE. The runtime mirrors E→W, SE→SW, and NE→NW. N and S are not mirrored.

### Chunk 32 — The picture

**Goal.** The window opens at 1024×768, resizes up to 4K, and shows the same battlefield framing it shows today. Placeholders are still rectangles. They are drawn into the high-resolution picture, not blown up from a 320×256 buffer.

**Create / change**

- `cmd/fannon/main.go`: default window 1024×768. `SetWindowSizeLimits(1024, 768, 3840, 2160)`.
- A pure function, covered by `internal/app/scale_test.go`, replacing the integer-scale tests:

  | Window (DIP) | Device scale | S | Offscreen |
  |---|---|---|---|
  | 1024×768 | 1 | 3 | 960×768 |
  | 3840×2160 | 1 | 8.4375 | 2700×2160 |
  | 1024×768 | 2 | 6 | 1920×1536 |
  | 3840×2160 | 2 | 8.4375 (clamped) | 2700×2160 |

- `Layout` / `LayoutF` returns that offscreen. `DrawFinalScreen` clears the bars and blits the offscreen at 1:1 in the centre. Filter on that blit is nearest, because the scale is 1. Sprite filtering comes in Chunk 34.
- The cursor arrives in offscreen pixels. Divide by `S` and clamp to the 320×256 frame before the sim, the HUD hit tests, and edge scroll see it. A click in the bar never reaches the game.
- Every current draw (battle, HUD, overview, title, briefing, Boot Hill, the stub, the pointer) runs in offscreen pixels: logical coordinate times `S`. HUD text uses `ebiten/v2/text/v2` and the Go regular face (`golang.org/x/image/font/gofont/goregular`, the one new dependency) at a size that is 14 pixels when `S` is 3 and scales with `S`. Names are readable at the default window and at 4K.
- `docs/ARCHITECTURE.md`: replace the Display section with this contract. State that the world pixel space did not change.

**Done when.** `go test ./...`. `go run ./cmd/fannon` opens at 1024×768. Mission 1 still fits one view. Dragging the window wider adds bars and does not show more map. The cursor still selects a HUD name, and title → Boot Hill → briefing → battle all sit in the fitted picture.

**Do not.** Sprites, sheets, animation, changing `TileSize`, the camera, or mission JSON.

### Chunk 33 — Style board

**Goal.** You see the painted-miniature style at both sizes and accept it or send it back. Nothing in the game changes.

**Create**

- `assets/art/style/`. A contact sheet and the separate source images:
  - Snake, south-east three-quarter, idle, master size (about 96 pixels tall) and the same figure at the default-window size (36 pixels tall).
  - One jungle grass tile at 128 and at 48, plus a 2×2 of the 48 so a seam is visible.
  - One tree, a two-frame strip of shallow water, one snow tile, one door hut, one skidoo in three-quarter view, the pointer, and a corner of the status strip with a name set in the Chunk 32 type size.
- Same light, same brush, the palette named above. Soft edges. Transparent backgrounds on the figures. The grass tile is seamless.

**Done when.** The images are on disk and the session tells you the path. The game still draws rectangles.

**Do not.** Wire the images into the renderer. Do not start Chunk 34. Do not generate the rest of the cast. If you reject the board, the next session replaces this folder and stops again.

### Chunk 34 — Sprite stage

**Goal.** The renderer can play a sheet. Anything without a sheet still draws its rectangle. No production art yet.

**Create / change**

- `internal/render` loads PNG + JSON from an embedded `assets/art/`. One PNG per animation. Rows are the eight directions; columns are frames. JSON fields: `frameW`, `frameH`, `anchorX`, `anchorY` (the anchor is the sim point), `fps`, `loop`, and which rows are mirrors. A mirror row is not stored in the PNG.
- Draw scale is `S / 8`, filter linear. The anchor sits on the unit or tile point the rectangles use today.
- Pose, read from the sim, first match wins:

  | Condition | Pose |
  |---|---|
  | Dead, and the death cycle has not finished | death, once, then hold the corpse frame |
  | `Sinking` | sink, scrubbed by `Sink / SinkTime` (the 2 second clock), not a free loop |
  | `InWater` | swim, loop |
  | `GrenadeWind > 0` or `RocketWind > 0` or `SinceThrow < 0.25` | throw, once across that interval |
  | `SinceShot < 0.12` | shoot |
  | speed above 2 px/s | walk, loop |
  | else | idle, loop |

- `Unit.SinceShot` resets to 0 in `addMG` and counts up each step. `Unit.SinceThrow` resets in `launchGrenade` and `launchRocket` and counts up. Neither field changes combat. Death timing lives in the renderer, keyed by unit id.
- Draw order: ground, then a soft oval shadow under each body (drawn by the runtime, not baked into the painting), then trees, huts, crates, men, and vehicles sorted by foot Y, then grenades, tracers, and blasts, then the HUD and the pointer. A sprite may be taller than its cell. Its manifest anchor keeps the feet on the sim point.
- Scenery loops take a phase of `(tx*3 + ty*5)` frames, so neighbouring tiles do not sway together.
- Clock advances in the battle `Update`.

**Tests.** Facing 0 → E, π/2 → S, π → W (mirror), −π/2 → N. `addMG` zeroes `SinceShot`; one step of 1/60 increases it. `launchGrenade` zeroes `SinceThrow`. Phase offset differs for tile (0,0) and tile (1,0). A missing file selects the rectangle path.

**Done when.** `go test ./...`. The game looks as it did after Chunk 32, because no production sheet exists yet. A test sheet in the test data plays a frame.

**Do not.** Author the soldiers or the map. Do not change speeds, ranges, or mission data.

### Chunk 35 — Snake walks

**Goal.** Your men idle and walk. Everyone else is still a rectangle.

**Create.** Sheets for Snake, idle (4 frames, 8 fps, loop) and walk (6 frames, 12 fps, loop), five painted directions, master scale. Manifests as Chunk 34 describes. Edit from the accepted board; do not invent a second brush.

**Done when.** Mission 1: the two troopers stand, breathe, and walk in the direction they face. A man in the file faces along the file. Resize to 4K and the same sheets stay sharp. Eagle, enemies, and the map are unchanged.

**Do not.** Shoot, death, swim, or the other squads.

### Chunk 36 — Snake fights and falls

**Goal.** The same body shoots, throws, and dies.

**Create.** Shoot (2 frames, held for the 0.12 s window), throw (4 frames, played across the windup or the 0.25 s flourish), death (6 frames, about 0.4 s, once), corpse (1 frame). Same directions and the same body as Chunk 35.

**Done when.** Holding right plays the shot. A grenade plays the throw as it leaves. A dead trooper falls and stays down. The corpse is not the death’s first frame.

**Do not.** Swim, sink, or recolors.

### Chunk 37 — Snake in the water and the sand

**Goal.** Swimming and sinking use the same body.

**Create.** Swim (4 frames, loop) and sink (8 frames, scrubbed across the 2 seconds). The sink’s last frame is the moment he dies.

**Done when.** `go run ./cmd/fannon -river`: a man in deep water swims and does not walk on the surface. `go run ./cmd/fannon -hazards`: the tan pool plays the sink, then the corpse.

**Do not.** Enemy sheets.

### Chunk 38 — Eagle and Panther

**Goal.** The other two squads are the Snake sheets recoloured, not new people.

**Create.** Blue fatigues for Eagle, amber for Panther, every Snake animation from 35–37. Same pixels, same anchors, new colours. A recolor that muddies the face or the gun is redone from the Snake frame.

**Done when.** Split on the river sandbox: green Snake, blue Eagle, amber Panther, each animating. Merging back does not leave the wrong colour.

**Do not.** Repaint ranks onto the body. Rank stays HUD text.

### Chunk 39 — Enemy soldiers

**Goal.** Grunts, grenadiers, and rocketeers are their own bodies in the accepted style.

**Create.** For each: idle, shoot, death, corpse, swim, sink, five directions. Grenadier also has the throw, with the satchel readable. Rocketeer also has the launch, with the tube readable. No walk cycle. Posted men do not walk, and a civilian walk is Chunk 40.

**Done when.** Mission 1: a grunt idles facing east, turns, and the shot plays on his burst. An orange grenadier on Quicksand telegraphs the throw. A rocketeer on Valley of Ice aims the tube. Killing one plays death, then the corpse.

**Do not.** A second grunt costume. All grunts share one body.

### Chunk 40 — Civilians

**Goal.** The yellow wanderer is a painted person.

**Create.** Idle, walk, death, corpse. Pale shirt, dark trousers. No weapon.

**Done when.** `go run ./cmd/fannon -hazards`: he wanders, and shooting him plays the death. The phase rules are unchanged.

**Do not.** Spears, a second civilian costume.

### Chunk 41 — Ground

**Goal.** The flat green fill and the flat snow fill become looping ground.

**Create.** A seamless jungle grass loop (4 frames, about 8 fps) and a seamless snow loop at the same timing. 128×128 per frame. The battle picks snow when the phase terrain is `arctic`, grass otherwise.

**Done when.** Mission 1’s clearing moves. Mission 3’s field is snow. A 2×2 of either tile has no seam. Trees and water are still the old squares.

**Do not.** Animate water yet.

### Chunk 42 — Water, quicksand, ice

**Goal.** The three surfaces that should move do move.

**Create.** Loops, seamless, 128×128: shallow water, deep water, quicksand (a slow boil), ice (a slow sparkle). About 8 frames for water, 6 for sand and ice.

**Done when.** The river sandbox shows two different water loops and a still bridge. The hazard sandbox’s pool boils. Mission 3’s ice glints. Adjacent tiles of the same kind are not on the same frame.

**Do not.** Change swim speed or the sink timer.

### Chunk 43 — Trees and the hard ground

**Goal.** Cover reads as trees, and a forest does not sway in unison.

**Create.** Three tree silhouettes, each a sway loop (4 frames). A tree is taller than its tile; the anchor is the foot of the trunk on the blocked cell, and the canopy hangs up-screen. Cliff, ramp, and bridge are still paintings (one frame each). The mine is a still painting plus a one-frame glint loop. Cliff and ramp may share the rock colours of the style board.

**Done when.** Mission 1’s trees sway, three shapes, out of phase. A man walking south of a tree passes in front of the trunk; a man north of it passes behind the canopy. The cliff on Mission 3 and the bridge on the river sandbox are painted. The mine on the hazard sandbox is visible before it is stepped on.

**Do not.** New tile types.

### Chunk 44 — Huts, crates, and the skidoo

**Goal.** The things you interact with are painted.

**Create.** A door hut and a doorless hut at building size (a 2×2 hut is 256×256 of art) with a short chimney-smoke loop on both. A grenade crate and a rocket crate, distinct without a letter baked into the painting (the HUD still says G and R). A skidoo, idle and moving (4 frames), five directions, mirrored like the men. The enemy lamp stays a render overlay on the same body, not a second painting.

**Done when.** `go run ./cmd/fannon -hut` shows the door hut and the crate. `go run ./cmd/fannon -skidoo`: the skidoo idles, then the skis and the track move while you drive, and it faces as it turns. The enemy skidoo blinks. Boarding hides the troopers, as it does now.

**Do not.** A jeep skin, rubble, a destroyed-hut sprite.

### Chunk 45 — Fire and blasts

**Goal.** Shots and explosions are painted frames, short enough to match the sounds already in the game.

**Create.** A muzzle flash (tied to the 0.12 s shot window), an MG tracer, a grenade in the air, a rocket in the air, and a blast (about 0.2 s, played at the explosion). The blast is one animation shared by grenades, rockets, mines, and crates.

**Done when.** A burst flashes at the gun. A grenade is visible on the arc and the blast plays where it lands. A rocket and a mine use the same blast.

**Do not.** A second explosion style per weapon.

### Chunk 46 — Pointer and the status strip

**Goal.** The cursor and the HUD icons belong to the style. The strip itself stays a flat panel in the approved dark colour, because a single bitmap stretched between scale 3 and scale 8.4 will not fit the strip.

**Create.** Pointer, crosshair, and the board-vehicle cursor, painted at master scale and drawn in screen space at `S / 8`. Icons for grenade, rocket, on foot, in vehicle, and the map button, plus the three squad marks. The overview keeps its diagram layout and takes the accepted ground colours. Names, ranks, and counts stay the Chunk 32 text.

**Done when.** The battle cursor, the crosshair, and the board cursor are the new art. The strip shows the icons, the selected special is still obvious, and the overview colours match the map. Text is readable at 1024×768 and at the maximum window.

**Do not.** Move the strip or change which clicks it handles.

### Chunk 47 — Title, briefing, Boot Hill

**Goal.** The three menus are painted scenes. The troopers on the hill are the Snake sprite, not a new man.

**Create.** A title picture with the words “Fannon Codder” as part of the scene (this is the one place text may be in the painting; it does not get localised). A briefing backdrop. Boot Hill: sky, hill, a grave marker, save and load as icons. The queue of recruits reuses the trooper sheet. The grave count is still the real count, drawn as markers, not baked into the picture.

**Done when.** A new game walks title → Boot Hill → briefing → Mission 1, and each screen is the painted scene at both the default window and a window dragged toward 4K. Wipe a man and the grave count still goes up. Save and load still hit.

**Do not.** The title tune. High Scoring Heroes. A fullscreen toggle.

---

## After Chunk 47 (not this plan)

Keep these in `docs/CHUNKS.md` as a backlog so a future plan is easy:

- Jeeps (skin of Skidoo), tanks (shell, armour), Biggunz (static turret), choppers (altitude, land-on-head, heatseekers)
- Hostages, kidnap, factories, protect-civilians fail
- Desert / moors / underground tiles
- Missions 6–24 as data
- Wounded-squirm + “finish them” (manual); corpse-juggle (manual easter egg)
- Original-feeling title tune (new audio, not ripped). Battle sound effects are Chunks 24–31.
- Birds, snowmen, igloos as flavour
- High Scoring Heroes table
- A fullscreen toggle
- Headless sim replay for regression of each phase

---

## Suggested session order vs. “I want to see a game faster”

If you want a **playable toy on day one**, do **01 → 06** in order (Mission 1, nameless greens vs reds). Then 07–09 make it feel like Cannon Fodder (Boot Hill). Then 11–15 (Mission 2) is the first time it is actually the game.

Do not skip 03–05; Mission 1 is the control tutor.

Chunks 01–22 and 24–45 are in. Chunk 23 is dropped. The sound chunks are done. Chunk 32 (the picture) is in. The Chunk 33 style board in `assets/art/style/` is accepted. Chunk 34 (the sprite stage) is in. Chunk 35 (Snake walks) is in. Chunk 36 (Snake fights and falls) is in. Chunk 37 (Snake in the water and the sand) is in. Chunk 38 (Eagle and Panther) is in. Chunk 39 (enemy soldiers) is in. Chunk 40 (civilians) is in. Chunk 41 (ground) is in. Chunk 42 (water, quicksand, ice) is in. Chunk 43 (trees and the hard ground) is in. Chunk 44 (huts, crates, and the skidoo) is in. Chunk 45 (fire and blasts) is in. Next is Chunk 46.

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
| Integer 320×256 world | The tactical frame. Missions, camera, and audio distances stay in these pixels. |
| Picture is 1024×768, cap 4K (Chunks 32–47) | A sharper drawing of that same frame. The window does not reveal more of the map. |
| Small subsystems, not vertical-slice dumps | Matches limited tokens; each session has a kill-condition. |
| Stop after Mission 5 | Vehicles/hazards needed for “it feels like CF” are in; tanks/choppers are a second season. |
| Grunts hold, turn, then burst (Chunks 21–22) | Shoot-on-sight made Mission 1 a meat grinder. The Amiga window is “aim first.” No chase, no A*. |
| Sim emits cues, audio plays them | Ebitengine stays out of `internal/sim`. A new sound is a kind, an emit, and a clip. |

## Risks

- **Feel will be wrong until you watch a longplay while playtesting.** Budget time to compare pointer-scroll, walk speed, and MG chatter after Chunks 06, 15, 20.
- **Right-click** may be eaten by the window manager. If so, add a fallback (`Ctrl` = fire) in Chunk 04 without removing right-click.
- **Both-buttons grenade** is fiddly on some mice; keep it and also accept `Space` as “special at pointer” from Chunk 14.
- **Soft-locks** (exploding all crates) are authentic; still make Mission 2 phase 2 have a crate you do not have to shoot-walk through.
- **Scope creep** (A*, chase AI, a second art style mid-stream) will blow the budget. Grunt feel is Chunks 21–22 only: hold the post, turn, burst. The picture is Chunks 32–47, one chunk at a time. Architecture.md is the brake.
- **Do not rebalance Mission 1 by deleting grunts or shortening the gun.** The south man walks in because approach is 140 px and he spawns at 80. Chunk 21 stops the walk; Chunk 22 stops the laser.

## First message to Grok after you accept this plan

Chunks 01–31 are already in. Paste the **graphics preamble** plus **Chunk 32** only. When that window looks right, come back and paste Chunk 33 in a new session, then stop for the style board.

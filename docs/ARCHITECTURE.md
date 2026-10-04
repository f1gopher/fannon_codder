# Architecture

Current behavior. Read this before changing code. If you change behavior, update this file.

Already done: `docs/CHUNKS.md`. Not done: `docs/BACKLOG.md`. Bindings: `controls.md`, generated from `internal/app/controls.go` by `go test ./internal/app -run TestControlsDoc`.

## Session rules

- Do the change you were asked for. Leave the next backlog item alone.
- `internal/sim` must not import Ebitengine or `internal/audio`. The sim appends `Cue` values. The battle plays them after `Step`.
- `Update` samples input, steps the world once, then advances the picture clock. No gameplay in `Draw`. `ebiten.TPS` is 60.
- A new rule gets a named test in `internal/sim` or `internal/campaign`.
- `go test ./...` and `go build ./cmd/fannon` pass.
- Launch from the repo root. Mission JSON loads relative to the working directory. Main package is `cmd/fannon`.

## Legal

Original maps, art, and audio. Do not extract or ship Amiga graphics, samples, map data, sprite sheets, or the theme. Mission titles may homage the original.

## Packages

- `cmd/fannon` — window, `RunGame`
- `internal/app` — `ebiten.Game`, scenes, scale, pause, quit
- `internal/input` — pointer
- `internal/sim` — world, combat, AI, objectives
- `internal/campaign` — recruits, names, ranks, heroes, save
- `internal/render` — sheets, or a coloured rectangle when a sheet is missing
- `internal/audio` — the only package that opens the audio device
- `internal/data` — mission JSON
- `data/missions` — one JSON per phase, listed by `campaign.json`
- `assets/art` — one PNG and one JSON per animation

`SpawnUnit` appends to `World.Units` and can move earlier pointers. Re-fetch with `World.Unit(id)` across `Step`. `AddSkidoo` does the same to vehicles. Re-fetch with `vehicleByID`.

## Display

The world frame is 320×256. A tile is 16 px. A trooper body is about 8 px. The status strip is 52×256 (`HUDWidth`). The playfield is 268×256. The camera does not zoom. Audio distances are world pixels. A larger window is a sharper picture of that same frame.

The window opens at 1024×768 device pixels. Limits are 1024×768 through 3840×2160. The frame fits uniformly, by height. 1024×768 is scale 3 (picture 960×768, 32 px of bar each side). 3840×2160 is scale 8.4375 (picture 2700×2160).

`Layout` returns an offscreen of 320×S by 256×S. `S` is that fit times `Monitor().DeviceScaleFactor()`, clamped so the offscreen never exceeds 3840×2160. `DrawFinalScreen` clears the bars and blits the offscreen 1:1, nearest, centered. Draws use logical coordinates times `S`. The cursor is divided by `S` and clamped to the frame. A click in the bar does not reach the game.

A painted sheet draws at `S/8` with a linear filter. 8 is source pixels per world pixel. HUD text is `ebiten/v2/text/v2` with Go Regular. The size is 14 px when `S` is 3, otherwise `14×S/3`.

## World

`World` holds the map, units, projectiles, buildings, pickups, vehicles, camera, objectives, and rng.

`Unit` has id, name, rank, side (player, enemy, civilian), hp (alive, wounded, dead), position, velocity, facing, water flag, squad, vehicle, and kills.

Squads are Snake, Eagle, and Panther. Members are in rank order. A squad holds grenades, rockets, and an active flag. `SpawnPlayerSquad` stores the given positions as the file trail, so the men stay in that line until a new move order. Followers stand `FileSpacing` (10) apart along it.

Tiles: grass, tree, waterShallow, waterDeep, ice, quicksand, cliff, bridge, mine.

A building with a door spawns grunts until destroyed. Cap is `maxDoorSpawns` (6). Doors do not spawn grenadiers or rocketeers. A doorless hut is scenery and is not a `destroy_enemy_buildings` target. MG fire does not damage a hut.

A crate holds 4 grenades or 4 rockets (`CrateAmount`). A player shot, grenade, or rocket explodes it. Enemy fire leaves it.

The only vehicle is the skidoo. A jeep would be the same struct with another skin.

Objectives in use are `kill_all_enemy` and `destroy_enemy_buildings`. `protect_civilians` parses and does not change the phase.

## Combat

Player MG range, rate, and spread come from `GunStatsFor`. Private is 80 px, 8/s, 0.12 rad. General is 150 px, 14/s, 0.02 rad. The steps between are linear. Enemies use the grunt table in `internal/sim/ai.go`, not this curve. Movement, grenades, rockets, and specials do not scale with rank.

`NewEmpty` sets `WoundChance` to `WoundOdds` (0.34). A `World` literal in a test leaves it at 0, so those shots still kill. `rollWound` treats 0 as always kill and 1 as always wound.

An MG hit on a standing enemy or civilian kills him, or wounds him. A player is never wounded: the round kills him. A wounded man leaves the file, cannot shoot, and squirms on the corpse frame until another MG round, a blast, or a ram finishes him. The yell and the kill point happen on the finish.

Player MG passes through a standing friendly and finishes a wounded friendly. Explosives and rams kill anyone who is not already dead, including the owner.

`Living()` means alive. A wounded enemy still blocks kill-all. The phase is lost when no player is `Living()`.

Shooting an enemy or civilian corpse shoves it about one tile. A dead player stays put. Keep the shove that short. In `combat.go`, `juggleKick` is 36 px/s along the shot and does not stack past that, `juggleHop` is 64 px/s upward and refreshes, gravity is 420, bounce is 0.28, ground drag is 200, and rest is 12. Three seconds after death (`juggleLife`), further MG hits leave the body still. A hop already in the air finishes. Corpses do not sink or walk off cliffs. The shadow stays on the ground while the sprite lifts.

Rounds are fast (`MGSpeed` 500). The cone is the miss chance.

## Enemy infantry

There is no pathfinding and no patrol. `steerToward` slides on a blocked axis. Feel constants live in `internal/sim/ai.go`.

Mission 1, and any sandbox (`World.Mission` 0), leaves map grunts posted. They spawn facing east and stay on that tile. `Phase.World` sets `World.Mission` from the phase and marks non-rocketeer enemies `Aggressive` when the mission is 2 or higher.

An aggressive grunt or grenadier acquires like the original troop check. Past `EnemySight` (200 px, `0xC8`) he stays put. Inside `EnemySightBlind` (40 px, `0x28`) he closes even when the straight walk is blocked. Between those he closes only when `pathClear` (trees, cliffs, standing huts). He then stops inside gun range. A heard shot does not pull him. Rocketeers never close. Grenade and rocket windups use that same acquire.

Fire contact is `EnemyMGRange` (70) and clear LOS. He turns at `EnemyTurnRate` (4 rad/s). The first round waits until `SpotT` passes `EnemyReact` (0.55 s) plus one jitter of `EnemyReactJitter` (0.15 s), and facing error is inside `EnemyFaceTol` (0.35 rad). Then he fires `EnemyBurst` (3) rounds at `EnemyMGRoF` (4/s), stays silent for `EnemyBurstPause` (0.75 s) while still turning, and repeats. The cone is `EnemyMGSpread` (0.20 rad), not `World.Spread`. Losing range or LOS calls `resetGruntContact`. The next sighting owes a full reaction.

`EnemyHear` is 40 px, the same blind contact. An infantry MG round calls `wakeFromMG` and sets `HearID` on an idle posted grunt. An existing `ReactAt` or `HearID` is left alone. `Step` runs AI before fire, so the id latches when the round spawns and `SpotT` advances on the next tick. Hearing alone never fires and never starts a walk. Grenades, rockets, and vehicle guns do not wake anyone. A dead shooter with no LOS clears the clock.

Only `spawnDoorGrunt` sets `HasPost`. He walks about `DoorPostTiles` (3) south onto a walkable cell, slides on blocked axes, and does not shoot until he arrives or a step cannot move. The post then clears and `Aggressive` is set, on every mission. Sinking clears the post and does not set `Aggressive`. Enemies loaded from JSON, and a plain `SpawnUnit`, do not take that walk. The pre-placed grunt in `NewHutWorld` is a plain spawn.

A grenadier carries `GrenadierBombs` (2), winds up for `GrenadierWindup` (0.8 s), and waits `GrenadierCooldown` (5 s). Inside `GrenadierMinRange` he uses the grunt gun. Grenades arc over trees. A rocketeer holds still. `RocketeerWindup` is 0.55 s. `RocketRange` is 200. A rocket needs LOS to start the windup.

During a windup he turns at the grunt rate and releases on the first frame inside `EnemyFaceTol`. The opening frame sets the clock and does not snap `Facing`. If the clock ends while he is still turning, `WindHeld` finishes the turn and releases. The windup does not restart.

A player squad you are not controlling uses that same turn, reaction, burst, and pause. Range, rate, and spread stay `GunStatsFor`. He does not hear and does not throw. The active squad fires on the first frame fire is held. `World.Spread == 0` still zeros the parked cone.

Mission 1 grunts are at tiles (17, 3), (1, 5), and (14, 10). The south man starts about 80 px from the squad. The other two are about 193 px out. None of that is inside `EnemyHear`, and mission 1 does not close, so the opener is three duels.

## Vehicles

`VehicleCapacity` is 8. Hold left to drive. A longer hold is faster, up to `VehicleMaxSpeed` (88) after `VehicleHoldFull` (1.4 s). Grass zeroes velocity on release. Ice keeps a skid (`iceSkidRate`). Right fires the mounted gun when `Armed` is set. Grenades and rockets do not fire from inside. Overlap within `VehicleRamR` (12) kills. Rockets and blasts destroy a skidoo. MG fire does not. Boarding hides the troopers. The enemy lamp is drawn on the same body.

An armed enemy skidoo drops the throttle inside `VehicleMGRange` (110) and outside the ram, turns the hull at `EnemyTurnRate`, then uses the grunt reaction, facing tolerance, burst, and pause at `VehicleMGRoF`. The burst clock is on the driver. An unarmed enemy skidoo still closes and rams. A blocked hull slides on one axis. The player's mounted gun still aims with `vehicleShoot`.

`destroyVehicle` clears `Alive`, kills the crew, and empties `Occupants`.

## Camera and HUD

The view is the playfield. Origin is (`HUDWidth`, 0), so a map as wide as the frame still pans by the strip. `LeaderMargin` is 32. `Camera.Contain` keeps the active leader, or the driven vehicle via `CameraFocus`, inside that margin. Edge scroll uses playfield coordinates.

The strip is opaque. Rank is a `ui/rank-*` flash above the name. Keys, low to high: pte, cpl, sgt, ssgt, sfc, msg, sgm, sp4, sp6, wo, cwo, cpt, maj, col, bg, gen. The abbreviation is only the missing-sheet fallback. Names and the G and R counts stay text. Crate paintings have no letter. The selected special gets a white stroke. The overview is a diagram. M, or the M icon, opens and closes it. A left click on the open map closes it. Arctic grass and tree cells use the snow colour.

## Campaign

Start with 15. Mission 1 deploys 2. After each mission, survivors gain one rank per phase they lived through, capped at General, then 15 recruits join. Every three completed missions, new recruits start one rank higher. Deploy takes the highest rank first. The field never holds more than six. No recruits left is game over.

One slot at `userConfig/fannon-codder/save.json`. Boot Hill saves only after a finished mission. The battle pause menu saves during a phase and writes the squad from that attempt back onto the queue in the file. The live queue is unchanged until you leave the phase. `Soldier.Kills` is `omitempty`, so an old save still loads.

`campaign.json` plays missions 1–5 (11 phases). After Mission 5, Boot Hill says the campaign continues.

- m01p01 deploy 2, kill-all.
- m02p01 Bridge Over the River Pie, deploy 3, kill-all. m02p02 Trash Enemy HQ, deploy 3, kill-all and one door hut.
- m03p01 Blast It's Cold, deploy 4, arctic, ice, cliff, four door huts, scarce crates.
- m04p01 Beachy Head, deploy 4, five huts, two crates, start grenades 0. m04p02 Pier Pressure, deploy 4, kill-all and four huts, start grenades 2 from here on. m04p03 Village People, deploy 5, civilians, doorless huts, one quicksand. m04p04 Quicksand, deploy 5, pools, mines, grenadiers.
- m05p01 Valley of Ice, deploy 3, rocketeers, both crate types, start rockets 0. m05p02 Barmy Bazookas, deploy 3. Rocket crate at tile (14, 7). m05p03 My Beautiful Skidoo, deploy 4, player and enemy skidoos, start rockets 1, destroy buildings only.

Sandboxes from `cmd/fannon`: `-cover`, `-river`, `-hut`, `-hazards`, `-skidoo`. `-skip-title` opens Mission 1. The river sandbox shows all three squad colours.

## Ranks and score

Sixteen ranks, Private (0) through General (15), in `internal/campaign/ranks.go`.

One enemy kill is one point for the player trooper who caused it. Grenades, rockets, the skidoo gun, and a ram score. The ram and the mounted gun credit the first living occupant (`vehicleDriver`). Civilians, friendlies, and a non-player owner do not score. `creditKill` is in `combat.go`. Phase kills fold into `Soldier.Kills` when the phase is tallied.

Boot Hill opens the table with H (`drawHeroes`). While it is open, click or Enter closes it. The table shows the best `HeroShow` (12), living or fallen, by kills, then rank, then name. A tie prefers the living man. The fallen list keeps `HeroKeep` (64). The same name keeps the higher score. Logic is `internal/campaign/heroes.go`.

## Hazards and water

Civilians wander and do not shoot. Killing one does not fail the phase. Quicksand traps on entry and kills after `SinkTime` (2 s). A mine explodes like a grenade when a living unit's centre steps on it, then the tile is grass.

Shallow water is `ShallowSpeedMul` (0.5) and can still fire. Deep water is `DeepSpeedMul` (1/3). Swimmers cannot fire or throw. A bridge is land. Ice is grass speed on foot. The skidoo skids.

## Pause and quit

P, or the 16 px strip button at (20, 238), toggles `Battle.paused`. Pause stops orders, `World.Step`, `render.Advance`, and the skidoo hum. The playfield says PAUSED. The button gets a white stroke. Entering pause clears fire and drive. A finished phase does not pause. There is no pause sheet. The icon is two cream bars.

Escape during a phase that is still playing opens the pause menu on the same plaque. Up and Down move the row. The pointer highlights the row it is over, and a left click chooses it. Enter chooses the highlighted row. Escape again continues. The rows are restart level, main menu, save game, load game, and quit. Restart and the main menu return the squad that started the attempt to the queue and drop deaths from that attempt. Save writes the campaign slot with that squad still on the queue and stays on the menu. Load reads the slot and opens Boot Hill, or the menu says “No save”. Quit ends the process. A pause that was already on stays on when the menu closes.

Escape on the title, Boot Hill, briefing, and a finished phase still opens the quit prompt and freezes the screen. Y or Enter ends the process. N or a second Escape cancels. A cleared phase and a wipe leave on click or Enter.

Pause, the quit prompt, PHASE COMPLETE, and PHASE FAILED share `render.Notice`. It draws `ui/notice`, a painted olive plaque with a cream rim, and cream letters with an olive edge, in the middle of the frame. The first line is the headline. A missing plaque is a flat olive fill. The quit lines are QUIT THE GAME?, Y or Enter quit, and N or Escape stay. A clear says PHASE COMPLETE. A wipe says PHASE FAILED. Both add Click or Enter.

## Audio

PCM is 16-bit stereo at 44100 Hz. One mixer per process. A second `NewMixer` panics. A kind with no clip is silent. Ebitengine players have no playback-rate API.

`Cue` is kind, world position, and id. The battle calls `SetListener` at `CameraFocus`, or the playfield centre when there is no living leader, then `PlayCues(TakeCues())`. A world cue is full inside `HearNear` (48 px) and silent at `HearFar` (320 px). Past that it takes no voice. There is no pan. `PlayKind` ignores position.

| Kind | Emit |
| --- | --- |
| `CueGun` | `addMG`, on foot and mounted |
| `CueBoom` | once at the top of `explode` |
| `CueDeath` | `kill`, and only on the living-to-dead step. Pitch is `id % 3` |
| `CueThrow` | `launchGrenade` |
| `CueRocket` | `launchRocket`, at the tube |
| `CueSplash` | on-foot `InWater` goes false to true, and the unit was already sampled. Standing in water, shallow to deep, and leaving are silent |
| `CueSink` | quicksand first sticks. The death is still `CueDeath` |
| `CuePickup` | a player takes a crate |
| `CueBoard`, `CueExit` | `enterVehicle`, `dismount`, once for the squad |

`CueClick`, `CueWin`, and `CueFail` are `PlayKind` only. The title and the briefing click. A clear plays the win sting. A wipe plays the fail sting. Boot Hill and the strip icons are silent. The world never emits those three.

The skidoo hum is one looping voice on the mixer. It plays while the listener is in a living occupied skidoo, or the nearest other occupied skidoo is inside `HearFar`. His own vehicle wins. Empty, destroyed, or out of range is silent. Pause stops it. Pitch runs from `EngineIdlePitch` (0.75) to `EngineTopPitch` (1.45) at `VehicleMaxSpeed`. A `loopReader` resamples a synthesised cycle. Scene changes call `SetEngine(false, 0)` before `Leave`.

Clips are original synthesis. There is no title tune.

## Sprites

`internal/render` embeds `assets/art`. The key is the path without the extension. The style board has no JSON, so it is not drawn. Cell size, anchor, fps, loop, rows, and mirrors are in the JSON beside each PNG. A mirror names a stored row and is not stored.

Directions are E, SE, S, SW, W, NW, N, NE. Facing 0 is east, π/2 is south, π is west, and −π/2 is north. Stored rows are E, SE, S, N, NE. The runtime mirrors E to W, SE to SW, and NE to NW.

The anchor sits on the sim point. Ground frames pin their top-left to the cell. `drawTopLeft` bleeds a 1 px edge, disables mipmaps, and overlaps one device pixel. Ebitengine v2.10.2 `DrawImageOptions` has no wrap address, and linear filtering otherwise samples the atlas padding.

Draw order: ground, a soft oval under each painted body, then trees, huts, crates, men, and vehicles by foot Y, then grenades, tracers, and blasts, then birds, then the HUD and the pointer. Rectangles cast no oval.

Pose, first match: wounded (corpse frame, rocked), death (once, then the corpse; a dead player skips this and holds the corpse frame), sink (scrubbed by `Sink/SinkTime`), swim, throw (the windup, or 0.25 s after launch), shoot (`SinceShot < 0.12`), walk (faster than 2 px/s), idle. `SinceShot` resets in `addMG`. `SinceThrow` resets in `launchGrenade` and `launchRocket`. Both count up in `Step` and do not change combat. Gunfire draws no muzzle flash, on foot or mounted. Death time is per unit id in the renderer. The battle `Update` calls `render.Advance` after `Step`.

A missing walk sheet uses that actor's idle. Any other missing sheet, or a facing the sheet does not store, draws the coloured rectangle.

Actor keys: `snake`, `eagle`, `panther`, `grunt`, `grenadier`, `rocketeer`, `civilian`. Eagle and Panther are Snake recolors. Enemies have no walk sheet.

Ground sheets are 128×128. Grass and snow keep extra frames on disk. The battle draws frame 0 only. Arctic terrain uses snow for grass and tree cells. Shallow and deep water, quicksand, and ice do loop. Neighbours of those use `SceneryFrame` (`tx*3+ty*5`). Cliff, ramp, and bridge are one frame. The mine is a short glint on grass.

`tree/sway` stores three silhouettes on rows E, SE, and S. The row is `(tx+ty*2) mod 3`. The anchor is the trunk foot.

Flavour in `internal/render/flavour.go` does not block movement or fire. Birds cross every map. Grass grows scrub. Arctic maps grow snowmen and igloos.

`fx/blast` is the only explosion, about 0.2 s. The sim marker still lasts `blastTime` (0.35 s) and draws nothing after the last frame. The orange outline is only the missing-sheet fallback. Grenades, rockets, mines, and crates share it.

Menus `menu/title`, `menu/briefing`, and `menu/hill` are one frame pinned to the top-left of the 320×256 frame. The title painting includes the words Fannon Codder. Boot Hill draws `snake/idle` facing south-east for the queue and one `menu/grave` per death. Save and load use the old click rectangles. Names, the mission line, and the briefing copy stay text.

The pointer and the board cursor are tinted gold so they stay readable on snow. The crosshair, shown while firing, is red. The black outline is left alone. The exit mark on a skidoo you already occupy stays the small bitmap, in the same gold.

## Accepted picture limits

Leave these unless asked. They are not open bugs.

- Snake swim is a smoother brush than idle. North and north-east swim repeat a pose. The north sink still shows the slung rifle.
- Enemy idle is a one-pixel bob. Swim is one side pose on every row. The rocketeer tube is the east painting on every row. Grenadier north-east reuses south-east. Grenadier and rocketeer shoot, death, swim, and sink reuse the grunt body.
- Civilian front and back walks alternate one stride with the standing pose. The north-east corpse reuses the south-east body.
- Skidoo north and south are top-down, and those ski posts are blockier than the side views.
- `menu/load` still has a thin warm edge from keying.

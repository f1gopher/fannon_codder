# Architecture

Current behavior. Read this before changing code. If you change behavior, update this file.

Already done: `docs/CHUNKS.md`. Not done: `docs/BACKLOG.md`. Bindings: `controls.md`, generated from `internal/app/controls.go` by `go test ./internal/app -run TestControlsDoc`.

## Session rules

- Do the change you were asked for. Leave the next backlog item alone.
- `internal/sim` must not import Ebitengine or `internal/audio`. The sim appends `Cue` values. The battle plays them after `Step`.
- `Update` samples input, steps the world once, then advances the picture clock. No gameplay in `Draw`. `ebiten.TPS` is 60.
- A new rule gets a named test in `internal/sim` or `internal/campaign`.
- `go test ./...`, `go build ./cmd/fannon`, and `go build ./cmd/mapedit` pass.
- Launch from the repo root. Mission JSON loads relative to the working directory. Main package is `cmd/fannon`.

## Legal

Original maps, art, and audio. Do not extract or ship Amiga graphics, samples, map data, sprite sheets, or the theme. Mission titles may homage the original.

## Packages

- `cmd/fannon` — window, `RunGame`
- `cmd/mapedit` — read-only map editor window
- `internal/app` — `ebiten.Game`, scenes, scale, pause, quit
- `internal/input` — pointer
- `internal/sim` — world, combat, AI, objectives
- `internal/campaign` — recruits, names, ranks, heroes, save
- `internal/render` — sheets, or a coloured rectangle when a sheet is missing. The battle ground call is a short adapter.
- `internal/terrain` — ground painter. Plates, water shader, canopies. The battle and the editor both call it.
- `internal/mapedit` — phase list and scrolling viewport
- `internal/audio` — the only package that opens the audio device
- `internal/data` — mission JSON
- `data/missions` — one JSON per phase, listed by `campaign.json`
- `assets/art` — one PNG and one JSON per animation

`SpawnUnit` appends to `World.Units` and can move earlier pointers. Re-fetch with `World.Unit(id)` across `Step`. `AddSkidoo` does the same to vehicles. Re-fetch with `vehicleByID`.

## Map editor

`go run ./cmd/mapedit` from the repo root. It lists `campaign.json` in order. A left rail picks the phase. Under the list, the rail shows the selected phase's terrain and its size. The rest of the window is that whole map at the game's picture scale. It opens on the painted terrain. Escape closes it. Nothing is written. A gold diamond marks the player start tile. A red diamond marks each placed enemy. Huts, crates, and birds are not drawn. Wheel and the arrow keys scroll in every view. A left drag scrolls too, except in the regions view.

A phase is either a tile grid or freeform regions. Regions are a closed ring of world-pixel points. The map starts as that terrain's open ground, and each region paints over it. A later region wins on the overlap. A tree region is a canopy. Water, ice, sand, rock, a ramp, a timber bridge, and mines paint as those materials. `PrepareChart` in `internal/terrain` does that paint. The battle calls it for a region phase.

Every campaign phase is a region map. The size line is world pixels (`384 × 256` on Mission 1). V switches to those region outlines, labelled with the type name. That is the edit view. A left drag there samples a freehand ring. Releasing closes it with the straight line back to the first sample, and the new region is a tree. It stays in the session, including across phase changes, and is not written to the file. The view choice stays when the phase changes. A phase that still has `map.tiles` uses a cell grid labelled with `data.TileRune`, and its size line is tiles. In the battle, tree slices are still Y-sorted with the actors. Movement, fire, and mines still use the tile grid: a cell whose centre lies inside a region keeps that material, and open ground stays grass. Jungle paints that grass from the world position. Snow paints its open ground from the world position. Desert paints its open ground from the world position. Forest paints its open ground from the world position. A spent mine drops its painted mark. The overview draws the same rings. Sandboxes stay tile grids. `m01p01` is jungle ground and four tree regions.

## Display

The world frame is 320×256 of content. A tile is 16 px. A trooper body is about 8 px. The status strip is 52×256 (`HUDWidth`). The camera does not zoom. Audio distances are world pixels. A larger window is a sharper picture of the same world pixels.

The window opens at 1024×768 device pixels. Limits are 1024×768 through 3840×2160. Scale fits the 256-tall frame by height. 1024×768 is scale 3. 3840×2160 is scale 8.4375. The picture then grows sideways to the window, for menus and for battles. 1024×768 is 1024×768. 1920×1080 is 1920×1080 at scale 1080/256. A window taller than 320×256 still fits by width, with bars above and below. The offscreen never exceeds 3840×2160.

`Layout` returns that offscreen. `S` is the fit times `Monitor().DeviceScaleFactor()`. `DrawFinalScreen` clears any remaining bars and blits the offscreen 1:1, nearest, centered. Draws use logical coordinates times `S`. The cursor is divided by `S` and clamped to the picture. A click in a bar does not reach the game.

A painted sheet draws at `S/8` with a linear filter. 8 is source pixels per world pixel. HUD text is `ebiten/v2/text/v2` with Go Regular. The size is 14 px when `S` is 3, otherwise `14×S/3`.

## World

`World` holds the map, units, projectiles, buildings, pickups, vehicles, camera, objectives, and rng.

`Unit` has id, name, rank, side (player, enemy, civilian), hp (alive, wounded, dead), position, velocity, facing, water flag, squad, vehicle, and kills.

Squads are Snake, Eagle, and Panther. Members are in rank order. A squad holds grenades, rockets, and an active flag. A split selects the new squad. `SpawnPlayerSquad` stores the given positions as the file trail, so the men stay in that line until a new move order. Followers stand `FileSpacing` (10) apart along it. A trail shorter than the file keeps that spacing on past the last crumb, so men who just joined do not stop on the same point. The active squad faces `AimX, AimY` every step, standing or walking. A man in a vehicle does not. Another squad faces its walk. Once it has arrived and has no enemy in range, each man on foot takes a different sector around the heading he stopped on. A man who can see an enemy turns onto him. One man keeps that heading.

Tiles: grass, tree, waterShallow, waterDeep, ice, quicksand, cliff, bridge, mine.

A building with a door spawns grunts until destroyed. Cap is `maxDoorSpawns` (6). Doors do not spawn grenadiers or rocketeers. A doorless hut is scenery and is not a `destroy_enemy_buildings` target. MG fire does not damage a hut.

A crate holds 4 grenades or 4 rockets (`CrateAmount`). A player shot, grenade, or rocket explodes it. Enemy fire leaves it.

The only vehicle is the skidoo. A jeep would be the same struct with another skin.

A battle opens with the player squad still aboard a helicopter. It flies in from the north-west along the nose of `chopper/fly`. One rope hangs straight down from the door, or two side by side when more than one soldier is aboard. They slide down those ropes, walk out into a wedge north of the leader, and kneel facing north. The leader stays on the player start. A spot in water, a tree, or a cliff keeps that man's spawn instead. They stay kneeling in that wedge until the squad is given a move. A shot, from that wedge or from any other kneeling squad, keeps the kneel. Orders, guns, and `World.Step` wait until the arrival kneel is done. The ropes reel in and the aircraft leaves to the south-east while the phase plays. Pause holds the arrival. The rotor is a mixer loop until the aircraft is gone, and a scene change stops it. The helicopter is not a vehicle: nothing boards it, and it does not fight.

Objectives in use are `kill_all_enemy` and `destroy_enemy_buildings`. `protect_civilians` parses and does not change the phase.

## Combat

Player MG range and spread come from `GunStatsFor`. Private is 80 px and 0.12 rad. General is 150 px and 0.02 rad. Those two steps are linear. The squad shares one gun, in the Cannon Fodder rotation: the leader shoots between the others. Ranks 0–3 fire every 0.36 s. Rank 4 and up, including a General, fire every 0.30 s. The clock uses the leader's rank. Each round still uses that man's own range and spread. A parked squad keeps the grunt burst and pause, at that same rank rate. Enemies use the grunt table in `internal/sim/ai.go`. Movement, grenades, rockets, and specials do not scale with rank.

`NewEmpty` sets `WoundChance` to `WoundOdds` (0.34). A `World` literal in a test leaves it at 0, so those shots still kill. `rollWound` treats 0 as always kill and 1 as always wound.

An MG hit on a standing enemy or civilian kills him, or wounds him. A player is never wounded: the round kills him. The round that kills an enemy or a player throws him about one tile along the shot (`killSlide` 32 px/s, `killHop` 70 px/s). A grenade, rocket, or other blast throws him away from the centre. On the centre that is about six tiles (`blastSlide` 120 px/s, `blastHop` 110 px/s). The throw shrinks in a straight line with distance and is nothing at the edge of the blast. He lands and lies on the ground. A civilian stays where he fell. A wounded man leaves the file, cannot shoot, and squirms on the corpse frame until another MG round, a blast, or a ram finishes him. The yell and the kill point happen on the finish. That finishing round throws an enemy the same short way.

Player MG passes through a standing friendly and finishes a wounded friendly. Explosives and rams kill anyone who is not already dead, including the owner.

`Living()` means alive. A wounded enemy still blocks kill-all. The phase is lost when no player is `Living()`.

The phase is not called while a round, rocket, grenade, blast, or flying roof is still going. Men stop walking and nobody starts a new burst. When that has finished, no living player is a wipe. Objectives met with someone still up is a win, including when a late shot has already killed a squadmate. `World.Settle` then keeps the shoot and throw clocks, an ice skid, and hopping bodies going. The squad is tallied once nothing left can change who lived. Click or Enter finishes anything still in the air, then leaves. The result is up once the phase is called. A win says a large gold Victory! and a wipe keeps the plaque. The quit prompt still freezes the screen.

Shooting an enemy or civilian corpse shoves it about one tile. A dead player stays put. Keep the shove that short. In `combat.go`, `juggleKick` is 36 px/s along the shot and does not stack past that, `juggleHop` is 64 px/s upward and refreshes, gravity is 420, bounce is 0.28, ground drag is 200, and rest is 12. Three seconds after death (`juggleLife`), further MG hits leave the body still and the round passes through. A dead player never stops a round. A hop already in the air finishes. Corpses do not sink or walk off cliffs. The shadow stays on the ground while the sprite lifts.

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

A player squad you are not controlling uses that same turn, reaction, burst, and pause once it has no move order. Selecting another squad leaves a click-move running until the leader arrives. Range, rate, and spread stay `GunStatsFor`. He does not hear and does not throw. The active squad fires one round on the first frame fire is held, then waits out the shared clock. `World.Spread == 0` still zeros the parked cone.

Mission 1 grunts are at tiles (17, 3), (1, 5), and (14, 10). The south man starts about 80 px from the squad. The other two are about 193 px out. None of that is inside `EnemyHear`, and mission 1 does not close, so the opener is three duels.

## Vehicles

`VehicleCapacity` is 8. Hold left to drive. A longer hold is faster, up to `VehicleMaxSpeed` (88) after `VehicleHoldFull` (1.4 s). Grass zeroes velocity on release. Ice keeps a skid (`iceSkidRate`). Right fires the mounted gun when `Armed` is set. Grenades and rockets do not fire from inside. Overlap within `VehicleRamR` (12) kills. Rockets and blasts destroy a skidoo. MG fire does not. Boarding hides the troopers. The enemy lamp is drawn on the same body.

An armed enemy skidoo drops the throttle inside `VehicleMGRange` (110) and outside the ram, turns the hull at `EnemyTurnRate`, then uses the grunt reaction, facing tolerance, burst, and pause at `VehicleMGRoF`. The burst clock is on the driver. An unarmed enemy skidoo still closes and rams. A blocked hull slides on one axis. The player's mounted gun still aims with `vehicleShoot`.

`destroyVehicle` clears `Alive`, kills the crew, and empties `Occupants`.

## Camera and HUD

The view is the playfield. Origin is (`HUDWidth`, 0), so a map as wide as the picture still pans by the strip. On a wide window the playfield is wider than 268×256 and shows more of the map. A phase map is at least 24×16 tiles (384×256). That covers a 1440×900 battle, whose playfield is about 358×256. `LeaderMargin` is 32. `Camera.Contain` keeps the active leader, or the driven vehicle via `CameraFocus`, inside that margin. Edge scroll uses playfield coordinates.

The strip is opaque. Rank is a `ui/rank-*` flash on the left of the name, on the same line. Keys, low to high: pte, cpl, sgt, ssgt, sfc, msg, sgm, sp4, sp6, wo, cwo, cpt, maj, col, bg, gen. The abbreviation is only the missing-sheet fallback. Names and the G and R counts stay text. Crate paintings have no letter. The selected special gets a white stroke. The overview is a diagram. M, or the M icon, opens and closes it. A left click on the open map closes it. On a region phase it draws those rings. On a tile phase it draws cells. Overview ground, trees, and water follow the phase terrain. Snow grass and snow trees use the snow colour.

## Campaign

Start with 15. Mission 1 deploys 2. After each mission, survivors gain one rank per phase they lived through, capped at General, then 15 recruits join. Every three completed missions, new recruits start one rank higher. Deploy takes the highest rank first. The field never holds more than six. No recruits left is game over.

One slot at `userConfig/fannon-codder/save.json`. Boot Hill saves only after a finished mission. The battle pause menu saves during a phase and writes the squad from that attempt back onto the queue in the file. The live queue is unchanged until you leave the phase. `Soldier.Kills` is `omitempty`, so an old save still loads.

`campaign.json` plays missions 1–5 (11 phases). After Mission 5, Boot Hill says the campaign continues.

- m01p01 deploy 2, jungle, kill-all, one grenade crate. The file is open ground plus four tree regions. Every campaign phase is stored the same way. The battle paints those rings and still moves on the same cells.
- m02p01 Bridge Over the River Pie, deploy 3, jungle, kill-all. m02p02 Trash Enemy HQ, deploy 3, jungle, kill-all and one door hut.
- m03p01 Blast It's Cold, deploy 4, snow, ice, cliff, four door huts, scarce crates.
- m04p01 Beachy Head, deploy 4, jungle, five huts, two crates, start grenades 0. m04p02 Pier Pressure, deploy 4, jungle, kill-all and four huts, start grenades 2 from here on. m04p03 Village People, deploy 5, jungle, civilians, doorless huts, one quicksand. m04p04 Quicksand, deploy 5, jungle, pools, mines, grenadiers.
- m05p01 Valley of Ice, deploy 3, snow, rocketeers, both crate types, start rockets 0. m05p02 Barmy Bazookas, deploy 3, snow. Rocket crate at tile (14, 7). m05p03 My Beautiful Skidoo, deploy 4, snow, player and enemy skidoos, start rockets 1, destroy buildings only.

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

Escape on the title, Boot Hill, briefing, the debrief, and a finished phase still opens the quit prompt and freezes the screen. Y or Enter ends the process. N or a second Escape cancels. A cleared phase and a wipe leave on click or Enter.

A victory click opens the debrief before Boot Hill. Achievements lists each survivor’s name, rank, and kills. Remembrance lists each man who died on that attempt, the same way. A page with nobody on it is skipped. Another click or Enter moves on. The click plays the same tick as the briefing. A defeat still goes straight to Boot Hill. A sandbox still returns to the title.

Pause, the quit prompt, and PHASE FAILED share `render.Notice`. It draws `ui/notice`, a painted olive plaque with a cream rim, and cream letters with an olive edge, in the middle of the frame. The first line is the headline. A missing plaque is a flat olive fill. The quit lines are QUIT THE GAME?, Y or Enter quit, and N or Escape stay. A wipe says PHASE FAILED and adds Click or Enter. A clear draws gold `Victory!` in the middle of the frame, with the same Click or Enter hint under it. Living player soldiers who are on their feet stop, face south toward the camera, and play `salute` (snake, eagle, and panther). Swimmers, men who are sinking, and vehicle crews stay as they were.

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

`CueClick`, `CueWin`, `CueFail`, `CueBrief`, and `CueHill` are `PlayKind` only. The title and the briefing click. A clear starts the victory march when Victory! appears. A wipe starts a slow minor-key march when PHASE FAILED appears. Boot Hill starts a slow A-minor lament and loops it. The briefing starts an E-minor march and loops it. Leaving either screen, or leaving a result, fades that bed out over a second. The lament does not play under the briefing, and the briefing does not play under the battle. A result march fades under the next screen instead of being cut. The title has no tune. Strip icons are silent. The world never emits those five. The skidoo hum stops when the phase is no longer playing, so it does not sit under a march or a menu loop.

The skidoo hum is one looping voice on the mixer. It plays while the listener is in a living occupied skidoo, or the nearest other occupied skidoo is inside `HearFar`. His own vehicle wins. Empty, destroyed, or out of range is silent. Pause stops it. Pitch runs from `EngineIdlePitch` (0.75) to `EngineTopPitch` (1.45) at `VehicleMaxSpeed`. A `loopReader` resamples a synthesised cycle. Scene changes call `SetEngine(false, 0)` before `Leave`.

The arrival helicopter has its own loop, `SetRotor`. It plays from the first moment of a battle until the aircraft has flown off. Pause and a scene change stop it. It does not use the skidoo pitch.

The gunshot is a short muzzle crack, a low thump, and three duller slaps off a street. The whole report is about 240 ms. Clips are original synthesis. The victory march, the defeat march, the briefing, and Boot Hill are new music for the same small band: two trumpets, horn stabs, a walking bass, and drums. Boot Hill leads with a flugelhorn. The defeat is in D minor, slower, and the lead falls an octave at the end. There is no title tune.

## Sprites

`internal/render` embeds `assets/art`. The key is the path without the extension. The style board has no JSON, so it is not drawn. Cell size, anchor, fps, loop, rows, and mirrors are in the JSON beside each PNG. A mirror names a stored row and is not stored.

Directions are E, SE, S, SW, W, NW, N, NE. Facing 0 is east, π/2 is south, π is west, and −π/2 is north. Stored rows are E, SE, S, N, NE. The runtime mirrors E to W, SE to SW, and NE to NW.

The anchor sits on the sim point. Ground frames pin their top-left to the cell. `drawTopLeft` bleeds a 1 px edge, disables mipmaps, and overlaps one device pixel. Ebitengine v2.10.2 `DrawImageOptions` has no wrap address, and linear filtering otherwise samples the atlas padding.

Draw order: ground, a soft oval under each painted body, then trees, huts, crates, men, and vehicles by foot Y, then grenades, tracers, and blasts, then birds, then the HUD and the pointer. Rectangles cast no oval.

Pose, first match: wounded (corpse frame, rocked), death (once, then the corpse; a dead player skips this and holds the corpse frame), sink (scrubbed by `Sink/SinkTime`), swim, throw (the windup, or 0.25 s after launch), shoot (`SinceShot < 0.12`), walk (faster than 2 px/s), salute on a win, kneel, idle. A player soldier who is not in the squad you are controlling kneels while he is standing still. The controlled squad, and anyone walking, shooting, or in the water, keeps the usual pose. Soldier idle holds frame 0, so the helmet does not grow and shrink. While a soldier is idle or kneeling, his head glances on its own clock. The leader of the squad you control keeps his head on the pointer. He may nod, look up, or hop. The others look left and right. About once a minute a man hops, stares up, or, if he is a player and not that leader, turns to the camera and salutes. Civilians still play their idle cycle. A missing kneel sheet draws idle. The grenadier's front cel sits on a painted box, so that facing does not split the helmet. `SinceShot` resets in `addMG`. `SinceThrow` resets in `launchGrenade` and `launchRocket`. Both count up in `Step` and do not change combat. Gunfire draws no muzzle flash, on foot or mounted. Death time is per unit id in the renderer. The battle `Update` calls `render.Advance` after `Step`.

A missing walk sheet uses that actor's idle. Any other missing sheet, or a facing the sheet does not store, draws the coloured rectangle.

Actor keys: `snake`, `eagle`, `panther`, `grunt`, `grenadier`, `rocketeer`, `civilian`. Eagle and Panther are Snake recolors. Enemies have no walk sheet.

The painter lives in `internal/terrain`. `Field` and the map editor both call it with the phase `terrain`: jungle, forest, desert, or snow. "grass" paints as jungle. "arctic" paints as snow. A campaign battle paints the region rings. A sandbox, or a phase that still has `map.tiles`, paints the tile grid. Touching cells of the same material are one area on that grid. Jungle open ground is computed from the world position: olive shade, yellow sun, and a short south-east grain. It does not sample plates/jungle/ground.png. Snow open ground is computed the same way: blue shade, white crests, and a short south-east grain. It does not sample plates/snow/ground.png. Desert open ground is computed the same way: pale gold banks, lighter crests, and a short south-east grain. It does not sample plates/desert/ground.png. Forest open ground is computed the same way: dark moss, deeper shade, and a short south-east grain. It does not sample plates/forest/ground.png. Every kind keeps its own canopy, shallow, and deep plates, so editing one kind does not change the others. Water does not sample those shallow and deep plates. It is pale at the shore and deepens offshore, and each kind keeps its own two blues. Ice, dirt, and quicksand stay shared. A snow cliff's edge waves a few pixels, the way the shore does, and the face is the cliff palette with a thin snow lip and a shadow at the foot. It does not sample the rock plate. Other terrains still do. A grass cell on snow paints as snow. Ice stays cyan. Quicksand paints as sand. Land that meets water gets a soft shore: warm sand on jungle, forest, and desert, and a pale frozen lip on snow. The line waves in and out, and corners are rounded, so a straight run of cells is not a ruled edge. The sand blends into the pale water, then that water fades into the deep. A narrow channel stays paler than open water at the same distance from one coast. The lip is paint. The cell underneath is still grass or water. A bridge is a timber deck on the water under it. Boards run across the span, the outer edges are darker rails, and a soft shadow falls on the water to the south-east. The ends sit a little onto the bank. Snow dusts the boards. The shimmer stays off the deck. Ramps are the gap in a cliff face. Mines are a mark on the open ground and the picture rebuilds when the mine becomes grass. Water, sand, and ice shimmer from `water.kage`. The other materials stay still. Missions 1, 2, and 4 are jungle. Missions 3 and 5 are snow. Forest and desert paint when a phase names them.

Trees in one connected stand share a canopy. The outline is a cluster of crowns with a ragged fringe that fades into the ground, and a soft shadow falls to the south-east. Each tree cell draws its own slice of that canopy and Y-sorts on the south edge of the cell. The old ground sheets and `tree/sway` stay on disk. The battle does not sample them.

Flavour in `internal/render/flavour.go` does not block movement or fire. Birds cross every map. Jungle, forest, and desert stay clear. Snow maps grow snowmen and igloos. `prop/scrub` stays on disk and is not drawn.

`fx/blast` is the only explosion painting, about 0.2 s. A grenade, rocket, mine, or crate marker still lasts `blastTime` (0.35 s) and draws nothing after the last frame. Those four also throw a render-only particle cloud (a small flash, two rings, many tiny sparks, embers, debris, and small rising smoke) that lasts about 1.85 s after the marker. On soil, sand, or snow they leave a soft stain that multiplies the paint already there, so the ground's colour and grain stay. It lasts until the next battle. Water and a shed fireball leave none. Ground debris is larger, lands within about one soldier height of the blast, and fades out by about 1.2–1.5 s. It matches the ground under the blast: water spray, beach or desert sand, snow, or soil. The orange outline is only the missing-sheet fallback. A destroyed shed adds a second blast at the hut: the same sheet, looped slower, about 1.25 s, drawn larger and drifting up. The roof (eaves and chimney cut from `hut/door` or `hut/plain`) tumbles for `RoofFlight` (1.2 s) to a point `RoofReach` (64 px) past the hut, away from the blast. On the way down it kills soldiers inside `RoofHitR` (16 px). The landed roof stays. That flight is pending, so a finished phase waits for it.

Menus `menu/title`, `menu/briefing`, and `menu/hill` are 16:9 paintings, 3648×2048 source pixels (456×256 world pixels). `DrawMenu` centers each one. A window narrower than 16:9 crops the sides, so the picture still meets both edges. The title painting includes the words Fannon Codder. Boot Hill draws `snake/idle` facing south-east for the queue and one `menu/grave` per death. Load stays at the top left. Save stays at the top right of the picture. Names, the mission line, and the briefing copy stay text, over the blank panel in the middle of the map.

The pointer and the board cursor are tinted gold so they stay readable on snow. The crosshair, shown while firing, is red. The black outline is left alone. The exit mark on a skidoo you already occupy stays the small bitmap, in the same gold.

## Accepted picture limits

Leave these unless asked. They are not open bugs.

- Snake swim is a smoother brush than idle. North and north-east swim repeat a pose. The north sink still shows the slung rifle.
- Enemy swim is one side pose on every row. The rocketeer tube is the east painting on every row. Grenadier north-east reuses south-east. Grenadier and rocketeer shoot, death, swim, and sink reuse the grunt body.
- Civilian front and back walks alternate one stride with the standing pose. The north-east corpse reuses the south-east body.
- Skidoo north and south are top-down, and those ski posts are blockier than the side views.
- `menu/load` still has a thin warm edge from keying.

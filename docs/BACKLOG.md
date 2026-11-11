# Not done

These are not in the game. Do one item, then stop. Do not start one of these to finish something else.

Original assets only. No Amiga graphics, samples, map binaries, or the theme tune. Do not recreate a Sensible Software phase. A new map in the same style is allowed.

When an item lands, add one line to `docs/CHUNKS.md` and delete it here. Update `docs/ARCHITECTURE.md` if behavior changed.

New behavior gets a named test in `internal/sim`. `internal/sim` does not import Ebitengine or `internal/audio`.

Each chunk below lists the sheets and the sounds it adds. The chunk stays here until those play in the battle. A coloured rectangle or a borrowed sheet is only for the sim test. Show one contact sheet of the new cels and wait before writing `assets/art`.

Sheets are one PNG and one JSON. Eight source pixels make one world pixel. Stored rows are E, SE, S, N, NE. The runtime mirrors E to W, SE to SW, and NE to NW. The anchor sits on the sim point. People keep the toy-soldier body already approved: two arms, two legs. No muzzle flash.

The sim only emits a `Cue`. Clips are original synthesis in `internal/audio`. No sampled audio. A launch cue is not the impact. Impacts stay `CueBoom`, and `fx/blast` stays the only explosion painting. `CueDeath` stays the soldier yell. `CueRocket` stays the bazooka. `CueBoard` and `CueExit` stay the mount and dismount for every vehicle a soldier can enter. The skidoo hum stays the skidoo.

No pathfinding. Vehicles and escorts use the axis slide already in `steerToward`. Do not retune grunt sight, hearing, or Mission 1 aggression. The last chunk replaces the Mission 1 map. It does not change those numbers.

The place bar is full (`barStamps`, keys 1–9). The first new placeable adds a second row. Later chunks add an icon on that row. Shift+1 onward selects that row in the order the icons were added. Clicking an icon still selects it.

## Editor

Every chunk that adds a thing also teaches the map editor to add it, move it, and delete it. The gestures stay the ones already there.

- Add. Choose the icon, then left-click empty ground. The click selects the new thing.
- Edit. Left-drag moves it by whole tiles. A building keeps its footprint on the map. W on a selected patrol or chopper adds a route point, and that point drags on its own.
- Delete. Delete or Backspace removes the selection. A selected route point goes, and the owner stays. The player start stays.

These keys are editor-only. Do not add them to `controls.go` or `controls.md`.

- F on a jeep, skidoo, or tank flips `side` between player and enemy.
- A on a jeep, skidoo, tank, or chopper flips `armed`.
- H on a civilian flips `home`.

S writes every list the editor can edit, and the per-phase memo keeps those lists. Escape still drops unsaved work. The first chunk that edits a list loads what the phase already has, so a save does not blank civilians or the Mission 5 skidoos. The icon is one frame of that thing's sheet. The editor draws the battle sheet. Wrecks are battle-only. A named mapedit test places, drags, and deletes the new thing, then round-trips it through S.

## Already in

Do not rebuild these.

- Grunts: posted, patrol, Mission 2+ chase, burst fire, short hearing.
- Grenadiers and rocketeers. Door huts spawn grunts only. Plain huts are scenery.
- Civilians wander, flee a nearby shot or blast, and do not shoot. `protect_civilians` loses the phase when one dies. A wound does not. `m04p03` does not list it.
- One vehicle, the skidoo. Player drive and gun. Enemy driver, ram, and armed burst. Rockets and blasts kill it. MG fire does not.
- The opening helicopter drops the squad and leaves. It is not a vehicle and does not fight.
- Crates, barrels, mines, water, bridges, ice, cliffs, quicksand.
- Objectives that change the phase: `kill_all_enemy`, `destroy_enemy_buildings`.

Cannon Fodder 1 has one enemy trooper. The extra people are the hostage, the enemy leader, and the civilian. Grenadiers and rocketeers stay as they are. Cannon Fodder 2 switches are out of this list.

## Gaps, in order

Do these from the top. One chunk a session. The big map is last and uses every chunk above it.

### 7. Jeep

`vehicles[].kind` of `jeep` is the skidoo struct, speeds, gun, ram, and enemy AI, with another sheet. `skidoo` stays the snow vehicle. Armed and unarmed, player and enemy, already work. Snow phases keep their skidoos.

Editor. A jeep icon writes `vehicles[].kind` of `jeep`, side player, armed true. A skidoo icon writes `skidoo` with the same defaults, so snow phases can be edited too. Load the vehicles already on a phase, including Mission 5, before the first save. Drag moves one. Delete removes one. F flips side. A flips armed. An unarmed jeep is the plain sheet.

Pictures. Match the skidoo cell: 240×176, anchor 120,88, five stored rows. `jeep/idle` is one frame per row at 4 fps, looping, a slight body rock. `jeep/move` is four frames at 8 fps, wheels turning. Armed and plain are two pairs: `jeep/idle` and `jeep/move` have a pintle gun, `jeep/plain-idle` and `jeep/plain-move` do not. The gun is painted in the cel. No muzzle flash. It vanishes when destroyed, as the skidoo does. The blast is the wreck.

Sound. A new engine loop, not the skidoo clip. Four-stroke hum, pitched from idle to full by `VehicleMaxSpeed`, same rule as `SetEngine`. The skidoo keeps its own loop. The mounted gun stays `CueGun`. Board and leave stay `CueBoard` and `CueExit`. A ram stays silent. Destruction stays `CueBoom`.

### 8. Tank and shell

`kind` `tank` is a new vehicle. Slower than the skidoo, about half `VehicleMaxSpeed`. It does not skid. The player boards and drives it like a skidoo. Right click fires a shell, not the mounted gun. Grenades and rockets still do not fire from inside.

A shell is a projectile that explodes like a small rocket. It destroys a skidoo, a jeep, a tank, a reinforced door, a factory, and a computer. MG fire and grenades do not destroy a tank. A rocket does.

An enemy tank starts with a driver, as the skidoo does. It closes, stops outside ram range, and fires shells. It does not use the infantry burst. No pathfinding.

Editor. A tank icon writes `vehicles[].kind` of `tank`, side enemy, armed true. Drag, delete, F, and A are the jeep rules. A player tank is allowed.

Pictures. `tank/idle` and `tank/move`, 320×220, anchor 160,200, five stored rows. Bigger than the jeep. Idle is one frame. Move is four frames at 8 fps, tracks turning. The barrel points with the hull. `tank/wreck` is one burned hull, same cell, and it stays for the phase. `fx/shell` is one east cel, 32×16, anchor 16,8, drawn along its velocity. No muzzle flash. The impact is `fx/blast` and the same particle cloud as a rocket, drawn a little smaller.

Sound. Launch emits `CueShell`: a low cough, shorter and duller than `CueRocket`, about 120 ms. The impact emits `CueBoom` and does not also emit `CueShell`. A grenade that fails to hurt it emits `CueClank`. The engine is a new low loop, slower than the jeep, pitched by the tank's own top speed. Board and leave reuse `CueBoard` and `CueExit`. The crew's death yell still plays when the tank dies.

### 9. Big gun

`kind` `turret` does not move and cannot be boarded. It turns at `EnemyTurnRate` and fires the tank shell at a player in the open, using the grunt reaction once, then a steady reload. No ram.

MG fire and grenades do not destroy it. A rocket or a shell does. That is the hardened big gun.

Enemy only.

Editor. A turret icon writes `vehicles[].kind` of `turret`, side enemy. Drag and delete work. F and A do nothing, and the status line says so. There is no player turret and no unarmed turret.

Pictures. `turret/idle`, 192×192, anchor 96,168, five stored rows, one frame. A concrete ring and a barrel that points with facing. `turret/wreck` is the same cell with the barrel dropped, and it stays. The shell in flight is `fx/shell`. The impact is `fx/blast`. No muzzle flash.

Sound. Launch emits `CueShell`. Impact emits `CueBoom`. A grenade that does nothing emits `CueClank`. No engine and no board cue. Turning is silent.

### 10. Combat helicopter

This is not the opening drop. The drop stays as it is. Nothing boards this one.

`kind` `chopper` flies. Ground, water, trees, and huts do not block it. It is drawn above the men. MG fire misses it. A rocket or a shell destroys it. On destruction the wreck falls and the blast can kill.

`armed` true fires a homing rocket at the nearest living player in range. The rocket turns toward him and explodes on contact. `armed` false does not shoot and does not drop grenades. Either kind kills a man it passes low over.

It flies a `line` of tiles when one is given, like a patrol, and turns to fire when a player is in range. With no line it holds its tile and turns.

Enemy only.

Editor. A chopper icon writes `vehicles[].kind` of `chopper`, side enemy, armed true. Drag and delete work. A flips armed. F does nothing. W adds a route point, and Delete on that point removes the point. No line means it holds its tile.

Pictures. `chopper/attack` and `chopper/plain`, about 200×140, anchor on the belly, five stored rows, one frame each. Smaller than `chopper/fly`, which stays the drop only. The rotor is the same code disc as the drop, not frames in the sheet. Attack paints stub pylons. Plain does not. `fx/homing` is a small east missile, 40×16, two frames at 8 fps, looping, drawn along its velocity. `chopper/wreck` is one tilted body, same cell as the attack sheet. It falls with the crash and stays where it hits. No muzzle flash. The hit is `fx/blast`.

Sound. Armed launch emits `CueHoming`, a thin rising whistle, one shot, about 200 ms. Impact emits `CueBoom`. The fall emits `CueCrash` once, a descending whine, and the boom plays when it hits, not at the start of the fall. An unarmed chopper has no launch cue. The rotor reuses `SetRotor`. The drop keeps that voice until the aircraft has left. After that, the nearest living combat chopper inside `HearFar` starts it. Pause stops it. A chopper does not use the jeep or tank hum.

### 11. Replace Mission 1

Do this only after chunks 1–10 are in the game.

Replace `data/missions/m01p01.json` with a new jungle phase. Keep mission 1, phase 1, deploy 2, and the campaign slot. Keep the title "It's a Jungle Out There". Rewrite the briefing in our own words so it names the objectives. Start grenades and start rockets stay 0. The map supplies crates.

Size is 64×40 tiles (1024×640 world pixels), regions only, no `tiles`. One terrain, jungle. No ice, snow, cliffs, or skidoo. The jeep is the jungle skidoo.

The phase requires every objective those chunks added, plus the two that already work:

- `kill_all_enemy`
- `destroy_enemy_buildings`
- `protect_civilians`
- `rescue_hostages`
- `kidnap_leader`
- `get_civilian_home`
- `destroy_factory`
- `destroy_computer`

Lay the route out so two privates meet each toy after they can deal with it:

1. South start. Grenade crate, a barrel chain, posted grunts, a stream and a bridge.
2. A village with wandering civilians, one civilian who must be taken home, and plain huts.
3. Hostages, guards, and a rescue tent.
4. Door huts, then a rocket crate, then a reinforced door.
5. A player jeep and an enemy jeep on a long open run.
6. The leader and a handover.
7. An enemy tank and a big gun. Shells from the tank, or rockets, are the way through.
8. The factory and the computer past that gun.
9. An armed chopper over the north end, plus a grenadier, a rocketeer, a patrol, mines, and one quicksand pool.

Spread them. Do not pile every object on the start.

No new sheets and no new cues. Draw what chunks 1–10 added. The drop still uses `chopper/fly` and `SetRotor`. The briefing stays text on the existing painting.

Build this phase in the editor with the tools from the chunks above. Add, drag, and delete until the route is right, then S. Do not hand-write the placements the editor can save. Regions are still not written by S. Paint the rings in the file as the other region phases are stored.

The overview marks each new thing, as small shapes on the diagram, not the battle sprites: tent, handover, home, reinforced door, factory, computer, jeep, tank, turret, chopper, hostage, and leader. A wreck stays marked until the phase ends.

Update the tests and docs that lock the old 24×16 phase: `internal/data/mission_test.go`, `internal/data/region_test.go`, `internal/data/region_cells_test.go`, `internal/mapedit/editor_test.go`, `internal/mapedit/place_test.go`, and the Mission 1 lines in `docs/ARCHITECTURE.md`. `TestPhasesFillTheScreen` is a minimum and still holds.

## Still later

Not part of the gap list. Leave them until the user asks.

- Title tune. New music.
- Missions 6–24 as phase JSON under `data/missions`, listed from `campaign.json`. Use the kinds above. Do not copy Sensible layouts.
- Desert, moors, and underground tiles.
- Fullscreen toggle.
- Headless sim replay of each phase. No record format exists. `go test` is the check until this exists.
- Map editor: write region rings, and draw region types other than tree. S already writes `playerStart`, `enemies`, `buildings`, `pickups`, and `barrels`. A region phase would write `map.regions`. A tile phase would write `map.tiles` with `data.TileRune`. A left drag in the regions view adds a tree region for the session and does not write it. Every campaign phase is already regions. The battle paints those rings and still moves on the cells.

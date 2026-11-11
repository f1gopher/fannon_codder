# Not done

These are not in the game. Do one item, then stop. Do not start one of these to finish something else.

Original assets only. No Amiga graphics, samples, map binaries, or the theme tune. Do not recreate a Sensible Software phase. A new map in the same style is allowed.

When an item lands, add one line to `docs/CHUNKS.md` and delete it here. Update `docs/ARCHITECTURE.md` if behavior changed.

New behavior gets a named test in `internal/sim`. `internal/sim` does not import Ebitengine or `internal/audio`.

Each chunk below lists the sheets and the sounds it adds. The chunk stays here until those play in the battle. A coloured rectangle or a borrowed sheet is only for the sim test. Show one contact sheet of the new cels and wait before writing `assets/art`.

Sheets are one PNG and one JSON. Eight source pixels make one world pixel. Stored rows are E, SE, S, N, NE. The runtime mirrors E to W, SE to SW, and NE to NW. The anchor sits on the sim point. People keep the toy-soldier body already approved: two arms, two legs. No muzzle flash.

The sim only emits a `Cue`. Clips are original synthesis in `internal/audio`. No sampled audio. A launch cue is not the impact. Impacts stay `CueBoom`, and `fx/blast` stays the only explosion painting. `CueDeath` stays the soldier yell. `CueRocket` stays the bazooka. `CueBoard` and `CueExit` stay the mount and dismount for every vehicle a soldier can enter. The skidoo hum stays the skidoo.

No pathfinding. Vehicles and escorts use the axis slide already in `steerToward`. Do not retune grunt sight, hearing, or Mission 1 aggression. The Mission 1 map is the 64×40 jungle route. Do not retune those numbers to make that map easier.

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
- The skidoo, and the jeep as that same vehicle with another sheet and a four-stroke hum. Player drive and gun. Enemy driver, ram, and armed burst. Rockets and blasts kill it. MG fire does not.
- The opening helicopter drops the squad and leaves. It is not a vehicle and does not fight.
- Crates, barrels, mines, water, bridges, ice, cliffs, quicksand.
- Objectives that change the phase: `kill_all_enemy`, `destroy_enemy_buildings`.

Cannon Fodder 1 has one enemy trooper. The extra people are the hostage, the enemy leader, and the civilian. Grenadiers and rocketeers stay as they are. Cannon Fodder 2 switches are out of this list.

## Gaps, in order

The eleven gap chunks are in the game. Mission 1 is the 64×40 jungle route.

## Still later

Not part of the gap list. Leave them until the user asks.

- Title tune. New music.
- Missions 6–24 as phase JSON under `data/missions`, listed from `campaign.json`. Use the kinds above. Do not copy Sensible layouts.
- Desert, moors, and underground tiles.
- Fullscreen toggle.
- Headless sim replay of each phase. No record format exists. `go test` is the check until this exists.
- Map editor: write region rings, and draw region types other than tree. S already writes `playerStart`, `enemies`, `buildings`, `pickups`, and `barrels`. A region phase would write `map.regions`. A tile phase would write `map.tiles` with `data.TileRune`. A left drag in the regions view adds a tree region for the session and does not write it. Every campaign phase is already regions. The battle paints those rings and still moves on the cells.

# Done

Chunks 01–51 and the later changes below are in the game. This is a record of what shipped. How it works now is `docs/ARCHITECTURE.md`. An early line is not a spec.

When you finish an item from `docs/BACKLOG.md`, add one line under "After the numbered plan" and delete it there. Update `docs/ARCHITECTURE.md` if behavior changed.

## 01–20 Campaign through Mission 5

- 01 Bootstrap. Module, window, game loop.
- 02 Title and battle scenes, custom pointer, camera type.
- 03 Player squad walks in file.
- 04 Squad machine gun, corpses, player MG ignores a standing friendly.
- 05 Grunts, kill-all win, wipe loses the phase.
- 06 Mission 1 JSON. Trees block walking.
- 07 Names, sixteen ranks, recruit pool, left status strip.
- 08 Boot Hill, briefing, fail and retry.
- 09 Promotions, +15 recruits, save and load.
- 10 Rank changes MG range, spread, and rate. Grenade and rocket icons.
- 11 Maps larger than the view. Trees block shots. Pointer edge scroll.
- 12 Shallow water, deep water, bridges.
- 13 Split and merge Snake, Eagle, and Panther.
- 14 Door huts, grenades, exploding crates, destroy-buildings objective.
- 15 Mission 2, two phases.
- 16 Status strip finished. Overview map.
- 17 Mission 3. Ice, cliffs, scarce grenade crates.
- 18 Civilians, quicksand, mines.
- 19 Mission 4, four phases. Starting grenades. Grenadiers.
- 20 Bazookas, rocketeers, skidoo, Mission 5.

## 21–23 Posted grunts

- 21 Hold the spawn tile. Limited turn. Reaction before the first shot.
- 22 Burst of three, then a pause. Wide cone.
- 23 A nearby infantry MG round starts that reaction. He still needs a clear shot to fire.

## 24–31 Sound

- 24 Cue bus and gunshot. The sim records cues. Audio plays them.
- 25 One boom for grenades, rockets, mines, and crates.
- 26 Death yell, three pitches. A second kill is silent.
- 27 Grenade whoosh and rocket whoosh.
- 28 Splash, quicksand gulp, crate pickup, board, dismount.
- 29 World cues get quieter with distance.
- 30 Skidoo hum. A loop, not a cue.
- 31 Menu click, win sting, fail sting.

## 32–47 Picture

- 32 Window opens at 1024×768, caps at 4K. Same 320×256 frame.
- 33 Style board accepted. `assets/art/style/` is not drawn. Later art follows it.
- 34 Sprite stage. PNG plus JSON. Missing sheet stays a rectangle.
- 35 Snake idle and walk.
- 36 Snake shoot, throw, death, corpse.
- 37 Snake swim and sink.
- 38 Eagle and Panther are Snake recolors.
- 39 Grunt, grenadier, and rocketeer sheets. No walk cycle.
- 40 Civilian sheets.
- 41 Grass and snow paintings. The battle draws frame 0 only.
- 42 Shallow water, deep water, quicksand, and ice loops.
- 43 Trees, cliff, ramp, bridge, mine.
- 44 Huts, crates, skidoo.
- 45 Muzzle flash, tracer, grenade, rocket, one shared blast.
- 46 Pointer, crosshair, board cursor, status-strip icons.
- 47 Painted title, briefing, and Boot Hill.

## 48–51 The rest of the numbered plan

- 48 A squad you are not controlling turns, waits, then bursts. Range and rate stay that man's rank.
- 49 A grunt from a door walks straight out, then posts. Map-placed men do not.
- 50 Grenadiers and rocketeers turn through the windup, then release.
- 51 An armed enemy skidoo holds gun range and bursts. An unarmed one still rams.

## After the numbered plan

- Mission 2 and later grunts and grenadiers close in. Mission 1 and sandboxes stay posted. Rocketeers stay put. No pathfinding. `a37048e`
- A missing walk sheet draws idle, so a walking enemy is not a rectangle. `409fabd`
- Grass and snow stay on frame 0. `a8ff4fa`
- Ground tiles hide the seam from atlas padding. `ebab65c`
- Rank insignia on the status strip. `f0e5d87`
- High Scoring Heroes on Boot Hill. `19a2322`
- Birds, scrub, snowmen, igloos. They do not block movement or fire. `bd3d66c`
- Pause. `5b97066`
- An MG hit can wound. Another round, a blast, or a ram finishes him. `edeaab7`
- Shooting a corpse shoves it about one tile. Kick is capped. The hop refreshes.

# Controls

Everything you use to play Fannon Codder. The pointer is the mouse, in the game’s own screen, not the window’s pixel size.

Closing the window also ends the game.

This page is generated from the control list in `internal/app/controls.go`. When a binding changes, `go test ./internal/app -run TestControlsDoc` rewrites this file and checks that every key and mouse button the game reads is described here.

## Any screen

- **Escape** — Opens “Quit the game?”. The screen underneath stays frozen until you answer, so a battle press does not also give an order.
- **Y or Enter, while the quit prompt is open** — Quits the game.
- **N or Escape, while the quit prompt is open** — Closes the prompt and returns to the screen you were on.

## Title

- **Left click, or Enter** — Leaves the title for Boot Hill.

## Boot Hill

- **Left click LOAD** — Loads the saved campaign. If there is no save, the hill says “No save”.
- **Left click SAVE** — Saves the campaign after at least one mission has been finished. Earlier than that, the hill says “Save after a mission”.
- **H** — Opens or closes the High Scoring Heroes table. One point is a kill. Grenades, rockets, and the skidoo gun score for the trooper who used them.
- **Left click anywhere else, or Enter** — Opens the next briefing, or the “not implemented yet” card when that mission has no map. While the heroes table is open, this closes the table instead.

## Briefing

- **Left click, or Enter** — Deploys the squad and starts the phase.

## Unfinished mission

- **Left click, or Enter** — Returns to Boot Hill.

## Battle — moving and looking

- **Move the pointer to a playfield edge** — Scrolls the view that way. The view also follows the active leader so the squad stays on screen. The left status strip is not part of the playfield.
- **Left click on open ground** — Orders the active squad to move there. Holding the button and dragging updates the destination while they are already moving.
- **Left click the squad letter on the status strip (S, E, or P), or press 1, 2, or 3** — Selects that squad: 1 Snake, 2 Eagle, 3 Panther. Clicking a trooper in a squad that is not active selects that squad too.
- **Left click a trooper in the active squad** — Highlights or unhighlights that trooper. Highlighted troopers are the ones a split hands to the new squad.
- **Left click the troop logo, or the active squad’s letter again** — Splits the highlighted troopers into the next free squad. With nobody highlighted, the logo still tries to split.

## Battle — firing

- **Hold the right mouse button, or hold Ctrl** — The active squad fires toward the pointer. The cursor becomes a crosshair. Either Ctrl key works. This does nothing while the pointer is on the status strip or the map overlay is open.
- **C** — Switches the special weapon between grenades and the bazooka.
- **Space, or hold the right button and left-click** — Uses the selected special weapon at the pointer. The leader throws a grenade or fires a rocket. This is not a move order. It does nothing on the status strip or while the map overlay is open.
- **Left click the G or R icon on the status strip** — Selects grenades (G) or rockets (R). If any trooper is highlighted, the same click cycles how much of that ammo a split gives the new squad: none, half, then all.

## Battle — vehicles and map

- **Left click a skidoo when the cursor shows board** — The active leader boards that skidoo.
- **Left click when the cursor shows exit** — The leader dismounts.
- **Left click or hold, while the leader is in a vehicle** — Drives toward the pointer. Releasing the button stops the drive order.
- **P, or left click the pause button** — Pauses the battle. Men, guns, vehicles, and the skidoo hum stop. Press P or click the button again to continue. A finished phase does not pause.
- **Left click the M icon** — Opens or closes the map overlay. A left click on the open map closes it.

## Battle — end of a phase

- **Left click, or Enter, after victory or defeat** — Leaves the battlefield for Boot Hill. In a debug sandbox, this returns to the title instead.


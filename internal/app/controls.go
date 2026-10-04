package app

import "strings"

// binding is one player-facing control. Tokens are the ebiten names the game
// actually reads (KeyEnter, MouseButtonLeft). controls.md is rendered from
// this list; TestControlsDoc rewrites that file and rejects any keyboard or
// mouse button used in the program that is not named here.
type binding struct {
	section string
	input   string
	action  string
	tokens  []string
}

// controls is the only place a binding is described. Add a row when you read
// a new key or mouse button, then run:
//
//	go test ./internal/app -run TestControlsDoc
//
// The test rewrites controls.md at the repository root.
func controls() []binding {
	return []binding{
		{
			section: "Any screen",
			input:   "Escape",
			action:  "Opens “Quit the game?”. The screen underneath stays frozen until you answer, so a battle press does not also give an order.",
			tokens:  []string{"KeyEscape"},
		},
		{
			section: "Any screen",
			input:   "Y or Enter, while the quit prompt is open",
			action:  "Quits the game.",
			tokens:  []string{"KeyY", "KeyEnter"},
		},
		{
			section: "Any screen",
			input:   "N or Escape, while the quit prompt is open",
			action:  "Closes the prompt and returns to the screen you were on.",
			tokens:  []string{"KeyN", "KeyEscape"},
		},
		{
			section: "Title",
			input:   "Left click, or Enter",
			action:  "Leaves the title for Boot Hill.",
			tokens:  []string{"MouseButtonLeft", "KeyEnter"},
		},
		{
			section: "Boot Hill",
			input:   "Left click LOAD",
			action:  "Loads the saved campaign. If there is no save, the hill says “No save”.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Boot Hill",
			input:   "Left click SAVE",
			action:  "Saves the campaign after at least one mission has been finished. Earlier than that, the hill says “Save after a mission”.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Boot Hill",
			input:   "H",
			action:  "Opens or closes the High Scoring Heroes table. One point is a kill. Grenades, rockets, and the skidoo gun score for the trooper who used them.",
			tokens:  []string{"KeyH"},
		},
		{
			section: "Boot Hill",
			input:   "Left click anywhere else, or Enter",
			action:  "Opens the next briefing, or the “not implemented yet” card when that mission has no map. While the heroes table is open, this closes the table instead.",
			tokens:  []string{"MouseButtonLeft", "KeyEnter"},
		},
		{
			section: "Briefing",
			input:   "Left click, or Enter",
			action:  "Deploys the squad and starts the phase.",
			tokens:  []string{"MouseButtonLeft", "KeyEnter"},
		},
		{
			section: "Unfinished mission",
			input:   "Left click, or Enter",
			action:  "Returns to Boot Hill.",
			tokens:  []string{"MouseButtonLeft", "KeyEnter"},
		},
		{
			section: "Battle — moving and looking",
			input:   "Move the pointer to a playfield edge",
			action:  "Scrolls the view that way. The view also follows the active leader so the squad stays on screen. The left status strip is not part of the playfield.",
			tokens:  nil,
		},
		{
			section: "Battle — moving and looking",
			input:   "Left click on open ground",
			action:  "Orders the active squad to move there. Holding the button and dragging updates the destination while they are already moving.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Battle — moving and looking",
			input:   "Left click the squad letter on the status strip (S, E, or P), or press 1, 2, or 3",
			action:  "Selects that squad: 1 Snake, 2 Eagle, 3 Panther. Clicking a trooper in a squad that is not active selects that squad too.",
			tokens:  []string{"MouseButtonLeft", "Key1", "Key2", "Key3"},
		},
		{
			section: "Battle — moving and looking",
			input:   "Left click a trooper in the active squad",
			action:  "Highlights or unhighlights that trooper. Highlighted troopers are the ones a split hands to the new squad.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Battle — moving and looking",
			input:   "Left click the troop logo, or the active squad’s letter again",
			action:  "Splits the highlighted troopers into the next free squad. With nobody highlighted, the logo still tries to split.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Battle — firing",
			input:   "Hold the right mouse button, or hold Ctrl",
			action:  "The active squad fires toward the pointer. The cursor becomes a crosshair. Either Ctrl key works. This does nothing while the pointer is on the status strip or the map overlay is open.",
			tokens:  []string{"MouseButtonRight", "KeyControlLeft", "KeyControlRight"},
		},
		{
			section: "Battle — firing",
			input:   "C",
			action:  "Switches the special weapon between grenades and the bazooka.",
			tokens:  []string{"KeyC"},
		},
		{
			section: "Battle — firing",
			input:   "Space, or hold the right button and left-click",
			action:  "Uses the selected special weapon at the pointer. The leader throws a grenade or fires a rocket. This is not a move order. It does nothing on the status strip or while the map overlay is open.",
			tokens:  []string{"KeySpace", "MouseButtonRight", "MouseButtonLeft"},
		},
		{
			section: "Battle — firing",
			input:   "Left click the G or R icon on the status strip",
			action:  "Selects grenades (G) or rockets (R). If any trooper is highlighted, the same click cycles how much of that ammo a split gives the new squad: none, half, then all.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Battle — vehicles and map",
			input:   "Left click a skidoo when the cursor shows board",
			action:  "The active leader boards that skidoo.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Battle — vehicles and map",
			input:   "Left click when the cursor shows exit",
			action:  "The leader dismounts.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Battle — vehicles and map",
			input:   "Left click or hold, while the leader is in a vehicle",
			action:  "Drives toward the pointer. Releasing the button stops the drive order.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Battle — vehicles and map",
			input:   "Left click the M icon",
			action:  "Opens or closes the map overlay. A left click on the open map closes it.",
			tokens:  []string{"MouseButtonLeft"},
		},
		{
			section: "Battle — end of a phase",
			input:   "Left click, or Enter, after victory or defeat",
			action:  "Leaves the battlefield for Boot Hill. In a debug sandbox, this returns to the title instead.",
			tokens:  []string{"MouseButtonLeft", "KeyEnter"},
		},
	}
}

// controlsMarkdown is the player guide at the repository root.
func controlsMarkdown() string {
	var b strings.Builder
	b.WriteString("# Controls\n\n")
	b.WriteString("Everything you use to play Fannon Codder. The pointer is the mouse, in the game’s own screen, not the window’s pixel size.\n\n")
	b.WriteString("Closing the window also ends the game.\n\n")
	b.WriteString("This page is generated from the control list in `internal/app/controls.go`. ")
	b.WriteString("When a binding changes, `go test ./internal/app -run TestControlsDoc` rewrites this file and checks that every key and mouse button the game reads is described here.\n\n")

	section := ""
	for _, row := range controls() {
		if row.section != section {
			if section != "" {
				b.WriteByte('\n')
			}
			section = row.section
			b.WriteString("## ")
			b.WriteString(section)
			b.WriteString("\n\n")
		}
		b.WriteString("- **")
		b.WriteString(row.input)
		b.WriteString("** — ")
		b.WriteString(row.action)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	return b.String()
}

# Not done

These are not in the game. Do one item, then stop. Do not start one of these to finish something else.

Original assets only. No Amiga graphics, samples, map binaries, or the theme tune.

When an item lands, add one line to `docs/CHUNKS.md` and delete it here. Update `docs/ARCHITECTURE.md` if behavior changed.

- Title tune. New music.
- Missions 6–24 as phase JSON under `data/missions`, listed from `campaign.json`. Add the vehicles and objectives below only as a mission needs them.
- Jeep, as a skin on the skidoo. Same type.
- Tanks. Shell and armour.
- Static turrets.
- Choppers. Altitude, landing on a man, homing rockets.
- Hostages, kidnap, and factories.
- `protect_civilians` as a real fail. It parses today and does not change the phase.
- Desert, moors, and underground tiles.
- Fullscreen toggle.
- Headless sim replay of each phase. No record format exists. `go test` is the check until this exists.

Leave these alone unless the user asks: pathfinding, patrols, and retuning Mission 1 by deleting grunts or shortening the gun.

package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fannon-codder/internal/sim"
)

// Campaign lists phase files in play order.
type Campaign struct {
	Phases []string `json:"phases"`
}

// Phase is one mission phase loaded from JSON.
type Phase struct {
	Mission     int       `json:"mission"`
	Phase       int       `json:"phase"`
	Title       string    `json:"title"`
	Briefing    string    `json:"briefing"`
	Deploy      int       `json:"deploy"`
	Terrain     string    `json:"terrain"`
	Map         Tilemap   `json:"map"`
	PlayerStart [2]int    `json:"playerStart"`
	Enemies     []Enemy   `json:"enemies"`
	Objectives  []string  `json:"objectives"`
}

type Tilemap struct {
	W     int      `json:"w"`
	H     int      `json:"h"`
	Tiles []string `json:"tiles"`
}

type Enemy struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Kind string `json:"kind"`
}

func missionsDir() string {
	candidates := []string{
		filepath.Join("data", "missions"),
		filepath.Join("..", "data", "missions"),
		filepath.Join("..", "..", "data", "missions"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return filepath.Join("data", "missions")
}

func LoadCampaign() (*Campaign, error) {
	b, err := os.ReadFile(filepath.Join(missionsDir(), "campaign.json"))
	if err != nil {
		return nil, err
	}
	var c Campaign
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if len(c.Phases) == 0 {
		return nil, fmt.Errorf("campaign has no phases")
	}
	return &c, nil
}

func LoadPhase(name string) (*Phase, error) {
	b, err := os.ReadFile(filepath.Join(missionsDir(), name))
	if err != nil {
		return nil, err
	}
	var p Phase
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// LoadFirstWorld loads campaign.json and builds the first phase.
func LoadFirstWorld() (*sim.World, error) {
	c, err := LoadCampaign()
	if err != nil {
		return nil, err
	}
	p, err := LoadPhase(c.Phases[0])
	if err != nil {
		return nil, err
	}
	return p.World()
}

// World builds a sim.World from this phase (original-feeling map, not a rip).
func (p *Phase) World() (*sim.World, error) {
	m, err := p.Map.toSim()
	if err != nil {
		return nil, err
	}
	w := sim.NewEmpty()
	w.Map = m
	mw, mh := m.PixelSize()
	w.Camera.MapW = mw
	w.Camera.MapH = mh
	w.Objectives = w.Objectives[:0]
	for _, s := range p.Objectives {
		o, ok := sim.ParseObjective(s)
		if !ok {
			return nil, fmt.Errorf("unknown objective %q", s)
		}
		w.Objectives = append(w.Objectives, o)
	}
	if len(w.Objectives) == 0 {
		w.Objectives = []sim.Objective{sim.KillAllEnemy}
	}

	start := sim.TileCenter(p.PlayerStart[0], p.PlayerStart[1])
	n := p.Deploy
	if n < 1 {
		n = 1
	}
	positions := make([]sim.Vec2, n)
	for i := 0; i < n; i++ {
		positions[i] = sim.Vec2{X: start.X - float64(i)*sim.FileSpacing, Y: start.Y}
	}
	w.SpawnPlayerSquad(sim.SquadSnake, positions)

	for _, e := range p.Enemies {
		w.SpawnUnit(sim.SideEnemy, sim.TileCenter(e.X, e.Y))
	}
	return w, nil
}

func (tm Tilemap) toSim() (sim.Map, error) {
	if tm.W <= 0 || tm.H <= 0 {
		return sim.Map{}, fmt.Errorf("invalid map size %dx%d", tm.W, tm.H)
	}
	if len(tm.Tiles) != tm.H {
		return sim.Map{}, fmt.Errorf("map h=%d but %d rows", tm.H, len(tm.Tiles))
	}
	m := sim.Map{W: tm.W, H: tm.H, Tiles: make([]sim.Tile, tm.W*tm.H)}
	for y, row := range tm.Tiles {
		if len(row) != tm.W {
			return sim.Map{}, fmt.Errorf("row %d len %d want %d", y, len(row), tm.W)
		}
		for x, ch := range row {
			t, err := parseTile(ch)
			if err != nil {
				return sim.Map{}, fmt.Errorf("tile %d,%d: %w", x, y, err)
			}
			m.Tiles[y*tm.W+x] = t
		}
	}
	return m, nil
}

func parseTile(ch rune) (sim.Tile, error) {
	switch ch {
	case '.', 'P', 'x', ' ':
		return sim.TileGrass, nil
	case 'T', '#':
		return sim.TileTree, nil
	case '~':
		return sim.TileWaterShallow, nil
	case 'W':
		return sim.TileWaterDeep, nil
	case 'B', '=':
		return sim.TileBridge, nil
	default:
		return 0, fmt.Errorf("unknown tile %q", string(ch))
	}
}

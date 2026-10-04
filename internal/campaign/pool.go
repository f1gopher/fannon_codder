package campaign

import "fmt"

const (
	// StartingRecruits is the Boot Hill queue at a new game (Amiga: 15).
	StartingRecruits = 15
	// RecruitsPerMission join after a full mission (wired in chunk 09).
	RecruitsPerMission = 15
)

// Soldier is one named man in the pool or on the field.
type Soldier struct {
	Name              string `json:"name"`
	Rank              Rank   `json:"rank"`
	Kills             int    `json:"kills,omitempty"`
	PhasesThisMission int    `json:"phasesThisMission,omitempty"`
}

// Pool is the unused recruit queue. Deployed men leave it.
type Pool struct {
	Recruits []Soldier
	nextName int // index into the name list for future +15 (chunk 09)
	names    []string
}

// NewGamePool builds 15 Privates from the start of the name list.
func NewGamePool() *Pool {
	return NewGamePoolNames(LoadNames())
}

// NewGamePoolNames is like NewGamePool with an explicit name list (tests).
func NewGamePoolNames(names []string) *Pool {
	if len(names) == 0 {
		names = []string{"Jools", "Jops", "Stoo"}
	}
	n := StartingRecruits
	if n > len(names) {
		n = len(names)
	}
	p := &Pool{names: names, nextName: n}
	for i := 0; i < n; i++ {
		p.Recruits = append(p.Recruits, Soldier{Name: names[i], Rank: Private})
	}
	return p
}

func (p *Pool) NextNameIndex() int {
	if p == nil {
		return 0
	}
	return p.nextName
}

func (p *Pool) Remaining() int {
	if p == nil {
		return 0
	}
	return len(p.Recruits)
}

// Deploy takes n men: highest rank first, FIFO among equal ranks.
func (p *Pool) Deploy(n int) []Soldier {
	if p == nil || n <= 0 {
		return nil
	}
	if n > len(p.Recruits) {
		n = len(p.Recruits)
	}
	out := make([]Soldier, 0, n)
	for i := 0; i < n; i++ {
		idx := p.pickIndex()
		out = append(out, p.Recruits[idx])
		p.Recruits = append(p.Recruits[:idx], p.Recruits[idx+1:]...)
	}
	return out
}

// Return puts survivors back at the front of the unused queue.
func (p *Pool) Return(men []Soldier) {
	if p == nil || len(men) == 0 {
		return
	}
	p.Recruits = append(append([]Soldier{}, men...), p.Recruits...)
}

// CompleteMission promotes anyone who survived phases this mission, then
// adds RecruitsPerMission new men. missionsCompleted is the count AFTER this mission.
func (p *Pool) CompleteMission(missionsCompleted int) {
	if p == nil {
		return
	}
	for i := range p.Recruits {
		n := p.Recruits[i].PhasesThisMission
		if n <= 0 {
			continue
		}
		p.Recruits[i].Rank = p.Recruits[i].Rank.Add(n)
		p.Recruits[i].PhasesThisMission = 0
	}
	intake := Rank(missionsCompleted / 3) // extra training every 3 missions
	p.AddRecruits(RecruitsPerMission, intake)
}

// AddRecruits appends n men of the given rank, drawing names from the list.
func (p *Pool) AddRecruits(n int, rank Rank) {
	if p == nil || n <= 0 {
		return
	}
	for i := 0; i < n; i++ {
		p.Recruits = append(p.Recruits, Soldier{Name: p.nextRecruitName(), Rank: rank})
	}
}

func (p *Pool) nextRecruitName() string {
	if len(p.names) == 0 {
		p.nextName++
		return fmt.Sprintf("Recruit%d", p.nextName)
	}
	idx := p.nextName % len(p.names)
	gen := p.nextName / len(p.names)
	p.nextName++
	base := p.names[idx]
	if gen == 0 {
		return base
	}
	return fmt.Sprintf("%s%d", base, gen+1)
}

// Restore rebuilds a pool from a save (names list is reloaded from disk).
func RestorePool(recruits []Soldier, nextName int) *Pool {
	p := &Pool{
		Recruits: append([]Soldier{}, recruits...),
		nextName: nextName,
		names:    LoadNames(),
	}
	return p
}

func (p *Pool) pickIndex() int {
	best := 0
	for i := 1; i < len(p.Recruits); i++ {
		if p.Recruits[i].Rank > p.Recruits[best].Rank {
			best = i
		}
	}
	return best
}

package campaign

const (
	// StartingRecruits is the Boot Hill queue at a new game (Amiga: 15).
	StartingRecruits = 15
	// RecruitsPerMission join after a full mission (wired in chunk 09).
	RecruitsPerMission = 15
)

// Soldier is one named man in the pool or on the field.
type Soldier struct {
	Name string
	Rank Rank
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

func (p *Pool) pickIndex() int {
	best := 0
	for i := 1; i < len(p.Recruits); i++ {
		if p.Recruits[i].Rank > p.Recruits[best].Rank {
			best = i
		}
	}
	return best
}

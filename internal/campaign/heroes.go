package campaign

import "sort"

const (
	// HeroShow is how many names the High Scoring Heroes table draws.
	HeroShow = 12
	// HeroKeep is how many fallen scores the bureau remembers.
	HeroKeep = 64
)

// Hero is one line on the High Scoring Heroes table.
type Hero struct {
	Name  string `json:"name"`
	Rank  Rank   `json:"rank"`
	Kills int    `json:"kills"`
}

// Remember keeps a fallen man's score. A later entry with the same name
// replaces a lower one. The bureau drops anyone outside the kept list.
func Remember(fallen []Hero, h Hero) []Hero {
	if h.Kills <= 0 || h.Name == "" {
		return fallen
	}
	out := make([]Hero, 0, len(fallen)+1)
	for _, cur := range fallen {
		if cur.Name == h.Name {
			if cur.Kills > h.Kills {
				h = cur
			}
			continue
		}
		out = append(out, cur)
	}
	out = append(out, h)
	sortHeroes(out)
	if len(out) > HeroKeep {
		out = out[:HeroKeep]
	}
	return out
}

// HighScorers is the table: living men with a score, plus the fallen the
// bureau still remembers. The best HeroShow stay, highest kills first.
func HighScorers(living []Soldier, fallen []Hero) []Hero {
	best := make(map[string]Hero, len(living)+len(fallen))
	for _, s := range living {
		if s.Kills <= 0 || s.Name == "" {
			continue
		}
		best[s.Name] = Hero{Name: s.Name, Rank: s.Rank, Kills: s.Kills}
	}
	for _, h := range fallen {
		if h.Kills <= 0 || h.Name == "" {
			continue
		}
		if cur, ok := best[h.Name]; ok && cur.Kills >= h.Kills {
			continue
		}
		best[h.Name] = h
	}
	out := make([]Hero, 0, len(best))
	for _, h := range best {
		out = append(out, h)
	}
	sortHeroes(out)
	if len(out) > HeroShow {
		out = out[:HeroShow]
	}
	return out
}

func sortHeroes(h []Hero) {
	sort.Slice(h, func(i, j int) bool {
		if h[i].Kills != h[j].Kills {
			return h[i].Kills > h[j].Kills
		}
		if h[i].Rank != h[j].Rank {
			return h[i].Rank > h[j].Rank
		}
		return h[i].Name < h[j].Name
	})
}

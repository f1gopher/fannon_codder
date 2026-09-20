package campaign

import (
	"path/filepath"
	"testing"
)

func TestCompleteMissionNoDeathsMatchesFAQ(t *testing.T) {
	p := NewGamePoolNames(manyNames(40))
	men := p.Deploy(2)
	if p.Remaining() != 13 {
		t.Fatalf("after deploy remaining=%d, want 13", p.Remaining())
	}
	for i := range men {
		men[i].PhasesThisMission = 1
	}
	p.Return(men)
	p.CompleteMission(1)
	// 13 unused + 2 survivors + 15 new = 30
	if p.Remaining() != 30 {
		t.Fatalf("after M1 remaining=%d, want 30", p.Remaining())
	}
	nCpl := 0
	for _, s := range p.Recruits {
		if s.Rank == Corporal {
			nCpl++
			if s.Name != "Jools" && s.Name != "Jops" {
				t.Fatalf("unexpected corporal %s", s.Name)
			}
		}
	}
	if nCpl != 2 {
		t.Fatalf("corporals=%d, want 2 (Jools and Jops)", nCpl)
	}
	// M2 deploys 3 (highest rank first): 2 Cpl + 1 Pte → 27 remaining
	dep := p.Deploy(3)
	if len(dep) != 3 {
		t.Fatalf("deployed %d", len(dep))
	}
	if dep[0].Rank != Corporal || dep[1].Rank != Corporal {
		t.Fatalf("first two should be corporals, got %s %s", dep[0].Rank, dep[1].Rank)
	}
	if p.Remaining() != 27 {
		t.Fatalf("M2 remaining=%d, want 27", p.Remaining())
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p := NewGamePoolNames(manyNames(20))
	men := p.Deploy(2)
	men[0].PhasesThisMission = 1
	p.Return(men[:1])
	p.CompleteMission(1)
	path := filepath.Join(t.TempDir(), "save.json")
	sg := SaveGame{
		PhaseIndex:        1,
		MissionsCompleted: 1,
		Graves:            1,
		Recruits:          p.Recruits,
		NextName:          p.nextName,
	}
	if err := Save(path, sg); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Graves != 1 || got.MissionsCompleted != 1 {
		t.Fatalf("meta %+v", got)
	}
	restored := RestorePool(got.Recruits, got.NextName)
	if restored.Remaining() != p.Remaining() {
		t.Fatalf("pool %d vs %d", restored.Remaining(), p.Remaining())
	}
}

func manyNames(n int) []string {
	out := make([]string, n)
	base := []string{"Jools", "Jops", "Stoo", "Jon", "Richard"}
	for i := 0; i < n; i++ {
		if i < len(base) {
			out[i] = base[i]
			continue
		}
		out[i] = extraPrivates(n)[i%len(extraPrivates(n))] + string(rune('0'+i/10))
	}
	return out
}

package campaign

import "testing"

func TestPoolDeployTwoLeavesThirteen(t *testing.T) {
	p := NewGamePool()
	if p.Remaining() != StartingRecruits {
		t.Fatalf("start remaining=%d, want %d", p.Remaining(), StartingRecruits)
	}
	men := p.Deploy(2)
	if len(men) != 2 {
		t.Fatalf("deployed %d, want 2", len(men))
	}
	if p.Remaining() != 13 {
		t.Fatalf("remaining=%d, want 13", p.Remaining())
	}
}

func TestFreshPoolDeploysSameFirstTwo(t *testing.T) {
	a := NewGamePool().Deploy(2)
	b := NewGamePool().Deploy(2)
	if a[0].Name != b[0].Name || a[1].Name != b[1].Name {
		t.Fatalf("fresh pool should be deterministic: %v vs %v", namesOf(a), namesOf(b))
	}
	if a[0].Name != "Jools" || a[1].Name != "Jops" {
		t.Fatalf("first two should be Jools, Jops; got %s, %s", a[0].Name, a[1].Name)
	}
	if a[0].Rank != Private || a[1].Rank != Private {
		t.Fatal("new conscripts are Privates")
	}
}

func TestHighestRankDeploysFirst(t *testing.T) {
	p := NewGamePoolNames([]string{"Ann", "Bob", "Cal"})
	p.Recruits[1].Rank = Sergeant
	men := p.Deploy(1)
	if men[0].Name != "Bob" {
		t.Fatalf("got %s, want Bob (highest rank, FIFO among equals would be Ann if ranks tied)", men[0].Name)
	}
	if p.Recruits[0].Name != "Ann" {
		t.Fatalf("FIFO among remaining: want Ann first, got %s", p.Recruits[0].Name)
	}
}

func TestLoadNamesHasSensibleHeadAndLength(t *testing.T) {
	names := LoadNames()
	if len(names) < 400 {
		t.Fatalf("got %d names, want at least 400", len(names))
	}
	if names[0] != "Jools" || names[1] != "Jops" || names[2] != "Stoo" {
		t.Fatalf("head=%v, want Jools, Jops, Stoo", names[:3])
	}
}

func namesOf(s []Soldier) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = s[i].Name
	}
	return out
}

package campaign

import "testing"

func TestRememberDropsAnyoneBelowTheKeptList(t *testing.T) {
	var fallen []Hero
	for i := 0; i < HeroKeep+5; i++ {
		fallen = Remember(fallen, Hero{Name: nameAt(i), Rank: Private, Kills: i + 1})
	}
	if len(fallen) != HeroKeep {
		t.Fatalf("kept %d, want %d", len(fallen), HeroKeep)
	}
	if fallen[0].Kills != HeroKeep+5 {
		t.Fatalf("top kills %d", fallen[0].Kills)
	}
	if fallen[len(fallen)-1].Kills != 6 {
		t.Fatalf("floor kills %d, want 6", fallen[len(fallen)-1].Kills)
	}
}

func TestRememberKeepsTheHigherScoreForOneName(t *testing.T) {
	fallen := Remember(nil, Hero{Name: "Jools", Rank: Private, Kills: 2})
	fallen = Remember(fallen, Hero{Name: "Jools", Rank: Corporal, Kills: 1})
	if len(fallen) != 1 || fallen[0].Kills != 2 || fallen[0].Rank != Private {
		t.Fatalf("%+v", fallen)
	}
	fallen = Remember(fallen, Hero{Name: "Jools", Rank: Corporal, Kills: 4})
	if len(fallen) != 1 || fallen[0].Kills != 4 || fallen[0].Rank != Corporal {
		t.Fatalf("%+v", fallen)
	}
}

func TestHighScorersShowsTheLivingAndTheFallen(t *testing.T) {
	living := []Soldier{
		{Name: "Jools", Rank: Corporal, Kills: 3},
		{Name: "Jops", Rank: Private, Kills: 0},
	}
	fallen := []Hero{{Name: "Stoo", Rank: Private, Kills: 5}}
	got := HighScorers(living, fallen)
	if len(got) != 2 || got[0].Name != "Stoo" || got[1].Name != "Jools" {
		t.Fatalf("%+v", got)
	}
}

func TestHighScorersStopsAtTwelve(t *testing.T) {
	var living []Soldier
	for i := 0; i < 20; i++ {
		living = append(living, Soldier{Name: nameAt(i), Rank: Private, Kills: i + 1})
	}
	got := HighScorers(living, nil)
	if len(got) != HeroShow {
		t.Fatalf("showed %d", len(got))
	}
	if got[0].Kills != 20 || got[HeroShow-1].Kills != 9 {
		t.Fatalf("range %d..%d", got[0].Kills, got[HeroShow-1].Kills)
	}
}

func nameAt(i int) string {
	return string(rune('A'+(i%26))) + string(rune('a'+((i/26)%26)))
}

package app

import "testing"

func TestArrivalHoldsTheWorld(t *testing.T) {
	b := NewBattle(NewProgress())
	h := &switchHost{}
	id := b.world.Squads[0].MemberIDs[0]
	u := b.world.Unit(id)
	x0, y0 := u.X, u.Y
	for i := 0; i < 30; i++ {
		if err := b.Update(h); err != nil {
			t.Fatal(err)
		}
	}
	u = b.world.Unit(id)
	if u.X != x0 || u.Y != y0 {
		t.Fatalf("moved to %v,%v during the drop", u.X, u.Y)
	}
	if b.insert == nil || b.insert.Landed() {
		t.Fatal("the squad should still be coming in")
	}
}

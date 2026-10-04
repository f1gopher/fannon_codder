package sim

import "testing"

func TestRoofFliesOffAndCrushes(t *testing.T) {
	w := NewEmpty()
	w.AI = false
	w.Objectives = nil
	w.AddDoorHut(4, 4)
	b := &w.Buildings[0]
	cx, cy := b.X+b.W/2, b.Y+b.H/2
	w.explode(cx-10, cy, 0)
	if b.Alive {
		t.Fatal("hut should be gone")
	}
	if len(w.Roofs) != 1 || !w.Roofs[0].Flying {
		t.Fatalf("roofs=%d flying=%v", len(w.Roofs), len(w.Roofs) == 1 && w.Roofs[0].Flying)
	}
	roof := w.Roofs[0]
	if roof.TX == roof.SX && roof.TY == roof.SY {
		t.Fatal("roof should leave the hut")
	}
	// The blast was west of centre, so the slab flies east.
	if roof.TX <= roof.SX {
		t.Fatalf("landing x=%v, start x=%v", roof.TX, roof.SX)
	}
	under := w.SpawnUnit(SideEnemy, Vec2{X: roof.TX, Y: roof.TY})
	clear := w.SpawnUnit(SideEnemy, Vec2{X: roof.SX - 80, Y: roof.SY - 80})
	underID, clearID := under.ID, clear.ID
	var hutBlast bool
	for _, e := range w.Explosions {
		if e.Life > blastTime && e.Scale > 1 {
			hutBlast = true
		}
	}
	if !hutBlast {
		t.Fatal("shed blast should outlast a grenade flash")
	}
	w.Status = Won
	if !w.Pending() {
		t.Fatal("a roof still in the air should hold the phase")
	}
	w.Status = Playing
	for i := 0; i < int(RoofFlight*60)+5; i++ {
		w.Step(1.0 / 60)
	}
	if w.Roofs[0].Flying || w.Roofs[0].Height != 0 || w.Roofs[0].Angle != 0 {
		t.Fatalf("roof still flying height=%v t=%v angle=%v", w.Roofs[0].Height, w.Roofs[0].T, w.Roofs[0].Angle)
	}
	if w.Unit(underID).Living() {
		t.Fatal("the landing should crush the soldier under it")
	}
	if !w.Unit(clearID).Living() {
		t.Fatal("a soldier clear of the landing should live")
	}
	w.Status = Won
	if w.Pending() {
		t.Fatal("a landed roof should not hold the phase")
	}
}

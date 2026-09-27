package sim

const (
	// EngineIdlePitch is the loop rate at a standstill.
	// EngineTopPitch is the rate at VehicleMaxSpeed.
	// The mixer plays the engine clip at this rate. 1 is the clip as synthesised.
	EngineIdlePitch = 0.75
	EngineTopPitch  = 1.45
)

// EngineHum is the one skidoo loop the mixer should play.
// Pitch is EngineIdlePitch at rest and EngineTopPitch at VehicleMaxSpeed.
type EngineHum struct {
	Run   bool
	Pitch float64
}

// HearEngine decides the loop from the listener's own skidoo and the nearest
// other one. A dead or empty vehicle is silent. The listener's own vehicle
// wins. Another skidoo counts only inside radius (the distance where a cue
// goes silent).
func HearEngine(ownAlive, ownOccupied bool, ownSpeed float64, nearAlive, nearOccupied bool, nearSpeed, nearDist, radius float64) EngineHum {
	if ownAlive && ownOccupied {
		return EngineHum{Run: true, Pitch: enginePitch(ownSpeed)}
	}
	if nearAlive && nearOccupied && nearDist < radius {
		return EngineHum{Run: true, Pitch: enginePitch(nearSpeed)}
	}
	return EngineHum{}
}

func enginePitch(speed float64) float64 {
	if speed < 0 {
		speed = 0
	}
	t := speed / VehicleMaxSpeed
	if t > 1 {
		t = 1
	}
	return EngineIdlePitch + (EngineTopPitch-EngineIdlePitch)*t
}

// EngineHumAt is HearEngine for this world. lx, ly is the listener.
// radius is the Chunk 29 silence distance.
func (w *World) EngineHumAt(lx, ly, radius float64) EngineHum {
	if w == nil {
		return EngineHum{}
	}
	ownAlive, ownOccupied := false, false
	ownSpeed := 0.0
	own := w.leaderVehicle()
	if own != nil {
		ownAlive = own.Alive
		ownOccupied = w.vehicleOccupied(own)
		ownSpeed = hypot(own.VX, own.VY)
	}
	nearAlive, nearOccupied := false, false
	nearSpeed, nearDist := 0.0, 0.0
	best := 0.0
	found := false
	for i := range w.Vehicles {
		v := &w.Vehicles[i]
		if own != nil && v.ID == own.ID {
			continue
		}
		if !v.Alive || !w.vehicleOccupied(v) {
			continue
		}
		d := hypot(v.X-lx, v.Y-ly)
		if d >= radius {
			continue
		}
		if found && d >= best {
			continue
		}
		found = true
		best = d
		nearAlive = true
		nearOccupied = true
		nearSpeed = hypot(v.VX, v.VY)
		nearDist = d
	}
	return HearEngine(ownAlive, ownOccupied, ownSpeed, nearAlive, nearOccupied, nearSpeed, nearDist, radius)
}

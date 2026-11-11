package sim

const (
	// EngineIdlePitch is the loop rate at a standstill.
	// EngineTopPitch is the rate at VehicleMaxSpeed.
	// The mixer plays the engine clip at this rate. 1 is the clip as synthesised.
	EngineIdlePitch = 0.75
	EngineTopPitch  = 1.45
)

// EngineHum is the one vehicle loop the mixer should play.
// Pitch is EngineIdlePitch at rest and EngineTopPitch at VehicleMaxSpeed.
// Voice is KindSkidoo or KindJeep. A jeep does not use the skidoo clip.
type EngineHum struct {
	Run   bool
	Pitch float64
	Voice string
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
	return enginePitchTop(speed, VehicleMaxSpeed)
}

func enginePitchTop(speed, top float64) float64 {
	if speed < 0 {
		speed = 0
	}
	if top <= 0 {
		top = VehicleMaxSpeed
	}
	t := speed / top
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
	nearVoice := ""
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
		nearVoice = engineVoice(v.Kind)
	}
	h := HearEngine(ownAlive, ownOccupied, ownSpeed, nearAlive, nearOccupied, nearSpeed, nearDist, radius)
	if !h.Run {
		return h
	}
	if ownAlive && ownOccupied && own != nil {
		h.Voice = engineVoice(own.Kind)
		h.Pitch = enginePitchTop(ownSpeed, vehicleTop(own.Kind))
		return h
	}
	h.Voice = nearVoice
	h.Pitch = enginePitchTop(nearSpeed, vehicleTop(nearVoice))
	return h
}

func engineVoice(kind string) string {
	switch kind {
	case KindJeep:
		return KindJeep
	case KindTank:
		return KindTank
	default:
		return KindSkidoo
	}
}

func vehicleTop(kind string) float64 {
	if kind == KindTank {
		return VehicleMaxSpeed * TankSpeedMul
	}
	return VehicleMaxSpeed
}

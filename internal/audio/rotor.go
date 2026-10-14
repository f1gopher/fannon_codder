package audio

import "math"

const (
	rotorVolume = 0.30
	// rotorFrames is 0.2 s. The blade pulse and the whine both fit it
	// a whole number of times, so the loop joins on a zero crossing.
	rotorFrames = SampleRate / 5
	rotorBlade  = 10
	rotorWhine  = 40
)

// RotorPCM is one slice of the arrival helicopter, synthesised to loop.
func RotorPCM() []byte {
	out := make([]byte, rotorFrames*4)
	for i := 0; i < rotorFrames; i++ {
		t := float64(i) / SampleRate
		cycle := 2 * math.Pi * t
		// A squared sine is a soft blade thump, zero at both ends.
		pulse := math.Sin(cycle * rotorBlade)
		pulse *= pulse
		s := pulse*0.42 +
			math.Sin(cycle*rotorWhine)*0.20 +
			math.Sin(cycle*rotorWhine*2)*0.05
		putStereo(out, i, s)
	}
	return out
}

func (m *Mixer) initRotor() {
	if m == nil || m.ctx == nil {
		return
	}
	src := newLoopReader(RotorPCM())
	p, err := m.ctx.NewPlayer(src)
	if err != nil {
		return
	}
	p.SetVolume(rotorVolume)
	m.rotor = src
	m.rotorPlayer = p
}

// SetRotor starts or stops the helicopter loop. Calling it every frame
// does not restart the loop.
func (m *Mixer) SetRotor(run bool) {
	if m == nil || m.rotor == nil || m.rotorPlayer == nil {
		return
	}
	if run == m.rotorOn {
		return
	}
	m.rotorOn = run
	if run {
		m.rotorPlayer.Play()
		return
	}
	m.rotorPlayer.Pause()
}

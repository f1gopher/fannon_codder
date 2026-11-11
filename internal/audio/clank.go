package audio

import "math"

// clankSamples is about 80 ms: a dull metal tick, not a blast.
const clankSamples = SampleRate * 80 / 1000

// Clank is the sound of a grenade or a barrel hitting a reinforced door.
// It is synthesised, not taken from the Amiga game.
func Clank() []byte {
	out := make([]byte, clankSamples*4)
	var noise lcg = 0xC1A4
	var phase float64
	for i := 0; i < clankSamples; i++ {
		t := float64(i) / SampleRate
		freq := 210 * math.Exp(-t*18)
		phase += freq / SampleRate
		body := math.Sin(2*math.Pi*phase) * math.Exp(-t*28)
		tick := noise.signed() * math.Exp(-t*90)
		env := 1.0
		if t < 0.002 {
			env = t / 0.002
		}
		remain := float64(clankSamples-1-i) / SampleRate
		if remain < 0.012 {
			env *= remain / 0.012
		}
		putStereo(out, i, (body*0.55+tick*0.22)*env)
	}
	return out
}

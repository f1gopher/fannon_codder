package audio

import "math"

// boomSamples is 180 ms. A crate chain can set off several blasts in one
// frame, so the mixer keeps a few voices and restarts the oldest.
const boomSamples = SampleRate * 180 / 1000

// Boom is an original low blast: a slow noise body plus a falling sub thump.
// It is synthesised, not taken from the Amiga game.
func Boom() []byte {
	out := make([]byte, boomSamples*4)
	var noise lcg = 0xB00B
	var lp, phase float64
	for i := 0; i < boomSamples; i++ {
		t := float64(i) / SampleRate
		white := noise.signed()
		lp += 0.06 * (white - lp)

		freq := 32 + 70*math.Exp(-t*14)
		phase += freq / SampleRate
		thump := math.Sin(2*math.Pi*phase) * math.Exp(-t*9)

		env := math.Exp(-t * 9)
		if t < 0.004 {
			env *= t / 0.004
		}
		remain := float64(boomSamples-1-i) / SampleRate
		if remain < 0.015 {
			env *= remain / 0.015
		}
		putStereo(out, i, (lp*0.45+thump*0.72)*env)
	}
	return out
}

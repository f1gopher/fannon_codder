package audio

import "math"

const (
	clickSamples = SampleRate * 30 / 1000
	winSamples   = SampleRate * 360 / 1000
	failSamples  = SampleRate * 420 / 1000
)

// Click is a short UI tick. The title and the briefing play it.
func Click() []byte {
	return burst(clickSamples, 0xC11C, func(t, u float64, n *lcg) float64 {
		tick := math.Sin(2*math.Pi*1480*t) * math.Exp(-t*36)
		snap := n.signed() * math.Exp(-t*90) * 0.18
		return (tick*0.7 + snap) * bell(u)
	})
}

// Win is two rising notes for a cleared phase.
func Win() []byte {
	return twoNotes(winSamples, 523, 659)
}

// Fail is two falling notes for a wipe or an escape.
func Fail() []byte {
	return twoNotes(failSamples, 392, 262)
}

func twoNotes(samples int, f0, f1 float64) []byte {
	out := make([]byte, samples*4)
	half := samples / 2
	var phase float64
	for i := 0; i < samples; i++ {
		freq := f0
		local := float64(i) / float64(half)
		if i >= half {
			freq = f1
			local = float64(i-half) / float64(samples-half)
		}
		phase += freq / SampleRate
		env := bell(local)
		s := math.Sin(2*math.Pi*phase)*0.62 + math.Sin(4*math.Pi*phase)*0.12
		putStereo(out, i, s*env)
	}
	return out
}

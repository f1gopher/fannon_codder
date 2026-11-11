package audio

import "math"

// throwSamples is 90 ms. The boom on landing is a separate clip.
const throwSamples = SampleRate * 90 / 1000

// rocketSamples is 140 ms. Brighter and longer than the grenade whoosh.
const rocketSamples = SampleRate * 140 / 1000

// shellSamples is about 120 ms. A low cough, shorter and duller than the bazooka.
const shellSamples = SampleRate * 120 / 1000

// Throw is an original soft whoosh for a bomb leaving the hand.
func Throw() []byte {
	out := make([]byte, throwSamples*4)
	var noise lcg = 0xA770
	var slow, band float64
	dur := float64(throwSamples-1) / SampleRate
	for i := 0; i < throwSamples; i++ {
		t := float64(i) / SampleRate
		white := noise.signed()
		slow += 0.05 * (white - slow)
		band += 0.28 * ((white - slow) - band)
		u := t / dur
		env := math.Sin(math.Pi * u)
		putStereo(out, i, band*0.95*env)
	}
	return out
}

// Rocket is an original rising whoosh for a bazooka round leaving the tube.
func Rocket() []byte {
	out := make([]byte, rocketSamples*4)
	var noise lcg = 0xB4E0
	var lp, phase float64
	dur := float64(rocketSamples-1) / SampleRate
	for i := 0; i < rocketSamples; i++ {
		t := float64(i) / SampleRate
		u := t / dur
		white := noise.signed()
		cutoff := 0.08 + 0.62*u
		lp += cutoff * (white - lp)
		air := white - lp
		freq := 160 + 1100*u*u
		phase += freq / SampleRate
		tone := math.Sin(2*math.Pi*phase) * 0.22
		env := math.Sin(math.Pi * u)
		putStereo(out, i, (air*0.62+tone)*env)
	}
	return out
}

// Shell is the tank gun's launch. Low, short, and duller than Rocket.
func Shell() []byte {
	out := make([]byte, shellSamples*4)
	var noise lcg = 0x51C0
	var lp, phase float64
	dur := float64(shellSamples-1) / SampleRate
	for i := 0; i < shellSamples; i++ {
		t := float64(i) / SampleRate
		u := t / dur
		white := noise.signed()
		lp += 0.12 * (white - lp)
		body := white - lp
		phase += (90 + 40*(1-u)) / SampleRate
		thud := math.Sin(2*math.Pi*phase) * 0.55
		env := math.Sin(math.Pi*u) * (1 - u*0.35)
		putStereo(out, i, (body*0.28+thud)*env)
	}
	return out
}

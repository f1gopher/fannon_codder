package audio

import (
	"encoding/binary"
	"math"
)

// gunCrack is the muzzle event, about 40 ms. gunSamples is the whole report,
// crack plus the slaps coming back off a street. A Private fires every 125 ms,
// so the tail of one shot still overlaps the next.
const (
	gunCrack   = SampleRate * 40 / 1000
	gunSamples = SampleRate * 240 / 1000
)

// Gunshot is an original street report: a low bark, a chest thump,
// then three duller slaps. It is synthesised, not taken from a film or the Amiga game.
func Gunshot() []byte {
	dry := make([]float64, gunSamples)
	var noise lcg = 0xA11CE
	var prev, bark, body float64
	var phase, sub float64
	for i := 0; i < gunSamples; i++ {
		t := float64(i) / SampleRate
		white := noise.signed()
		diff := white - prev
		prev = white
		// Two poles take the hiss off the crack so it barks instead of tinkling.
		bark += 0.18 * (diff - bark)
		body += 0.045 * (white - body)

		freq := 62 + 140*math.Exp(-t*28)
		phase += freq / SampleRate
		sub += (freq * 0.5) / SampleRate
		thump := math.Sin(2*math.Pi*phase)*0.72 + math.Sin(2*math.Pi*sub)*0.55

		muzzle := math.Exp(-t * 90)
		chest := math.Exp(-t * 22)
		air := math.Exp(-t * 14)
		// A fraction of a millisecond of rise keeps the speaker from clicking.
		rise := 1.0
		if t < 0.0006 {
			rise = t / 0.0006
		}
		// A few milliseconds of raw edge, then the weight takes over.
		edge := math.Exp(-t * 280)
		dry[i] = (diff*0.36*edge + bark*0.30*muzzle + body*0.40*air + thump*0.9*chest) * rise
	}

	mix := make([]float64, gunSamples)
	copy(mix, dry)
	// Near glass, then the far side of the street. Later slaps are duller.
	addSlap(mix, dry, 38, 0.36, 0.08)
	addSlap(mix, dry, 86, 0.22, 0.05)
	addSlap(mix, dry, 148, 0.13, 0.035)

	out := make([]byte, gunSamples*4)
	const tail = 0.012
	for i := 0; i < gunSamples; i++ {
		s := mix[i]
		remain := float64(gunSamples-1-i) / SampleRate
		if remain < tail {
			s *= remain / tail
		}
		putStereo(out, i, s*0.55)
	}
	return out
}

// addSlap delays the crack and dulls it, the way a wall returns a shot.
func addSlap(dst, src []float64, delayMs int, gain, dull float64) {
	delay := SampleRate * delayMs / 1000
	if delay >= len(src) || dull <= 0 || dull >= 1 {
		return
	}
	var lp float64
	for i := delay; i < len(src); i++ {
		lp += dull * (src[i-delay] - lp)
		dst[i] += lp * gain
	}
}

func putStereo(dst []byte, i int, s float64) {
	if s > 1 {
		s = 1
	} else if s < -1 {
		s = -1
	}
	v := uint16(int16(s * 32767))
	d := dst[i*4:]
	binary.LittleEndian.PutUint16(d[0:2], v)
	binary.LittleEndian.PutUint16(d[2:4], v)
}

type lcg uint32

func (r *lcg) signed() float64 {
	*r = *r*1664525 + 1013904223
	u := float64(*r>>8) / float64(1<<24)
	return u*2 - 1
}

package audio

import (
	"encoding/binary"
	"math"
)

// gunSamples is 40 ms. A Private fires every 125 ms, so voices overlap
// only while a squad shoots on the same frame.
const gunSamples = SampleRate * 40 / 1000

// Gunshot is an original crack: a short noise burst plus a falling thump.
// It is synthesised, not taken from the Amiga game.
func Gunshot() []byte {
	out := make([]byte, gunSamples*4)
	var noise lcg = 0xA11CE
	var prev, lp, phase float64
	for i := 0; i < gunSamples; i++ {
		t := float64(i) / SampleRate
		white := noise.signed()
		crack := white - prev
		prev = white
		lp += 0.35 * (crack - lp)

		freq := 140 + 760*math.Exp(-t*60)
		phase += freq / SampleRate
		thump := math.Sin(2*math.Pi*phase) * math.Exp(-t*55)

		env := math.Exp(-t * 62)
		if t < 0.0008 {
			env *= t / 0.0008
		}
		remain := float64(gunSamples-1-i) / SampleRate
		if remain < 0.003 {
			env *= remain / 0.003
		}
		putStereo(out, i, (lp*0.32+thump*0.28)*env)
	}
	return out
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

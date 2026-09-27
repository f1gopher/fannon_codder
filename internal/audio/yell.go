package audio

import (
	"encoding/binary"
	"math"
)

// yellSamples is 220 ms. Three playback rates share this one shout.
const yellSamples = SampleRate * 220 / 1000

// deathRates spread one yell by about a minor third so a wiped squad
// is not the same sample retriggered. Index is DeathVariant.
var deathRates = [3]float64{0.86, 1, 1.16}

// DeathVariant picks the yell pitch from the unit id stored on the cue.
func DeathVariant(id int) int {
	if id < 0 {
		id = -id
	}
	return id % 3
}

// Yell is an original short shout: a falling voiced stack, not a recording.
func Yell() []byte {
	out := make([]byte, yellSamples*4)
	var noise lcg = 0xD1E
	var phase float64
	for i := 0; i < yellSamples; i++ {
		t := float64(i) / SampleRate
		f0 := 150 + 80*math.Exp(-t*7) + 12*math.Sin(2*math.Pi*5*t)
		phase += f0 / SampleRate
		s := 0.0
		for h := 1; h <= 6; h++ {
			s += math.Sin(2*math.Pi*phase*float64(h)) / float64(h)
		}
		s /= 2.45
		breath := noise.signed() * 0.08 * math.Exp(-t*25)
		env := math.Exp(-t * 5)
		if t < 0.012 {
			env *= t / 0.012
		}
		remain := float64(yellSamples-1-i) / SampleRate
		if remain < 0.02 {
			env *= remain / 0.02
		}
		putStereo(out, i, (s+breath)*env)
	}
	return out
}

// Pitch resamples a 16-bit stereo clip. A rate above 1 is higher and shorter.
// Rate 1 returns the same bytes.
func Pitch(pcm []byte, rate float64) []byte {
	if len(pcm) < 8 || rate <= 0 || math.Abs(rate-1) < 1e-9 {
		out := make([]byte, len(pcm))
		copy(out, pcm)
		return out
	}
	inN := len(pcm) / 4
	outN := int(math.Round(float64(inN) / rate))
	if outN < 2 {
		outN = 2
	}
	out := make([]byte, outN*4)
	last := inN - 1
	for i := 0; i < outN; i++ {
		src := float64(i) * rate
		i0 := int(src)
		if i0 >= last {
			i0 = last - 1
		}
		if i0 < 0 {
			i0 = 0
		}
		frac := src - float64(i0)
		s0 := pcmSample(pcm, i0)
		s1 := pcmSample(pcm, i0+1)
		putStereo(out, i, s0*(1-frac)+s1*frac)
	}
	return out
}

// DeathPCM is the yell at the pitch for this unit id.
func DeathPCM(id int) []byte {
	return Pitch(Yell(), deathRates[DeathVariant(id)])
}

func pcmSample(pcm []byte, i int) float64 {
	v := int16(binary.LittleEndian.Uint16(pcm[i*4:]))
	return float64(v) / 32767
}

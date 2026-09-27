package audio

import "math"

const (
	splashSamples = SampleRate * 110 / 1000
	sinkSamples   = SampleRate * 200 / 1000
	pickupSamples = SampleRate * 70 / 1000
	boardSamples  = SampleRate * 80 / 1000
	exitSamples   = SampleRate * 55 / 1000
)

// Splash is a short watery hit. Wading and swimming share it.
func Splash() []byte {
	return burst(splashSamples, 0x5A71, func(t, u float64, n *lcg) float64 {
		white := n.signed()
		freq := 90 + 380*math.Exp(-t*16)
		return (white*0.35 + math.Sin(2*math.Pi*freq*t)*0.45) * math.Exp(-t*10) * bell(u)
	})
}

// Sink is the gulp when quicksand first sticks. The yell at the end is separate.
func Sink() []byte {
	return burst(sinkSamples, 0x51A0, func(t, u float64, n *lcg) float64 {
		white := n.signed()
		freq := 40 + 70*math.Exp(-t*4)
		return (white*0.18 + math.Sin(2*math.Pi*freq*t)*0.7) * math.Exp(-t*5) * bell(u)
	})
}

// Pickup is a short bright blip for taking a crate.
func Pickup() []byte {
	return burst(pickupSamples, 0xA1C, func(t, u float64, _ *lcg) float64 {
		ping := math.Sin(2*math.Pi*740*t) * math.Exp(-t*28)
		if t > 0.028 {
			ping += math.Sin(2*math.Pi*1180*(t-0.028)) * math.Exp(-(t-0.028)*30) * 0.8
		}
		return ping * 0.7 * bell(u)
	})
}

// Board is the clack of climbing onto a skidoo.
func Board() []byte {
	return burst(boardSamples, 0xB0A2, func(t, u float64, n *lcg) float64 {
		click := n.signed() * math.Exp(-t*40)
		thud := math.Sin(2*math.Pi*110*t) * math.Exp(-t*18)
		return (click*0.45 + thud*0.55) * bell(u)
	})
}

// Exit is a lighter clack for stepping off.
func Exit() []byte {
	return burst(exitSamples, 0xE417, func(t, u float64, n *lcg) float64 {
		click := n.signed() * math.Exp(-t*55)
		tick := math.Sin(2*math.Pi*640*t) * math.Exp(-t*35)
		return (click*0.35 + tick*0.4) * bell(u)
	})
}

func bell(u float64) float64 {
	if u < 0 || u > 1 {
		return 0
	}
	return math.Sin(math.Pi * u)
}

func burst(samples int, seed uint32, sample func(t, u float64, n *lcg) float64) []byte {
	out := make([]byte, samples*4)
	noise := lcg(seed)
	dur := float64(samples-1) / SampleRate
	for i := 0; i < samples; i++ {
		t := float64(i) / SampleRate
		putStereo(out, i, sample(t, t/dur, &noise))
	}
	return out
}

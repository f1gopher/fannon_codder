package audio

import "math"

const (
	clickSamples = SampleRate * 30 / 1000
	saveSamples  = SampleRate * 280 / 1000
)

// winSamples and failSamples are the march lengths. victory.go and defeat.go set them.
// briefSamples and hillSamples are the looping menu beds.
var (
	winSamples   int
	failSamples  int
	briefSamples int
	hillSamples  int
)

// Save is the bright sting when a hostage enters the tent.
// It stays under half a second and is not the victory march.
func Save() []byte {
	return burst(saveSamples, 0x5A7E, func(t, u float64, n *lcg) float64 {
		a := math.Sin(2*math.Pi*660*t) * math.Exp(-t*6)
		b := math.Sin(2*math.Pi*990*t) * math.Exp(-t*7)
		c := math.Sin(2*math.Pi*1320*t) * math.Exp(-t*8)
		air := n.signed() * math.Exp(-t*20) * 0.04
		return (a*0.45 + b*0.35 + c*0.22 + air) * bell(u)
	})
}

// Click is a short UI tick. The title and the briefing play it.
func Click() []byte {
	return burst(clickSamples, 0xC11C, func(t, u float64, n *lcg) float64 {
		tick := math.Sin(2*math.Pi*1480*t) * math.Exp(-t*36)
		snap := n.signed() * math.Exp(-t*90) * 0.18
		return (tick*0.7 + snap) * bell(u)
	})
}

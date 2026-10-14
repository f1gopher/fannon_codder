package audio

import "math"

const (
	clickSamples = SampleRate * 30 / 1000
)

// winSamples and failSamples are the march lengths. victory.go and defeat.go set them.
// briefSamples and hillSamples are the looping menu beds.
var (
	winSamples   int
	failSamples  int
	briefSamples int
	hillSamples  int
)

// Click is a short UI tick. The title and the briefing play it.
func Click() []byte {
	return burst(clickSamples, 0xC11C, func(t, u float64, n *lcg) float64 {
		tick := math.Sin(2*math.Pi*1480*t) * math.Exp(-t*36)
		snap := n.signed() * math.Exp(-t*90) * 0.18
		return (tick*0.7 + snap) * bell(u)
	})
}

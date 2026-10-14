package audio

import "math"

// Boot Hill loop. A slow A-minor lament for the same band, with a flugelhorn
// lead instead of the parade trumpet. Sparse drums. The tune is original.
// It is not Narcissus.
const (
	hillBPM   = 66.0
	hillBeats = 16.0
)

func init() {
	hillSamples = int(math.Round(float64(SampleRate) * (hillBeats * 60 / hillBPM)))
}

// Hill is the looping Boot Hill bed.
func Hill() []byte {
	return renderLoop(scoreHill(), 60/hillBPM, hillSamples, drumPlan{
		kickAmp:    0.2,
		snareOn3:   true,
		snareAmp:   0.1,
		snareUntil: 1e9,
		hatsUntil:  hillBeats,
		hatAmp:     0.018,
		seed:       0xB007, seedR: 0x4111,
	})
}

func scoreHill() []bandNote {
	// Flugel lead and a quiet second. Bars: Am, G, F, Am.
	type pair struct {
		beat, dur, lead, harm float64
	}
	line := []pair{
		{0, 2, 76, 72}, {2, 2, 74, 69},
		{4, 2, 72, 67}, {6, 1.6, 71, 67},
		{8, 2, 72, 69}, {10, 2, 69, 64},
		{12, 2, 71, 67}, {14, 1.6, 69, 64},
	}
	var notes []bandNote
	for _, p := range line {
		notes = append(notes,
			bandNote{p.beat, p.dur, p.lead, 0.26, -0.12, 3, 0},
			bandNote{p.beat, p.dur, p.harm, 0.12, 0.22, 3, 0},
		)
	}
	type stab struct {
		beat, dur float64
		a, b, c   float64
	}
	stabs := []stab{
		{0, 3.5, 57, 60, 64},
		{4, 3.5, 55, 59, 62},
		{8, 3.5, 53, 57, 60},
		{12, 3.4, 57, 60, 64},
	}
	for _, s := range stabs {
		for _, m := range []float64{s.a, s.b, s.c} {
			notes = append(notes, bandNote{s.beat, s.dur, m, 0.05, 0, 1, 0})
		}
	}
	bass := []bandNote{
		{0, 3.5, 45, 0.3, 0, 2, 0},
		{4, 3.5, 43, 0.28, 0, 2, 0},
		{8, 3.5, 41, 0.28, 0, 2, 0},
		{12, 3.4, 45, 0.3, 0, 2, 0},
	}
	return append(notes, bass...)
}

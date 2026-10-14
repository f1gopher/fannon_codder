package audio

import "math"

// Defeat march. Same band as the victory, in D minor and at a funeral pace.
// The tune is original. It is not the Sensible fail sting.
const (
	failBPM   = 104.0
	failBeats = 16.5
	failTail  = 0.7
)

func init() {
	failSamples = int(math.Round(float64(SampleRate) * (failBeats*60/failBPM + failTail)))
}

// Fail is the wipe march. The lead trumpet falls an octave on the last chord.
func Fail() []byte {
	return renderBand(scoreDefeat(), 60/failBPM, failSamples, drumPlan{
		kickAmp:  0.36,
		snareOn2: true, snareAmp: 0.3,
		snareUntil: 12, rollFrom: 12, rollTo: 14,
		hatsUntil: 14, hatAmp: 0.03,
		seed: 0xD1E0, seedR: 0x0A11,
	})
}

func scoreDefeat() []bandNote {
	// Lead and second trumpet. Bars: Dm, Gm, Dm, then a fall onto Dm.
	type pair struct {
		beat, dur, lead, harm float64
	}
	line := []pair{
		{0, 1, 81, 77}, {1, 0.5, 79, 76}, {1.5, 0.5, 77, 74},
		{2, 1, 76, 72}, {3, 1, 74, 69},

		{4, 1, 74, 70}, {5, 0.5, 72, 70}, {5.5, 0.5, 70, 67},
		{6, 1, 69, 65}, {7, 1, 67, 62},

		{8, 1, 69, 65}, {9, 0.5, 67, 64}, {9.5, 0.5, 65, 62},
		{10, 1, 64, 60}, {11, 1, 62, 57},

		{12, 1.2, 74, 69}, {13.2, 0.8, 72, 67},
	}
	var notes []bandNote
	for _, p := range line {
		notes = append(notes,
			bandNote{p.beat, p.dur, p.lead, 0.2, -0.18, 0, 0},
			bandNote{p.beat, p.dur, p.harm, 0.13, 0.28, 0, 0},
		)
	}
	notes = append(notes,
		bandNote{14, 2.5, 74, 0.22, -0.18, 0, -12},
		bandNote{14, 2.5, 69, 0.12, 0.28, 0, 0},
	)
	type stab struct {
		beat, dur float64
		a, b, c   float64
	}
	stabs := []stab{
		{1, 0.9, 62, 65, 69}, {3, 0.9, 62, 65, 69},
		{5, 0.9, 55, 58, 62}, {7, 0.9, 55, 58, 62},
		{9, 0.9, 62, 65, 69}, {11, 0.9, 62, 65, 69},
		{14, 2.3, 62, 65, 69},
	}
	for _, s := range stabs {
		amp := 0.07
		if s.dur > 1 {
			amp = 0.1
		}
		for _, m := range []float64{s.a, s.b, s.c} {
			notes = append(notes, bandNote{s.beat, s.dur, m, amp, 0, 1, 0})
		}
	}
	bass := []bandNote{
		{0, 0.95, 38, 0.34, 0, 2, 0}, {1, 0.95, 45, 0.28, 0, 2, 0},
		{2, 0.95, 41, 0.3, 0, 2, 0}, {3, 0.95, 45, 0.28, 0, 2, 0},
		{4, 0.95, 43, 0.34, 0, 2, 0}, {5, 0.95, 50, 0.26, 0, 2, 0},
		{6, 0.95, 46, 0.28, 0, 2, 0}, {7, 0.95, 43, 0.3, 0, 2, 0},
		{8, 0.95, 38, 0.34, 0, 2, 0}, {9, 0.95, 45, 0.28, 0, 2, 0},
		{10, 0.95, 41, 0.3, 0, 2, 0}, {11, 0.95, 40, 0.28, 0, 2, 0},
		{12, 1.4, 38, 0.36, 0, 2, 0}, {14, 2.3, 38, 0.4, 0, 2, 0},
	}
	return append(notes, bass...)
}

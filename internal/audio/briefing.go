package audio

import "math"

// Briefing loop. Same parade band as the marches, in E minor at a march
// pace: a trumpet call, horn stabs, a pulsing bass, and a snare cadence.
// It is the sound of orders being given. It is not the Sensible Objective cue.
const (
	briefBPM   = 108.0
	briefBeats = 16.0
)

func init() {
	briefSamples = int(math.Round(float64(SampleRate) * (briefBeats * 60 / briefBPM)))
}

// Brief is the looping briefing bed.
func Brief() []byte {
	return renderLoop(scoreBrief(), 60/briefBPM, briefSamples, drumPlan{
		kick2: true, kickAmp: 0.36, kick2Amp: 0.22,
		snareOn1: true, snareOn3: true, snareAmp: 0.28,
		snareUntil: 1e9,
		hatsUntil:  briefBeats, hatAmp: 0.04,
		seed: 0xB21EF, seedR: 0x0B1E,
	})
}

func scoreBrief() []bandNote {
	// Lead and second trumpet. Bars: Em, Am, C, B, back to Em.
	type pair struct {
		beat, dur, lead, harm float64
	}
	line := []pair{
		{0, 0.5, 76, 71}, {0.5, 0.5, 79, 74}, {1, 0.5, 81, 76}, {1.5, 0.5, 79, 74},
		{2, 1, 76, 71}, {3, 1, 74, 71},

		{4, 0.5, 72, 67}, {4.5, 0.5, 74, 69}, {5, 0.5, 76, 72}, {5.5, 0.5, 74, 69},
		{6, 1, 72, 67}, {7, 1, 71, 67},

		{8, 0.5, 72, 67}, {8.5, 0.5, 76, 72}, {9, 1, 79, 74},
		{10, 0.5, 76, 72}, {10.5, 0.5, 74, 71}, {11, 1, 72, 67},

		{12, 0.5, 83, 78}, {12.5, 0.5, 81, 76}, {13, 0.5, 79, 74}, {13.5, 0.5, 78, 74},
		{14, 1.5, 76, 71},
	}
	var notes []bandNote
	for _, p := range line {
		notes = append(notes,
			bandNote{p.beat, p.dur, p.lead, 0.18, -0.16, 0, 0},
			bandNote{p.beat, p.dur, p.harm, 0.11, 0.24, 0, 0},
		)
	}
	type stab struct {
		beat, dur float64
		a, b, c   float64
	}
	stabs := []stab{
		{0, 3.4, 64, 67, 71},
		{4, 3.4, 60, 64, 67},
		{8, 3.4, 60, 64, 67},
		{12, 3.4, 59, 66, 71},
	}
	for _, s := range stabs {
		for _, m := range []float64{s.a, s.b, s.c} {
			notes = append(notes, bandNote{s.beat, s.dur, m, 0.055, 0, 1, 0})
		}
	}
	bass := []bandNote{
		{0, 0.9, 40, 0.32, 0, 2, 0}, {1, 0.9, 47, 0.26, 0, 2, 0},
		{2, 0.9, 43, 0.28, 0, 2, 0}, {3, 0.9, 40, 0.3, 0, 2, 0},
		{4, 0.9, 45, 0.32, 0, 2, 0}, {5, 0.9, 52, 0.24, 0, 2, 0},
		{6, 0.9, 48, 0.26, 0, 2, 0}, {7, 0.9, 45, 0.28, 0, 2, 0},
		{8, 0.9, 48, 0.32, 0, 2, 0}, {9, 0.9, 55, 0.22, 0, 2, 0},
		{10, 0.9, 52, 0.24, 0, 2, 0}, {11, 0.9, 48, 0.26, 0, 2, 0},
		{12, 0.9, 47, 0.32, 0, 2, 0}, {13, 0.9, 42, 0.26, 0, 2, 0},
		{14, 0.9, 43, 0.26, 0, 2, 0}, {15, 0.7, 40, 0.3, 0, 2, 0},
	}
	return append(notes, bass...)
}

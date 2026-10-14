package audio

import "math"

// Victory march. Original writing for a small parade band: two trumpets,
// horn stabs, a walking bass, and drums. It is in the jaunty spirit of a
// Cannon Fodder phase clear. It is not the Sensible tune.
const (
	winBPM   = 150.0
	winBeats = 24.5
	winTail  = 0.75
)

func init() {
	winSamples = int(math.Round(float64(SampleRate) * (winBeats*60/winBPM + winTail)))
}

type bandNote struct {
	beat  float64
	dur   float64
	midi  float64
	amp   float64
	pan   float64
	kind  int     // 0 trumpet, 1 horn, 2 bass
	slide float64 // semitones across the note. Negative falls.
}

type drumPlan struct {
	kick2      bool
	kickAmp    float64
	kick2Amp   float64
	snareOn1   bool
	snareOn3   bool
	snareOn2   bool
	snareAmp   float64
	snareUntil float64
	rollFrom   float64
	rollTo     float64
	hatsUntil  float64
	openHats   bool
	hatAmp     float64
	crashes    []float64
	crashAmp   float64
	seed       uint32
	seedR      uint32
}

// Win is the phase-clear march.
func Win() []byte {
	return renderBand(scoreMarch(), 60/winBPM, winSamples, drumPlan{
		kick2: true, kickAmp: 0.42, kick2Amp: 0.28,
		snareOn1: true, snareOn3: true, snareAmp: 0.34,
		snareUntil: 20, rollFrom: 20, rollTo: 22,
		hatsUntil: 22.2, openHats: true, hatAmp: 0.045,
		crashes: []float64{0, 22}, crashAmp: 0.16,
		seed: 0xC0FFEE, seedR: 0xBEEF,
	})
}

func renderBand(notes []bandNote, beat float64, n int, drums drumPlan) []byte {
	return renderBandN(notes, beat, n, 0, drums)
}

// renderLoop folds a short tail back onto the start so a menu bed can repeat
// without a click where the reverb would have been cut.
func renderLoop(notes []bandNote, beat float64, n int, drums drumPlan) []byte {
	tail := int(math.Round(0.08 * float64(SampleRate)))
	return renderBandN(notes, beat, n, tail, drums)
}

func renderBandN(notes []bandNote, beat float64, n, tail int, drums drumPlan) []byte {
	span := n + tail
	left := make([]float64, span)
	right := make([]float64, span)
	ph := make([][2]float64, len(notes))
	noise := lcg(drums.seed)
	noiseR := lcg(drums.seedR)
	var kickP, hatZ, hatZR float64
	for i := 0; i < span; i++ {
		t := float64(i) / SampleRate
		var l, r float64
		for k, note := range notes {
			s := bandVoice(t, beat, note, &ph[k])
			if s == 0 {
				continue
			}
			pl, pr := panOf(note.pan)
			l += s * pl
			r += s * pr
		}
		l, r, kickP, hatZ, hatZR = addDrums(t, beat, l, r, &noise, &noiseR, kickP, hatZ, hatZR, drums)
		left[i], right[i] = l, r
	}
	room(left, right)
	if tail > 0 {
		for i := 0; i < tail; i++ {
			w := 1 - float64(i)/float64(tail)
			left[i] += left[n+i] * w
			right[i] += right[n+i] * w
		}
		left = left[:n]
		right = right[:n]
	}
	peak := 1e-6
	for i := 0; i < n; i++ {
		if a := math.Abs(left[i]); a > peak {
			peak = a
		}
		if a := math.Abs(right[i]); a > peak {
			peak = a
		}
	}
	gain := 0.78 / peak
	out := make([]byte, n*4)
	var dcL, dcR float64
	for i := 0; i < n; i++ {
		l := math.Tanh(left[i] * gain)
		r := math.Tanh(right[i] * gain)
		l, dcL = blockDC(l, dcL)
		r, dcR = blockDC(r, dcR)
		putLR(out, i, l, r)
	}
	return out
}

func scoreMarch() []bandNote {
	// Lead and second trumpet. Bars: F, Bb, F, C, C, F.
	type pair struct {
		beat, dur, lead, harm float64
	}
	line := []pair{
		{0, 0.5, 72, 69}, {0.5, 0.5, 76, 72}, {1, 0.5, 77, 72}, {1.5, 0.5, 77, 69},
		{2, 0.5, 81, 77}, {2.5, 0.5, 79, 76}, {3, 1, 77, 72},

		{4, 0.5, 79, 76}, {4.5, 0.5, 81, 77}, {5, 0.5, 82, 79}, {5.5, 0.5, 81, 77},
		{6, 0.5, 79, 74}, {6.5, 0.5, 77, 74}, {7, 1, 77, 70},

		{8, 0.5, 81, 77}, {8.5, 0.5, 84, 81}, {9, 1, 84, 81},
		{10, 0.5, 82, 79}, {10.5, 0.5, 86, 82}, {11, 0.5, 84, 81}, {11.5, 0.5, 82, 79},

		{12, 0.5, 79, 76}, {12.5, 0.5, 76, 72}, {13, 0.5, 79, 76}, {13.5, 0.5, 81, 76},
		{14, 0.5, 79, 76}, {14.5, 0.5, 76, 72}, {15, 1, 76, 72},

		{16, 0.5, 79, 76}, {16.5, 0.5, 81, 76}, {17, 0.5, 79, 76}, {17.5, 0.5, 76, 72},
		{18, 0.5, 74, 67}, {18.5, 0.5, 76, 72}, {19, 0.5, 79, 76}, {19.5, 0.5, 76, 72},

		{20, 0.5, 77, 72}, {20.5, 0.5, 81, 77}, {21, 0.5, 84, 81}, {21.5, 0.5, 81, 77},
		{22, 2.5, 77, 72},
	}
	var notes []bandNote
	for _, p := range line {
		notes = append(notes,
			bandNote{p.beat, p.dur, p.lead, 0.2, -0.18, 0, 0},
			bandNote{p.beat, p.dur, p.harm, 0.13, 0.28, 0, 0},
		)
	}
	// Horns on the off-beats. The last stab is the button chord.
	type stab struct {
		beat, dur float64
		a, b, c   float64
	}
	stabs := []stab{
		{1, 0.7, 65, 69, 72}, {3, 0.7, 65, 69, 72},
		{5, 0.7, 58, 62, 65}, {7, 0.7, 58, 62, 65},
		{9, 0.7, 65, 69, 72}, {11, 0.7, 58, 62, 65},
		{13, 0.7, 60, 64, 67}, {15, 0.7, 60, 64, 67},
		{17, 0.7, 60, 64, 67}, {19, 0.7, 60, 64, 67},
		{22, 2.2, 65, 69, 72},
	}
	for _, s := range stabs {
		amp := 0.07
		if s.dur > 1 {
			amp = 0.11
		}
		for _, m := range []float64{s.a, s.b, s.c} {
			notes = append(notes, bandNote{s.beat, s.dur, m, amp, 0, 1, 0})
		}
	}
	// Walking bass under the six bars, then a held root.
	bass := []bandNote{
		{0, 0.9, 41, 0.34, 0, 2, 0}, {1, 0.9, 48, 0.3, 0, 2, 0},
		{2, 0.45, 45, 0.3, 0, 2, 0}, {2.5, 0.45, 48, 0.28, 0, 2, 0}, {3, 0.9, 41, 0.32, 0, 2, 0},
		{4, 0.9, 46, 0.34, 0, 2, 0}, {5, 0.9, 41, 0.3, 0, 2, 0},
		{6, 0.45, 53, 0.26, 0, 2, 0}, {6.5, 0.45, 50, 0.26, 0, 2, 0}, {7, 0.9, 46, 0.32, 0, 2, 0},
		{8, 0.9, 41, 0.34, 0, 2, 0}, {9, 0.9, 48, 0.3, 0, 2, 0},
		{10, 0.9, 46, 0.32, 0, 2, 0}, {11, 0.9, 41, 0.3, 0, 2, 0},
		{12, 0.9, 48, 0.32, 0, 2, 0}, {13, 0.9, 43, 0.3, 0, 2, 0},
		{14, 0.45, 52, 0.26, 0, 2, 0}, {14.5, 0.45, 48, 0.26, 0, 2, 0}, {15, 0.9, 43, 0.3, 0, 2, 0},
		{16, 0.9, 48, 0.32, 0, 2, 0}, {17, 0.9, 52, 0.28, 0, 2, 0},
		{18, 0.9, 55, 0.24, 0, 2, 0}, {19, 0.9, 52, 0.28, 0, 2, 0},
		{20, 1.4, 41, 0.36, 0, 2, 0}, {22, 2.2, 41, 0.4, 0, 2, 0},
	}
	return append(notes, bass...)
}

func bandVoice(t, beat float64, n bandNote, ph *[2]float64) float64 {
	u := t - n.beat*beat
	dur := n.dur * beat
	rel := 0.09
	if n.kind == 2 {
		rel = 0.04
	} else if n.kind == 3 {
		rel = 0.14
	}
	if u < 0 || u >= dur+rel {
		return 0
	}
	env := bandEnv(u, dur, rel, n.kind)
	hz := midiHz(n.midi)
	if n.slide != 0 && dur > 0 {
		p := u / dur
		if p > 1 {
			p = 1
		}
		hz *= math.Pow(2, n.slide*p/12)
	} else if n.kind != 2 && n.kind != 3 && u < 0.035 {
		hz *= math.Pow(2, -0.03*(1-u/0.035))
	}
	if n.kind == 0 && n.slide == 0 && u > 0.07 {
		hz *= 1 + 0.0045*math.Sin(2*math.Pi*5.1*(u-0.07))
	}
	if n.kind == 3 && n.slide == 0 && u > 0.14 {
		hz *= 1 + 0.0032*math.Sin(2*math.Pi*4.4*(u-0.14))
	}
	open := 1 - math.Exp(-u*22)
	cut := 1.8 + 6.5*open
	if n.kind == 1 {
		cut = 1.4 + 3.5*open
	}
	var s float64
	switch n.kind {
	case 2:
		ph[0] += hz / SampleRate
		ph[1] += hz * 2 / SampleRate
		s = math.Sin(2*math.Pi*ph[0])*0.85 + math.Sin(2*math.Pi*ph[1])*0.15
	case 3:
		// Flugelhorn: a sine with a little warmth. The hill lead uses it
		// so the memorial is not another parade trumpet.
		ph[0] += hz / SampleRate
		s = math.Sin(2*math.Pi*ph[0])*0.82 + sawRoll(ph[0], 2.4)*0.18
	default:
		ph[0] += hz / SampleRate
		ph[1] += hz * 1.004 / SampleRate
		s = sawRoll(ph[0], cut)*0.62 + sawRoll(ph[1], cut)*0.38
		if n.kind == 0 {
			s += math.Sin(2*math.Pi*ph[0]*0.5) * 0.22
		}
	}
	return s * env * n.amp
}

func bandEnv(u, dur, rel float64, kind int) float64 {
	attack := 0.014
	if kind == 3 {
		attack = 0.09
	}
	e := 1.0
	if u < attack {
		e = u / attack
	} else if kind == 2 {
		e = math.Exp(-(u - attack) * 5.5)
	} else if kind == 3 {
		e = 0.88 + 0.12*math.Exp(-(u-attack)*0.8)
	} else {
		e = 0.62 + 0.38*math.Exp(-(u-attack)*2.4)
	}
	if u > dur {
		e *= 1 - (u-dur)/rel
	}
	if e < 0 {
		return 0
	}
	return e
}

// sawRoll is a saw smoothed by keeping fewer harmonics as cut falls.
// cut is in harmonic numbers, about 2 when a note starts and about 8 once it is open.
func sawRoll(phase, cut float64) float64 {
	var s float64
	for k := 1; k <= 8; k++ {
		fk := float64(k)
		g := 1 / (1 + (fk/cut)*(fk/cut))
		s += math.Sin(2*math.Pi*phase*fk) / fk * g
	}
	return s * 0.72
}

func addDrums(t, beat, l, r float64, n, nR *lcg, kickP, hatZ, hatZR float64, d drumPlan) (float64, float64, float64, float64, float64) {
	bt := t / beat
	if frac, ok := beatHit(bt, 0, 0.5); ok {
		if frac < 1.0/SampleRate {
			kickP = 0
		}
		drop := 180*math.Exp(-frac*beat*20) + 48
		kickP += drop / SampleRate
		s := math.Sin(2*math.Pi*kickP) * math.Exp(-frac*beat*9)
		l += s * d.kickAmp
		r += s * d.kickAmp
	}
	if d.kick2 {
		if frac, ok := beatHit(bt, 2, 0.5); ok {
			if frac < 1.0/SampleRate {
				kickP = 0
			}
			drop := 160*math.Exp(-frac*beat*20) + 46
			kickP += drop / SampleRate
			s := math.Sin(2*math.Pi*kickP) * math.Exp(-frac*beat*10)
			l += s * d.kick2Amp
			r += s * d.kick2Amp
		}
	}
	// Snare until the roll, then sixteenths into the last chord.
	if bt < d.snareUntil {
		if d.snareOn1 {
			if frac, ok := beatHit(bt, 1, 0.4); ok {
				l, r = l+snare(frac*beat, n)*d.snareAmp, r+snare(frac*beat, nR)*d.snareAmp
			}
		}
		if d.snareOn3 {
			if frac, ok := beatHit(bt, 3, 0.4); ok {
				l, r = l+snare(frac*beat, n)*d.snareAmp, r+snare(frac*beat, nR)*d.snareAmp
			}
		}
		if d.snareOn2 {
			if frac, ok := beatHit(bt, 2, 0.4); ok {
				l, r = l+snare(frac*beat, n)*d.snareAmp, r+snare(frac*beat, nR)*d.snareAmp
			}
		}
	} else if bt < d.rollTo {
		base := math.Floor(bt*4) / 4
		frac := bt - base
		if frac < 0.22 {
			span := d.rollTo - d.rollFrom
			if span < 0.1 {
				span = 0.1
			}
			amp := d.snareAmp * (0.45 + 0.7*(bt-d.rollFrom)/span)
			l += snare(frac*beat, n) * amp
			r += snare(frac*beat, nR) * amp
		}
	}
	// Closed hats on the eighths. The last eighth of each bar is open on a win.
	if bt < d.hatsUntil {
		eighth := math.Floor(bt*2) / 2
		frac := bt - eighth
		if frac < 0.35 {
			open := d.openHats && int(math.Floor(bt+1e-6))%4 == 3 && frac > 0 && eighth-math.Floor(eighth) > 0.4
			decay := 38.0
			amp := d.hatAmp
			if open {
				decay = 9
				amp = d.hatAmp * 1.55
			}
			z := n.signed()
			zR := nR.signed()
			hatZ += 0.35 * (z - hatZ)
			hatZR += 0.35 * (zR - hatZR)
			h := (z - hatZ) * math.Exp(-frac*beat*decay) * amp
			hR := (zR - hatZR) * math.Exp(-frac*beat*decay) * amp
			l += h
			r += hR
		}
	}
	for _, hit := range d.crashes {
		u := t - hit*beat
		if u < 0 || u > 1.3 {
			continue
		}
		metal := math.Sin(2*math.Pi*310*u)*0.35 + math.Sin(2*math.Pi*470*u)*0.25 + math.Sin(2*math.Pi*780*u)*0.15
		s := (n.signed()*0.55 + metal) * math.Exp(-u*3.2) * d.crashAmp
		l += s
		r += s * 0.85
	}
	return l, r, kickP, hatZ, hatZR
}

func snare(u float64, n *lcg) float64 {
	if u < 0 {
		return 0
	}
	body := math.Sin(2*math.Pi*188*u) * math.Exp(-u*45)
	return body*0.45 + n.signed()*math.Exp(-u*16)*0.7
}

// beatHit is the time since the start of a beat whose index mod 4 equals mod.
func beatHit(beats float64, mod int, window float64) (float64, bool) {
	if beats < 0 {
		return 0, false
	}
	base := math.Floor(beats)
	if int(base)%4 != mod {
		return 0, false
	}
	frac := beats - base
	if frac >= window {
		return 0, false
	}
	return frac, true
}

func panOf(p float64) (float64, float64) {
	return math.Sqrt(0.5 * (1 - p)), math.Sqrt(0.5 * (1 + p))
}

// room is a short stereo slap so the band is not dry.
func room(left, right []float64) {
	dL := int(math.Round(0.081 * SampleRate))
	dR := int(math.Round(0.097 * SampleRate))
	bufL := make([]float64, dL)
	bufR := make([]float64, dR)
	for i := range left {
		wetL := bufL[i%dL]
		wetR := bufR[i%dR]
		bufL[i%dL] = left[i]*0.42 + wetL*0.38
		bufR[i%dR] = right[i]*0.42 + wetR*0.38
		left[i] += wetL * 0.34
		right[i] += wetR * 0.34
	}
}

func blockDC(x, state float64) (float64, float64) {
	y := x - state
	state += y * 0.002
	return y, state
}

func midiHz(n float64) float64 {
	return 440 * math.Pow(2, (n-69)/12)
}

func putLR(dst []byte, i int, left, right float64) {
	putOne(dst[i*4:], left)
	putOne(dst[i*4+2:], right)
}

func putOne(d []byte, s float64) {
	if s > 1 {
		s = 1
	} else if s < -1 {
		s = -1
	}
	v := uint16(int16(s * 32767))
	d[0] = byte(v)
	d[1] = byte(v >> 8)
}

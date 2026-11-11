package audio

import (
	"encoding/binary"
	"math"
	"sync"

	"fannon-codder/internal/sim"
)

const (
	engineVolume = 0.34
	// engineFrames is 0.2 s. 50 Hz fits it an integer number of times,
	// so the clip wraps without a click.
	engineFrames = SampleRate / 5
	engineHz     = 50
)

// EnginePCM is one revolution of the skidoo hum, synthesised to loop.
// The mixer changes pitch by reading it faster or slower. Ebitengine
// players have no playback-rate control, so the rate lives in the reader.
func EnginePCM() []byte {
	out := make([]byte, engineFrames*4)
	for i := 0; i < engineFrames; i++ {
		t := float64(i) / SampleRate
		cycle := 2 * math.Pi * t
		am := 0.78 + 0.22*math.Sin(cycle*engineHz/2)
		s := math.Sin(cycle*engineHz)*0.48 +
			math.Sin(cycle*engineHz*2)*0.2 +
			math.Sin(cycle*engineHz*3)*0.07
		putStereo(out, i, s*am)
	}
	return out
}

// JeepPCM is one cycle of a four-stroke hum, lower than the skidoo.
// A power stroke is loud; the next revolution is the quiet half.
// The clip is 0.2 s so 40 Hz wraps on a sample boundary.
func JeepPCM() []byte {
	const hz = 40.0
	out := make([]byte, engineFrames*4)
	for i := 0; i < engineFrames; i++ {
		t := float64(i) / SampleRate
		cycle := 2 * math.Pi * t
		// Positive on the firing revolution, negative on the exhausted one.
		stroke := math.Sin(cycle * hz / 2)
		gain := 0.42
		if stroke > 0 {
			gain = 1
		}
		s := math.Sin(cycle*hz)*0.55 +
			math.Sin(cycle*hz*2)*0.22 +
			math.Sin(cycle*hz*3)*0.06
		putStereo(out, i, s*gain*0.9)
	}
	return out
}

// loopReader loops stereo 16-bit PCM. rate is how fast the play head moves:
// 1 plays the clip as stored, 2 plays it an octave higher.
type loopReader struct {
	pcm  []int16
	n    int
	mu   sync.Mutex
	rate float64
	pos  float64
}

func newLoopReader(pcm []byte) *loopReader {
	frames := len(pcm) / 4
	samples := make([]int16, frames*2)
	for i := 0; i < frames*2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(pcm[i*2:]))
	}
	return &loopReader{pcm: samples, n: frames, rate: 1}
}

func (r *loopReader) setRate(rate float64) {
	if r == nil {
		return
	}
	if rate < 0.05 {
		rate = 0.05
	}
	r.mu.Lock()
	r.rate = rate
	r.mu.Unlock()
}

func (r *loopReader) Read(p []byte) (int, error) {
	if r == nil || r.n < 2 {
		return 0, nil
	}
	frames := len(p) / 4
	if frames == 0 {
		return 0, nil
	}
	r.mu.Lock()
	rate := r.rate
	pos := r.pos
	r.mu.Unlock()
	for i := 0; i < frames; i++ {
		i0 := int(math.Floor(pos)) % r.n
		if i0 < 0 {
			i0 += r.n
		}
		i1 := i0 + 1
		if i1 >= r.n {
			i1 = 0
		}
		frac := pos - math.Floor(pos)
		for ch := 0; ch < 2; ch++ {
			s0 := float64(r.pcm[i0*2+ch])
			s1 := float64(r.pcm[i1*2+ch])
			s := s0*(1-frac) + s1*frac
			binary.LittleEndian.PutUint16(p[i*4+ch*2:], uint16(int16(s)))
		}
		pos += rate
		for pos >= float64(r.n) {
			pos -= float64(r.n)
		}
	}
	r.mu.Lock()
	r.pos = pos
	r.mu.Unlock()
	return frames * 4, nil
}

func (m *Mixer) initEngine() {
	if m == nil || m.ctx == nil {
		return
	}
	src := newLoopReader(EnginePCM())
	p, err := m.ctx.NewPlayer(src)
	if err != nil {
		return
	}
	p.SetVolume(engineVolume)
	m.engine = src
	m.enginePlayer = p
	js := newLoopReader(JeepPCM())
	jp, err := m.ctx.NewPlayer(js)
	if err != nil {
		return
	}
	jp.SetVolume(engineVolume)
	m.jeep = js
	m.jeepPlayer = jp
	ts := newLoopReader(TankPCM())
	tp, err := m.ctx.NewPlayer(ts)
	if err != nil {
		return
	}
	tp.SetVolume(engineVolume)
	m.tank = ts
	m.tankPlayer = tp
}

// TankPCM is one cycle of a low diesel loop, slower than the jeep's four-stroke.
// 25 Hz wraps inside the 0.2 s clip (five cycles).
func TankPCM() []byte {
	const hz = 25.0
	out := make([]byte, engineFrames*4)
	for i := 0; i < engineFrames; i++ {
		t := float64(i) / SampleRate
		cycle := 2 * math.Pi * t
		throb := 0.72 + 0.28*math.Sin(cycle*hz/2)
		s := math.Sin(cycle*hz)*0.62 +
			math.Sin(cycle*hz*2)*0.16 +
			math.Sin(cycle*hz*0.5)*0.12
		putStereo(out, i, s*throb*0.95)
	}
	return out
}

// SetEngine starts or stops one vehicle loop and sets its playback rate.
// voice is sim.KindJeep, sim.KindTank, or the skidoo. The other loops stay paused.
// pitch is EngineIdlePitch at rest and EngineTopPitch at that vehicle's top speed.
// Calling it every frame does not restart the loop.
func (m *Mixer) SetEngine(run bool, pitch float64, voice string) {
	if m == nil || m.engine == nil || m.enginePlayer == nil {
		return
	}
	reader := m.engine
	player := m.enginePlayer
	switch voice {
	case sim.KindJeep:
		if m.jeep == nil || m.jeepPlayer == nil {
			return
		}
		reader = m.jeep
		player = m.jeepPlayer
	case sim.KindTank:
		if m.tank == nil || m.tankPlayer == nil {
			return
		}
		reader = m.tank
		player = m.tankPlayer
	default:
		voice = sim.KindSkidoo
	}
	if !run {
		if m.engineOn {
			m.pauseEngines()
		}
		m.engineOn = false
		m.engineVoice = ""
		return
	}
	reader.setRate(pitch)
	if m.engineOn && m.engineVoice == voice {
		return
	}
	m.pauseEngines()
	m.engineOn = true
	m.engineVoice = voice
	player.Play()
}

func (m *Mixer) pauseEngines() {
	if m.enginePlayer != nil {
		m.enginePlayer.Pause()
	}
	if m.jeepPlayer != nil {
		m.jeepPlayer.Pause()
	}
	if m.tankPlayer != nil {
		m.tankPlayer.Pause()
	}
}

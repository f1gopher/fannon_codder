package audio

import (
	"encoding/binary"
	"math"
	"sync"
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
}

// SetEngine starts or stops the one skidoo loop and sets its playback rate.
// pitch is EngineIdlePitch at rest and EngineTopPitch at full speed.
// Calling it every frame does not restart the loop.
func (m *Mixer) SetEngine(run bool, pitch float64) {
	if m == nil || m.engine == nil || m.enginePlayer == nil {
		return
	}
	if run {
		m.engine.setRate(pitch)
	}
	if run == m.engineOn {
		return
	}
	m.engineOn = run
	if run {
		m.enginePlayer.Play()
		return
	}
	m.enginePlayer.Pause()
}

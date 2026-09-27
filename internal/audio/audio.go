// Package audio plays sim cues. The simulation never imports this package.
// One Mixer owns the process-wide Ebitengine audio context.
//
// A cue kind with no clip loaded is silent. Add a sound by loading 16-bit
// little-endian stereo PCM at SampleRate; Play already walks every cue.
package audio

import (
	"github.com/hajimehoshi/ebiten/v2/audio"

	"fannon-codder/internal/sim"
)

const (
	// SampleRate is the one rate for every clip, including later music.
	SampleRate = 44100
	gunVoices  = 8
	gunVolume  = 0.45
)

// Mixer is the clip registry and the voice pool.
type Mixer struct {
	ctx   *audio.Context
	slots map[sim.CueKind]*clip
}

type clip struct {
	players []*audio.Player
	next    int
	volume  float64
}

// NewMixer opens the audio device and loads the gunshot.
// Calling it twice in one process panics: Ebitengine allows one context.
func NewMixer() *Mixer {
	m := &Mixer{
		ctx:   audio.NewContext(SampleRate),
		slots: map[sim.CueKind]*clip{},
	}
	m.Load(sim.CueGun, Gunshot(), gunVoices, gunVolume)
	return m
}

// Load registers a clip. pcm is 16-bit little-endian stereo at SampleRate.
// voices is how many copies may overlap; the oldest restarts when all are busy.
// volume is the clip's full-loudness level in [0, 1]. Loading a kind again replaces its voices.
func (m *Mixer) Load(kind sim.CueKind, pcm []byte, voices int, volume float64) {
	if m == nil || kind == sim.CueNone || len(pcm) == 0 || voices < 1 {
		return
	}
	if volume < 0 {
		volume = 0
	}
	players := make([]*audio.Player, voices)
	for i := range players {
		p := m.ctx.NewPlayerFromBytes(pcm)
		p.SetVolume(volume)
		players[i] = p
	}
	m.slots[kind] = &clip{players: players, volume: volume}
}

// Play starts one voice per cue. Missing kinds are skipped.
// Cue position is stored for a later distance chunk and does not change volume yet.
func (m *Mixer) Play(cues []sim.Cue) {
	if m == nil {
		return
	}
	for _, c := range cues {
		m.play(c)
	}
}

// PlayKind is for stings that do not come from the world (a click, a phase result).
func (m *Mixer) PlayKind(kind sim.CueKind) {
	m.play(sim.Cue{Kind: kind})
}

func (m *Mixer) play(c sim.Cue) {
	if m == nil {
		return
	}
	cl := m.slots[c.Kind]
	if cl == nil || len(cl.players) == 0 {
		return
	}
	p := cl.players[cl.next]
	cl.next++
	if cl.next >= len(cl.players) {
		cl.next = 0
	}
	if err := p.Rewind(); err != nil {
		return
	}
	p.SetVolume(cl.volume)
	p.Play()
}

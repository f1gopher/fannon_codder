// Package audio plays sim cues. The simulation never imports this package.
// One Mixer owns the process-wide Ebitengine audio context.
//
// A cue kind with no clip loaded is silent. Add a sound by loading 16-bit
// little-endian stereo PCM at SampleRate; Play already walks every cue.
package audio

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"

	"fannon-codder/internal/sim"
)

const (
	// SampleRate is the one rate for every clip, including later music.
	SampleRate   = 44100
	gunVoices    = 8
	gunVolume    = 0.45
	boomVoices   = 4
	boomVolume   = 0.7
	deathVoices  = 4
	deathVolume  = 0.55
	throwVoices  = 4
	throwVolume  = 0.42
	rocketVoices = 3
	rocketVolume = 0.5
	splashVolume = 0.55
	sinkVolume   = 0.5
	pickupVolume = 0.45
	boardVolume  = 0.4
	exitVolume   = 0.4
)

// Mixer is the clip registry and the voice pool.
type Mixer struct {
	ctx   *audio.Context
	slots map[sim.CueKind]*clip
	// lx, ly are the listener in world pixels (the active leader, or the
	// camera centre when there is no leader).
	lx, ly float64
}

type clip struct {
	banks  []voiceBank
	volume float64
}

// voiceBank is one pitch of a clip. A gun or a blast has one bank.
// A death yell has three, chosen from the unit id.
type voiceBank struct {
	players []*audio.Player
	next    int
}

// NewMixer opens the audio device and loads the battle clips.
// Calling it twice in one process panics: Ebitengine allows one context.
func NewMixer() *Mixer {
	m := &Mixer{
		ctx:   audio.NewContext(SampleRate),
		slots: map[sim.CueKind]*clip{},
	}
	m.Load(sim.CueGun, Gunshot(), gunVoices, gunVolume)
	m.Load(sim.CueBoom, Boom(), boomVoices, boomVolume)
	m.load(sim.CueDeath, [][]byte{DeathPCM(0), DeathPCM(1), DeathPCM(2)}, deathVoices, deathVolume)
	m.Load(sim.CueThrow, Throw(), throwVoices, throwVolume)
	m.Load(sim.CueRocket, Rocket(), rocketVoices, rocketVolume)
	m.Load(sim.CueSplash, Splash(), 1, splashVolume)
	m.Load(sim.CueSink, Sink(), 1, sinkVolume)
	m.Load(sim.CuePickup, Pickup(), 1, pickupVolume)
	m.Load(sim.CueBoard, Board(), 1, boardVolume)
	m.Load(sim.CueExit, Exit(), 1, exitVolume)
	return m
}

// Load registers a clip. pcm is 16-bit little-endian stereo at SampleRate.
// voices is how many copies may overlap; the oldest restarts when all are busy.
// volume is the clip's full-loudness level in [0, 1]. Loading a kind again replaces its voices.
func (m *Mixer) Load(kind sim.CueKind, pcm []byte, voices int, volume float64) {
	m.load(kind, [][]byte{pcm}, voices, volume)
}

// load registers one clip. Each entry of pcms is a pitch variant with its own voices.
func (m *Mixer) load(kind sim.CueKind, pcms [][]byte, voices int, volume float64) {
	if m == nil || kind == sim.CueNone || len(pcms) == 0 || voices < 1 {
		return
	}
	if volume < 0 {
		volume = 0
	}
	banks := make([]voiceBank, 0, len(pcms))
	for _, pcm := range pcms {
		if len(pcm) == 0 {
			continue
		}
		players := make([]*audio.Player, voices)
		for i := range players {
			p := m.ctx.NewPlayerFromBytes(pcm)
			p.SetVolume(volume)
			players[i] = p
		}
		banks = append(banks, voiceBank{players: players})
	}
	if len(banks) == 0 {
		return
	}
	m.slots[kind] = &clip{banks: banks, volume: volume}
}

// SetListener is the world point volume is measured from.
// The battle sets it each frame: the active leader, or the camera centre
// when that squad has no leader.
func (m *Mixer) SetListener(x, y float64) {
	if m == nil {
		return
	}
	m.lx, m.ly = x, y
}

// Play starts one voice per cue. Missing kinds are skipped.
// Volume is the clip's own level times DistanceGain from the listener.
func (m *Mixer) Play(cues []sim.Cue) {
	if m == nil {
		return
	}
	for _, c := range cues {
		m.play(c, true)
	}
}

// PlayKind is for stings that do not come from the world (a click, a phase result).
// Those have no position, so they stay at the clip's full volume.
func (m *Mixer) PlayKind(kind sim.CueKind) {
	m.play(sim.Cue{Kind: kind}, false)
}

func (m *Mixer) play(c sim.Cue, positional bool) {
	if m == nil {
		return
	}
	cl := m.slots[c.Kind]
	if cl == nil || len(cl.banks) == 0 {
		return
	}
	bi := 0
	if len(cl.banks) > 1 {
		bi = DeathVariant(c.ID) % len(cl.banks)
	}
	b := &cl.banks[bi]
	if len(b.players) == 0 {
		return
	}
	vol := cl.volume
	if positional {
		vol *= DistanceGain(math.Hypot(c.X-m.lx, c.Y-m.ly))
	}
	if vol <= 0 {
		return
	}
	p := b.players[b.next]
	b.next++
	if b.next >= len(b.players) {
		b.next = 0
	}
	if err := p.Rewind(); err != nil {
		return
	}
	p.SetVolume(vol)
	p.Play()
}

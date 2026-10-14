// Package audio plays sim cues. The simulation never imports this package.
// One Mixer owns the process-wide Ebitengine audio context.
//
// A cue kind with no clip loaded is silent. Add a sound by loading 16-bit
// little-endian stereo PCM at SampleRate; Play already walks every cue.
package audio

import (
	"bytes"
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
	clickVolume  = 0.5
	winVolume    = 0.8
	failVolume   = 0.8
	briefVolume  = 0.75
	hillVolume   = 0.7
)

// Mixer is the clip registry and the voice pool.
type Mixer struct {
	ctx   *audio.Context
	slots map[sim.CueKind]*clip
	// lx, ly are the listener in world pixels (the active leader, or the
	// camera centre when there is no leader).
	lx, ly       float64
	engine       *loopReader
	enginePlayer *audio.Player
	engineOn     bool
	rotor        *loopReader
	rotorPlayer  *audio.Player
	rotorOn      bool
	fade         *stingFade
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
	m.load(sim.CueDeath, [][]byte{DeathPCM(0), DeathPCM(1), DeathPCM(2)}, deathVoices, deathVolume, false)
	m.Load(sim.CueThrow, Throw(), throwVoices, throwVolume)
	m.Load(sim.CueRocket, Rocket(), rocketVoices, rocketVolume)
	m.Load(sim.CueSplash, Splash(), 1, splashVolume)
	m.Load(sim.CueSink, Sink(), 1, sinkVolume)
	m.Load(sim.CuePickup, Pickup(), 1, pickupVolume)
	m.Load(sim.CueBoard, Board(), 1, boardVolume)
	m.Load(sim.CueExit, Exit(), 1, exitVolume)
	m.Load(sim.CueClick, Click(), 1, clickVolume)
	m.Load(sim.CueWin, Win(), 1, winVolume)
	m.Load(sim.CueFail, Fail(), 1, failVolume)
	m.loadLoop(sim.CueBrief, Brief(), briefVolume)
	m.loadLoop(sim.CueHill, Hill(), hillVolume)
	m.initEngine()
	m.initRotor()
	return m
}

// Load registers a clip. pcm is 16-bit little-endian stereo at SampleRate.
// voices is how many copies may overlap; the oldest restarts when all are busy.
// volume is the clip's full-loudness level in [0, 1]. Loading a kind again replaces its voices.
func (m *Mixer) Load(kind sim.CueKind, pcm []byte, voices int, volume float64) {
	m.load(kind, [][]byte{pcm}, voices, volume, false)
}

// loadLoop registers one clip that repeats until it is faded or paused.
func (m *Mixer) loadLoop(kind sim.CueKind, pcm []byte, volume float64) {
	m.load(kind, [][]byte{pcm}, 1, volume, true)
}

// load registers one clip. Each entry of pcms is a pitch variant with its own voices.
// loop keeps a menu bed playing for as long as that screen is up.
func (m *Mixer) load(kind sim.CueKind, pcms [][]byte, voices int, volume float64, loop bool) {
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
		players := make([]*audio.Player, 0, voices)
		for i := 0; i < voices; i++ {
			var p *audio.Player
			if loop {
				src := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
				var err error
				p, err = m.ctx.NewPlayer(src)
				if err != nil {
					continue
				}
			} else {
				p = m.ctx.NewPlayerFromBytes(pcm)
			}
			p.SetVolume(volume)
			players = append(players, p)
		}
		if len(players) == 0 {
			continue
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
	// A new march or menu bed leaves any fade that was easing this same
	// player out, or the new screen would go quiet with it.
	if bedKind(c.Kind) {
		m.dropFromFade(p)
	}
	switch c.Kind {
	case sim.CueBrief:
		m.pauseUnlessFading(sim.CueHill)
	case sim.CueHill:
		m.pauseUnlessFading(sim.CueBrief)
	}
	if err := p.Rewind(); err != nil {
		return
	}
	p.SetVolume(vol)
	p.Play()
}

// MarchFade is how long a march or a menu loop takes to die once the
// player leaves the screen that started it.
const MarchFade = 1.0

type stingFade struct {
	players []*audio.Player
	from    []float64
	left    float64
	dur     float64
}

// FadeMarches eases out a victory march, a defeat march, the briefing, or
// Boot Hill, whichever of those is still playing. A click or a gunshot is
// left alone. Calling it with nothing playing does nothing.
func (m *Mixer) FadeMarches(seconds float64) {
	if m == nil || seconds <= 0 {
		return
	}
	var ps []*audio.Player
	var from []float64
	for _, kind := range []sim.CueKind{sim.CueWin, sim.CueFail, sim.CueBrief, sim.CueHill} {
		cl := m.slots[kind]
		if cl == nil {
			continue
		}
		for bi := range cl.banks {
			for _, p := range cl.banks[bi].players {
				if p != nil && p.IsPlaying() {
					ps = append(ps, p)
					from = append(from, cl.volume)
				}
			}
		}
	}
	if len(ps) == 0 {
		return
	}
	m.fade = &stingFade{players: ps, from: from, left: seconds, dur: seconds}
}

// Tick advances the march fade. The game calls it once a frame.
func (m *Mixer) Tick(dt float64) {
	if m == nil || m.fade == nil {
		return
	}
	m.fade.left -= dt
	level := stingLevel(m.fade.left, m.fade.dur)
	for i, p := range m.fade.players {
		p.SetVolume(m.fade.from[i] * level)
		if level == 0 {
			p.Pause()
		}
	}
	if level == 0 {
		m.fade = nil
	}
}

func bedKind(kind sim.CueKind) bool {
	switch kind {
	case sim.CueWin, sim.CueFail, sim.CueBrief, sim.CueHill:
		return true
	default:
		return false
	}
}

// dropFromFade forgets p so a restart is not turned down by the old fade.
func (m *Mixer) dropFromFade(p *audio.Player) {
	if m == nil || m.fade == nil || p == nil {
		return
	}
	n := 0
	for i, q := range m.fade.players {
		if q == p {
			continue
		}
		m.fade.players[n] = q
		m.fade.from[n] = m.fade.from[i]
		n++
	}
	m.fade.players = m.fade.players[:n]
	m.fade.from = m.fade.from[:n]
	if n == 0 {
		m.fade = nil
	}
}

func (m *Mixer) inFade(p *audio.Player) bool {
	if m == nil || m.fade == nil || p == nil {
		return false
	}
	for _, q := range m.fade.players {
		if q == p {
			return true
		}
	}
	return false
}

// pauseUnlessFading stops a menu loop that would otherwise keep playing
// under the one that just started. A loop already fading is left to finish.
func (m *Mixer) pauseUnlessFading(kind sim.CueKind) {
	if m == nil {
		return
	}
	cl := m.slots[kind]
	if cl == nil {
		return
	}
	for bi := range cl.banks {
		for _, p := range cl.banks[bi].players {
			if p == nil || m.inFade(p) {
				continue
			}
			p.Pause()
		}
	}
}

// stingLevel is 1 at the start of a fade and 0 at the end, with a soft curve.
func stingLevel(left, dur float64) float64 {
	if dur <= 0 || left <= 0 {
		return 0
	}
	u := left / dur
	if u > 1 {
		u = 1
	}
	return u * u * (3 - 2*u)
}

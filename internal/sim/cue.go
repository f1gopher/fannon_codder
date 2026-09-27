package sim

// CueKind identifies a sound. Playback lives outside the sim: a kind with
// no clip loaded is silent, so a later chunk adds a constant, an emit, and
// a clip without changing the bus.
type CueKind uint8

const (
	CueNone CueKind = iota
	CueGun
	CueBoom
	CueDeath
	CueThrow
	CueRocket
)

// maxCues bounds the queue when a frame is not drained. The newest shots
// stay; the battle takes the queue every update.
const maxCues = 64

// Cue is one sound to play. X and Y are world pixels at the source.
// ID is the unit for a kind that varies by who made it (the death yell).
// Volume ignores position until a later chunk.
type Cue struct {
	Kind CueKind
	X, Y float64
	ID   int
}

func (w *World) emit(kind CueKind, x, y float64) {
	w.pushCue(Cue{Kind: kind, X: x, Y: y})
}

func (w *World) emitFrom(kind CueKind, x, y float64, id int) {
	w.pushCue(Cue{Kind: kind, X: x, Y: y, ID: id})
}

func (w *World) pushCue(c Cue) {
	if len(w.Cues) < maxCues {
		w.Cues = append(w.Cues, c)
		return
	}
	copy(w.Cues, w.Cues[1:])
	w.Cues[maxCues-1] = c
}

// TakeCues returns the sounds emitted since the last take and clears the queue.
func (w *World) TakeCues() []Cue {
	if len(w.Cues) == 0 {
		return nil
	}
	out := w.Cues
	w.Cues = nil
	return out
}

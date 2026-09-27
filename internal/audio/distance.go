package audio

// A cue is full volume inside HearNear and silent at HearFar.
// Between them the clip's own volume falls in a straight line.
// Ebitengine players have volume and no pan, so distance is loudness only.
const (
	HearNear = 48
	HearFar  = 320
)

// DistanceGain is the scale applied to a clip's own volume.
// 0 px and HearNear are full. HearFar and beyond are silent.
// Halfway from HearNear to HearFar is half.
func DistanceGain(dist float64) float64 {
	if dist <= HearNear {
		return 1
	}
	if dist >= HearFar {
		return 0
	}
	return (HearFar - dist) / (HearFar - HearNear)
}

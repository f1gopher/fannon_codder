package audio

import "testing"

func TestDistanceGain(t *testing.T) {
	if g := DistanceGain(0); g != 1 {
		t.Fatalf("0 px gain=%v, want 1", g)
	}
	if g := DistanceGain(HearNear); g != 1 {
		t.Fatalf("%v px gain=%v, want 1", HearNear, g)
	}
	if g := DistanceGain(HearFar); g != 0 {
		t.Fatalf("%v px gain=%v, want 0", HearFar, g)
	}
	if g := DistanceGain(HearFar + 40); g != 0 {
		t.Fatalf("past the edge gain=%v, want 0", g)
	}
	mid := float64(HearNear+HearFar) / 2
	if g := DistanceGain(mid); g != 0.5 {
		t.Fatalf("halfway (%v px) gain=%v, want 0.5", mid, g)
	}
	// The crack itself is not scaled. Playback multiplies this gain later.
	if len(Gunshot()) != gunSamples*4 {
		t.Fatalf("gun pcm len=%d, want %d", len(Gunshot()), gunSamples*4)
	}
}

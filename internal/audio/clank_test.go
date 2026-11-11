package audio

import (
	"encoding/binary"
	"testing"
)

func TestClankIsAShortTick(t *testing.T) {
	pcm := Clank()
	again := Clank()
	if len(pcm) != clankSamples*4 || len(pcm) != len(again) {
		t.Fatalf("len=%d", len(pcm))
	}
	ms := clankSamples * 1000 / SampleRate
	if ms < 70 || ms > 90 {
		t.Fatalf("clank is %d ms", ms)
	}
	var peak int
	for i := 0; i < clankSamples; i++ {
		if pcm[i*4] != again[i*4] || pcm[i*4+2] != again[i*4+2] {
			t.Fatal("clank synthesis is not stable")
		}
		l := int(int16(binary.LittleEndian.Uint16(pcm[i*4:])))
		if l < 0 {
			l = -l
		}
		if l > peak {
			peak = l
		}
	}
	if peak < 1000 {
		t.Fatalf("clank peak %d is silent", peak)
	}
}

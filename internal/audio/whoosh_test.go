package audio

import (
	"encoding/binary"
	"testing"
)

func TestWhooshShapes(t *testing.T) {
	throw := Throw()
	rocket := Rocket()
	checkWhoosh(t, "throw", throw, throwSamples)
	checkWhoosh(t, "rocket", rocket, rocketSamples)
	if throwSamples >= rocketSamples || rocketSamples >= boomSamples {
		t.Fatal("the grenade whoosh should be shorter than the rocket, and both shorter than the blast")
	}
	if crossings(throw) >= crossings(rocket) {
		t.Fatal("the rocket should be brighter than the grenade whoosh")
	}
}

func checkWhoosh(t *testing.T, name string, pcm []byte, samples int) {
	t.Helper()
	again := Throw()
	if name == "rocket" {
		again = Rocket()
	}
	if len(pcm) != len(again) || len(pcm) != samples*4 {
		t.Fatalf("%s len=%d, want %d", name, len(pcm), samples*4)
	}
	for i := range pcm {
		if pcm[i] != again[i] {
			t.Fatalf("%s synthesis is not stable", name)
		}
	}
	var peak int
	for i := 0; i < samples; i++ {
		l := int16(binary.LittleEndian.Uint16(pcm[i*4:]))
		r := int16(binary.LittleEndian.Uint16(pcm[i*4+2:]))
		if l != r {
			t.Fatalf("%s sample %d is stereo %d/%d", name, i, l, r)
		}
		mag := int(l)
		if mag < 0 {
			mag = -mag
		}
		if mag > peak {
			peak = mag
		}
	}
	if peak < 4000 || peak > 30000 {
		t.Fatalf("%s peak %d, want a whoosh with headroom", name, peak)
	}
	start := int16(binary.LittleEndian.Uint16(pcm[0:2]))
	end := int16(binary.LittleEndian.Uint16(pcm[len(pcm)-2:]))
	if start > 200 || start < -200 || end > 200 || end < -200 {
		t.Fatalf("%s endpoints %d..%d should be near silence", name, start, end)
	}
}

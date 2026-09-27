package audio

import (
	"encoding/binary"
	"testing"
)

func TestGunshotShape(t *testing.T) {
	a := Gunshot()
	b := Gunshot()
	if len(a) != len(b) || len(a) != gunSamples*4 {
		t.Fatalf("len=%d, want %d", len(a), gunSamples*4)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("gunshot synthesis is not stable")
		}
	}
	var first, last int64
	half := gunSamples / 2
	var peak int
	for i := 0; i < gunSamples; i++ {
		l := int16(binary.LittleEndian.Uint16(a[i*4:]))
		r := int16(binary.LittleEndian.Uint16(a[i*4+2:]))
		if l != r {
			t.Fatalf("sample %d is stereo %d/%d; pan comes later", i, l, r)
		}
		mag := int(l)
		if mag < 0 {
			mag = -mag
		}
		if mag > peak {
			peak = mag
		}
		sq := int64(l) * int64(l)
		if i < half {
			first += sq
		} else {
			last += sq
		}
	}
	if peak < 8000 || peak > 30000 {
		t.Fatalf("peak %d, want a loud crack with headroom", peak)
	}
	if first <= last {
		t.Fatal("the crack should decay")
	}
	start := int16(binary.LittleEndian.Uint16(a[0:2]))
	end := int16(binary.LittleEndian.Uint16(a[len(a)-2:]))
	if start > 200 || start < -200 || end > 200 || end < -200 {
		t.Fatalf("endpoints %d..%d should be near silence", start, end)
	}
}

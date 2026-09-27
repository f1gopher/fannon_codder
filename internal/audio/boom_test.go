package audio

import (
	"encoding/binary"
	"testing"
)

func TestBoomShape(t *testing.T) {
	a := Boom()
	b := Boom()
	if len(a) != len(b) || len(a) != boomSamples*4 {
		t.Fatalf("len=%d, want %d", len(a), boomSamples*4)
	}
	if boomSamples <= gunSamples*3 {
		t.Fatalf("boom is %d samples, gun is %d; the blast should be much longer", boomSamples, gunSamples)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("boom synthesis is not stable")
		}
	}
	var first, last int64
	half := boomSamples / 2
	var peak int
	for i := 0; i < boomSamples; i++ {
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
		t.Fatalf("peak %d, want a loud blast with headroom", peak)
	}
	if first <= last {
		t.Fatal("the blast should decay")
	}
	start := int16(binary.LittleEndian.Uint16(a[0:2]))
	end := int16(binary.LittleEndian.Uint16(a[len(a)-2:]))
	if start > 200 || start < -200 || end > 200 || end < -200 {
		t.Fatalf("endpoints %d..%d should be near silence", start, end)
	}
	if crossings(a) >= crossings(Gunshot())/3 {
		t.Fatal("the blast should sit well below the gun crack")
	}
}

func crossings(pcm []byte) int {
	n := 0
	prev := int16(0)
	samples := len(pcm) / 4
	for i := 0; i < samples; i++ {
		s := int16(binary.LittleEndian.Uint16(pcm[i*4:]))
		if (prev < 0 && s >= 0) || (prev >= 0 && s < 0) {
			n++
		}
		prev = s
	}
	return n
}

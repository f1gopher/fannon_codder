package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestYellShape(t *testing.T) {
	a := Yell()
	b := Yell()
	if len(a) != len(b) || len(a) != yellSamples*4 {
		t.Fatalf("len=%d, want %d", len(a), yellSamples*4)
	}
	if yellSamples <= gunSamples*3 {
		t.Fatal("a yell should be much longer than the crack")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("yell synthesis is not stable")
		}
	}
	var first, last int64
	half := yellSamples / 2
	var peak int
	for i := 0; i < yellSamples; i++ {
		l := int16(binary.LittleEndian.Uint16(a[i*4:]))
		r := int16(binary.LittleEndian.Uint16(a[i*4+2:]))
		if l != r {
			t.Fatalf("sample %d is stereo %d/%d", i, l, r)
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
		t.Fatalf("peak %d, want a shout with headroom", peak)
	}
	if first <= last {
		t.Fatal("the yell should decay")
	}
	start := int16(binary.LittleEndian.Uint16(a[0:2]))
	end := int16(binary.LittleEndian.Uint16(a[len(a)-2:]))
	if start > 200 || start < -200 || end > 200 || end < -200 {
		t.Fatalf("endpoints %d..%d should be near silence", start, end)
	}
}

func TestDeathPitchesFollowUnitID(t *testing.T) {
	if DeathVariant(1) != 1 || DeathVariant(2) != 2 || DeathVariant(3) != 0 {
		t.Fatalf("variants %d %d %d", DeathVariant(1), DeathVariant(2), DeathVariant(3))
	}
	if DeathVariant(4) != DeathVariant(1) || DeathVariant(-2) != DeathVariant(2) {
		t.Fatal("pitch should repeat every third id")
	}
	base := Yell()
	same := DeathPCM(1)
	low := DeathPCM(3)
	high := DeathPCM(2)
	if !bytes.Equal(same, base) {
		t.Fatal("variant 1 should be the synthesised yell")
	}
	if len(low) <= len(base) || len(high) >= len(base) {
		t.Fatalf("lengths low=%d base=%d high=%d", len(low), len(base), len(high))
	}
	if bytes.Equal(low, high) || bytes.Equal(low, base) {
		t.Fatal("the three pitches should differ")
	}
	if !bytes.Equal(DeathPCM(4), DeathPCM(1)) {
		t.Fatal("the same variant should be the same clip")
	}
}

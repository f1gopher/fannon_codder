package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestRotorPCMLoops(t *testing.T) {
	a := RotorPCM()
	b := RotorPCM()
	if !bytes.Equal(a, b) {
		t.Fatal("rotor synthesis is not stable")
	}
	frames := len(a) / 4
	if frames != rotorFrames {
		t.Fatalf("frames=%d, want %d", frames, rotorFrames)
	}
	period := SampleRate / rotorBlade
	if frames%period != 0 {
		t.Fatalf("frames %d is not a whole number of blade cycles", frames)
	}
	if s := int16(binary.LittleEndian.Uint16(a[0:2])); s != 0 {
		t.Fatalf("loop start %d, want silence so the join is clean", s)
	}
}

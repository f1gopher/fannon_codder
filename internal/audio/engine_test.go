package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEnginePCMLoops(t *testing.T) {
	a := EnginePCM()
	b := EnginePCM()
	if !bytes.Equal(a, b) {
		t.Fatal("engine synthesis is not stable")
	}
	frames := len(a) / 4
	if frames != engineFrames {
		t.Fatalf("frames=%d, want %d", frames, engineFrames)
	}
	// 50 Hz fits the buffer an integer number of times, so it wraps cleanly.
	period := SampleRate / engineHz
	if frames%period != 0 {
		t.Fatalf("frames %d is not a whole number of %d-sample cycles", frames, period)
	}
	if s := int16(binary.LittleEndian.Uint16(a[0:2])); s != 0 {
		t.Fatalf("loop start %d, want silence so the join is clean", s)
	}
}

func TestLoopReaderRate(t *testing.T) {
	pcm := make([]byte, 4*4)
	for i := 0; i < 4; i++ {
		putStereo(pcm, i, float64(i+1)/10)
	}
	r := newLoopReader(pcm)
	got := make([]byte, len(pcm))
	n, err := r.Read(got)
	if err != nil || n != len(pcm) {
		t.Fatalf("read %d, %v", n, err)
	}
	if !bytes.Equal(got, pcm) {
		t.Fatal("rate 1 should copy the clip")
	}
	r.setRate(2)
	n, err = r.Read(got)
	if err != nil || n != len(pcm) {
		t.Fatalf("read %d, %v", n, err)
	}
	want := make([]byte, len(pcm))
	putStereo(want, 0, 0.1)
	putStereo(want, 1, 0.3)
	putStereo(want, 2, 0.1)
	putStereo(want, 3, 0.3)
	if !bytes.Equal(got, want) {
		t.Fatal("rate 2 should step two samples per frame and wrap")
	}
}

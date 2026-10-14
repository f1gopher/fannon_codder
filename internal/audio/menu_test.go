package audio

import (
	"bytes"
	"testing"
)

func TestMenuBedsAreTheirOwnMusic(t *testing.T) {
	brief, hill := Brief(), Hill()
	win, fail, click := Win(), Fail(), Click()
	if bytes.Equal(brief, hill) || bytes.Equal(brief, win) || bytes.Equal(hill, fail) || bytes.Equal(brief, click) {
		t.Fatal("briefing and Boot Hill should be their own clips")
	}
	if !bytes.Equal(brief, Brief()) || !bytes.Equal(hill, Hill()) {
		t.Fatal("menu synthesis is not stable")
	}
	if len(brief) != briefSamples*4 || len(hill) != hillSamples*4 {
		t.Fatalf("lengths brief=%d hill=%d", len(brief), len(hill))
	}
	if briefSamples < SampleRate*6 || hillSamples < SampleRate*8 {
		t.Fatalf("loops are short: brief %d hill %d", briefSamples, hillSamples)
	}
	if rms(brief[SampleRate*4:SampleRate*8]) < 0.04 {
		t.Fatal("briefing is quiet in the middle")
	}
	if rms(hill[SampleRate*4:SampleRate*8]) < 0.03 {
		t.Fatal("Boot Hill is quiet in the middle")
	}
	if edge := rms(brief[len(brief)-800:]); edge > 0.08 {
		t.Fatalf("briefing loop ends loud (%v), so the repeat will click", edge)
	}
	if edge := rms(hill[len(hill)-800:]); edge > 0.08 {
		t.Fatalf("Boot Hill loop ends loud (%v), so the repeat will click", edge)
	}
	// The downbeat should be louder than the bar of rest before the repeat.
	if rms(brief[:SampleRate/2]) < rms(brief[len(brief)-SampleRate/2:]) {
		t.Fatal("briefing downbeat is quieter than the end of the loop")
	}
	if rms(hill[:SampleRate/2]) < rms(hill[len(hill)-SampleRate/2:]) {
		t.Fatal("Boot Hill downbeat is quieter than the end of the loop")
	}
}

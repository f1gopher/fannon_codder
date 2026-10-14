package audio

import (
	"bytes"
	"math"
	"testing"
)

func TestStingsAreNotTheGun(t *testing.T) {
	gun := Gunshot()
	click, win, fail := Click(), Win(), Fail()
	if bytes.Equal(click, gun) || bytes.Equal(win, gun) || bytes.Equal(fail, gun) {
		t.Fatal("a scene sting is the gunshot")
	}
	if bytes.Equal(click, win) || bytes.Equal(win, fail) || bytes.Equal(click, fail) {
		t.Fatal("click, win, and fail should be different clips")
	}
	if !bytes.Equal(click, Click()) || !bytes.Equal(win, Win()) || !bytes.Equal(fail, Fail()) {
		t.Fatal("sting synthesis is not stable")
	}
	if len(click) >= len(win) || len(click) >= len(fail) {
		t.Fatal("the click should be shorter than the result stings")
	}
	if len(click) != clickSamples*4 || len(win) != winSamples*4 || len(fail) != failSamples*4 {
		t.Fatalf("lengths click=%d win=%d fail=%d", len(click), len(win), len(fail))
	}
	if winSamples < SampleRate*5 || failSamples < SampleRate*5 {
		t.Fatalf("marches are short: win %d fail %d", winSamples, failSamples)
	}
	if rms(win[:SampleRate*4]) < 0.04 || rms(win[len(win)-SampleRate*4:]) < 0.02 {
		t.Fatal("victory march is quiet at the start or the finish")
	}
	if rms(fail[:SampleRate*4]) < 0.04 || rms(fail[len(fail)-SampleRate*4:]) < 0.02 {
		t.Fatal("defeat march is quiet at the start or the finish")
	}
}

func TestMarchFadeEasesOut(t *testing.T) {
	if stingLevel(MarchFade, MarchFade) != 1 {
		t.Fatal("a fade starts at full level")
	}
	if stingLevel(0, MarchFade) != 0 || stingLevel(-0.1, MarchFade) != 0 {
		t.Fatal("a fade ends at silence")
	}
	mid := stingLevel(MarchFade/2, MarchFade)
	if mid < 0.4 || mid > 0.6 {
		t.Fatalf("mid fade = %v", mid)
	}
	var m *Mixer
	m.FadeMarches(MarchFade)
	m.Tick(0.1)
}

func rms(pcm []byte) float64 {
	if len(pcm) < 4 {
		return 0
	}
	var sum float64
	n := 0
	for i := 0; i+3 < len(pcm); i += 4 {
		v := int16(pcm[i]) | int16(pcm[i+1])<<8
		s := float64(v) / 32767
		sum += s * s
		n++
	}
	return math.Sqrt(sum / float64(n))
}

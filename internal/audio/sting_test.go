package audio

import (
	"bytes"
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
}

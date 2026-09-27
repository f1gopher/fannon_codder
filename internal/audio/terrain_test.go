package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestTerrainClips(t *testing.T) {
	clips := []struct {
		name    string
		pcm     []byte
		samples int
		again   func() []byte
	}{
		{"splash", Splash(), splashSamples, Splash},
		{"sink", Sink(), sinkSamples, Sink},
		{"pickup", Pickup(), pickupSamples, Pickup},
		{"board", Board(), boardSamples, Board},
		{"exit", Exit(), exitSamples, Exit},
	}
	seen := map[int]string{}
	for _, c := range clips {
		if prev, ok := seen[c.samples]; ok {
			t.Fatalf("%s and %s share a length", c.name, prev)
		}
		seen[c.samples] = c.name
		again := c.again()
		if !bytes.Equal(c.pcm, again) || len(c.pcm) != c.samples*4 {
			t.Fatalf("%s len=%d stable=%v", c.name, len(c.pcm), bytes.Equal(c.pcm, again))
		}
		var peak int
		for i := 0; i < c.samples; i++ {
			l := int16(binary.LittleEndian.Uint16(c.pcm[i*4:]))
			r := int16(binary.LittleEndian.Uint16(c.pcm[i*4+2:]))
			if l != r {
				t.Fatalf("%s sample %d is stereo %d/%d", c.name, i, l, r)
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
			t.Fatalf("%s peak %d", c.name, peak)
		}
		start := int16(binary.LittleEndian.Uint16(c.pcm[0:2]))
		end := int16(binary.LittleEndian.Uint16(c.pcm[len(c.pcm)-2:]))
		if start > 200 || start < -200 || end > 200 || end < -200 {
			t.Fatalf("%s endpoints %d..%d", c.name, start, end)
		}
	}
}

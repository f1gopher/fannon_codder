package terrain

import "image/color"

// Palette is the flat paint for one terrain: foam, planks, and the mine mark.
// Ground, water, and trees come from that terrain's own plates.
type Palette struct {
	GrassLo, GrassHi     color.RGBA
	SnowLo, SnowHi       color.RGBA
	TreeLo, TreeHi       color.RGBA
	ShallowLo, ShallowHi color.RGBA
	DeepLo, DeepHi       color.RGBA
	IceLo, IceHi         color.RGBA
	SandLo, SandHi       color.RGBA
	CliffLo, CliffHi     color.RGBA
	RampLo, RampHi       color.RGBA
	Foam                 color.RGBA
	Flower               color.RGBA
	Bloom                color.RGBA
	Stone                color.RGBA
	Plank                color.RGBA
	PlankGap             color.RGBA
	Mine                 color.RGBA
	MineMark             color.RGBA
}

func junglePalette() Palette {
	return Palette{
		GrassLo:   color.RGBA{R: 40, G: 92, B: 36, A: 255},
		GrassHi:   color.RGBA{R: 118, G: 162, B: 48, A: 255},
		SnowLo:    color.RGBA{R: 214, G: 222, B: 226, A: 255},
		SnowHi:    color.RGBA{R: 246, G: 250, B: 252, A: 255},
		TreeLo:    color.RGBA{R: 16, G: 64, B: 28, A: 255},
		TreeHi:    color.RGBA{R: 64, G: 132, B: 46, A: 255},
		ShallowLo: color.RGBA{R: 28, G: 150, B: 158, A: 255},
		ShallowHi: color.RGBA{R: 96, G: 206, B: 196, A: 255},
		DeepLo:    color.RGBA{R: 10, G: 58, B: 96, A: 255},
		DeepHi:    color.RGBA{R: 22, G: 104, B: 132, A: 255},
		IceLo:     color.RGBA{R: 158, G: 214, B: 224, A: 255},
		IceHi:     color.RGBA{R: 198, G: 236, B: 248, A: 255},
		SandLo:    color.RGBA{R: 188, G: 142, B: 62, A: 255},
		SandHi:    color.RGBA{R: 220, G: 176, B: 84, A: 255},
		CliffLo:   color.RGBA{R: 96, G: 88, B: 74, A: 255},
		CliffHi:   color.RGBA{R: 154, G: 144, B: 124, A: 255},
		RampLo:    color.RGBA{R: 112, G: 84, B: 48, A: 255},
		RampHi:    color.RGBA{R: 168, G: 128, B: 70, A: 255},
		Foam:      color.RGBA{R: 226, G: 242, B: 236, A: 255},
		Flower:    color.RGBA{R: 232, G: 206, B: 128, A: 255},
		Bloom:     color.RGBA{R: 176, G: 72, B: 48, A: 255},
		Stone:     color.RGBA{R: 90, G: 84, B: 74, A: 255},
		Plank:     color.RGBA{R: 128, G: 86, B: 48, A: 255},
		PlankGap:  color.RGBA{R: 78, G: 52, B: 30, A: 255},
		Mine:      color.RGBA{R: 70, G: 56, B: 36, A: 255},
		MineMark:  color.RGBA{R: 28, G: 22, B: 16, A: 255},
	}
}

func forestPalette() Palette {
	p := junglePalette()
	p.Foam = color.RGBA{R: 214, G: 224, B: 196, A: 255}
	p.Plank = color.RGBA{R: 96, G: 72, B: 44, A: 255}
	p.PlankGap = color.RGBA{R: 54, G: 40, B: 26, A: 255}
	p.TreeLo = color.RGBA{R: 18, G: 52, B: 36, A: 255}
	p.TreeHi = color.RGBA{R: 42, G: 96, B: 64, A: 255}
	return p
}

func desertPalette() Palette {
	p := junglePalette()
	p.Foam = color.RGBA{R: 236, G: 232, B: 210, A: 255}
	p.Plank = color.RGBA{R: 176, G: 144, B: 96, A: 255}
	p.PlankGap = color.RGBA{R: 120, G: 92, B: 58, A: 255}
	p.CliffLo = color.RGBA{R: 150, G: 112, B: 72, A: 255}
	p.CliffHi = color.RGBA{R: 196, G: 160, B: 112, A: 255}
	p.TreeLo = color.RGBA{R: 96, G: 78, B: 36, A: 255}
	p.TreeHi = color.RGBA{R: 168, G: 140, B: 72, A: 255}
	return p
}

func snowPalette() Palette {
	p := junglePalette()
	p.Foam = color.RGBA{R: 246, G: 250, B: 252, A: 255}
	p.CliffLo = color.RGBA{R: 78, G: 86, B: 94, A: 255}
	p.CliffHi = color.RGBA{R: 132, G: 142, B: 150, A: 255}
	p.TreeLo = color.RGBA{R: 22, G: 58, B: 48, A: 255}
	p.TreeHi = color.RGBA{R: 78, G: 122, B: 112, A: 255}
	p.Stone = color.RGBA{R: 70, G: 78, B: 86, A: 255}
	p.Plank = color.RGBA{R: 132, G: 136, B: 140, A: 255}
	p.PlankGap = color.RGBA{R: 78, G: 84, B: 90, A: 255}
	return p
}

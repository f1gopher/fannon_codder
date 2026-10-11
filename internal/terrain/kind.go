package terrain

import "image/color"

// Kind is the terrain a phase paints with. The four kinds do not share
// ground, water, or trees. Ice, rock, dirt, and quicksand stay one material.
type Kind uint8

const (
	Jungle Kind = iota
	Forest
	Desert
	Snow
)

// KindOf maps a phase file value onto a kind. "grass" and "" are jungle.
// "arctic" is snow. "forrest" is accepted as forest.
func KindOf(s string) Kind {
	switch s {
	case "forest", "forrest":
		return Forest
	case "desert":
		return Desert
	case "snow", "arctic":
		return Snow
	default:
		return Jungle
	}
}

func (k Kind) String() string {
	switch k {
	case Forest:
		return "forest"
	case Desert:
		return "desert"
	case Snow:
		return "snow"
	default:
		return "jungle"
	}
}

// look is one kind's plates and the paint that is not in those plates.
// Ground, canopy, shallow, and deep are that kind's files alone.
type look struct {
	kind     Kind
	pal      Palette
	ground   *painted
	canopy   *painted
	shallow  *painted
	deep     *painted
	shore    *painted
	open     surface
	shoreMix float64
	bank     color.RGBA
	bankMix  float64
	leaf     color.RGBA
	shadow   color.RGBA
	crown    color.RGBA
	crownMix float64
	snowCaps bool
	rockMix  float64
	rockTint color.RGBA
}

func lookOf(terrain string) look {
	switch KindOf(terrain) {
	case Forest:
		return forestLook
	case Desert:
		return desertLook
	case Snow:
		return snowLook
	default:
		return jungleLook
	}
}

var (
	jungleLook = look{
		kind:     Jungle,
		pal:      junglePalette(),
		open:     surfGrass,
		shoreMix: 0.85,
		bank:     color.RGBA{R: 14, G: 96, B: 128, A: 255},
		bankMix:  0.42,
		leaf:     color.RGBA{R: 70, G: 108, B: 46, A: 255},
		shadow:   color.RGBA{R: 78, G: 98, B: 52, A: 255},
		crown:    color.RGBA{R: 46, G: 78, B: 38, A: 255},
		crownMix: 0.16,
	}
	forestLook = look{
		kind:     Forest,
		pal:      forestPalette(),
		open:     surfGrass,
		shoreMix: 0.75,
		bank:     color.RGBA{R: 18, G: 64, B: 48, A: 255},
		bankMix:  0.46,
		leaf:     color.RGBA{R: 48, G: 82, B: 56, A: 255},
		shadow:   color.RGBA{R: 40, G: 62, B: 36, A: 255},
		crown:    color.RGBA{R: 30, G: 58, B: 48, A: 255},
		crownMix: 0.18,
	}
	desertLook = look{
		kind:     Desert,
		pal:      desertPalette(),
		open:     surfGrass,
		shoreMix: 0.55,
		bank:     color.RGBA{R: 40, G: 96, B: 110, A: 255},
		bankMix:  0.36,
		leaf:     color.RGBA{R: 140, G: 120, B: 64, A: 255},
		shadow:   color.RGBA{R: 120, G: 96, B: 60, A: 255},
		crown:    color.RGBA{R: 120, G: 96, B: 48, A: 255},
		crownMix: 0.14,
		rockMix:  0.25,
		rockTint: color.RGBA{R: 196, G: 160, B: 120, A: 255},
	}
	snowLook = look{
		kind:     Snow,
		pal:      snowPalette(),
		open:     surfSnow,
		shoreMix: 0.40,
		bank:     color.RGBA{R: 70, G: 110, B: 130, A: 255},
		bankMix:  0.34,
		leaf:     color.RGBA{R: 120, G: 150, B: 130, A: 255},
		shadow:   color.RGBA{R: 118, G: 136, B: 146, A: 255},
		crown:    color.RGBA{R: 150, G: 168, B: 172, A: 255},
		crownMix: 0.12,
		snowCaps: true,
		rockMix:  0.35,
		rockTint: color.RGBA{R: 168, G: 184, B: 196, A: 255},
	}
)

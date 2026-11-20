package mapedit

import (
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	typeNone = iota
	typeTitle
	typeBrief
	typeSize
	typeMission
	titleMax   = 80
	briefMax   = 400
	sizeMax    = 32
	missionMax = 60
)

func (e *Editor) typing() bool {
	return e.typeKind != typeNone
}

func (e *Editor) startType(kind int) {
	if e.phase == nil || e.err != nil {
		return
	}
	e.typeKind = kind
	e.typeNote = ""
	switch kind {
	case typeTitle:
		e.draft = e.phase.Title
	case typeBrief:
		e.draft = e.phase.Briefing
	case typeSize:
		e.draft = e.sizeDraft()
	case typeMission:
		e.draft = e.missionName()
	default:
		e.draft = ""
	}
}

func (e *Editor) cancelType() {
	e.typeKind = typeNone
	e.draft = ""
	e.typeNote = ""
}

func (e *Editor) commitType() {
	if e.phase == nil || e.typeKind == typeNone {
		return
	}
	text := strings.TrimSpace(strings.ReplaceAll(e.draft, "\n", " "))
	switch e.typeKind {
	case typeTitle:
		if text != e.phase.Title {
			e.phase.Title = text
			if e.index >= 0 && e.index < len(e.phases) {
				e.phases[e.index].title = text
			}
			e.dirty = true
			e.rebuildRows()
		}
	case typeBrief:
		if text != e.phase.Briefing {
			e.phase.Briefing = text
			e.dirty = true
		}
	case typeSize:
		if !e.applySize(text) {
			return
		}
	case typeMission:
		e.renameMission(text)
	}
	e.cancelType()
}

func (e *Editor) typeKeys() {
	if !e.typing() {
		if inpututil.IsKeyJustPressed(ebiten.KeyT) {
			e.startType(typeTitle)
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyB) {
			e.startType(typeBrief)
			return
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			e.startType(typeSize)
		}
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		e.cancelType()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		e.commitType()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		r := []rune(e.draft)
		if len(r) > 0 {
			e.draft = string(r[:len(r)-1])
		}
		return
	}
	limit := briefMax
	switch e.typeKind {
	case typeTitle:
		limit = titleMax
	case typeSize:
		limit = sizeMax
	case typeMission:
		limit = missionMax
	}
	for _, ch := range ebiten.AppendInputChars(nil) {
		if ch == 0 || unicode.IsControl(ch) {
			continue
		}
		if len([]rune(e.draft)) >= limit {
			break
		}
		e.draft += string(ch)
	}
}

func (e *Editor) typePrompt() string {
	if e.typeNote != "" {
		return e.typeNote
	}
	switch e.typeKind {
	case typeTitle:
		return "Title"
	case typeSize:
		if e.freeform {
			return "Size in pixels, a multiple of 16"
		}
		return "Size in tiles"
	default:
		return "Briefing"
	}
}

func clipLine(s string, n int) string {
	r := []rune(s)
	if n < 1 || len(r) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	return string(r[:n-1]) + "…"
}

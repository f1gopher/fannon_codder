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
	titleMax = 80
	briefMax = 400
)

func (e *Editor) typing() bool {
	return e.typeKind != typeNone
}

func (e *Editor) startType(kind int) {
	if e.phase == nil || e.err != nil {
		return
	}
	e.typeKind = kind
	if kind == typeTitle {
		e.draft = e.phase.Title
	} else {
		e.draft = e.phase.Briefing
	}
}

func (e *Editor) cancelType() {
	e.typeKind = typeNone
	e.draft = ""
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
		}
	case typeBrief:
		if text != e.phase.Briefing {
			e.phase.Briefing = text
			e.dirty = true
		}
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
	if e.typeKind == typeTitle {
		limit = titleMax
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
	if e.typeKind == typeTitle {
		return "Title"
	}
	return "Briefing"
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

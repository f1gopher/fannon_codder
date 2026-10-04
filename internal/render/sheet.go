package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"math"
	"path"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"fannon-codder/assets/art"
)

// ArtScale is source pixels per world pixel. Sheets are drawn at S/ArtScale.
const ArtScale = 8

// Directions stored in a manifest, clockwise from east. Y grows down the
// screen, so south is π/2. W, SW, and NW are usually mirrors and are then
// absent from the PNG.
var dirOrder = []string{"E", "SE", "S", "SW", "W", "NW", "N", "NE"}

// Sheet is one animation. Cel images are copies of the frames, so a linear
// filter cannot bleed into the next frame.
type Sheet struct {
	FrameW, FrameH int
	AnchorX        int
	AnchorY        int
	FPS            float64
	Loop           bool
	Frames         int
	place          map[string]placement
	cels           [][]*ebiten.Image
}

type placement struct {
	row    int
	mirror bool
}

type manifest struct {
	FrameW  int               `json:"frameW"`
	FrameH  int               `json:"frameH"`
	AnchorX int               `json:"anchorX"`
	AnchorY int               `json:"anchorY"`
	FPS     float64           `json:"fps"`
	Loop    bool              `json:"loop"`
	Rows    []string          `json:"rows"`
	Mirrors map[string]string `json:"mirrors"`
}

// Library is the sheets found under an fs.FS. The key is the path without
// the extension, slash-separated: snake/idle, ground/grass, tree/sway.
type Library struct {
	sheets map[string]*Sheet
}

// Men use snake, eagle, panther, grunt, grenadier, rocketeer, or civilian,
// plus idle, walk, shoot, throw, death, corpse, swim, or sink.
// Ground uses ground/grass, ground/snow, ground/water-shallow,
// ground/water-deep, ground/quicksand, ground/ice, ground/bridge,
// ground/cliff, ground/ramp, ground/mine.
// Props use tree/sway, hut/door, hut/plain, crate/grenade, crate/rocket,
// skidoo/idle, skidoo/move, prop/scrub, prop/snowman, prop/igloo.
// Birds use bird/flap (east stored, west mirrored) and cross above the field.
// Fire uses fx/tracer, fx/grenade, fx/rocket, and fx/blast. fx/flash is unused.
// The cursor and the status strip use ui/pointer, ui/crosshair, ui/board,
// ui/grenade, ui/rocket, ui/foot, ui/vehicle, ui/map, ui/mark-snake,
// ui/mark-eagle, ui/mark-panther, and ui/rank-* (one flash per player rank).
// Those draw in screen space at S/8.
// The menus use menu/title, menu/briefing, and menu/hill as full-frame
// paintings (2560×2048, anchor at the top-left), plus menu/grave,
// menu/save, and menu/load. The hill queue is the Snake idle sheet.
//
// Ground frames pin their top-left to the cell and ignore the anchor.
// Every other anchor sits on the sim point: a trooper or crate or skidoo
// position, a hut's bottom centre, or the bottom centre of a tree's cell.
// Scenery loops add tx*3+ty*5 to the frame index.

func Load(fsys fs.FS) (*Library, error) {
	lib := &Library{sheets: map[string]*Sheet{}}
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".json") {
			return err
		}
		key := strings.TrimSuffix(p, ".json")
		sh, err := loadSheet(fsys, key)
		if err != nil {
			return err
		}
		lib.sheets[key] = sh
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lib, nil
}

func (l *Library) Get(key string) *Sheet {
	if l == nil {
		return nil
	}
	return l.sheets[path.Clean(key)]
}

func loadSheet(fsys fs.FS, key string) (*Sheet, error) {
	raw, err := fs.ReadFile(fsys, key+".json")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	if m.FrameW < 1 || m.FrameH < 1 {
		return nil, fmt.Errorf("%s: frameW and frameH must be positive", key)
	}
	if len(m.Rows) == 0 {
		return nil, fmt.Errorf("%s: rows is empty", key)
	}
	place := map[string]placement{}
	for i, name := range m.Rows {
		if !validDir(name) {
			return nil, fmt.Errorf("%s: unknown row %q", key, name)
		}
		if _, dup := place[name]; dup {
			return nil, fmt.Errorf("%s: row %s is listed twice", key, name)
		}
		place[name] = placement{row: i}
	}
	for name, src := range m.Mirrors {
		if !validDir(name) {
			return nil, fmt.Errorf("%s: unknown mirror %q", key, name)
		}
		if _, stored := place[name]; stored {
			return nil, fmt.Errorf("%s: %s is stored and also a mirror", key, name)
		}
		srcPlace, ok := place[src]
		if !ok || srcPlace.mirror {
			return nil, fmt.Errorf("%s: mirror %s → %s, and %s is not a stored row", key, name, src, src)
		}
		place[name] = placement{row: srcPlace.row, mirror: true}
	}
	pngBytes, err := fs.ReadFile(fsys, key+".png")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	decoded, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	b := decoded.Bounds()
	if b.Dx()%m.FrameW != 0 || b.Dy() != len(m.Rows)*m.FrameH {
		return nil, fmt.Errorf("%s: png is %dx%d, want a multiple of %d wide and %d tall",
			key, b.Dx(), b.Dy(), m.FrameW, len(m.Rows)*m.FrameH)
	}
	frames := b.Dx() / m.FrameW
	cels := make([][]*ebiten.Image, len(m.Rows))
	for row := range m.Rows {
		cels[row] = make([]*ebiten.Image, frames)
		for frame := 0; frame < frames; frame++ {
			x0 := b.Min.X + frame*m.FrameW
			y0 := b.Min.Y + row*m.FrameH
			cels[row][frame] = crop(decoded, image.Rect(x0, y0, x0+m.FrameW, y0+m.FrameH))
		}
	}
	return &Sheet{
		FrameW: m.FrameW, FrameH: m.FrameH,
		AnchorX: m.AnchorX, AnchorY: m.AnchorY,
		FPS: m.FPS, Loop: m.Loop, Frames: frames,
		place: place, cels: cels,
	}, nil
}

func crop(src image.Image, r image.Rectangle) *ebiten.Image {
	dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			dst.Set(x, y, src.At(r.Min.X+x, r.Min.Y+y))
		}
	}
	return ebiten.NewImageFromImage(dst)
}

func validDir(name string) bool {
	for _, d := range dirOrder {
		if d == name {
			return true
		}
	}
	return false
}

// Direction is the sheet row for a sim facing. 0 is E, π/2 is S, π is W,
// and −π/2 is N.
func Direction(ang float64) string {
	n := int(math.Round(ang / (math.Pi / 4)))
	n %= len(dirOrder)
	if n < 0 {
		n += len(dirOrder)
	}
	return dirOrder[n]
}

// image is the frame for a direction. ok is false when the sheet does not
// carry that facing. mirror means the caller flips the cel.
func (s *Sheet) image(dir string, frame int) (img *ebiten.Image, mirror bool, ok bool) {
	if s == nil || s.Frames == 0 {
		return nil, false, false
	}
	p, found := s.place[dir]
	if !found {
		return nil, false, false
	}
	if frame < 0 {
		frame = 0
	}
	if frame >= s.Frames {
		frame = s.Frames - 1
	}
	return s.cels[p.row][frame], p.mirror, true
}

// still is the first stored row, used by scenery that has no facing.
func (s *Sheet) still(frame int) (*ebiten.Image, bool) {
	return s.Row(0, frame)
}

// RowCount is the number of stored rows. tree/sway keeps one silhouette per row.
func (s *Sheet) RowCount() int {
	if s == nil {
		return 0
	}
	return len(s.cels)
}

// Row is one stored row. The index wraps, so a forest can pick a silhouette from the tile.
func (s *Sheet) Row(row, frame int) (*ebiten.Image, bool) {
	if s == nil || len(s.cels) == 0 || s.Frames == 0 {
		return nil, false
	}
	row = mod(row, len(s.cels))
	if frame < 0 {
		frame = 0
	}
	if frame >= s.Frames {
		frame = s.Frames - 1
	}
	return s.cels[row][frame], true
}

// FrameAt is the column for a clock. A loop wraps. A one-shot holds the
// last frame.
func FrameAt(t, fps float64, n int, loop bool) int {
	if n <= 1 {
		return 0
	}
	if fps <= 0 {
		return 0
	}
	i := int(math.Floor(t * fps))
	if loop {
		return mod(i, n)
	}
	if i >= n {
		return n - 1
	}
	if i < 0 {
		return 0
	}
	return i
}

// FrameScrub maps p in 0..1 across the frames once. p of 1 is the last frame.
func FrameScrub(p float64, n int) int {
	if n <= 1 {
		return 0
	}
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	i := int(p * float64(n))
	if i >= n {
		i = n - 1
	}
	return i
}

// SceneryFrame is the column for a looping tile. Neighbours differ by
// tx*3+ty*5, so a forest does not sway in step.
func SceneryFrame(tx, ty, n int, t, fps float64) int {
	if n <= 1 {
		return 0
	}
	base := 0
	if fps > 0 {
		base = int(math.Floor(t * fps))
	}
	return mod(base+tx*3+ty*5, n)
}

func mod(a, n int) int {
	if n <= 0 {
		return 0
	}
	m := a % n
	if m < 0 {
		m += n
	}
	return m
}

var (
	sheetsOnce     sync.Once
	embedded       *Library
	sheetsOverride *Library
)

func activeSheets() *Library {
	if sheetsOverride != nil {
		return sheetsOverride
	}
	sheetsOnce.Do(func() {
		lib, err := Load(art.Files)
		if err != nil {
			panic(err)
		}
		embedded = lib
	})
	return embedded
}

// SetSheets replaces the embedded library. nil restores it. Tests use this.
func SetSheets(l *Library) {
	sheetsOverride = l
}

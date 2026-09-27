package app

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestControlsDoc(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "controls.md")
	want := controlsMarkdown()
	got, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if string(got) != want {
		if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("rewrote %s from internal/app/controls.go; review the new page and commit it", path)
	}

	used := usedInputTokens(t, root)
	listed := map[string]bool{}
	for _, row := range controls() {
		for _, tok := range row.tokens {
			listed[tok] = true
		}
	}
	var missing []string
	for tok := range used {
		if !listed[tok] {
			missing = append(missing, tok)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("game reads %s but controls() does not name them; add a binding, then rerun this test so controls.md updates", strings.Join(missing, ", "))
	}
	var stale []string
	for tok := range listed {
		if !used[tok] {
			stale = append(stale, tok)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Fatalf("controls() names %s but the game no longer reads them; drop the token so controls.md stays true", strings.Join(stale, ", "))
	}
}

var inputToken = regexp.MustCompile(`ebiten\.(Key[A-Za-z0-9]+|MouseButton(?:Left|Right|Middle|0|1|2))`)

func usedInputTokens(t *testing.T, root string) map[string]bool {
	t.Helper()
	found := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if base == "vendor" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range inputToken.FindAllSubmatch(body, -1) {
			found[string(m[1])] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

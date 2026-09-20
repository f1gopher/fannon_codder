package campaign

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadNames reads data/names.txt. Falls back to a short built-in list.
func LoadNames() []string {
	for _, dir := range []string{
		filepath.Join("data"),
		filepath.Join("..", "data"),
		filepath.Join("..", "..", "data"),
	} {
		path := filepath.Join(dir, "names.txt")
		if names, err := readNameFile(path); err == nil && len(names) > 0 {
			return names
		}
	}
	return fallbackNames
}

func readNameFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var names []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		n := strings.TrimSpace(sc.Text())
		if n == "" || strings.HasPrefix(n, "#") {
			continue
		}
		names = append(names, n)
	}
	return names, sc.Err()
}

var fallbackNames = []string{"Jools", "Jops", "Stoo"}

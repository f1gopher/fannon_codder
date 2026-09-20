package campaign

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SaveGame is the one Boot Hill slot.
type SaveGame struct {
	PhaseIndex         int       `json:"phaseIndex"`
	MissionsCompleted  int       `json:"missionsCompleted"`
	Graves             int       `json:"graves"`
	GameOver           bool      `json:"gameOver"`
	AwaitingStub       bool      `json:"awaitingStub"`
	Recruits           []Soldier `json:"recruits"`
	NextName           int       `json:"nextName"`
}

func DefaultSavePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "fannon-codder", "save.json"), nil
}

func Save(path string, s SaveGame) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func Load(path string) (SaveGame, error) {
	var s SaveGame
	b, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(b, &s)
	return s, err
}

func SaveExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

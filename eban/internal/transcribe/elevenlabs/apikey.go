package elevenlabs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadAPIKey resolves the API key from $ELEVENLABS_API_KEY or the
// ~/.config/eban/apikey file.
func LoadAPIKey() (string, error) {
	if key := strings.TrimSpace(os.Getenv("ELEVENLABS_API_KEY")); key != "" {
		return key, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("resolve home directory")
	}
	return LoadAPIKeyFile(filepath.Join(home, ".config", "eban", "apikey"))
}

// LoadAPIKeyFile reads and trims an API key from a single file.
func LoadAPIKeyFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("api key not found: set ELEVENLABS_API_KEY or create %s (chmod 600)", path)
	}
	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", fmt.Errorf("api key file %s is empty", path)
	}
	return key, nil
}

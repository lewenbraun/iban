package elevenlabs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAPIKeyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "apikey")

	if _, err := LoadAPIKeyFile(filepath.Join(dir, "missing")); err == nil {
		t.Error("expected error for missing key file")
	}
	if err := os.WriteFile(path, []byte("  sk-test-123 \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	key, err := LoadAPIKeyFile(path)
	if err != nil {
		t.Fatalf("LoadAPIKeyFile: %v", err)
	}
	if key != "sk-test-123" {
		t.Errorf("key = %q, want %q", key, "sk-test-123")
	}
	if err := os.WriteFile(path, []byte("   \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAPIKeyFile(path); err == nil {
		t.Error("expected error for empty key file")
	}
}

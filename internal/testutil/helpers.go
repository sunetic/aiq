package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aiq/aiq/internal/session"
)

// CreateTempDir creates a temporary directory for testing
// Returns the directory path and a cleanup function
func CreateTempDir(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

// CreateTempSession creates a temporary session file for testing
// Returns the session and the file path
func CreateTempSession(t *testing.T, dataSource string, dbType string) (*session.Session, string) {
	sess := session.NewSession(dataSource, dbType)
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "test_session.json")
	return sess, sessionPath
}

// LoadTestFixture loads a test fixture file by name
// Returns the file content as bytes
func LoadTestFixture(name string) ([]byte, error) {
	// Fixtures are stored in internal/testutil/fixtures/
	fixturePath := filepath.Join("internal", "testutil", "fixtures", name)
	return os.ReadFile(fixturePath)
}

// MustLoadTestFixture loads a test fixture and panics on error
// Useful in tests where fixture loading failure should fail the test
func MustLoadTestFixture(t *testing.T, name string) []byte {
	data, err := LoadTestFixture(name)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
	return data
}

// WriteTestFixture writes data to a test fixture file
// Useful for creating fixtures during test development
func WriteTestFixture(name string, data []byte) error {
	fixturePath := filepath.Join("internal", "testutil", "fixtures", name)
	dir := filepath.Dir(fixturePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fixturePath, data, 0644)
}

// LoadJSONFixture loads a JSON fixture and unmarshals it into v
func LoadJSONFixture(name string, v interface{}) error {
	data, err := LoadTestFixture(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// MustLoadJSONFixture loads a JSON fixture and unmarshals it, panicking on error
func MustLoadJSONFixture(t *testing.T, name string, v interface{}) {
	data := MustLoadTestFixture(t, name)
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("Failed to unmarshal fixture %s: %v", name, err)
	}
}

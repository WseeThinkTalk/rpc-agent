package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "test_env")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	envFile := filepath.Join(tempDir, ".env")
	err = os.WriteFile(envFile, []byte("TEST_DOTENV_VAR=dotenv_value\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write temp .env: %v", err)
	}

	subDir := filepath.Join(tempDir, "sub_service")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	err = os.Chdir(subDir)
	if err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}
	defer os.Chdir(cwd)

	os.Unsetenv("TEST_DOTENV_VAR")

	LoadEnv()

	val := os.Getenv("TEST_DOTENV_VAR")
	if val != "dotenv_value" {
		t.Errorf("expected TEST_DOTENV_VAR to be 'dotenv_value', got '%s'", val)
	}
}

package env

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	dir, err := os.Getwd()
	if err != nil {
		_ = godotenv.Load()
		return
	}

	startDir := dir
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				fmt.Printf("[env] Failed to load .env from %s: %v\n", envPath, err)
			} else {
				fmt.Printf("[env] .env loaded from %s (ALIPAY_PRIVATE_KEY=%d chars)\n", envPath, len(os.Getenv("ALIPAY_PRIVATE_KEY")))
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	fmt.Printf("[env] .env file not found, searched from %s up to %s\n", startDir, dir)
	_ = godotenv.Load()
}

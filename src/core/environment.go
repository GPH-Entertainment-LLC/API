package core

import (
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadLocalEnvironment(currDir string) {
	// loading environment
	godotenv.Load(filepath.Join(currDir, "config", "local.env"))
}

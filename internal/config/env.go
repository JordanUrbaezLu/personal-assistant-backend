package config

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/joho/godotenv"
)

// Load loads the environment variables from the given env file in the module root.
func Load(envFile string) {
	err := godotenv.Load(dir(envFile))
	if err != nil {
		panic(fmt.Errorf("Error loading %s: %w", envFile, err))
	}
}

// dir searches upward from the current directory until it finds go.mod.
// Then it returns the full path to the env file in that directory.
func dir(envFile string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found module root
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found"))
		}
		currentDir = parent
	}

	return filepath.Join(currentDir, envFile)
}

package config

import (
	"errors"
	"os"
	"path/filepath"
)

const configFileName = "runefact.toml"

// ErrProjectNotFound is returned when no runefact.toml can be located.
var ErrProjectNotFound = errors.New("runefact.toml not found (run 'runefact init' to create a project)")

// FindProjectRoot walks up from the current working directory looking for runefact.toml.
func FindProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return FindProjectRootFrom(wd)
}

// FindProjectRootFrom walks up from startDir looking for runefact.toml.
func FindProjectRootFrom(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}

	for {
		if IsProjectRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrProjectNotFound
		}
		dir = parent
	}
}

// IsProjectRoot reports whether dir contains a runefact.toml file.
func IsProjectRoot(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, configFileName))
	return err == nil && !info.IsDir()
}

// GetConfigPath returns the full path to runefact.toml within the given project root.
func GetConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, configFileName)
}

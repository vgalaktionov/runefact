package build

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

const cacheFileName = ".runefact-cache.json"

// BuildCache tracks content hashes to skip unchanged files.
type BuildCache struct {
	Hashes map[string]string `json:"hashes"`
	path   string
}

// LoadCache reads the build cache from the project root.
func LoadCache(projectRoot string) *BuildCache {
	c := &BuildCache{
		Hashes: make(map[string]string),
		path:   filepath.Join(projectRoot, cacheFileName),
	}

	data, err := os.ReadFile(c.path)
	if err != nil {
		return c
	}
	json.Unmarshal(data, c)
	if c.Hashes == nil {
		c.Hashes = make(map[string]string)
	}
	return c
}

// NeedsRebuild returns true if the file has changed since last build.
func (c *BuildCache) NeedsRebuild(file string) bool {
	content, err := os.ReadFile(file)
	if err != nil {
		return true
	}

	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	if c.Hashes[file] == hashStr {
		return false
	}
	c.Hashes[file] = hashStr
	return true
}

// Save persists the cache to disk.
func (c *BuildCache) Save() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

// InvalidateDependents marks all files with the given palette dependency as needing rebuild.
func (c *BuildCache) InvalidateDependents(paletteFile string) {
	// Remove palette hash to trigger rebuild.
	delete(c.Hashes, paletteFile)
	// Note: dependents are invalidated indirectly — if the palette's hash changed,
	// we can't know which sprites depend on it without parsing. For correctness,
	// palette changes always trigger a full rebuild of sprites/maps.
}

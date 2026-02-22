package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var runeExtensions = map[string]bool{
	".palette": true,
	".sprite":  true,
	".map":     true,
	".inst":    true,
	".sfx":     true,
	".track":   true,
}

// IsRuneFile reports whether a file path has a recognized rune extension.
func IsRuneFile(path string) bool {
	return runeExtensions[filepath.Ext(path)]
}

// RebuildFunc is called with a list of changed file paths.
type RebuildFunc func(changed []string) error

// Watcher watches rune files for changes and triggers rebuilds.
type Watcher struct {
	fsw       *fsnotify.Watcher
	debounce  time.Duration
	onRebuild RebuildFunc
	deps      *DependencyTracker
	done      chan struct{}
}

// New creates a new Watcher.
func New(debounce time.Duration, onRebuild RebuildFunc) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		fsw:       fsw,
		debounce:  debounce,
		onRebuild: onRebuild,
		deps:      NewDependencyTracker(),
		done:      make(chan struct{}),
	}, nil
}

// WatchDir recursively watches a directory for rune file changes.
func (w *Watcher) WatchDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return w.fsw.Add(path)
		}
		return nil
	})
}

// Start begins watching for file changes. Blocks until Stop is called.
func (w *Watcher) Start() {
	var mu sync.Mutex
	pending := map[string]struct{}{}
	var timer *time.Timer

	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if !IsRuneFile(event.Name) {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			mu.Lock()
			pending[event.Name] = struct{}{}

			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(w.debounce, func() {
				mu.Lock()
				files := make([]string, 0, len(pending))
				for f := range pending {
					files = append(files, f)
				}
				pending = map[string]struct{}{}
				mu.Unlock()

				// Expand dependencies.
				expanded := w.deps.ExpandDependencies(files)

				log.Printf("Rebuilding: %v", expanded)
				if err := w.onRebuild(expanded); err != nil {
					log.Printf("Build error: %v", err)
				}
			})
			mu.Unlock()

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)

		case <-w.done:
			return
		}
	}
}

// Stop signals the watcher to stop.
func (w *Watcher) Stop() error {
	close(w.done)
	return w.fsw.Close()
}

// DependencyTracker tracks cross-file dependencies for incremental rebuilds.
type DependencyTracker struct {
	// paletteDeps maps palette name -> sprite files that use it
	paletteDeps map[string][]string
	// spriteDeps maps sprite file -> map files that use it
	spriteDeps map[string][]string
	// instDeps maps instrument name -> sfx/track files that use it
	instDeps map[string][]string
}

// NewDependencyTracker creates a new empty tracker.
func NewDependencyTracker() *DependencyTracker {
	return &DependencyTracker{
		paletteDeps: make(map[string][]string),
		spriteDeps:  make(map[string][]string),
		instDeps:    make(map[string][]string),
	}
}

// ExpandDependencies takes changed files and returns all files that need rebuilding.
func (dt *DependencyTracker) ExpandDependencies(changed []string) []string {
	seen := map[string]bool{}
	var result []string

	var add func(path string)
	add = func(path string) {
		if seen[path] {
			return
		}
		seen[path] = true
		result = append(result, path)

		ext := filepath.Ext(path)
		base := strings.TrimSuffix(filepath.Base(path), ext)

		switch ext {
		case ".palette":
			for _, dep := range dt.paletteDeps[base] {
				add(dep)
			}
		case ".sprite":
			for _, dep := range dt.spriteDeps[base] {
				add(dep)
			}
		case ".inst":
			for _, dep := range dt.instDeps[base] {
				add(dep)
			}
		}
	}

	for _, f := range changed {
		add(f)
	}
	return result
}

// RegisterPaletteDep records that a sprite file depends on a palette.
func (dt *DependencyTracker) RegisterPaletteDep(paletteName, spriteFile string) {
	dt.paletteDeps[paletteName] = append(dt.paletteDeps[paletteName], spriteFile)
}

// RegisterSpriteDep records that a map file depends on a sprite file.
func (dt *DependencyTracker) RegisterSpriteDep(spriteName, mapFile string) {
	dt.spriteDeps[spriteName] = append(dt.spriteDeps[spriteName], mapFile)
}

// RegisterInstrumentDep records that a track/sfx file depends on an instrument.
func (dt *DependencyTracker) RegisterInstrumentDep(instName, audioFile string) {
	dt.instDeps[instName] = append(dt.instDeps[instName], audioFile)
}

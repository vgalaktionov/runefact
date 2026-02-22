package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestIsRuneFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"player.sprite", true},
		{"default.palette", true},
		{"world.map", true},
		{"piano.inst", true},
		{"laser.sfx", true},
		{"bgm.track", true},
		{"readme.txt", false},
		{"image.png", false},
		{"", false},
		{"noext", false},
	}
	for _, tt := range tests {
		if got := IsRuneFile(tt.path); got != tt.want {
			t.Errorf("IsRuneFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestDependencyTracker_ExpandDependencies(t *testing.T) {
	dt := NewDependencyTracker()

	// palette "default" → sprite "player.sprite"
	dt.RegisterPaletteDep("default", "/assets/player.sprite")
	// sprite "player" → map "world.map"
	dt.RegisterSpriteDep("player", "/assets/world.map")
	// instrument "piano" → track "bgm.track"
	dt.RegisterInstrumentDep("piano", "/assets/bgm.track")

	t.Run("palette change cascades to sprite and map", func(t *testing.T) {
		result := dt.ExpandDependencies([]string{"/assets/default.palette"})
		// palette → sprite → map (transitive)
		if len(result) != 3 {
			t.Fatalf("expected 3 files, got %d: %v", len(result), result)
		}
		has := map[string]bool{}
		for _, f := range result {
			has[f] = true
		}
		if !has["/assets/default.palette"] {
			t.Error("missing default.palette")
		}
		if !has["/assets/player.sprite"] {
			t.Error("missing player.sprite")
		}
		if !has["/assets/world.map"] {
			t.Error("missing world.map")
		}
	})

	t.Run("instrument change cascades to track", func(t *testing.T) {
		result := dt.ExpandDependencies([]string{"/assets/piano.inst"})
		if len(result) != 2 {
			t.Fatalf("expected 2 files, got %d: %v", len(result), result)
		}
		has := map[string]bool{}
		for _, f := range result {
			has[f] = true
		}
		if !has["/assets/piano.inst"] {
			t.Error("missing piano.inst")
		}
		if !has["/assets/bgm.track"] {
			t.Error("missing bgm.track")
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		result := dt.ExpandDependencies([]string{"/assets/default.palette", "/assets/player.sprite"})
		seen := map[string]int{}
		for _, f := range result {
			seen[f]++
		}
		for f, count := range seen {
			if count > 1 {
				t.Errorf("duplicate: %s appeared %d times", f, count)
			}
		}
	})

	t.Run("unrelated file not expanded", func(t *testing.T) {
		result := dt.ExpandDependencies([]string{"/assets/laser.sfx"})
		if len(result) != 1 {
			t.Fatalf("expected 1 file, got %d: %v", len(result), result)
		}
		if result[0] != "/assets/laser.sfx" {
			t.Errorf("unexpected file: %s", result[0])
		}
	})
}

func TestWatcher_StartStop(t *testing.T) {
	dir := t.TempDir()

	// Create a rune file to trigger a change.
	testFile := filepath.Join(dir, "test.sprite")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var rebuilt [][]string

	w, err := New(50*time.Millisecond, func(changed []string) error {
		mu.Lock()
		rebuilt = append(rebuilt, changed)
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := w.WatchDir(dir); err != nil {
		t.Fatal(err)
	}

	go w.Start()

	// Give the watcher time to start.
	time.Sleep(100 * time.Millisecond)

	// Modify the file.
	if err := os.WriteFile(testFile, []byte("updated"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for debounce + processing.
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	count := len(rebuilt)
	mu.Unlock()

	if count == 0 {
		t.Error("expected at least one rebuild callback, got none")
	}

	if err := w.Stop(); err != nil {
		t.Errorf("Stop() error: %v", err)
	}
}

func TestWatcher_IgnoresNonRuneFiles(t *testing.T) {
	dir := t.TempDir()

	txtFile := filepath.Join(dir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("initial"), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	rebuildCount := 0

	w, err := New(50*time.Millisecond, func(changed []string) error {
		mu.Lock()
		rebuildCount++
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := w.WatchDir(dir); err != nil {
		t.Fatal(err)
	}

	go w.Start()
	time.Sleep(100 * time.Millisecond)

	// Modify a non-rune file.
	if err := os.WriteFile(txtFile, []byte("updated"), 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	count := rebuildCount
	mu.Unlock()

	if count != 0 {
		t.Errorf("expected no rebuilds for .txt file, got %d", count)
	}

	if err := w.Stop(); err != nil {
		t.Errorf("Stop() error: %v", err)
	}
}

func TestWatcher_Debounce(t *testing.T) {
	dir := t.TempDir()

	spriteFile := filepath.Join(dir, "test.sprite")
	if err := os.WriteFile(spriteFile, []byte("v0"), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	rebuildCount := 0

	w, err := New(200*time.Millisecond, func(changed []string) error {
		mu.Lock()
		rebuildCount++
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := w.WatchDir(dir); err != nil {
		t.Fatal(err)
	}

	go w.Start()
	time.Sleep(100 * time.Millisecond)

	// Rapid-fire writes.
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(spriteFile, []byte("v"+string(rune('1'+i))), 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(30 * time.Millisecond)
	}

	// Wait for debounce to settle.
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := rebuildCount
	mu.Unlock()

	// With 200ms debounce and 30ms between writes, all 5 writes should
	// collapse into 1-2 rebuilds at most.
	if count > 2 {
		t.Errorf("expected debounced rebuilds (<=2), got %d", count)
	}
	if count == 0 {
		t.Error("expected at least one rebuild")
	}

	if err := w.Stop(); err != nil {
		t.Errorf("Stop() error: %v", err)
	}
}

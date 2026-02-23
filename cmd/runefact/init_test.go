package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestScaffoldFiles(t *testing.T) {
	files := scaffoldFiles("test-game")
	if len(files) == 0 {
		t.Fatal("scaffoldFiles returned no files")
	}

	// Check runefact.toml is first and contains project name.
	if files[0].path != "runefact.toml" {
		t.Errorf("first file = %q, want runefact.toml", files[0].path)
	}
	if got := files[0].content; len(got) == 0 {
		t.Error("runefact.toml is empty")
	}

	// Verify all expected paths are present.
	want := map[string]bool{
		"runefact.toml":                   false,
		"assets/palettes/default.palette": false,
		"assets/sprites/player.sprite":    false,
		"assets/sprites/tiles.sprite":     false,
		"assets/maps/level1.map":          false,
		"assets/instruments/lead.inst":    false,
		"assets/instruments/bass.inst":    false,
		"assets/sfx/jump.sfx":             false,
		"assets/sfx/coin.sfx":             false,
		"assets/tracks/demo.track":        false,
	}
	for _, f := range files {
		if _, ok := want[f.path]; !ok {
			t.Errorf("unexpected file: %s", f.path)
		}
		want[f.path] = true
	}
	for path, found := range want {
		if !found {
			t.Errorf("missing file: %s", path)
		}
	}
}

func TestRunInit_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	flagInitName = "test-project"
	flagInitForce = false
	flagQuiet = true

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// Check files exist.
	for _, name := range []string{
		"runefact.toml",
		"assets/palettes/default.palette",
		"assets/sprites/player.sprite",
		"assets/sprites/tiles.sprite",
		"assets/maps/level1.map",
		"assets/instruments/lead.inst",
		"assets/instruments/bass.inst",
		"assets/sfx/jump.sfx",
		"assets/sfx/coin.sfx",
		"assets/tracks/demo.track",
		".mcp.json",
		".claude/settings.local.json",
	} {
		// MCP config files are created by setupMCPConfig, not scaffoldFiles.
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
		}
	}

	// Check dirs exist.
	for _, d := range []string{
		"assets/palettes",
		"assets/sprites",
		"assets/maps",
		"assets/instruments",
		"assets/sfx",
		"assets/tracks",
		".claude",
	} {
		info, err := os.Stat(filepath.Join(dir, d))
		if err != nil {
			t.Errorf("expected dir %s to exist: %v", d, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", d)
		}
	}
}

func TestRunInit_ErrorsWithoutForce(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Create existing runefact.toml.
	if err := os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte("[project]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	flagInitForce = false
	err := runInit(nil, nil)
	if err == nil {
		t.Fatal("expected error when runefact.toml exists without --force")
	}
}

func TestRunInit_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Create existing runefact.toml.
	if err := os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte("[project]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	flagInitName = "overwrite-test"
	flagInitForce = true
	flagQuiet = true

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit with --force: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("runefact.toml is empty after force overwrite")
	}
}

func TestSetupMCPConfig_MergesExisting(t *testing.T) {
	dir := t.TempDir()

	// Create pre-existing .mcp.json with another server.
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	existingMCP := `{"mcpServers":{"other-server":{"type":"stdio","command":"other"}}}`
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(existingMCP), 0644); err != nil {
		t.Fatal(err)
	}

	// Create pre-existing claude settings with other permissions.
	existingClaude := `{"permissions":{"allow":["Bash(*)"]},"enabledMcpjsonServers":["other-server"]}`
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.local.json"), []byte(existingClaude), 0644); err != nil {
		t.Fatal(err)
	}

	written := setupMCPConfig(dir)
	if len(written) != 2 {
		t.Fatalf("expected 2 files written, got %d: %v", len(written), written)
	}

	// Verify .mcp.json has both servers.
	data, _ := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	var mcpDoc map[string]any
	if err := json.Unmarshal(data, &mcpDoc); err != nil {
		t.Fatalf("parsing .mcp.json: %v", err)
	}
	servers, _ := mcpDoc["mcpServers"].(map[string]any)
	if _, ok := servers["other-server"]; !ok {
		t.Error("existing other-server was clobbered")
	}
	if _, ok := servers["runefact"]; !ok {
		t.Error("runefact server was not added")
	}

	// Verify claude settings merged.
	data, _ = os.ReadFile(filepath.Join(dir, ".claude", "settings.local.json"))
	var claudeDoc map[string]any
	if err := json.Unmarshal(data, &claudeDoc); err != nil {
		t.Fatalf("parsing claude settings: %v", err)
	}

	enabledRaw, _ := claudeDoc["enabledMcpjsonServers"].([]any)
	var enabled []string
	for _, v := range enabledRaw {
		if s, ok := v.(string); ok {
			enabled = append(enabled, s)
		}
	}
	if !slices.Contains(enabled, "other-server") {
		t.Error("existing other-server was removed from enabledMcpjsonServers")
	}
	if !slices.Contains(enabled, "runefact") {
		t.Error("runefact was not added to enabledMcpjsonServers")
	}

	perms, _ := claudeDoc["permissions"].(map[string]any)
	allowRaw, _ := perms["allow"].([]any)
	var allow []string
	for _, v := range allowRaw {
		if s, ok := v.(string); ok {
			allow = append(allow, s)
		}
	}
	if !slices.Contains(allow, "Bash(*)") {
		t.Error("existing Bash(*) permission was removed")
	}
	if !slices.Contains(allow, "mcp__runefact__*") {
		t.Error("mcp__runefact__* permission was not added")
	}
}

func TestSetupMCPConfig_SkipsExistingRunefact(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}

	// Already has runefact configured.
	existingMCP := `{"mcpServers":{"runefact":{"type":"stdio","command":"custom-path"}}}`
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(existingMCP), 0644); err != nil {
		t.Fatal(err)
	}

	written := setupMCPConfig(dir)

	// Should have written settings.local.json but NOT .mcp.json.
	for _, f := range written {
		if f == ".mcp.json" {
			t.Error(".mcp.json should not be rewritten when runefact already exists")
		}
	}

	// Verify the command was NOT overwritten.
	data, _ := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	var mcpDoc map[string]any
	_ = json.Unmarshal(data, &mcpDoc)
	servers, _ := mcpDoc["mcpServers"].(map[string]any)
	rf, _ := servers["runefact"].(map[string]any)
	if cmd, _ := rf["command"].(string); cmd != "custom-path" {
		t.Errorf("runefact command was clobbered: got %q, want %q", cmd, "custom-path")
	}
}

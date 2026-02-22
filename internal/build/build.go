package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vgalaktionov/runefact/internal/audio"
	"github.com/vgalaktionov/runefact/internal/config"
	"github.com/vgalaktionov/runefact/internal/instrument"
	"github.com/vgalaktionov/runefact/internal/manifest"
	"github.com/vgalaktionov/runefact/internal/palette"
	"github.com/vgalaktionov/runefact/internal/sfx"
	"github.com/vgalaktionov/runefact/internal/sprite"
	"github.com/vgalaktionov/runefact/internal/tilemap"
	"github.com/vgalaktionov/runefact/internal/track"
)

// Scope defines which asset types to build.
type Scope string

const (
	ScopeAll     Scope = "all"
	ScopeSprites Scope = "sprites"
	ScopeMaps    Scope = "maps"
	ScopeAudio   Scope = "audio"
)

// Options controls what gets built.
type Options struct {
	Scope     Scope
	Files     []string // specific files to build (empty = all)
	OutputDir string
}

// Result contains the output of a build.
type Result struct {
	Artifacts    []string
	Errors       []error
	Warnings     []string
	ManifestPath string
}

// Build compiles rune files into game-ready artifacts.
func Build(opts Options, cfg *config.ProjectConfig, projectRoot string) *Result {
	result := &Result{}

	if opts.OutputDir == "" {
		opts.OutputDir = filepath.Join(projectRoot, cfg.Project.Output)
	}
	if opts.Scope == "" {
		opts.Scope = ScopeAll
	}

	assetsDir := filepath.Join(projectRoot, "assets")
	md := &manifest.ManifestData{Package: cfg.Project.Package}

	// Phase 1: Parse all palettes.
	palettes := map[string]*palette.Palette{}
	if paletteDir := filepath.Join(assetsDir, "palettes"); dirExists(paletteDir) {
		files := discoverFiles(paletteDir, ".palette", opts.Files)
		for _, f := range files {
			p, err := palette.LoadPalette(f)
			if err != nil {
				result.Errors = append(result.Errors, err)
				continue
			}
			palettes[p.Name] = p
		}
	}

	// Phase 2: Parse and render sprites.
	if opts.Scope == ScopeAll || opts.Scope == ScopeSprites {
		if spriteDir := filepath.Join(assetsDir, "sprites"); dirExists(spriteDir) {
			files := discoverFiles(spriteDir, ".sprite", opts.Files)
			for _, f := range files {
				sf, err := sprite.LoadSpriteFile(f)
				if err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				pal, ok := palettes[sf.PaletteRef]
				if !ok && sf.PaletteRef != "" {
					result.Errors = append(result.Errors, fmt.Errorf("%s: palette %q not found", f, sf.PaletteRef))
					continue
				}
				if pal == nil {
					pal = &palette.Palette{Colors: map[string]palette.Color{}}
				}

				resolved, err := sf.Resolve(pal)
				if err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				img, meta, err := sprite.RenderSpriteSheet(resolved)
				if err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				baseName := strings.TrimSuffix(filepath.Base(f), ".sprite")
				relPath := filepath.Join("sprites", baseName+".png")
				outPath := filepath.Join(opts.OutputDir, relPath)

				if err := sprite.WritePNG(img, outPath); err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				result.Artifacts = append(result.Artifacts, outPath)
				md.AddSpriteSheet(filepath.Base(f), relPath, meta)
			}
		}
	}

	// Phase 3: Parse and render maps.
	if opts.Scope == ScopeAll || opts.Scope == ScopeMaps {
		if mapDir := filepath.Join(assetsDir, "maps"); dirExists(mapDir) {
			files := discoverFiles(mapDir, ".map", opts.Files)
			for _, f := range files {
				mf, warnings, err := tilemap.LoadMapFile(f)
				if err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}
				for _, w := range warnings {
					result.Warnings = append(result.Warnings, w.Message)
				}

				j := mf.ToJSON()
				baseName := strings.TrimSuffix(filepath.Base(f), ".map")
				relPath := filepath.Join("maps", baseName+".json")
				outPath := filepath.Join(opts.OutputDir, relPath)

				if err := tilemap.WriteJSON(j, outPath); err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				result.Artifacts = append(result.Artifacts, outPath)
				md.AddMap(filepath.Base(f), relPath)
			}
		}
	}

	// Phase 4: Parse instruments (needed by audio).
	instruments := map[string]*instrument.Instrument{}
	if instDir := filepath.Join(assetsDir, "instruments"); dirExists(instDir) {
		files := discoverFiles(instDir, ".inst", opts.Files)
		for _, f := range files {
			inst, err := instrument.LoadInstrument(f)
			if err != nil {
				result.Errors = append(result.Errors, err)
				continue
			}
			instruments[inst.Name] = inst
		}
	}

	// Phase 5: Render SFX and tracks.
	if opts.Scope == ScopeAll || opts.Scope == ScopeAudio {
		if sfxDir := filepath.Join(assetsDir, "sfx"); dirExists(sfxDir) {
			files := discoverFiles(sfxDir, ".sfx", opts.Files)
			for _, f := range files {
				s, err := sfx.LoadSFX(f)
				if err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				samples, audioWarnings := s.Render(cfg.Defaults.SampleRate)
				for _, w := range audioWarnings {
					result.Warnings = append(result.Warnings, w.Message)
				}

				baseName := strings.TrimSuffix(filepath.Base(f), ".sfx")
				relPath := filepath.Join("audio", baseName+".wav")
				outPath := filepath.Join(opts.OutputDir, relPath)

				if err := audio.WriteWAV(outPath, samples, cfg.Defaults.SampleRate, cfg.Defaults.BitDepth); err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				result.Artifacts = append(result.Artifacts, outPath)
				md.AddAudio(filepath.Base(f), relPath)
			}
		}

		if trackDir := filepath.Join(assetsDir, "tracks"); dirExists(trackDir) {
			files := discoverFiles(trackDir, ".track", opts.Files)
			for _, f := range files {
				tr, err := track.LoadTrack(f)
				if err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				samples, err := tr.Render(instruments, cfg.Defaults.SampleRate)
				if err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				baseName := strings.TrimSuffix(filepath.Base(f), ".track")
				relPath := filepath.Join("audio", baseName+".wav")
				outPath := filepath.Join(opts.OutputDir, relPath)

				if err := audio.WriteWAV(outPath, samples, cfg.Defaults.SampleRate, cfg.Defaults.BitDepth); err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}

				result.Artifacts = append(result.Artifacts, outPath)
				md.AddAudio(filepath.Base(f), relPath)
			}
		}
	}

	// Phase 6: Generate manifest.
	manifestPath := filepath.Join(opts.OutputDir, "manifest.go")
	if err := manifest.Generate(md, manifestPath); err != nil {
		result.Errors = append(result.Errors, err)
	} else {
		result.ManifestPath = manifestPath
		result.Artifacts = append(result.Artifacts, manifestPath)
	}

	return result
}

// Validate runs parsing without rendering — checks files for errors.
func Validate(opts Options, cfg *config.ProjectConfig, projectRoot string) *Result {
	result := &Result{}
	assetsDir := filepath.Join(projectRoot, "assets")

	// Parse palettes.
	palettes := map[string]*palette.Palette{}
	if paletteDir := filepath.Join(assetsDir, "palettes"); dirExists(paletteDir) {
		for _, f := range discoverFiles(paletteDir, ".palette", opts.Files) {
			p, err := palette.LoadPalette(f)
			if err != nil {
				result.Errors = append(result.Errors, err)
			} else {
				palettes[p.Name] = p
			}
		}
	}

	// Validate sprites.
	if opts.Scope == "" || opts.Scope == ScopeAll || opts.Scope == ScopeSprites {
		if spriteDir := filepath.Join(assetsDir, "sprites"); dirExists(spriteDir) {
			for _, f := range discoverFiles(spriteDir, ".sprite", opts.Files) {
				sf, err := sprite.LoadSpriteFile(f)
				if err != nil {
					result.Errors = append(result.Errors, err)
					continue
				}
				pal := palettes[sf.PaletteRef]
				if pal == nil {
					pal = &palette.Palette{Colors: map[string]palette.Color{}}
				}
				if _, err := sf.Resolve(pal); err != nil {
					result.Errors = append(result.Errors, err)
				}
			}
		}
	}

	// Validate maps.
	if opts.Scope == "" || opts.Scope == ScopeAll || opts.Scope == ScopeMaps {
		if mapDir := filepath.Join(assetsDir, "maps"); dirExists(mapDir) {
			for _, f := range discoverFiles(mapDir, ".map", opts.Files) {
				_, warnings, err := tilemap.LoadMapFile(f)
				if err != nil {
					result.Errors = append(result.Errors, err)
				}
				for _, w := range warnings {
					result.Warnings = append(result.Warnings, w.Message)
				}
			}
		}
	}

	// Validate instruments.
	if instDir := filepath.Join(assetsDir, "instruments"); dirExists(instDir) {
		for _, f := range discoverFiles(instDir, ".inst", opts.Files) {
			if _, err := instrument.LoadInstrument(f); err != nil {
				result.Errors = append(result.Errors, err)
			}
		}
	}

	// Validate SFX.
	if opts.Scope == "" || opts.Scope == ScopeAll || opts.Scope == ScopeAudio {
		if sfxDir := filepath.Join(assetsDir, "sfx"); dirExists(sfxDir) {
			for _, f := range discoverFiles(sfxDir, ".sfx", opts.Files) {
				if _, err := sfx.LoadSFX(f); err != nil {
					result.Errors = append(result.Errors, err)
				}
			}
		}

		if trackDir := filepath.Join(assetsDir, "tracks"); dirExists(trackDir) {
			for _, f := range discoverFiles(trackDir, ".track", opts.Files) {
				if _, err := track.LoadTrack(f); err != nil {
					result.Errors = append(result.Errors, err)
				}
			}
		}
	}

	return result
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func discoverFiles(dir, ext string, filter []string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
			continue
		}
		fullPath := filepath.Join(dir, e.Name())
		if len(filter) > 0 && !matchesFilter(fullPath, e.Name(), filter) {
			continue
		}
		// Validate UTF-8 encoding early.
		if err := checkUTF8(fullPath); err != nil {
			continue // skip non-UTF-8 files silently; parsers will report errors
		}
		files = append(files, fullPath)
	}
	return files
}

func matchesFilter(fullPath, name string, filter []string) bool {
	for _, f := range filter {
		if f == name || f == fullPath || filepath.Base(f) == name {
			return true
		}
	}
	return false
}

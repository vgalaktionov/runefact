package config

import (
	"errors"
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

// ProjectConfig represents the runefact.toml project configuration.
type ProjectConfig struct {
	Project  ProjectSection  `toml:"project"`
	Defaults DefaultsSection `toml:"defaults"`
	Preview  PreviewSection  `toml:"preview"`
}

// ProjectSection contains project-level settings.
type ProjectSection struct {
	Name    string `toml:"name"`
	Output  string `toml:"output"`
	Package string `toml:"package"`
}

// DefaultsSection contains default asset parameters.
type DefaultsSection struct {
	SpriteSize int `toml:"sprite_size"`
	SampleRate int `toml:"sample_rate"`
	BitDepth   int `toml:"bit_depth"`
}

// PreviewSection contains live previewer settings.
type PreviewSection struct {
	WindowWidth  int     `toml:"window_width"`
	WindowHeight int     `toml:"window_height"`
	Background   string  `toml:"background"`
	PixelScale   int     `toml:"pixel_scale"`
	AudioVolume  float64 `toml:"audio_volume"`
}

// LoadConfig reads and parses a runefact.toml file.
func LoadConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return ParseConfig(data)
}

// ParseConfig parses runefact.toml content and applies defaults.
func ParseConfig(data []byte) (*ProjectConfig, error) {
	cfg := &ProjectConfig{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	applyDefaults(cfg)
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func applyDefaults(cfg *ProjectConfig) {
	if cfg.Project.Output == "" {
		cfg.Project.Output = "build/assets"
	}
	if cfg.Project.Package == "" {
		cfg.Project.Package = "assets"
	}
	if cfg.Defaults.SpriteSize == 0 {
		cfg.Defaults.SpriteSize = 16
	}
	if cfg.Defaults.SampleRate == 0 {
		cfg.Defaults.SampleRate = 44100
	}
	if cfg.Defaults.BitDepth == 0 {
		cfg.Defaults.BitDepth = 16
	}
	if cfg.Preview.WindowWidth == 0 {
		cfg.Preview.WindowWidth = 800
	}
	if cfg.Preview.WindowHeight == 0 {
		cfg.Preview.WindowHeight = 600
	}
	if cfg.Preview.Background == "" {
		cfg.Preview.Background = "#1a1a2e"
	}
	if cfg.Preview.PixelScale == 0 {
		cfg.Preview.PixelScale = 4
	}
	if cfg.Preview.AudioVolume == 0 {
		cfg.Preview.AudioVolume = 0.5
	}
}

func validate(cfg *ProjectConfig) error {
	var errs []error
	if cfg.Defaults.SpriteSize < 1 {
		errs = append(errs, fmt.Errorf("defaults.sprite_size must be positive, got %d", cfg.Defaults.SpriteSize))
	}
	if cfg.Defaults.SampleRate < 1 {
		errs = append(errs, fmt.Errorf("defaults.sample_rate must be positive, got %d", cfg.Defaults.SampleRate))
	}
	if cfg.Defaults.BitDepth != 8 && cfg.Defaults.BitDepth != 16 && cfg.Defaults.BitDepth != 24 {
		errs = append(errs, fmt.Errorf("defaults.bit_depth must be 8, 16, or 24, got %d", cfg.Defaults.BitDepth))
	}
	if cfg.Preview.AudioVolume < 0 || cfg.Preview.AudioVolume > 1 {
		errs = append(errs, fmt.Errorf("preview.audio_volume must be 0.0-1.0, got %f", cfg.Preview.AudioVolume))
	}
	return errors.Join(errs...)
}

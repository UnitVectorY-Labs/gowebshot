package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Preset represents a named screen resolution preset.
type Preset string

const (
	PresetWidescreen Preset = "widescreen"
	PresetDesktop    Preset = "desktop"
	PresetSquare     Preset = "square"
	PresetPortrait   Preset = "portrait"
	PresetCustom     Preset = "custom"
)

type presetDimensions struct {
	Width  int
	Height int
}

// Presets maps preset name strings to their width and height.
var Presets = map[Preset]presetDimensions{
	PresetWidescreen: {Width: 1920, Height: 1080},
	PresetDesktop:    {Width: 1280, Height: 1024},
	PresetSquare:     {Width: 1080, Height: 1080},
	PresetPortrait:   {Width: 1080, Height: 1920},
}

// PresetNames returns the ordered list of available preset names.
// This is maintained as a literal slice because Go maps are unordered.
func PresetNames() []string {
	return []string{"widescreen", "desktop", "square", "portrait"}
}

// Config holds all configuration for taking a screenshot.
type Config struct {
	URL        string
	Dir        string
	Filename   string
	Preset     Preset
	Width      int
	Height     int
	Zoom       float64
	Scroll     int
	ChromePath string
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Preset:   PresetWidescreen,
		Width:    1920,
		Height:   1080,
		Zoom:     1.0,
		Scroll:   0,
		Filename: "screenshot.png",
		Dir:      "",
	}
}

// ApplyPreset sets Width and Height from the named preset.
func (c *Config) ApplyPreset(preset Preset) error {
	dims, ok := Presets[preset]
	if !ok {
		return fmt.Errorf("unknown preset: %q", preset)
	}
	c.Preset = preset
	c.Width = dims.Width
	c.Height = dims.Height
	return nil
}

// Validate checks that the configuration is complete and consistent.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.URL) == "" {
		return fmt.Errorf("url is required")
	}
	if c.Width <= 0 {
		return fmt.Errorf("width must be greater than 0")
	}
	if c.Height <= 0 {
		return fmt.Errorf("height must be greater than 0")
	}
	if c.Zoom <= 0 {
		return fmt.Errorf("zoom must be greater than 0")
	}
	if c.Scroll < 0 {
		return fmt.Errorf("scroll must be non-negative")
	}
	if strings.TrimSpace(c.Filename) == "" {
		return fmt.Errorf("filename is required")
	}
	// Only reject non-.png extensions; no extension is allowed (will default to .png at output time).
	if ext := filepath.Ext(c.Filename); ext != "" && ext != ".png" {
		return fmt.Errorf("output file must have .png extension")
	}
	return nil
}

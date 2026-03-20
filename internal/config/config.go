package config

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

// Crop describes how many pixels should be removed from each edge.
type Crop struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}

// Presets maps preset name strings to their width and height.
var Presets = map[Preset]presetDimensions{
	PresetWidescreen: {Width: 1920, Height: 1080},
	PresetDesktop:    {Width: 1440, Height: 900},
	PresetSquare:     {Width: 1200, Height: 1200},
	PresetPortrait:   {Width: 1080, Height: 1350},
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
	Crop       Crop
	Shift      bool
	Delay      time.Duration
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
		Crop:     Crop{},
		Shift:    false,
		Delay:    time.Second,
		Filename: "screenshot.png",
		Dir:      "",
	}
}

// ParseCrop parses a crop specification in top,bottom,left,right order.
func ParseCrop(value string) (Crop, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return Crop{}, fmt.Errorf("crop must be in top,bottom,left,right format")
	}

	values := make([]int, 4)
	for i, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return Crop{}, fmt.Errorf("crop values must be integers")
		}
		values[i] = n
	}

	return Crop{
		Top:    values[0],
		Bottom: values[1],
		Left:   values[2],
		Right:  values[3],
	}, nil
}

// IsZero reports whether the crop removes nothing.
func (c Crop) IsZero() bool {
	return c.Top == 0 && c.Bottom == 0 && c.Left == 0 && c.Right == 0
}

// Horizontal reports the total pixels removed from the left and right edges.
func (c Crop) Horizontal() int {
	return c.Left + c.Right
}

// Vertical reports the total pixels removed from the top and bottom edges.
func (c Crop) Vertical() int {
	return c.Top + c.Bottom
}

// CaptureWidth returns the width that should be captured before crop is applied.
func (c Config) CaptureWidth() int {
	if c.Shift {
		return c.Width + c.Crop.Horizontal()
	}
	return c.Width
}

// CaptureHeight returns the height that should be captured before crop is applied.
func (c Config) CaptureHeight() int {
	if c.Shift {
		return c.Height + c.Crop.Vertical()
	}
	return c.Height
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
	if c.Crop.Top < 0 || c.Crop.Bottom < 0 || c.Crop.Left < 0 || c.Crop.Right < 0 {
		return fmt.Errorf("crop values must be non-negative")
	}
	if c.Shift && c.Crop.IsZero() {
		return fmt.Errorf("shift requires a non-zero crop")
	}
	if !c.Shift {
		if c.Crop.Horizontal() >= c.Width {
			return fmt.Errorf("crop left + right must be less than width unless --shift is set")
		}
		if c.Crop.Vertical() >= c.Height {
			return fmt.Errorf("crop top + bottom must be less than height unless --shift is set")
		}
	}
	if c.Delay < 0 {
		return fmt.Errorf("delay must be non-negative")
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

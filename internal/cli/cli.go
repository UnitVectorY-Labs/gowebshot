package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
)

// ParseFlags parses command-line flags from args (typically os.Args[1:]).
// It returns the resolved Config, whether interactive mode should be used,
// and any error encountered during parsing.
func ParseFlags(args []string) (config.Config, bool, error) {
	if len(args) == 0 {
		return config.DefaultConfig(), true, nil
	}

	fs := flag.NewFlagSet("gowebshot", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	url := fs.String("url", "", "URL to capture")
	dir := fs.String("dir", "", "output directory")
	filename := fs.String("filename", "screenshot.png", "output filename")
	preset := fs.String("preset", "", "screen resolution preset ("+fmt.Sprintf("%v", config.PresetNames())+")")
	width := fs.Int("width", 0, "viewport width in pixels")
	height := fs.Int("height", 0, "viewport height in pixels")
	zoom := fs.Float64("zoom", 1.0, "page zoom level")
	scroll := fs.Int("scroll", 0, "pixels to scroll before capture")
	chrome := fs.String("chrome", "", "path to Chrome executable")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return config.Config{}, false, flag.ErrHelp
		}
		return config.Config{}, false, err
	}

	// Validate preset vs explicit dimensions.
	hasPreset := *preset != ""
	hasWidth := *width != 0
	hasHeight := *height != 0

	if hasPreset && (hasWidth || hasHeight) {
		return config.Config{}, false, fmt.Errorf("cannot use --preset with --width/--height")
	}
	if hasWidth != hasHeight {
		return config.Config{}, false, fmt.Errorf("--width and --height must be used together")
	}

	cfg := config.DefaultConfig()

	if hasPreset {
		if err := cfg.ApplyPreset(config.Preset(*preset)); err != nil {
			return config.Config{}, false, err
		}
	}

	if hasWidth && hasHeight {
		cfg.Preset = config.PresetCustom
		cfg.Width = *width
		cfg.Height = *height
	}

	cfg.URL = *url
	cfg.Dir = *dir
	cfg.Filename = *filename
	cfg.Zoom = *zoom
	cfg.Scroll = *scroll
	cfg.ChromePath = *chrome

	return cfg, false, nil
}

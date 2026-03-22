package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
)

// ParseFlags parses command-line flags from args (typically os.Args[1:]).
// It returns the resolved Config, whether interactive mode should be used,
// whether the version should be printed, and any error encountered during parsing.
func ParseFlags(args []string) (config.Config, bool, bool, error) {
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
	crop := fs.String("crop", "", "crop pixels as top,bottom,left,right")
	shift := fs.Bool("shift", false, "expand the capture area so crop keeps the requested output size")
	delay := fs.Duration("delay", time.Second, "delay after page load before capture (for example 500ms or 1s)")
	chrome := fs.String("chrome", "", "path to Chrome executable")
	version := fs.Bool("version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return config.Config{}, false, false, flag.ErrHelp
		}
		return config.Config{}, false, false, err
	}

	if *version {
		return config.Config{}, false, true, nil
	}

	urlProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "url" {
			urlProvided = true
		}
	})

	// Validate preset vs explicit dimensions.
	hasPreset := *preset != ""
	hasWidth := *width != 0
	hasHeight := *height != 0

	if hasPreset && (hasWidth || hasHeight) {
		return config.Config{}, false, false, fmt.Errorf("cannot use --preset with --width/--height")
	}
	if hasWidth != hasHeight {
		return config.Config{}, false, false, fmt.Errorf("--width and --height must be used together")
	}

	cfg := config.DefaultConfig()

	if hasPreset {
		if err := cfg.ApplyPreset(config.Preset(*preset)); err != nil {
			return config.Config{}, false, false, err
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
	if *crop != "" {
		parsedCrop, err := config.ParseCrop(*crop)
		if err != nil {
			return config.Config{}, false, false, err
		}
		cfg.Crop = parsedCrop
	}
	cfg.Shift = *shift
	cfg.Delay = *delay
	cfg.ChromePath = *chrome

	return cfg, !urlProvided, false, nil
}

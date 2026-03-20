package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Preset != PresetWidescreen {
		t.Errorf("expected preset %q, got %q", PresetWidescreen, cfg.Preset)
	}
	if cfg.Width != 1920 {
		t.Errorf("expected width 1920, got %d", cfg.Width)
	}
	if cfg.Height != 1080 {
		t.Errorf("expected height 1080, got %d", cfg.Height)
	}
	if cfg.Zoom != 1.0 {
		t.Errorf("expected zoom 1.0, got %f", cfg.Zoom)
	}
	if cfg.Scroll != 0 {
		t.Errorf("expected scroll 0, got %d", cfg.Scroll)
	}
	if cfg.Crop != (Crop{}) {
		t.Errorf("expected zero crop, got %+v", cfg.Crop)
	}
	if cfg.Shift {
		t.Error("expected shift to default to false")
	}
	if cfg.Delay != time.Second {
		t.Errorf("expected delay 1s, got %s", cfg.Delay)
	}
	if cfg.Filename != "screenshot.png" {
		t.Errorf("expected filename %q, got %q", "screenshot.png", cfg.Filename)
	}
	if cfg.Dir != "" {
		t.Errorf("expected empty dir, got %q", cfg.Dir)
	}
}

func TestPresetNames(t *testing.T) {
	names := PresetNames()
	expected := []string{"widescreen", "desktop", "square", "portrait"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d preset names, got %d", len(expected), len(names))
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected preset name[%d] = %q, got %q", i, name, names[i])
		}
	}
}

func TestApplyPreset(t *testing.T) {
	tests := []struct {
		preset Preset
		width  int
		height int
	}{
		{PresetWidescreen, 1920, 1080},
		{PresetDesktop, 1440, 900},
		{PresetSquare, 1200, 1200},
		{PresetPortrait, 1080, 1350},
	}

	for _, tt := range tests {
		t.Run(string(tt.preset), func(t *testing.T) {
			cfg := DefaultConfig()
			if err := cfg.ApplyPreset(tt.preset); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Width != tt.width {
				t.Errorf("expected width %d, got %d", tt.width, cfg.Width)
			}
			if cfg.Height != tt.height {
				t.Errorf("expected height %d, got %d", tt.height, cfg.Height)
			}
			if cfg.Preset != tt.preset {
				t.Errorf("expected preset %q, got %q", tt.preset, cfg.Preset)
			}
		})
	}
}

func TestApplyPresetInvalid(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.ApplyPreset("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown preset, got nil")
	}
}

func TestValidateURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = ""
	if err := cfg.Validate(); err == nil || err.Error() != "url is required" {
		t.Errorf("expected 'url is required', got %v", err)
	}

	cfg.URL = "   "
	if err := cfg.Validate(); err == nil || err.Error() != "url is required" {
		t.Errorf("expected 'url is required' for whitespace-only URL, got %v", err)
	}
}

func TestValidateWidth(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Width = 0
	if err := cfg.Validate(); err == nil || err.Error() != "width must be greater than 0" {
		t.Errorf("expected 'width must be greater than 0', got %v", err)
	}

	cfg.Width = -1
	if err := cfg.Validate(); err == nil || err.Error() != "width must be greater than 0" {
		t.Errorf("expected 'width must be greater than 0', got %v", err)
	}
}

func TestValidateHeight(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Height = 0
	if err := cfg.Validate(); err == nil || err.Error() != "height must be greater than 0" {
		t.Errorf("expected 'height must be greater than 0', got %v", err)
	}
}

func TestValidateZoom(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Zoom = 0
	if err := cfg.Validate(); err == nil || err.Error() != "zoom must be greater than 0" {
		t.Errorf("expected 'zoom must be greater than 0', got %v", err)
	}

	cfg.Zoom = -0.5
	if err := cfg.Validate(); err == nil || err.Error() != "zoom must be greater than 0" {
		t.Errorf("expected 'zoom must be greater than 0', got %v", err)
	}
}

func TestValidateScroll(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Scroll = -1
	if err := cfg.Validate(); err == nil || err.Error() != "scroll must be non-negative" {
		t.Errorf("expected 'scroll must be non-negative', got %v", err)
	}
}

func TestValidateDelay(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Delay = -1 * time.Second
	if err := cfg.Validate(); err == nil || err.Error() != "delay must be non-negative" {
		t.Errorf("expected 'delay must be non-negative', got %v", err)
	}
}

func TestParseCrop(t *testing.T) {
	crop, err := ParseCrop("10,20,30,40")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if crop != (Crop{Top: 10, Bottom: 20, Left: 30, Right: 40}) {
		t.Fatalf("unexpected crop: %+v", crop)
	}
}

func TestParseCropErrors(t *testing.T) {
	tests := []string{
		"10,20,30",
		"10,20,30,nope",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseCrop(input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestCaptureDimensions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Width = 800
	cfg.Height = 600
	cfg.Crop = Crop{Top: 10, Bottom: 20, Left: 30, Right: 40}

	if got := cfg.CaptureWidth(); got != 800 {
		t.Fatalf("expected capture width 800 without shift, got %d", got)
	}
	if got := cfg.CaptureHeight(); got != 600 {
		t.Fatalf("expected capture height 600 without shift, got %d", got)
	}

	cfg.Shift = true

	if got := cfg.CaptureWidth(); got != 870 {
		t.Fatalf("expected capture width 870 with shift, got %d", got)
	}
	if got := cfg.CaptureHeight(); got != 630 {
		t.Fatalf("expected capture height 630 with shift, got %d", got)
	}
}

func TestValidateFilename(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"

	cfg.Filename = ""
	if err := cfg.Validate(); err == nil || err.Error() != "filename is required" {
		t.Errorf("expected 'filename is required', got %v", err)
	}

	cfg.Filename = "   "
	if err := cfg.Validate(); err == nil || err.Error() != "filename is required" {
		t.Errorf("expected 'filename is required' for whitespace-only filename, got %v", err)
	}
}

func TestValidateFilenameExtension(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"

	cfg.Filename = "screenshot.jpg"
	if err := cfg.Validate(); err == nil || err.Error() != "output file must have .png extension" {
		t.Errorf("expected 'output file must have .png extension', got %v", err)
	}

	cfg.Filename = "screenshot.jpeg"
	if err := cfg.Validate(); err == nil || err.Error() != "output file must have .png extension" {
		t.Errorf("expected 'output file must have .png extension', got %v", err)
	}

	// .png extension should be accepted
	cfg.Filename = "screenshot.png"
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for .png extension, got %v", err)
	}

	// No extension should be accepted
	cfg.Filename = "screenshot"
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for no extension, got %v", err)
	}
}

func TestValidatePass(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config to pass, got %v", err)
	}
}

func TestValidateCropValues(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Crop.Top = -1
	if err := cfg.Validate(); err == nil || err.Error() != "crop values must be non-negative" {
		t.Fatalf("expected crop non-negative error, got %v", err)
	}
}

func TestValidateShiftRequiresCrop(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Shift = true
	if err := cfg.Validate(); err == nil || err.Error() != "shift requires a non-zero crop" {
		t.Fatalf("expected shift validation error, got %v", err)
	}
}

func TestValidateCropFitsViewportWithoutShift(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Width = 100
	cfg.Height = 80
	cfg.Crop = Crop{Left: 60, Right: 40}
	if err := cfg.Validate(); err == nil || err.Error() != "crop left + right must be less than width unless --shift is set" {
		t.Fatalf("expected width crop validation error, got %v", err)
	}

	cfg = DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Width = 100
	cfg.Height = 80
	cfg.Crop = Crop{Top: 30, Bottom: 50}
	if err := cfg.Validate(); err == nil || err.Error() != "crop top + bottom must be less than height unless --shift is set" {
		t.Fatalf("expected height crop validation error, got %v", err)
	}
}

func TestValidateCropWithShiftPasses(t *testing.T) {
	cfg := DefaultConfig()
	cfg.URL = "https://example.com"
	cfg.Width = 100
	cfg.Height = 80
	cfg.Crop = Crop{Top: 30, Bottom: 50}
	cfg.Shift = true
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected crop with shift to pass, got %v", err)
	}
}

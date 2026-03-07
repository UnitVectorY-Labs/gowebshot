package cli

import (
	"errors"
	"flag"
	"testing"
	"time"
)

func TestEmptyArgs_InteractiveMode(t *testing.T) {
	cfg, interactive, err := ParseFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !interactive {
		t.Fatal("expected interactive mode for empty args")
	}
	if cfg.Width != 1920 || cfg.Height != 1080 {
		t.Fatalf("expected default dimensions 1920x1080, got %dx%d", cfg.Width, cfg.Height)
	}
}

func TestURLFlagEnablesNonInteractiveMode(t *testing.T) {
	cfg, interactive, err := ParseFlags([]string{"--url", "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interactive {
		t.Fatal("expected non-interactive mode")
	}
	if cfg.URL != "https://example.com" {
		t.Fatalf("expected URL https://example.com, got %s", cfg.URL)
	}
}

func TestFlagsWithoutURLRemainInteractive(t *testing.T) {
	cfg, interactive, err := ParseFlags([]string{"--preset", "square", "--delay", "1500ms"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !interactive {
		t.Fatal("expected interactive mode when --url is omitted")
	}
	if cfg.Preset != "square" {
		t.Fatalf("expected preset square, got %s", cfg.Preset)
	}
	if cfg.Width != 1200 || cfg.Height != 1200 {
		t.Fatalf("expected 1200x1200 for square preset, got %dx%d", cfg.Width, cfg.Height)
	}
	if cfg.Delay != 1500*time.Millisecond {
		t.Fatalf("expected 1500ms delay, got %s", cfg.Delay)
	}
}

func TestPresetSquare(t *testing.T) {
	cfg, _, err := ParseFlags([]string{"--preset", "square"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Width != 1200 || cfg.Height != 1200 {
		t.Fatalf("expected 1200x1200 for square preset, got %dx%d", cfg.Width, cfg.Height)
	}
}

func TestPresetWithWidthErrors(t *testing.T) {
	_, _, err := ParseFlags([]string{"--preset", "square", "--width", "800"})
	if err == nil {
		t.Fatal("expected error when using --preset with --width")
	}
	expected := "cannot use --preset with --width/--height"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestWidthWithoutHeightErrors(t *testing.T) {
	_, _, err := ParseFlags([]string{"--width", "800"})
	if err == nil {
		t.Fatal("expected error when using --width without --height")
	}
	expected := "--width and --height must be used together"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestHelpReturnsErrHelp(t *testing.T) {
	_, _, err := ParseFlags([]string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected flag.ErrHelp, got %v", err)
	}
}

func TestDelayFlag(t *testing.T) {
	cfg, interactive, err := ParseFlags([]string{"--url", "https://example.com", "--delay", "1500ms"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interactive {
		t.Fatal("expected non-interactive mode")
	}
	if cfg.Delay != 1500*time.Millisecond {
		t.Fatalf("expected 1500ms delay, got %s", cfg.Delay)
	}
}

func TestExplicitURLFlagKeepsNonInteractiveModeEvenWhenEmpty(t *testing.T) {
	_, interactive, err := ParseFlags([]string{"--url", ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interactive {
		t.Fatal("expected non-interactive mode when --url is provided")
	}
}

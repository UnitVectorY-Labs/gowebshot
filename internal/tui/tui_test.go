package tui

import (
	"testing"
	"time"

	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
)

func newTestModel() model {
	return newModel(config.DefaultConfig())
}

func TestEditInsertRespectsCursor(t *testing.T) {
	m := newTestModel()
	m.filename = "screen.png"
	m.editingField = fieldFilename
	m.setCursorForField(fieldFilename, 6)

	m.editInsert("-new")

	if got, want := m.filename, "screen-new.png"; got != want {
		t.Fatalf("expected filename %q, got %q", want, got)
	}
}

func TestSyncPresetFromDimensionsMatchesNamedPreset(t *testing.T) {
	m := newTestModel()
	m.customWidth = "1440"
	m.customHeight = "900"

	m.syncPresetFromDimensions()

	if got, want := m.currentPreset(), string(config.PresetDesktop); got != want {
		t.Fatalf("expected preset %q, got %q", want, got)
	}
}

func TestSyncPresetFromDimensionsFallsBackToCustom(t *testing.T) {
	m := newTestModel()
	m.customWidth = "1439"
	m.customHeight = "900"

	m.syncPresetFromDimensions()

	if got, want := m.currentPreset(), string(config.PresetCustom); got != want {
		t.Fatalf("expected preset %q, got %q", want, got)
	}
}

func TestAdjustEditingValueClampsNumericFields(t *testing.T) {
	m := newTestModel()

	m.editingField = fieldZoom
	m.zoomPercent = "1"
	m.adjustEditingValue(-1)
	if got := m.zoomPercent; got != "1" {
		t.Fatalf("expected zoom to clamp at 1, got %q", got)
	}

	m.editingField = fieldScroll
	m.scroll = "0"
	m.adjustEditingValue(-1)
	if got := m.scroll; got != "0" {
		t.Fatalf("expected scroll to clamp at 0, got %q", got)
	}

	m.editingField = fieldDelay
	m.delay = "0s"
	m.adjustEditingValue(-1)
	if got := m.delay; got != "0s" {
		t.Fatalf("expected delay to clamp at 0s, got %q", got)
	}

	m.editingField = fieldCropTop
	m.cropTop = "0"
	m.adjustEditingValue(-1)
	if got := m.cropTop; got != "0" {
		t.Fatalf("expected crop top to clamp at 0, got %q", got)
	}
}

func TestBuildConfigConvertsZoomPercentAndDelay(t *testing.T) {
	m := newTestModel()
	m.url = "https://example.com"
	m.zoomPercent = "125"
	m.scroll = "12"
	m.delay = "1500ms"
	m.cropTop = "10"
	m.cropBottom = "20"
	m.cropLeft = "30"
	m.cropRight = "40"
	m.shift = true

	cfg := m.buildConfig()

	if cfg.Zoom != 1.25 {
		t.Fatalf("expected zoom 1.25, got %v", cfg.Zoom)
	}
	if cfg.Scroll != 12 {
		t.Fatalf("expected scroll 12, got %d", cfg.Scroll)
	}
	if cfg.Delay != 1500*time.Millisecond {
		t.Fatalf("expected delay 1500ms, got %s", cfg.Delay)
	}
	if cfg.Crop != (config.Crop{Top: 10, Bottom: 20, Left: 30, Right: 40}) {
		t.Fatalf("unexpected crop: %+v", cfg.Crop)
	}
	if !cfg.Shift {
		t.Fatal("expected shift to be enabled")
	}
}

func TestNewModelPrefillsInteractiveValues(t *testing.T) {
	initial := config.Config{
		URL:        "https://example.com",
		Dir:        "shots",
		Filename:   "example.png",
		Preset:     config.PresetCustom,
		Width:      800,
		Height:     600,
		Zoom:       1.25,
		Scroll:     42,
		Crop:       config.Crop{Top: 10, Bottom: 20, Left: 30, Right: 40},
		Shift:      true,
		Delay:      1500 * time.Millisecond,
		ChromePath: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	}

	m := newModel(initial)

	if m.url != initial.URL {
		t.Fatalf("expected URL %q, got %q", initial.URL, m.url)
	}
	if m.dir != initial.Dir {
		t.Fatalf("expected dir %q, got %q", initial.Dir, m.dir)
	}
	if m.filename != initial.Filename {
		t.Fatalf("expected filename %q, got %q", initial.Filename, m.filename)
	}
	if m.currentPreset() != string(config.PresetCustom) {
		t.Fatalf("expected preset %q, got %q", config.PresetCustom, m.currentPreset())
	}
	if m.customWidth != "800" || m.customHeight != "600" {
		t.Fatalf("expected dimensions 800x600, got %sx%s", m.customWidth, m.customHeight)
	}
	if m.zoomPercent != "125" {
		t.Fatalf("expected zoom percent 125, got %q", m.zoomPercent)
	}
	if m.scroll != "42" {
		t.Fatalf("expected scroll 42, got %q", m.scroll)
	}
	if m.cropTop != "10" || m.cropBottom != "20" || m.cropLeft != "30" || m.cropRight != "40" {
		t.Fatalf("unexpected crop values %q %q %q %q", m.cropTop, m.cropBottom, m.cropLeft, m.cropRight)
	}
	if !m.shift {
		t.Fatal("expected shift to be enabled")
	}
	if m.delay != "1.5s" {
		t.Fatalf("expected delay 1.5s, got %q", m.delay)
	}
	if m.chromePath != initial.ChromePath {
		t.Fatalf("expected chrome path %q, got %q", initial.ChromePath, m.chromePath)
	}
}

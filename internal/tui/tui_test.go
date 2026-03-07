package tui

import (
	"testing"
	"time"

	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
)

func newTestModel() model {
	defaults := config.DefaultConfig()
	return model{
		presetNames:  append(config.PresetNames(), string(config.PresetCustom)),
		customWidth:  "1920",
		customHeight: "1080",
		zoomPercent:  "100",
		scroll:       "0",
		delay:        defaults.Delay.String(),
		fieldCursors: make(map[fieldID]int),
	}
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
}

func TestBuildConfigConvertsZoomPercentAndDelay(t *testing.T) {
	m := newTestModel()
	m.url = "https://example.com"
	m.zoomPercent = "125"
	m.scroll = "12"
	m.delay = "1500ms"

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
}

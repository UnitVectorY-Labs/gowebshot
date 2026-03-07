package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/UnitVectorY-Labs/gowebshot/internal/capture"
	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
	"github.com/UnitVectorY-Labs/gowebshot/internal/output"
)

// ── Message types ──────────────────────────────────────────────────────────

type captureResultMsg struct {
	path string
	size int64
	err  error
}

// ── Model ──────────────────────────────────────────────────────────────────

type model struct {
	tabs      []string
	activeTab int

	fieldIndex int

	// Input tab
	url        string
	urlEditing bool

	// Output tab
	dir             string
	dirEditing      bool
	filename        string
	filenameEditing bool

	// Settings tab
	presetIndex    int
	presetNames    []string
	presetPicking  bool
	zoom           string
	zoomEditing    bool
	scroll         string
	scrollEditing  bool
	customWidth    string
	customWidthEditing  bool
	customHeight   string
	customHeightEditing bool

	chromePath string

	capturing      bool
	message        string
	messageIsError bool

	width  int
	height int

	escOnce bool
}

// ── Styling ────────────────────────────────────────────────────────────────

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // violet
	accentColor    = lipgloss.Color("#A78BFA") // light violet
	dimColor       = lipgloss.Color("#6B7280") // gray
	textColor      = lipgloss.Color("#E5E7EB") // light gray
	brightColor    = lipgloss.Color("#F9FAFB") // near-white
	successColor   = lipgloss.Color("#34D399") // green
	errorColor     = lipgloss.Color("#F87171") // red
	bgColor        = lipgloss.Color("#1F2937") // dark bg
	fieldBgColor   = lipgloss.Color("#374151") // field bg
	activeBgColor  = lipgloss.Color("#4C1D95") // active tab bg

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(brightColor).
			Background(primaryColor).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(dimColor).
				Background(fieldBgColor).
				Padding(0, 2)

	tabGapStyle = lipgloss.NewStyle().
			Background(bgColor).
			PaddingRight(0)

	// Content area
	contentStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(textColor)

	// Field styles
	labelStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Width(14)

	valueStyle = lipgloss.NewStyle().
			Foreground(textColor)

	editingValueStyle = lipgloss.NewStyle().
				Foreground(brightColor).
				Background(activeBgColor).
				Padding(0, 1)

	activeFieldStyle = lipgloss.NewStyle().
				Foreground(brightColor).
				Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Button
	buttonStyle = lipgloss.NewStyle().
			Foreground(brightColor).
			Background(primaryColor).
			Padding(0, 3).
			Bold(true)

	buttonActiveStyle = lipgloss.NewStyle().
				Foreground(brightColor).
				Background(accentColor).
				Padding(0, 3).
				Bold(true)

	// Message styles
	successMsgStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Section border
	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Help text
	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true)

	// Title
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
)

// ── Helpers ────────────────────────────────────────────────────────────────

func (m model) currentPreset() string {
	if m.presetIndex >= 0 && m.presetIndex < len(m.presetNames) {
		return m.presetNames[m.presetIndex]
	}
	return "widescreen"
}

func (m model) resolutionWidth() int {
	if m.currentPreset() == "custom" {
		w, _ := strconv.Atoi(m.customWidth)
		return w
	}
	p := config.Preset(m.currentPreset())
	if dims, ok := config.Presets[p]; ok {
		return dims.Width
	}
	return 1920
}

func (m model) resolutionHeight() int {
	if m.currentPreset() == "custom" {
		h, _ := strconv.Atoi(m.customHeight)
		return h
	}
	p := config.Preset(m.currentPreset())
	if dims, ok := config.Presets[p]; ok {
		return dims.Height
	}
	return 1080
}

func (m model) fieldCountForTab() int {
	switch m.activeTab {
	case 0: // Generate
		return 1 // just the button
	case 1: // Input
		return 1
	case 2: // Output
		return 2
	case 3: // Settings
		if m.currentPreset() == "custom" {
			return 5 // preset, zoom, scroll, width, height
		}
		return 3 // preset, zoom, scroll
	}
	return 0
}

func (m model) buildConfig() config.Config {
	z, err := strconv.ParseFloat(m.zoom, 64)
	if err != nil {
		z = 1.0
	}
	s, err := strconv.Atoi(m.scroll)
	if err != nil {
		s = 0
	}

	cfg := config.Config{
		URL:        m.url,
		Dir:        m.dir,
		Filename:   m.filename,
		Width:      m.resolutionWidth(),
		Height:     m.resolutionHeight(),
		Zoom:       z,
		Scroll:     s,
		ChromePath: m.chromePath,
	}

	preset := m.currentPreset()
	if preset != "custom" {
		cfg.Preset = config.Preset(preset)
	} else {
		cfg.Preset = config.PresetCustom
	}

	return cfg
}

func (m *model) stopAllEditing() {
	m.urlEditing = false
	m.dirEditing = false
	m.filenameEditing = false
	m.zoomEditing = false
	m.scrollEditing = false
	m.customWidthEditing = false
	m.customHeightEditing = false
	m.presetPicking = false
}

func (m model) isEditing() bool {
	return m.urlEditing || m.dirEditing || m.filenameEditing ||
		m.zoomEditing || m.scrollEditing ||
		m.customWidthEditing || m.customHeightEditing
}

func doCapture(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		data, err := capture.Capture(cfg)
		if err != nil {
			return captureResultMsg{err: err}
		}

		path, err := output.ResolvePath(cfg.Dir, cfg.Filename)
		if err != nil {
			return captureResultMsg{err: err}
		}

		if err := output.WriteFile(path, data); err != nil {
			return captureResultMsg{err: err}
		}

		return captureResultMsg{path: path, size: int64(len(data))}
	}
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// ── Init / Update / View ──────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case captureResultMsg:
		m.capturing = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Error: %v", msg.err)
			m.messageIsError = true
		} else {
			m.message = fmt.Sprintf("Screenshot saved to %s (%s)", msg.path, humanSize(msg.size))
			m.messageIsError = false
		}
		return m, nil

	case tea.KeyMsg:
		// Always allow Ctrl+C
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		// Block input while capturing
		if m.capturing {
			return m, nil
		}

		// Reset escOnce on any non-esc key
		if msg.Type != tea.KeyEsc {
			m.escOnce = false
		}

		// Handle text editing mode
		if m.isEditing() {
			return m.handleEditing(msg)
		}

		// Handle preset picker mode
		if m.presetPicking {
			return m.handlePresetPicker(msg)
		}

		return m.handleNormal(msg)
	}
	return m, nil
}

func (m model) handleEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.stopAllEditing()
		return m, nil
	case tea.KeyEnter:
		m.stopAllEditing()
		return m, nil
	case tea.KeyBackspace:
		m.editBackspace()
		return m, nil
	default:
		if msg.Type == tea.KeySpace {
			m.editInsert(" ")
		} else if len(msg.Runes) > 0 {
			m.editInsert(string(msg.Runes))
		}
		return m, nil
	}
}

func (m *model) editInsert(s string) {
	switch {
	case m.urlEditing:
		m.url += s
	case m.dirEditing:
		m.dir += s
	case m.filenameEditing:
		m.filename += s
	case m.zoomEditing:
		m.zoom += s
	case m.scrollEditing:
		m.scroll += s
	case m.customWidthEditing:
		m.customWidth += s
	case m.customHeightEditing:
		m.customHeight += s
	}
}

func (m *model) editBackspace() {
	del := func(s string) string {
		if len(s) > 0 {
			return s[:len(s)-1]
		}
		return s
	}
	switch {
	case m.urlEditing:
		m.url = del(m.url)
	case m.dirEditing:
		m.dir = del(m.dir)
	case m.filenameEditing:
		m.filename = del(m.filename)
	case m.zoomEditing:
		m.zoom = del(m.zoom)
	case m.scrollEditing:
		m.scroll = del(m.scroll)
	case m.customWidthEditing:
		m.customWidth = del(m.customWidth)
	case m.customHeightEditing:
		m.customHeight = del(m.customHeight)
	}
}

func (m model) handlePresetPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.presetPicking = false
		return m, nil
	case tea.KeyLeft:
		if m.presetIndex > 0 {
			m.presetIndex--
		} else {
			m.presetIndex = len(m.presetNames) - 1
		}
		m.syncPresetDimensions()
		return m, nil
	case tea.KeyRight:
		m.presetIndex = (m.presetIndex + 1) % len(m.presetNames)
		m.syncPresetDimensions()
		return m, nil
	case tea.KeyEnter:
		m.presetPicking = false
		return m, nil
	}
	return m, nil
}

func (m *model) syncPresetDimensions() {
	preset := m.currentPreset()
	if preset == "custom" {
		return
	}
	p := config.Preset(preset)
	if dims, ok := config.Presets[p]; ok {
		m.customWidth = strconv.Itoa(dims.Width)
		m.customHeight = strconv.Itoa(dims.Height)
	}
}

func (m model) handleNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyTab:
		m.activeTab = (m.activeTab + 1) % len(m.tabs)
		m.fieldIndex = 0
		return m, nil

	case tea.KeyShiftTab:
		m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		m.fieldIndex = 0
		return m, nil

	case tea.KeyRight:
		m.activeTab = (m.activeTab + 1) % len(m.tabs)
		m.fieldIndex = 0
		return m, nil

	case tea.KeyLeft:
		m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		m.fieldIndex = 0
		return m, nil

	case tea.KeyDown:
		maxIdx := m.fieldCountForTab() - 1
		if m.fieldIndex < maxIdx {
			m.fieldIndex++
		}
		return m, nil

	case tea.KeyUp:
		if m.fieldIndex > 0 {
			m.fieldIndex--
		}
		return m, nil

	case tea.KeyEnter:
		return m.activateField()

	case tea.KeySpace:
		if msg.String() == " " {
			return m.activateField()
		}

	case tea.KeyEsc:
		return m.handleEsc()
	}
	return m, nil
}

func (m model) activateField() (tea.Model, tea.Cmd) {
	switch m.activeTab {
	case 0: // Generate
		if m.fieldIndex == 0 {
			return m.triggerCapture()
		}
	case 1: // Input
		if m.fieldIndex == 0 {
			m.stopAllEditing()
			m.urlEditing = true
		}
	case 2: // Output
		m.stopAllEditing()
		switch m.fieldIndex {
		case 0:
			m.dirEditing = true
		case 1:
			m.filenameEditing = true
		}
	case 3: // Settings
		m.stopAllEditing()
		switch m.fieldIndex {
		case 0: // Preset
			m.presetPicking = true
		case 1: // Zoom
			m.zoomEditing = true
		case 2: // Scroll
			m.scrollEditing = true
		case 3: // Width (custom)
			m.customWidthEditing = true
		case 4: // Height (custom)
			m.customHeightEditing = true
		}
	}
	return m, nil
}

func (m model) triggerCapture() (tea.Model, tea.Cmd) {
	cfg := m.buildConfig()
	if err := cfg.Validate(); err != nil {
		errMsg := err.Error()
		m.messageIsError = true

		// Switch to the relevant tab based on error
		switch {
		case strings.Contains(errMsg, "url"):
			m.activeTab = 1
			m.fieldIndex = 0
			m.message = "URL is required"
		case strings.Contains(errMsg, "filename"):
			m.activeTab = 2
			m.fieldIndex = 1
			m.message = errMsg
		case strings.Contains(errMsg, "width"), strings.Contains(errMsg, "height"),
			strings.Contains(errMsg, "zoom"), strings.Contains(errMsg, "scroll"):
			m.activeTab = 3
			m.fieldIndex = 0
			m.message = errMsg
		default:
			m.message = errMsg
		}
		return m, nil
	}

	m.capturing = true
	m.message = "Capturing screenshot..."
	m.messageIsError = false
	return m, doCapture(cfg)
}

func (m model) handleEsc() (tea.Model, tea.Cmd) {
	// 1. If fieldIndex > 0, move to top
	if m.fieldIndex > 0 {
		m.fieldIndex = 0
		return m, nil
	}
	// 2. If not on Input tab, go to Input tab
	if m.activeTab != 1 {
		m.activeTab = 1
		m.fieldIndex = 0
		return m, nil
	}
	// 3. At top of Input tab: double-esc to quit
	if m.escOnce {
		return m, tea.Quit
	}
	m.escOnce = true
	m.message = "Press Esc again to quit"
	m.messageIsError = false
	return m, nil
}

// ── View ───────────────────────────────────────────────────────────────────

func (m model) View() string {
	// Terminal too small
	if m.width < 60 || m.height < 16 {
		msg := lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Render("Terminal too small. Resize to at least 60×16.")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
	}

	var b strings.Builder

	// Header
	header := titleStyle.Render("  gowebshot")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	// Content
	contentWidth := m.width - 6
	if contentWidth < 50 {
		contentWidth = 50
	}

	var content string
	switch m.activeTab {
	case 0:
		content = m.viewGenerate()
	case 1:
		content = m.viewInput()
	case 2:
		content = m.viewOutput()
	case 3:
		content = m.viewSettings()
	}

	boxStyle := sectionStyle.Width(contentWidth)
	b.WriteString(boxStyle.Render(content))
	b.WriteString("\n\n")

	// Message line
	if m.message != "" {
		if m.capturing {
			spinner := lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("⟳ ")
			b.WriteString(spinner + lipgloss.NewStyle().Foreground(accentColor).Render(m.message))
		} else if m.messageIsError {
			b.WriteString(errorMsgStyle.Render("✗ " + m.message))
		} else {
			b.WriteString(successMsgStyle.Render("✓ " + m.message))
		}
		b.WriteString("\n")
	}

	// Help bar
	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m model) renderTabBar() string {
	var tabs []string
	for i, t := range m.tabs {
		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render(t))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(t))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	// Draw a line under the tabs
	barWidth := m.width - 4
	if barWidth < 50 {
		barWidth = 50
	}
	line := lipgloss.NewStyle().
		Foreground(primaryColor).
		Render(strings.Repeat("─", barWidth))

	return row + "\n" + line
}

func (m model) renderHelp() string {
	if m.capturing {
		return helpStyle.Render("  Ctrl+C quit")
	}
	if m.isEditing() {
		return helpStyle.Render("  Enter confirm • Esc cancel • Type to edit")
	}
	if m.presetPicking {
		return helpStyle.Render("  ←/→ cycle preset • Enter confirm • Esc cancel")
	}
	return helpStyle.Render("  ←/→/Tab switch tabs • ↑/↓ move fields • Enter select • Esc back • Ctrl+C quit")
}

// ── Tab Views ──────────────────────────────────────────────────────────────

func (m model) viewGenerate() string {
	var lines []string

	urlVal := m.url
	if urlVal == "" {
		urlVal = lipgloss.NewStyle().Foreground(dimColor).Render("(not set)")
	}

	dirVal := m.dir
	if dirVal == "" {
		dirVal = lipgloss.NewStyle().Foreground(dimColor).Render("(current directory)")
	}

	preset := m.currentPreset()
	resolution := fmt.Sprintf("%s %dx%d", preset, m.resolutionWidth(), m.resolutionHeight())

	lines = append(lines,
		renderReadOnlyField("URL", urlVal),
		renderReadOnlyField("Directory", dirVal),
		renderReadOnlyField("Filename", m.filename),
		renderReadOnlyField("Resolution", resolution),
		renderReadOnlyField("Zoom", m.zoom+"×"),
		renderReadOnlyField("Scroll", m.scroll+"px"),
		"",
	)

	// Generate button
	if m.fieldIndex == 0 {
		lines = append(lines, "  "+buttonActiveStyle.Render("▸ Generate Screenshot"))
	} else {
		lines = append(lines, "  "+buttonStyle.Render("  Generate Screenshot"))
	}

	return strings.Join(lines, "\n")
}

func renderReadOnlyField(label, value string) string {
	l := labelStyle.Render(label + ":")
	v := valueStyle.Render(value)
	return l + " " + v
}

func (m model) viewInput() string {
	var lines []string

	lines = append(lines, m.renderEditableField("URL", m.url, m.urlEditing, 0))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Enter the URL of the page to capture."))

	return strings.Join(lines, "\n")
}

func (m model) viewOutput() string {
	var lines []string

	dirDisplay := m.dir
	if dirDisplay == "" {
		dirDisplay = lipgloss.NewStyle().Foreground(dimColor).Render("(current directory)")
	}
	lines = append(lines, m.renderEditableField("Directory", m.dir, m.dirEditing, 0))
	lines = append(lines, m.renderEditableField("Filename", m.filename, m.filenameEditing, 1))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Press Enter to edit, Esc to stop editing."))

	return strings.Join(lines, "\n")
}

func (m model) viewSettings() string {
	var lines []string

	// Preset row
	presetVal := m.renderPresetField()
	if m.fieldIndex == 0 {
		lines = append(lines, cursorStyle.Render("▸ ")+labelStyle.Render("Preset:")+
			" "+presetVal)
	} else {
		lines = append(lines, "  "+labelStyle.Render("Preset:")+" "+presetVal)
	}

	// Zoom
	lines = append(lines, m.renderEditableField("Zoom", m.zoom, m.zoomEditing, 1))

	// Scroll
	lines = append(lines, m.renderEditableField("Scroll", m.scroll, m.scrollEditing, 2))

	// Custom width/height
	if m.currentPreset() == "custom" {
		lines = append(lines, m.renderEditableField("Width", m.customWidth, m.customWidthEditing, 3))
		lines = append(lines, m.renderEditableField("Height", m.customHeight, m.customHeightEditing, 4))
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Press Enter to edit fields or open preset picker."))

	return strings.Join(lines, "\n")
}

func (m model) renderEditableField(label, value string, editing bool, idx int) string {
	cursor := "  "
	if m.fieldIndex == idx && !m.isEditing() && !m.presetPicking {
		cursor = cursorStyle.Render("▸ ")
	}

	l := labelStyle.Render(label + ":")

	if editing {
		displayed := value + "█"
		v := editingValueStyle.Render(displayed)
		return cursor + l + " " + v
	}

	displayVal := value
	if displayVal == "" && label == "Directory" {
		displayVal = lipgloss.NewStyle().Foreground(dimColor).Render("(current directory)")
	} else if displayVal == "" {
		displayVal = lipgloss.NewStyle().Foreground(dimColor).Render("(empty)")
	}

	if m.fieldIndex == idx {
		v := activeFieldStyle.Render(displayVal)
		return cursor + l + " " + v
	}

	v := valueStyle.Render(displayVal)
	return cursor + l + " " + v
}

func (m model) renderPresetField() string {
	name := m.currentPreset()

	if m.presetPicking {
		var parts []string
		for i, n := range m.presetNames {
			if i == m.presetIndex {
				parts = append(parts, lipgloss.NewStyle().
					Foreground(brightColor).
					Background(primaryColor).
					Padding(0, 1).
					Bold(true).
					Render(n))
			} else {
				parts = append(parts, lipgloss.NewStyle().
					Foreground(dimColor).
					Padding(0, 1).
					Render(n))
			}
		}
		return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
	}

	// Not picking: show current with arrows
	arrow := lipgloss.NewStyle().Foreground(accentColor).Render(" ◂▸")
	if m.fieldIndex == 0 {
		return activeFieldStyle.Render(name) + arrow
	}
	return valueStyle.Render(name) + arrow
}

// ── Run ────────────────────────────────────────────────────────────────────

// Run starts the interactive TUI with the given chrome path.
func Run(chromePath string) error {
	defaults := config.DefaultConfig()

	presetNames := append(config.PresetNames(), "custom")

	m := model{
		tabs:      []string{"Generate", "Input", "Output", "Settings"},
		activeTab: 1, // Start on Input tab
		presetNames:  presetNames,
		presetIndex:  0, // widescreen
		url:          "",
		dir:          defaults.Dir,
		filename:     defaults.Filename,
		zoom:         fmt.Sprintf("%g", defaults.Zoom),
		scroll:       strconv.Itoa(defaults.Scroll),
		customWidth:  strconv.Itoa(defaults.Width),
		customHeight: strconv.Itoa(defaults.Height),
		chromePath:   chromePath,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

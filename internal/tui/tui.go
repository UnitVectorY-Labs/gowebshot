package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/UnitVectorY-Labs/gowebshot/internal/capture"
	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
	"github.com/UnitVectorY-Labs/gowebshot/internal/output"
)

type captureResultMsg struct {
	path string
	size int64
	err  error
}

type fieldID int

const (
	fieldNone fieldID = iota
	fieldURL
	fieldDir
	fieldFilename
	fieldWidth
	fieldHeight
	fieldZoom
	fieldScroll
	fieldDelay
	fieldCropTop
	fieldCropBottom
	fieldCropLeft
	fieldCropRight
	fieldShift
)

const delayStep = 100 * time.Millisecond

type model struct {
	tabs      []string
	activeTab int

	fieldIndex int

	url      string
	dir      string
	filename string

	presetIndex   int
	presetNames   []string
	presetPicking bool

	customWidth  string
	customHeight string
	zoomPercent  string
	scroll       string
	delay        string
	cropTop      string
	cropBottom   string
	cropLeft     string
	cropRight    string
	shift        bool

	editingField fieldID
	fieldCursors map[fieldID]int

	chromePath string

	capturing      bool
	message        string
	messageIsError bool

	width  int
	height int

	escOnce bool
}

var (
	primaryColor  = lipgloss.Color("#F28C28")
	accentColor   = lipgloss.Color("#FFB869")
	dimColor      = lipgloss.Color("#7E95AD")
	textColor     = lipgloss.Color("#E8EEF5")
	brightColor   = lipgloss.Color("#FFF7ED")
	successColor  = lipgloss.Color("#67E8B2")
	errorColor    = lipgloss.Color("#FF8C82")
	bgColor       = lipgloss.Color("#081B2E")
	fieldBgColor  = lipgloss.Color("#102A44")
	activeBgColor = lipgloss.Color("#173A5E")

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(bgColor).
			Background(primaryColor).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Background(fieldBgColor).
				Padding(0, 2)

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(textColor)

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

	successMsgStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Background(bgColor).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)
)

func (m model) currentPreset() string {
	if m.presetIndex >= 0 && m.presetIndex < len(m.presetNames) {
		return m.presetNames[m.presetIndex]
	}
	return string(config.PresetWidescreen)
}

func presetDimensions(name string) (int, int, bool) {
	dims, ok := config.Presets[config.Preset(name)]
	if !ok {
		return 0, 0, false
	}
	return dims.Width, dims.Height, true
}

func parseIntOrDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func parseDurationOrDefault(value string, fallback time.Duration) time.Duration {
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func formatDurationValue(d time.Duration) string {
	return d.String()
}

func (m model) resolutionWidth() int {
	if width := parseIntOrDefault(m.customWidth, 0); width > 0 {
		return width
	}
	if width, _, ok := presetDimensions(m.currentPreset()); ok {
		return width
	}
	return config.DefaultConfig().Width
}

func (m model) resolutionHeight() int {
	if height := parseIntOrDefault(m.customHeight, 0); height > 0 {
		return height
	}
	if _, height, ok := presetDimensions(m.currentPreset()); ok {
		return height
	}
	return config.DefaultConfig().Height
}

func (m model) fieldCountForTab() int {
	switch m.activeTab {
	case 0:
		return 1
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 11
	default:
		return 0
	}
}

func (m model) buildConfig() config.Config {
	defaults := config.DefaultConfig()

	zoomPercent := parseIntOrDefault(m.zoomPercent, int(math.Round(defaults.Zoom*100)))
	if zoomPercent < 1 {
		zoomPercent = 1
	}

	scroll := parseIntOrDefault(m.scroll, defaults.Scroll)
	if scroll < 0 {
		scroll = 0
	}

	delay := parseDurationOrDefault(m.delay, defaults.Delay)
	if delay < 0 {
		delay = 0
	}

	cropTop := parseIntOrDefault(m.cropTop, defaults.Crop.Top)
	if cropTop < 0 {
		cropTop = 0
	}
	cropBottom := parseIntOrDefault(m.cropBottom, defaults.Crop.Bottom)
	if cropBottom < 0 {
		cropBottom = 0
	}
	cropLeft := parseIntOrDefault(m.cropLeft, defaults.Crop.Left)
	if cropLeft < 0 {
		cropLeft = 0
	}
	cropRight := parseIntOrDefault(m.cropRight, defaults.Crop.Right)
	if cropRight < 0 {
		cropRight = 0
	}

	cfg := config.Config{
		URL:      m.url,
		Dir:      m.dir,
		Filename: m.filename,
		Preset:   config.Preset(m.currentPreset()),
		Width:    m.resolutionWidth(),
		Height:   m.resolutionHeight(),
		Zoom:     float64(zoomPercent) / 100,
		Scroll:   scroll,
		Crop: config.Crop{
			Top:    cropTop,
			Bottom: cropBottom,
			Left:   cropLeft,
			Right:  cropRight,
		},
		Shift:      m.shift,
		Delay:      delay,
		ChromePath: m.chromePath,
	}

	return cfg
}

func (m *model) stopAllEditing() {
	m.editingField = fieldNone
	m.presetPicking = false
}

func (m model) isEditing() bool {
	return m.editingField != fieldNone
}

func isNumericField(field fieldID) bool {
	switch field {
	case fieldWidth, fieldHeight, fieldZoom, fieldScroll, fieldDelay:
		return true
	case fieldCropTop, fieldCropBottom, fieldCropLeft, fieldCropRight:
		return true
	default:
		return false
	}
}

func (m model) Init() tea.Cmd {
	return nil
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
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.capturing {
			return m, nil
		}
		if msg.Type != tea.KeyEsc {
			m.escOnce = false
		}
		if m.isEditing() {
			return m.handleEditing(msg)
		}
		if m.presetPicking {
			return m.handlePresetPicker(msg)
		}
		return m.handleNormal(msg)
	}

	return m, nil
}

func (m model) handleEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		m.stopAllEditing()
		return m, nil
	case tea.KeyBackspace:
		m.editBackspace()
		return m, nil
	case tea.KeyDelete:
		m.editDelete()
		return m, nil
	case tea.KeyLeft:
		m.moveCursor(-1)
		return m, nil
	case tea.KeyRight:
		m.moveCursor(1)
		return m, nil
	case tea.KeyHome:
		m.moveCursorToStart()
		return m, nil
	case tea.KeyEnd:
		m.moveCursorToEnd()
		return m, nil
	case tea.KeyUp:
		m.adjustEditingValue(1)
		return m, nil
	case tea.KeyDown:
		m.adjustEditingValue(-1)
		return m, nil
	default:
		if msg.Type == tea.KeySpace {
			m.editInsert(" ")
			return m, nil
		}
		if len(msg.Runes) > 0 {
			m.editInsert(string(msg.Runes))
		}
		return m, nil
	}
}

func (m *model) handlePresetPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.presetPicking = false
	case tea.KeyLeft, tea.KeyUp:
		if m.presetIndex > 0 {
			m.presetIndex--
		} else {
			m.presetIndex = len(m.presetNames) - 1
		}
		m.syncPresetDimensions()
	case tea.KeyRight, tea.KeyDown:
		m.presetIndex = (m.presetIndex + 1) % len(m.presetNames)
		m.syncPresetDimensions()
	case tea.KeyEnter:
		m.presetPicking = false
	}
	return *m, nil
}

func (m model) handleNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyTab, tea.KeyRight:
		m.activeTab = (m.activeTab + 1) % len(m.tabs)
		m.fieldIndex = 0
	case tea.KeyShiftTab, tea.KeyLeft:
		m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		m.fieldIndex = 0
	case tea.KeyDown:
		if maxIdx := m.fieldCountForTab() - 1; m.fieldIndex < maxIdx {
			m.fieldIndex++
		}
	case tea.KeyUp:
		if m.fieldIndex > 0 {
			m.fieldIndex--
		}
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
	case 0:
		if m.fieldIndex == 0 {
			return m.triggerCapture()
		}
	case 1:
		m.startEditing(fieldURL)
	case 2:
		switch m.fieldIndex {
		case 0:
			m.startEditing(fieldDir)
		case 1:
			m.startEditing(fieldFilename)
		}
	case 3:
		switch m.fieldIndex {
		case 0:
			m.stopAllEditing()
			m.presetPicking = true
		case 1:
			m.startEditing(fieldWidth)
		case 2:
			m.startEditing(fieldHeight)
		case 3:
			m.startEditing(fieldZoom)
		case 4:
			m.startEditing(fieldScroll)
		case 5:
			m.startEditing(fieldDelay)
		case 6:
			m.startEditing(fieldCropTop)
		case 7:
			m.startEditing(fieldCropBottom)
		case 8:
			m.startEditing(fieldCropLeft)
		case 9:
			m.startEditing(fieldCropRight)
		case 10:
			m.shift = !m.shift
		}
	}
	return m, nil
}

func (m model) triggerCapture() (tea.Model, tea.Cmd) {
	cfg := m.buildConfig()
	if err := cfg.Validate(); err != nil {
		errMsg := err.Error()
		m.message = errMsg
		m.messageIsError = true

		switch {
		case strings.Contains(errMsg, "url"):
			m.activeTab = 1
			m.fieldIndex = 0
			m.message = "URL is required"
		case strings.Contains(errMsg, "filename"):
			m.activeTab = 2
			m.fieldIndex = 1
		case strings.Contains(errMsg, "width"):
			m.activeTab = 3
			m.fieldIndex = 1
		case strings.Contains(errMsg, "height"):
			m.activeTab = 3
			m.fieldIndex = 2
		case strings.Contains(errMsg, "zoom"):
			m.activeTab = 3
			m.fieldIndex = 3
		case strings.Contains(errMsg, "scroll"):
			m.activeTab = 3
			m.fieldIndex = 4
		case strings.Contains(errMsg, "delay"):
			m.activeTab = 3
			m.fieldIndex = 5
		case strings.Contains(errMsg, "crop"):
			m.activeTab = 3
			m.fieldIndex = 6
		case strings.Contains(errMsg, "shift"):
			m.activeTab = 3
			m.fieldIndex = 10
		}
		return m, nil
	}

	m.capturing = true
	m.message = "Capturing screenshot..."
	m.messageIsError = false
	return m, doCapture(cfg)
}

func (m model) handleEsc() (tea.Model, tea.Cmd) {
	if m.fieldIndex > 0 {
		m.fieldIndex = 0
		return m, nil
	}
	if m.activeTab != 1 {
		m.activeTab = 1
		m.fieldIndex = 0
		return m, nil
	}
	if m.escOnce {
		return m, tea.Quit
	}
	m.escOnce = true
	m.message = "Press Esc again to quit"
	m.messageIsError = false
	return m, nil
}

func (m *model) startEditing(field fieldID) {
	m.stopAllEditing()
	m.editingField = field
	m.setCursorForField(field, len([]rune(m.valueForField(field))))
}

func (m *model) syncPresetDimensions() {
	if width, height, ok := presetDimensions(m.currentPreset()); ok {
		m.customWidth = strconv.Itoa(width)
		m.customHeight = strconv.Itoa(height)
	}
}

func (m *model) syncPresetFromDimensions() {
	width := parseIntOrDefault(m.customWidth, 0)
	height := parseIntOrDefault(m.customHeight, 0)
	if width <= 0 || height <= 0 {
		m.setPresetByName(string(config.PresetCustom))
		return
	}

	for _, name := range config.PresetNames() {
		presetWidth, presetHeight, ok := presetDimensions(name)
		if ok && presetWidth == width && presetHeight == height {
			m.setPresetByName(name)
			return
		}
	}

	m.setPresetByName(string(config.PresetCustom))
}

func (m *model) setPresetByName(name string) {
	for i, presetName := range m.presetNames {
		if presetName == name {
			m.presetIndex = i
			return
		}
	}
}

func (m model) valueForField(field fieldID) string {
	switch field {
	case fieldURL:
		return m.url
	case fieldDir:
		return m.dir
	case fieldFilename:
		return m.filename
	case fieldWidth:
		return m.customWidth
	case fieldHeight:
		return m.customHeight
	case fieldZoom:
		return m.zoomPercent
	case fieldScroll:
		return m.scroll
	case fieldDelay:
		return m.delay
	case fieldCropTop:
		return m.cropTop
	case fieldCropBottom:
		return m.cropBottom
	case fieldCropLeft:
		return m.cropLeft
	case fieldCropRight:
		return m.cropRight
	default:
		return ""
	}
}

func (m *model) setValueForField(field fieldID, value string) {
	switch field {
	case fieldURL:
		m.url = value
	case fieldDir:
		m.dir = value
	case fieldFilename:
		m.filename = value
	case fieldWidth:
		m.customWidth = value
	case fieldHeight:
		m.customHeight = value
	case fieldZoom:
		m.zoomPercent = value
	case fieldScroll:
		m.scroll = value
	case fieldDelay:
		m.delay = value
	case fieldCropTop:
		m.cropTop = value
	case fieldCropBottom:
		m.cropBottom = value
	case fieldCropLeft:
		m.cropLeft = value
	case fieldCropRight:
		m.cropRight = value
	}
}

func (m model) cursorForField(field fieldID) int {
	return m.fieldCursors[field]
}

func (m *model) setCursorForField(field fieldID, cursor int) {
	valueLen := len([]rune(m.valueForField(field)))
	if cursor < 0 {
		cursor = 0
	}
	if cursor > valueLen {
		cursor = valueLen
	}
	m.fieldCursors[field] = cursor
}

func (m *model) moveCursor(delta int) {
	field := m.editingField
	m.setCursorForField(field, m.cursorForField(field)+delta)
}

func (m *model) moveCursorToStart() {
	m.setCursorForField(m.editingField, 0)
}

func (m *model) moveCursorToEnd() {
	m.setCursorForField(m.editingField, len([]rune(m.valueForField(m.editingField))))
}

func (m *model) editInsert(insert string) {
	field := m.editingField
	if field == fieldNone {
		return
	}

	value := []rune(m.valueForField(field))
	cursor := m.cursorForField(field)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(value) {
		cursor = len(value)
	}

	insertRunes := []rune(insert)
	value = append(value[:cursor], append(insertRunes, value[cursor:]...)...)
	m.setValueForField(field, string(value))
	m.setCursorForField(field, cursor+len(insertRunes))

	if field == fieldWidth || field == fieldHeight {
		m.syncPresetFromDimensions()
	}
}

func (m *model) editBackspace() {
	field := m.editingField
	if field == fieldNone {
		return
	}

	value := []rune(m.valueForField(field))
	cursor := m.cursorForField(field)
	if cursor == 0 || len(value) == 0 {
		return
	}

	value = append(value[:cursor-1], value[cursor:]...)
	m.setValueForField(field, string(value))
	m.setCursorForField(field, cursor-1)

	if field == fieldWidth || field == fieldHeight {
		m.syncPresetFromDimensions()
	}
}

func (m *model) editDelete() {
	field := m.editingField
	if field == fieldNone {
		return
	}

	value := []rune(m.valueForField(field))
	cursor := m.cursorForField(field)
	if cursor >= len(value) {
		return
	}

	value = append(value[:cursor], value[cursor+1:]...)
	m.setValueForField(field, string(value))

	if field == fieldWidth || field == fieldHeight {
		m.syncPresetFromDimensions()
	}
}

func (m *model) adjustEditingValue(step int) {
	field := m.editingField
	if !isNumericField(field) {
		return
	}

	switch field {
	case fieldWidth:
		width := parseIntOrDefault(m.customWidth, m.resolutionWidth())
		width += step
		if width < 1 {
			width = 1
		}
		m.customWidth = strconv.Itoa(width)
		m.syncPresetFromDimensions()
		m.moveCursorToEnd()
	case fieldHeight:
		height := parseIntOrDefault(m.customHeight, m.resolutionHeight())
		height += step
		if height < 1 {
			height = 1
		}
		m.customHeight = strconv.Itoa(height)
		m.syncPresetFromDimensions()
		m.moveCursorToEnd()
	case fieldZoom:
		zoom := parseIntOrDefault(m.zoomPercent, 100)
		zoom += step
		if zoom < 1 {
			zoom = 1
		}
		m.zoomPercent = strconv.Itoa(zoom)
		m.moveCursorToEnd()
	case fieldScroll:
		scroll := parseIntOrDefault(m.scroll, 0)
		scroll += step
		if scroll < 0 {
			scroll = 0
		}
		m.scroll = strconv.Itoa(scroll)
		m.moveCursorToEnd()
	case fieldDelay:
		delay := parseDurationOrDefault(m.delay, config.DefaultConfig().Delay)
		delay += time.Duration(step) * delayStep
		if delay < 0 {
			delay = 0
		}
		m.delay = formatDurationValue(delay)
		m.moveCursorToEnd()
	case fieldCropTop:
		cropTop := parseIntOrDefault(m.cropTop, 0)
		cropTop += step
		if cropTop < 0 {
			cropTop = 0
		}
		m.cropTop = strconv.Itoa(cropTop)
		m.moveCursorToEnd()
	case fieldCropBottom:
		cropBottom := parseIntOrDefault(m.cropBottom, 0)
		cropBottom += step
		if cropBottom < 0 {
			cropBottom = 0
		}
		m.cropBottom = strconv.Itoa(cropBottom)
		m.moveCursorToEnd()
	case fieldCropLeft:
		cropLeft := parseIntOrDefault(m.cropLeft, 0)
		cropLeft += step
		if cropLeft < 0 {
			cropLeft = 0
		}
		m.cropLeft = strconv.Itoa(cropLeft)
		m.moveCursorToEnd()
	case fieldCropRight:
		cropRight := parseIntOrDefault(m.cropRight, 0)
		cropRight += step
		if cropRight < 0 {
			cropRight = 0
		}
		m.cropRight = strconv.Itoa(cropRight)
		m.moveCursorToEnd()
	}
}

func (m model) View() string {
	if m.width < 60 || m.height < 16 {
		msg := lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Render("Terminal too small. Resize to at least 60x16.")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("  gowebshot"))
	b.WriteString("\n\n")
	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

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

	b.WriteString(sectionStyle.Width(contentWidth).Render(contentStyle.Render(content)))
	b.WriteString("\n\n")

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

	barWidth := m.width - 4
	if barWidth < 50 {
		barWidth = 50
	}

	line := lipgloss.NewStyle().
		Foreground(primaryColor).
		Render(strings.Repeat("─", barWidth))

	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...) + "\n" + line
}

func (m model) renderHelp() string {
	if m.capturing {
		return helpStyle.Render("  Ctrl+C quit")
	}
	if m.isEditing() {
		text := "  Enter confirm • Esc cancel • ←/→ move cursor • Backspace/Delete edit"
		if isNumericField(m.editingField) {
			text += " • ↑/↓ adjust value"
		}
		return helpStyle.Render(text)
	}
	if m.presetPicking {
		return helpStyle.Render("  ←/→/↑/↓ cycle preset • Enter confirm • Esc cancel")
	}
	return helpStyle.Render("  ←/→/Tab switch tabs • ↑/↓ move fields • Enter select • Esc back • Ctrl+C quit")
}

func (m model) viewGenerate() string {
	var lines []string

	urlVal := m.url
	if strings.TrimSpace(urlVal) == "" {
		urlVal = lipgloss.NewStyle().Foreground(dimColor).Render("(not set)")
	}

	dirVal := m.dir
	if strings.TrimSpace(dirVal) == "" {
		dirVal = lipgloss.NewStyle().Foreground(dimColor).Render("(current directory)")
	}

	lines = append(lines,
		renderReadOnlyField("URL", urlVal),
		renderReadOnlyField("Directory", dirVal),
		renderReadOnlyField("Filename", m.filename),
		renderReadOnlyField("Preset", m.presetLabel(m.currentPreset())),
		renderReadOnlyField("Width", fmt.Sprintf("%d px", m.resolutionWidth())),
		renderReadOnlyField("Height", fmt.Sprintf("%d px", m.resolutionHeight())),
		renderReadOnlyField("Zoom", m.zoomPercent+"%"),
		renderReadOnlyField("Scroll", m.scroll+"px"),
		renderReadOnlyField("Crop", m.cropSummary()),
		renderReadOnlyField("Shift", m.enabledLabel(m.shift)),
		renderReadOnlyField("Delay", m.delay),
		"",
	)

	if m.fieldIndex == 0 {
		lines = append(lines, "  "+buttonActiveStyle.Render("▸ Generate Screenshot"))
	} else {
		lines = append(lines, "  "+buttonStyle.Render("  Generate Screenshot"))
	}

	return strings.Join(lines, "\n")
}

func renderReadOnlyField(label, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

func (m model) viewInput() string {
	return strings.Join([]string{
		m.renderEditableField("URL", m.url, m.editingField == fieldURL, 0),
		"",
		helpStyle.Render("  Enter the URL of the page to capture."),
	}, "\n")
}

func (m model) viewOutput() string {
	return strings.Join([]string{
		m.renderEditableField("Directory", m.dir, m.editingField == fieldDir, 0),
		m.renderEditableField("Filename", m.filename, m.editingField == fieldFilename, 1),
		"",
		helpStyle.Render("  Use ← and → while editing to move through long paths."),
	}, "\n")
}

func (m model) viewSettings() string {
	var lines []string

	presetValue := m.renderPresetField()
	if m.fieldIndex == 0 && !m.isEditing() && !m.presetPicking {
		lines = append(lines, cursorStyle.Render("▸ ")+labelStyle.Render("Preset:")+" "+presetValue)
	} else {
		lines = append(lines, "  "+labelStyle.Render("Preset:")+" "+presetValue)
	}

	lines = append(lines,
		m.renderEditableField("Width", m.customWidth, m.editingField == fieldWidth, 1),
		m.renderEditableField("Height", m.customHeight, m.editingField == fieldHeight, 2),
		m.renderEditableField("Zoom %", m.zoomPercent, m.editingField == fieldZoom, 3),
		m.renderEditableField("Scroll", m.scroll, m.editingField == fieldScroll, 4),
		m.renderEditableField("Delay", m.delay, m.editingField == fieldDelay, 5),
		m.renderEditableField("Crop Top", m.cropTop, m.editingField == fieldCropTop, 6),
		m.renderEditableField("Crop Bottom", m.cropBottom, m.editingField == fieldCropBottom, 7),
		m.renderEditableField("Crop Left", m.cropLeft, m.editingField == fieldCropLeft, 8),
		m.renderEditableField("Crop Right", m.cropRight, m.editingField == fieldCropRight, 9),
		m.renderToggleField("Shift", m.shift, 10),
		"",
		helpStyle.Render("  Selecting a preset loads its dimensions. Editing width or height switches to custom when needed."),
		helpStyle.Render("  Crop values are pixels removed after capture. Enable Shift to keep the requested output size."),
	)

	return strings.Join(lines, "\n")
}

func (m model) renderEditableField(label, value string, editing bool, idx int) string {
	cursor := "  "
	if m.fieldIndex == idx && !m.isEditing() && !m.presetPicking {
		cursor = cursorStyle.Render("▸ ")
	}

	if editing {
		return cursor + labelStyle.Render(label+":") + " " + m.renderEditingValue(m.editingField)
	}

	displayVal := value
	switch {
	case displayVal == "" && label == "Directory":
		displayVal = lipgloss.NewStyle().Foreground(dimColor).Render("(current directory)")
	case displayVal == "":
		displayVal = lipgloss.NewStyle().Foreground(dimColor).Render("(empty)")
	}

	if m.fieldIndex == idx {
		return cursor + labelStyle.Render(label+":") + " " + activeFieldStyle.Render(displayVal)
	}

	return cursor + labelStyle.Render(label+":") + " " + valueStyle.Render(displayVal)
}

func (m model) renderEditingValue(field fieldID) string {
	value := []rune(m.valueForField(field))
	cursor := m.cursorForField(field)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(value) {
		cursor = len(value)
	}

	display := string(value[:cursor]) + cursorStyle.Render("│") + string(value[cursor:])
	return editingValueStyle.Render(display)
}

func (m model) renderToggleField(label string, enabled bool, idx int) string {
	cursor := "  "
	if m.fieldIndex == idx && !m.isEditing() && !m.presetPicking {
		cursor = cursorStyle.Render("▸ ")
	}

	value := m.enabledLabel(enabled)
	if m.fieldIndex == idx {
		return cursor + labelStyle.Render(label+":") + " " + activeFieldStyle.Render(value)
	}

	return cursor + labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

func (m model) enabledLabel(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}

func (m model) cropSummary() string {
	return fmt.Sprintf("top %spx, bottom %spx, left %spx, right %spx", m.cropTop, m.cropBottom, m.cropLeft, m.cropRight)
}

func (m model) renderPresetField() string {
	if m.presetPicking {
		var parts []string
		for i, name := range m.presetNames {
			label := m.presetLabel(name)
			style := lipgloss.NewStyle().Foreground(dimColor).Padding(0, 1)
			if i == m.presetIndex {
				style = style.Foreground(brightColor).Background(primaryColor).Bold(true)
			}
			parts = append(parts, style.Render(label))
		}
		return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
	}

	arrow := lipgloss.NewStyle().Foreground(accentColor).Render(" ◂▸")
	label := m.presetLabel(m.currentPreset())
	if m.fieldIndex == 0 {
		return activeFieldStyle.Render(label) + arrow
	}
	return valueStyle.Render(label) + arrow
}

func (m model) presetLabel(name string) string {
	if name == string(config.PresetCustom) {
		return fmt.Sprintf("custom (%dx%d)", m.resolutionWidth(), m.resolutionHeight())
	}
	width, height, ok := presetDimensions(name)
	if !ok {
		return name
	}
	return fmt.Sprintf("%s (%dx%d)", name, width, height)
}

func newModel(initial config.Config) model {
	presetNames := append(config.PresetNames(), string(config.PresetCustom))

	m := model{
		tabs:         []string{"Generate", "Input", "Output", "Settings"},
		activeTab:    1,
		presetIndex:  0,
		presetNames:  presetNames,
		fieldCursors: make(map[fieldID]int),
	}

	m.url = initial.URL
	m.dir = initial.Dir
	m.filename = initial.Filename
	m.customWidth = strconv.Itoa(initial.Width)
	m.customHeight = strconv.Itoa(initial.Height)
	m.zoomPercent = strconv.Itoa(int(math.Round(initial.Zoom * 100)))
	m.scroll = strconv.Itoa(initial.Scroll)
	m.delay = formatDurationValue(initial.Delay)
	m.cropTop = strconv.Itoa(initial.Crop.Top)
	m.cropBottom = strconv.Itoa(initial.Crop.Bottom)
	m.cropLeft = strconv.Itoa(initial.Crop.Left)
	m.cropRight = strconv.Itoa(initial.Crop.Right)
	m.shift = initial.Shift
	m.chromePath = initial.ChromePath

	m.setPresetByName(string(initial.Preset))
	if initial.Preset == "" || m.currentPreset() != string(initial.Preset) {
		m.syncPresetFromDimensions()
	}

	return m
}

func Run(initial config.Config) error {
	m := newModel(initial)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

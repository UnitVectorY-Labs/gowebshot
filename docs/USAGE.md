---
layout: default
title: Usage
nav_order: 3
permalink: /usage
---

# Usage
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Modes

gowebshot has two execution modes:

- **Non-interactive mode** — Runs when any CLI flags are provided. Takes a single screenshot and exits.
- **Interactive TUI mode** — Runs when no arguments are given. Opens a terminal UI for repeated captures.

## Non-Interactive Mode

### Basic Usage

```bash
gowebshot --url https://example.com
```

This captures a screenshot of the specified URL at the default resolution (1920×1080) and saves it as `screenshot.png` in the current directory.

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--url` | URL to capture (required) | — |
| `--dir` | Output directory | Current directory |
| `--filename` | Output filename | `screenshot.png` |
| `--preset` | Resolution preset | `widescreen` |
| `--width` | Custom viewport width (requires `--height`) | — |
| `--height` | Custom viewport height (requires `--width`) | — |
| `--zoom` | Page zoom factor | `1.0` |
| `--scroll` | Vertical scroll in pixels | `0` |
| `--chrome` | Explicit path to Chrome/Chromium | Auto-discover |

### Resolution Presets

| Preset | Width | Height |
|--------|-------|--------|
| `widescreen` | 1920 | 1080 |
| `desktop` | 1440 | 900 |
| `square` | 1200 | 1200 |
| `portrait` | 1080 | 1350 |

### Examples

Capture at square resolution:

```bash
gowebshot --url https://example.com --preset square
```

Capture with custom dimensions:

```bash
gowebshot --url https://example.com --width 800 --height 600
```

Capture with zoom and scroll:

```bash
gowebshot --url https://example.com --zoom 1.5 --scroll 200
```

Save to a specific directory and filename:

```bash
gowebshot --url https://example.com --dir ./screenshots --filename homepage
```

Note: If the filename does not have an extension, `.png` is appended automatically.

### File Naming

If a file with the target name already exists, gowebshot automatically appends a numeric suffix to avoid overwriting:

- `screenshot.png` → `screenshot2.png` → `screenshot3.png` → ...

## Interactive TUI Mode

Launch without arguments to start the interactive mode:

```bash
gowebshot
```

### Tabs

The TUI provides four tabs:

- **Generate** — Shows a summary of the current configuration and a button to trigger capture.
- **Input** — Edit the URL to capture.
- **Output** — Edit the output directory and filename.
- **Settings** — Configure resolution preset, zoom, and scroll. Custom width/height fields appear when the "custom" preset is selected.

### Keyboard Controls

| Key | Action |
|-----|--------|
| `←`/`→` or `Tab`/`Shift+Tab` | Switch between tabs |
| `↑`/`↓` | Move between fields |
| `Enter` | Edit focused field or trigger action |
| `Space` | Trigger generate on Generate tab |
| `Esc` | Back (hierarchical: exit edit → close picker → move to top → return to Input → quit) |
| `Ctrl+C` | Quit immediately |
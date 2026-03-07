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

- **Non-interactive mode** — Runs when `--url` is provided. Takes a single screenshot and exits using defaults for any omitted flags.
- **Interactive TUI mode** — Runs when `--url` is omitted. Any other CLI flags pre-populate the TUI fields before it opens. See [TUI](/tui) for the interactive workflow.

## Non-Interactive Mode

### Basic Usage

```bash
gowebshot --url https://example.com
```

This captures a screenshot of the specified URL at the default resolution (1920×1080) and saves it as `screenshot.png` in the current directory.

If you omit `--url`, gowebshot opens the TUI instead. For example, this starts interactive mode with the square preset already selected:

```bash
gowebshot --preset square --delay 1500ms
```

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--url` | URL to capture. Its presence switches gowebshot into non-interactive mode. | — |
| `--dir` | Output directory | Current directory |
| `--filename` | Output filename | `screenshot.png` |
| `--preset` | Resolution preset | `widescreen` |
| `--width` | Custom viewport width (requires `--height`) | — |
| `--height` | Custom viewport height (requires `--width`) | — |
| `--zoom` | Page zoom factor | `1.0` |
| `--scroll` | Vertical scroll in pixels | `0` |
| `--delay` | Delay after page load before capture | `1s` |
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

Capture after letting the page settle for 2 seconds:

```bash
gowebshot --url https://example.com --delay 2s
```

Save to a specific directory and filename:

```bash
gowebshot --url https://example.com --dir ./screenshots --filename homepage
```

Note: If the filename does not have an extension, `.png` is appended automatically.

### File Naming

If a file with the target name already exists, gowebshot automatically appends a numeric suffix to avoid overwriting:

- `screenshot.png` → `screenshot2.png` → `screenshot3.png` → ...

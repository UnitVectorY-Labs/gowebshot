---
layout: default
title: TUI
nav_order: 4
permalink: /tui
---

# TUI
{: .no_toc }

## Launch

Run `gowebshot` without `--url` to open the interactive terminal UI:

```bash
gowebshot
```

Any other CLI flags are applied before the TUI opens. For example:

```bash
gowebshot --preset portrait --filename social-card
```

## Tabs

The TUI provides four tabs:

- **Generate** — Shows a summary of the current configuration and a button to trigger capture.
- **Input** — Edit the URL to capture.
- **Output** — Edit the output directory and filename.
- **Settings** — Choose a preset, edit width and height, set zoom percentage, scroll offset, crop each edge, optionally enable shift, and set the capture delay.

## Settings

- **Preset picker** — Cycle through preset names and their dimensions before confirming a selection.
- **Width / Height** — Always editable. Selecting a preset loads its dimensions; changing them manually switches the preset to `custom` unless they match a named preset exactly.
- **Zoom %** — Stored as a percentage in the TUI. While editing, `↑` and `↓` adjust it by 1%.
- **Scroll** — Stored in pixels. While editing, `↑` and `↓` adjust it by 1px and never allow negative values.
- **Crop Top / Bottom / Left / Right** — Pixels to trim from each edge after the screenshot is taken. While editing, `↑` and `↓` adjust them by 1px and never allow negative values.
- **Shift** — When enabled, gowebshot captures a larger viewport first so the final cropped PNG keeps the requested width and height.
- **Delay** — Wait time after page load and adjustments before capture. Default is `1s`.

## Keyboard Controls

| Key | Action |
|-----|--------|
| `←`/`→` or `Tab`/`Shift+Tab` | Switch between tabs |
| `↑`/`↓` | Move between fields |
| `Enter` | Edit focused field or trigger action |
| `Space` | Trigger generate on Generate tab |
| `Esc` | Back (exit edit or picker, move to the first field, return to Input, then quit on second press) |
| `Ctrl+C` | Quit immediately |

While editing a field:

| Key | Action |
|-----|--------|
| `←`/`→` | Move the cursor within the field |
| `Home` / `End` | Jump to the start or end of the field |
| `Backspace` / `Delete` | Remove characters around the cursor |
| `↑`/`↓` | Adjust numeric settings such as zoom, scroll, and delay |
| `Enter` | Confirm edits |
| `Esc` | Cancel editing |

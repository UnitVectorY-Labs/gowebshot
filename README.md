[![GitHub release](https://img.shields.io/github/release/UnitVectorY-Labs/gowebshot.svg)](https://github.com/UnitVectorY-Labs/gowebshot/releases/latest) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Active](https://img.shields.io/badge/Status-Active-green)](https://guide.unitvectorylabs.com/bestpractices/status/#active)
 [![Go Report Card](https://goreportcard.com/badge/github.com/UnitVectorY-Labs/gowebshot)](https://goreportcard.com/report/github.com/UnitVectorY-Labs/gowebshot)
 
# gowebshot

Simple command line application for capturing screen shots of webpages.

[![gowebshot CLI](./docs/cli.png)](cli.png)

## Features

- **Non-interactive mode** — Capture a screenshot with a single command using CLI flags.
- **Interactive TUI mode** — Configure and capture screenshots interactively with a keyboard-driven interface.
- **Preset resolutions** — Choose from widescreen (1920×1080), desktop (1440×900), square (1200×1200), or portrait (1080×1350).
- **Custom viewport** — Edit width and height directly, or start from a preset.
- **Zoom, scroll, and delay** — Apply page zoom, vertical scroll, and a configurable post-load delay before capture.
- **Auto-naming** — Automatically appends numeric suffixes to prevent overwriting existing files.
- **Chrome auto-discovery** — Finds Chrome/Chromium automatically, or accepts an explicit path.

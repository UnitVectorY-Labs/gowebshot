# gowebshot Agent Notes

This is a Go based library first with a companion command line application that follows idiomatic Go conventions. The purpose is to provide a method of validating a JSON Schema is strictly utilizing a subset of the JSON Schema specification as limited by various specific use cases, commonly utilized by LLMs when generating structured JSON output.

## Repo Conventions For The Agent
- Treat `docs/` as required source of truth alongside code. When behavior changes, update the matching doc pages in the same PR.
- Keep dependencies minimal and prefer stdlib. Add third-party packages only when they materially improve correctness or output quality.

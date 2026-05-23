package main

import (
	"fmt"
	"runtime"
	"testing"
)

func TestBuildVersionOutput_SemverWithoutV(t *testing.T) {
	got := buildVersionOutput("gowebshot", "1.2.3")
	want := fmt.Sprintf("gowebshot version v1.2.3 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

func TestBuildVersionOutput_SemverWithV(t *testing.T) {
	got := buildVersionOutput("gowebshot", "v1.2.3")
	want := fmt.Sprintf("gowebshot version v1.2.3 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

func TestBuildVersionOutput_DevVersion(t *testing.T) {
	got := buildVersionOutput("gowebshot", "dev")
	want := fmt.Sprintf("gowebshot version dev (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

func TestBuildVersionOutput_PrereleaseVersion(t *testing.T) {
	got := buildVersionOutput("gowebshot", "1.2.3-beta.1")
	want := fmt.Sprintf("gowebshot version v1.2.3-beta.1 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

func TestBuildVersionOutput_BuildMetadata(t *testing.T) {
	got := buildVersionOutput("gowebshot", "1.2.3+build.1")
	want := fmt.Sprintf("gowebshot version v1.2.3+build.1 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

func TestBuildVersionOutput_EmptyVersion(t *testing.T) {
	got := buildVersionOutput("gowebshot", "")
	want := fmt.Sprintf("gowebshot version  (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

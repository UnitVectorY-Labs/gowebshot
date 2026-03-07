package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePath_EmptyDirUsesCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	got, err := ResolvePath("", "shot")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Dir(got) != cwd {
		t.Errorf("expected dir %s, got %s", cwd, filepath.Dir(got))
	}
}

func TestResolvePath_CreatesMissingDirectories(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")

	_, err := ResolvePath(dir, "shot.png")
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("path is not a directory")
	}
}

func TestResolvePath_AppendsPngWhenNoExtension(t *testing.T) {
	dir := t.TempDir()

	got, err := ResolvePath(dir, "screenshot")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Ext(got) != ".png" {
		t.Errorf("expected .png extension, got %s", filepath.Ext(got))
	}
	if filepath.Base(got) != "screenshot.png" {
		t.Errorf("expected screenshot.png, got %s", filepath.Base(got))
	}
}

func TestResolvePath_RejectsNonPngExtensions(t *testing.T) {
	dir := t.TempDir()

	_, err := ResolvePath(dir, "shot.jpg")
	if err == nil {
		t.Fatal("expected error for non-.png extension")
	}
	if !strings.Contains(err.Error(), "output file must have .png extension") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolvePath_NumericSuffixing(t *testing.T) {
	dir := t.TempDir()

	// Create the initial file so the next resolve must pick a suffix.
	if err := os.WriteFile(filepath.Join(dir, "shot.png"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ResolvePath(dir, "shot.png")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(got) != "shot2.png" {
		t.Errorf("expected shot2.png, got %s", filepath.Base(got))
	}

	// Create shot2.png and resolve again to get shot3.png.
	if err := os.WriteFile(got, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	got, err = ResolvePath(dir, "shot.png")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(got) != "shot3.png" {
		t.Errorf("expected shot3.png, got %s", filepath.Base(got))
	}
}

func TestWriteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.png")
	data := []byte("fake-png-data")

	if err := WriteFile(path, data); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Errorf("expected %q, got %q", data, got)
	}
}

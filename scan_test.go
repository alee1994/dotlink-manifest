package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanSymlinksFindsLinks(t *testing.T) {
	dir := t.TempDir()

	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "link")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Fatal(err)
	}

	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(sub, "nested-link")
	if err := os.Symlink("../target.txt", nested); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := ScanSymlinks(dir, &out, JSONL); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, `"link":"`+link+`"`) {
		t.Errorf("missing top-level link entry, got:\n%s", got)
	}
	if !strings.Contains(got, `"target":"target.txt"`) {
		t.Errorf("missing top-level target, got:\n%s", got)
	}
	if !strings.Contains(got, `"link":"`+nested+`"`) {
		t.Errorf("missing nested link entry, got:\n%s", got)
	}
	if !strings.Contains(got, `"target":"../target.txt"`) {
		t.Errorf("missing nested target, got:\n%s", got)
	}
}

func TestScanSymlinksSkipsRegularFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plain.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := ScanSymlinks(dir, &out, JSONL); err != nil {
		t.Fatal(err)
	}
	if out.String() != "" {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestScanSymlinksMissingRoot(t *testing.T) {
	var out strings.Builder
	if err := ScanSymlinks(filepath.Join(t.TempDir(), "does-not-exist"), &out, JSONL); err == nil {
		t.Fatal("expected an error for a missing root directory")
	}
}

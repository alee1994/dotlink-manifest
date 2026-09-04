package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckManifestPassesWhenTargetsExist(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "vimrc")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "sub", ".vimrc")
	in := link + " -> ../vimrc\n"

	var out strings.Builder
	if err := CheckManifest(strings.NewReader(in), &out, Arrow); err != nil {
		t.Fatalf("expected no error for an existing target, got: %v", err)
	}
	if out.String() != "" {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestCheckManifestReportsDanglingLink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, ".vimrc")
	in := link + " -> vim/vimrc\n"

	var out strings.Builder
	err := CheckManifest(strings.NewReader(in), &out, Arrow)
	if err == nil {
		t.Fatal("expected an error for a dangling link")
	}
	if !strings.Contains(out.String(), link+" -> vim/vimrc") {
		t.Errorf("expected dangling entry in output, got %q", out.String())
	}
}

func TestCheckManifestResolvesTargetRelativeToLink(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "vimrc")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(sub, ".vimrc")
	in := link + " -> ../vimrc\n"

	var out strings.Builder
	if err := CheckManifest(strings.NewReader(in), &out, Arrow); err != nil {
		t.Fatalf("target is relative to the link's directory, not cwd: %v", err)
	}
}

func TestCheckManifestCountsMultipleDangling(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "a") + " -> missing-a\n" +
		filepath.Join(dir, "b") + " -> missing-b\n"

	var out strings.Builder
	err := CheckManifest(strings.NewReader(in), &out, Arrow)
	if err == nil {
		t.Fatal("expected an error for two dangling links")
	}
	if got := strings.Count(out.String(), "\n"); got != 2 {
		t.Errorf("got %d reported lines, want 2", got)
	}
}

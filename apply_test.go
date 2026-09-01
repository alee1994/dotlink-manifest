package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyManifestCreatesSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "sub", "vimrc")

	in := link + " -> ../vim/vimrc\n"
	if err := ApplyManifest(strings.NewReader(in), Arrow, false); err != nil {
		t.Fatal(err)
	}

	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("expected %s to be a symlink: %v", link, err)
	}
	if target != "../vim/vimrc" {
		t.Errorf("got target %q, want %q", target, "../vim/vimrc")
	}
}

func TestApplyManifestIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "vimrc")
	in := link + " -> vim/vimrc\n"

	if err := ApplyManifest(strings.NewReader(in), Arrow, false); err != nil {
		t.Fatal(err)
	}
	if err := ApplyManifest(strings.NewReader(in), Arrow, false); err != nil {
		t.Fatalf("second apply of an already-correct manifest should succeed, got: %v", err)
	}
}

func TestApplyManifestRefusesToClobberWithoutForce(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "vimrc")
	if err := os.WriteFile(link, []byte("real file, not a link"), 0o644); err != nil {
		t.Fatal(err)
	}

	in := link + " -> vim/vimrc\n"
	if err := ApplyManifest(strings.NewReader(in), Arrow, false); err == nil {
		t.Fatal("expected an error when Link already exists and isn't a symlink")
	}

	if err := ApplyManifest(strings.NewReader(in), Arrow, true); err != nil {
		t.Fatalf("expected -force to replace the existing file, got: %v", err)
	}
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("expected %s to be a symlink after -force: %v", link, err)
	}
	if target != "vim/vimrc" {
		t.Errorf("got target %q, want %q", target, "vim/vimrc")
	}
}

func TestApplyManifestRefusesToRepointWithoutForce(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "vimrc")
	if err := os.Symlink("old/vimrc", link); err != nil {
		t.Fatal(err)
	}

	in := link + " -> new/vimrc\n"
	if err := ApplyManifest(strings.NewReader(in), Arrow, false); err == nil {
		t.Fatal("expected an error when Link already points somewhere else")
	}

	if err := ApplyManifest(strings.NewReader(in), Arrow, true); err != nil {
		t.Fatalf("expected -force to repoint the existing symlink, got: %v", err)
	}
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if target != "new/vimrc" {
		t.Errorf("got target %q, want %q", target, "new/vimrc")
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory available")
	}

	got, err := expandHome("~/.vimrc")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".vimrc")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got, err := expandHome("/abs/path"); err != nil || got != "/abs/path" {
		t.Errorf("expandHome should leave an absolute path unchanged, got %q, err %v", got, err)
	}
}

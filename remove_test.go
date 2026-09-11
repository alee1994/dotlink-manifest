package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveManifestRemovesMatchingSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "vimrc")
	if err := os.Symlink("vim/vimrc", link); err != nil {
		t.Fatal(err)
	}

	in := link + " -> vim/vimrc\n"
	if err := RemoveManifest(strings.NewReader(in), Arrow, false); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Errorf("expected %s to be gone, lstat err: %v", link, err)
	}
}

func TestRemoveManifestIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "vimrc")
	in := link + " -> vim/vimrc\n"

	if err := RemoveManifest(strings.NewReader(in), Arrow, false); err != nil {
		t.Fatalf("removing an already-absent link should succeed, got: %v", err)
	}
}

func TestRemoveManifestRefusesMismatchedTargetWithoutForce(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "vimrc")
	if err := os.Symlink("other/vimrc", link); err != nil {
		t.Fatal(err)
	}

	in := link + " -> vim/vimrc\n"
	if err := RemoveManifest(strings.NewReader(in), Arrow, false); err == nil {
		t.Fatal("expected an error when Link points somewhere other than Target")
	}
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("link should still exist after the refused removal: %v", err)
	}

	if err := RemoveManifest(strings.NewReader(in), Arrow, true); err != nil {
		t.Fatalf("expected -force to remove the mismatched symlink, got: %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Errorf("expected %s to be gone after -force, lstat err: %v", link, err)
	}
}

func TestRemoveManifestRefusesNonSymlinkWithoutForce(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "vimrc")
	if err := os.WriteFile(link, []byte("real file, not a link"), 0o644); err != nil {
		t.Fatal(err)
	}

	in := link + " -> vim/vimrc\n"
	if err := RemoveManifest(strings.NewReader(in), Arrow, false); err == nil {
		t.Fatal("expected an error when Link exists and isn't a symlink")
	}

	if err := RemoveManifest(strings.NewReader(in), Arrow, true); err != nil {
		t.Fatalf("expected -force to remove the plain file, got: %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Errorf("expected %s to be gone after -force, lstat err: %v", link, err)
	}
}

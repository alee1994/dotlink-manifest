package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ApplyManifest reads entries in the given format from r and creates the
// symlinks they describe, one at a time. Like ConvertStream, it never
// buffers the whole manifest: each entry is decoded and applied before the
// next line is read.
//
// If force is false, a Link path that already exists and isn't already the
// correct symlink is treated as an error rather than being overwritten. A
// Link that's already a symlink to Target is left alone either way, so
// re-running ApplyManifest on an already-applied manifest is a no-op.
func ApplyManifest(r io.Reader, from Format, force bool) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}

		entry, err := decodeEntry(line, from)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}

		if err := applyEntry(entry, force); err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
	}
	return scanner.Err()
}

func applyEntry(e Entry, force bool) error {
	link, err := expandHome(e.Link)
	if err != nil {
		return err
	}

	if existing, err := os.Readlink(link); err == nil {
		if existing == e.Target {
			return nil
		}
		if !force {
			return fmt.Errorf("%s is already a symlink to %s, not %s (use -force to replace)", link, existing, e.Target)
		}
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("removing existing symlink %s: %w", link, err)
		}
	} else if _, statErr := os.Lstat(link); statErr == nil {
		if !force {
			return fmt.Errorf("%s already exists and isn't a symlink (use -force to replace)", link)
		}
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("removing existing file %s: %w", link, err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return fmt.Errorf("creating parent directory for %s: %w", link, err)
	}

	if err := os.Symlink(e.Target, link); err != nil {
		return fmt.Errorf("creating symlink %s -> %s: %w", link, e.Target, err)
	}
	return nil
}

// expandHome resolves a leading "~" or "~/" in a manifest path against the
// current user's home directory. os.Symlink has no notion of shell-style
// expansion, so a manifest written with "~/.vimrc" needs this before it can
// actually be created.
func expandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving ~: %w", err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// RemoveManifest reads entries in the given format from r and removes the
// symlinks they describe, one at a time, mirroring ApplyManifest.
//
// A Link that's already gone is left alone, so re-running RemoveManifest on
// an already-removed manifest is a no-op. A Link that's a symlink but
// doesn't point at Target, or that exists but isn't a symlink at all, is
// treated as an error rather than being removed - pass force to remove it
// regardless of what it currently is.
func RemoveManifest(r io.Reader, from Format, force bool) error {
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

		if err := removeEntry(entry, force); err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
	}
	return scanner.Err()
}

func removeEntry(e Entry, force bool) error {
	link, err := expandHome(e.Link)
	if err != nil {
		return err
	}

	if existing, err := os.Readlink(link); err == nil {
		if existing != e.Target && !force {
			return fmt.Errorf("%s is a symlink to %s, not %s (use -force to remove it anyway)", link, existing, e.Target)
		}
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("removing symlink %s: %w", link, err)
		}
		return nil
	}

	if _, err := os.Lstat(link); err == nil {
		if !force {
			return fmt.Errorf("%s exists and isn't a symlink (use -force to remove it anyway)", link)
		}
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("removing %s: %w", link, err)
		}
		return nil
	}

	return nil
}

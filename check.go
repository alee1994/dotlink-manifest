package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CheckManifest reads entries in the given format from r and writes one
// "link -> target" line to w for every entry whose target doesn't resolve
// to anything on disk. It returns an error if it found at least one, so a
// script can rely on dlm's exit code rather than parsing its output.
//
// A manifest can be checked whether or not it's been applied yet: this
// only asks whether Target exists, not whether Link is currently a
// symlink pointing at it.
func CheckManifest(r io.Reader, w io.Writer, from Format) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	bw := bufio.NewWriter(w)
	defer bw.Flush()

	lineNum := 0
	dangling := 0
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

		ok, err := targetExists(entry)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
		if !ok {
			dangling++
			if _, err := fmt.Fprintf(bw, "%s%s%s\n", entry.Link, arrowSep, entry.Target); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	if dangling > 0 {
		return fmt.Errorf("%d dangling link(s)", dangling)
	}
	return nil
}

// targetExists reports whether e.Target resolves to something on disk,
// resolving a relative target against the directory containing e.Link
// rather than the process's working directory - the same rule the
// filesystem uses when it actually follows the symlink.
func targetExists(e Entry) (bool, error) {
	link, err := expandHome(e.Link)
	if err != nil {
		return false, err
	}

	target := e.Target
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(link), target)
	}

	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

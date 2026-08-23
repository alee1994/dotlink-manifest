package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ScanSymlinks walks root and writes one entry per symlink it finds to w,
// encoded in the given format. Entries are written as the walk discovers
// them rather than collected first, so a tree with tens of thousands of
// links costs no more memory than a small one.
func ScanSymlinks(root string, w io.Writer, to Format) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink == 0 {
			return nil
		}
		target, err := os.Readlink(path)
		if err != nil {
			return fmt.Errorf("readlink %s: %w", path, err)
		}
		return encodeEntry(bw, Entry{Link: path, Target: target}, to)
	})
	if err != nil {
		return err
	}
	return bw.Flush()
}

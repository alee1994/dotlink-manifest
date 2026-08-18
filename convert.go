package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format identifies one of the two manifest formats dlm speaks.
type Format string

const (
	Arrow Format = "arrow"
	JSONL Format = "jsonl"
)

func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case Arrow, JSONL:
		return Format(s), nil
	default:
		return "", fmt.Errorf("unknown format %q (want arrow or jsonl)", s)
	}
}

// Entry is one symlink: Link points at Target.
type Entry struct {
	Link   string `json:"link"`
	Target string `json:"target"`
}

const arrowSep = " -> "

// ConvertStream reads entries in one format from r and writes them in
// another format to w, one line at a time. It never buffers more than a
// single entry, so it's safe to point at a symlink dump covering an
// entire home directory or filesystem.
func ConvertStream(r io.Reader, w io.Writer, from, to Format) error {
	scanner := bufio.NewScanner(r)
	// Paths are usually well under 64KB, but bump the ceiling well past
	// the default so an unusually long one doesn't just abort the run.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	bw := bufio.NewWriter(w)
	defer bw.Flush()

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

		if err := encodeEntry(bw, entry, to); err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return bw.Flush()
}

func decodeEntry(line string, from Format) (Entry, error) {
	switch from {
	case Arrow:
		i := strings.Index(line, arrowSep)
		if i < 0 {
			return Entry{}, fmt.Errorf("missing %q separator: %s", arrowSep, line)
		}
		return Entry{
			Link:   strings.TrimSpace(line[:i]),
			Target: strings.TrimSpace(line[i+len(arrowSep):]),
		}, nil
	case JSONL:
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return Entry{}, fmt.Errorf("invalid json: %w", err)
		}
		return e, nil
	default:
		return Entry{}, fmt.Errorf("unsupported input format %q", from)
	}
}

func encodeEntry(w io.Writer, e Entry, to Format) error {
	switch to {
	case Arrow:
		_, err := fmt.Fprintf(w, "%s%s%s\n", e.Link, arrowSep, e.Target)
		return err
	case JSONL:
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}
		_, err = w.Write(append(b, '\n'))
		return err
	default:
		return fmt.Errorf("unsupported output format %q", to)
	}
}

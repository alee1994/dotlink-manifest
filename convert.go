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

// Entry is one symlink: Link points at Target. Package is an optional
// stow-style grouping label - the name of the package (e.g. "vim",
// "shell") that owns this link, letting a manifest generated from a
// stow directory keep track of which links came from the same package
// without splitting into one manifest per package. The arrow format has
// no column for it, so a Package survives conversion to jsonl but is
// dropped when converting to arrow.
type Entry struct {
	Link    string `json:"link"`
	Target  string `json:"target"`
	Package string `json:"package,omitempty"`
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
		i, err := findArrowSep(line)
		if err != nil {
			return Entry{}, err
		}
		link, err := unescapeArrow(strings.TrimSpace(line[:i]))
		if err != nil {
			return Entry{}, err
		}
		target, err := unescapeArrow(strings.TrimSpace(line[i+len(arrowSep):]))
		if err != nil {
			return Entry{}, err
		}
		return Entry{Link: link, Target: target}, nil
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

// findArrowSep returns the index of the first arrowSep in line that isn't
// escaped, skipping over "\\" and "\ -> " so an escaped separator or
// backslash embedded in a path doesn't get mistaken for the real one.
func findArrowSep(line string) (int, error) {
	for i := 0; i < len(line); i++ {
		if line[i] != '\\' {
			if strings.HasPrefix(line[i:], arrowSep) {
				return i, nil
			}
			continue
		}
		if strings.HasPrefix(line[i+1:], arrowSep) {
			i += len(arrowSep)
			continue
		}
		if i+1 < len(line) && line[i+1] == '\\' {
			i++
			continue
		}
		return -1, fmt.Errorf(`stray backslash (use \\ for a literal backslash): %s`, line)
	}
	return -1, fmt.Errorf("missing %q separator: %s", arrowSep, line)
}

// unescapeArrow reverses escapeArrow: "\\" becomes "\" and an escaped
// separator becomes a literal arrowSep again.
func unescapeArrow(s string) (string, error) {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			sb.WriteByte(s[i])
			continue
		}
		switch {
		case strings.HasPrefix(s[i+1:], arrowSep):
			sb.WriteString(arrowSep)
			i += len(arrowSep)
		case i+1 < len(s) && s[i+1] == '\\':
			sb.WriteByte('\\')
			i++
		default:
			return "", fmt.Errorf(`stray backslash (use \\ for a literal backslash): %s`, s)
		}
	}
	return sb.String(), nil
}

// escapeArrow makes s safe to write as one side of an arrow-format line: a
// literal backslash is doubled, and a literal occurrence of arrowSep is
// prefixed with a backslash so it can't be confused with the real
// separator when the line is read back.
func escapeArrow(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, arrowSep, `\`+arrowSep)
	return s
}

func encodeEntry(w io.Writer, e Entry, to Format) error {
	switch to {
	case Arrow:
		_, err := fmt.Fprintf(w, "%s%s%s\n", escapeArrow(e.Link), arrowSep, escapeArrow(e.Target))
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

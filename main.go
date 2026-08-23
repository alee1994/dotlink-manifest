package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	from := flag.String("from", "", "input format: arrow or jsonl")
	to := flag.String("to", "", "output format: arrow or jsonl")
	in := flag.String("in", "", "input file (default stdin)")
	out := flag.String("out", "", "output file (default stdout)")
	scan := flag.String("scan", "", "scan this directory tree for symlinks instead of reading a manifest with -from/-in")
	flag.Parse()

	if err := run(*from, *to, *in, *out, *scan); err != nil {
		fmt.Fprintln(os.Stderr, "dlm:", err)
		os.Exit(1)
	}
}

func run(fromName, toName, inPath, outPath, scanPath string) error {
	to, err := ParseFormat(toName)
	if err != nil {
		return fmt.Errorf("-to: %w", err)
	}

	w := io.Writer(os.Stdout)
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	if scanPath != "" {
		if fromName != "" || inPath != "" {
			return fmt.Errorf("-scan can't be combined with -from or -in")
		}
		return ScanSymlinks(scanPath, w, to)
	}

	from, err := ParseFormat(fromName)
	if err != nil {
		return fmt.Errorf("-from: %w", err)
	}

	r := io.Reader(os.Stdin)
	if inPath != "" {
		f, err := os.Open(inPath)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	return ConvertStream(r, w, from, to)
}

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
	apply := flag.Bool("apply", false, "create the symlinks described by a manifest read with -from/-in, instead of converting")
	force := flag.Bool("force", false, "with -apply, replace a Link path that already exists")
	flag.Parse()

	if err := run(*from, *to, *in, *out, *scan, *apply, *force); err != nil {
		fmt.Fprintln(os.Stderr, "dlm:", err)
		os.Exit(1)
	}
}

func run(fromName, toName, inPath, outPath, scanPath string, apply, force bool) error {
	if apply {
		if toName != "" || outPath != "" || scanPath != "" {
			return fmt.Errorf("-apply can't be combined with -to, -out, or -scan")
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

		return ApplyManifest(r, from, force)
	}

	if force {
		return fmt.Errorf("-force only applies with -apply")
	}

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

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
	remove := flag.Bool("remove", false, "remove the symlinks described by a manifest read with -from/-in, instead of converting")
	check := flag.Bool("check", false, "report manifest entries read with -from/-in whose target doesn't exist, instead of converting")
	force := flag.Bool("force", false, "with -apply, replace a Link path that already exists; with -remove, remove it even if it doesn't match Target")
	flag.Parse()

	if err := run(*from, *to, *in, *out, *scan, *apply, *remove, *check, *force); err != nil {
		fmt.Fprintln(os.Stderr, "dlm:", err)
		os.Exit(1)
	}
}

func run(fromName, toName, inPath, outPath, scanPath string, apply, remove, check, force bool) error {
	if apply && check {
		return fmt.Errorf("-apply and -check can't be combined")
	}
	if apply && remove {
		return fmt.Errorf("-apply and -remove can't be combined")
	}
	if remove && check {
		return fmt.Errorf("-remove and -check can't be combined")
	}

	if apply {
		if toName != "" || outPath != "" || scanPath != "" {
			return fmt.Errorf("-apply can't be combined with -to, -out, or -scan")
		}
		from, err := ParseFormat(fromName)
		if err != nil {
			return fmt.Errorf("-from: %w", err)
		}

		r, closeR, err := openInput(inPath)
		if err != nil {
			return err
		}
		defer closeR()

		return ApplyManifest(r, from, force)
	}

	if remove {
		if toName != "" || outPath != "" || scanPath != "" {
			return fmt.Errorf("-remove can't be combined with -to, -out, or -scan")
		}
		from, err := ParseFormat(fromName)
		if err != nil {
			return fmt.Errorf("-from: %w", err)
		}

		r, closeR, err := openInput(inPath)
		if err != nil {
			return err
		}
		defer closeR()

		return RemoveManifest(r, from, force)
	}

	if force {
		return fmt.Errorf("-force only applies with -apply or -remove")
	}

	if check {
		if toName != "" || scanPath != "" {
			return fmt.Errorf("-check can't be combined with -to or -scan")
		}
		from, err := ParseFormat(fromName)
		if err != nil {
			return fmt.Errorf("-from: %w", err)
		}

		r, closeR, err := openInput(inPath)
		if err != nil {
			return err
		}
		defer closeR()

		w, closeW, err := openOutput(outPath)
		if err != nil {
			return err
		}
		defer closeW()

		return CheckManifest(r, w, from)
	}

	to, err := ParseFormat(toName)
	if err != nil {
		return fmt.Errorf("-to: %w", err)
	}

	w, closeW, err := openOutput(outPath)
	if err != nil {
		return err
	}
	defer closeW()

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

	r, closeR, err := openInput(inPath)
	if err != nil {
		return err
	}
	defer closeR()

	return ConvertStream(r, w, from, to)
}

// openInput opens inPath, or falls back to stdin when inPath is empty. The
// returned close func is always safe to defer, even for stdin.
func openInput(inPath string) (io.Reader, func(), error) {
	if inPath == "" {
		return os.Stdin, func() {}, nil
	}
	f, err := os.Open(inPath)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}

// openOutput opens outPath for writing, or falls back to stdout when
// outPath is empty. The returned close func is always safe to defer, even
// for stdout.
func openOutput(outPath string) (io.Writer, func(), error) {
	if outPath == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(outPath)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}

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
	flag.Parse()

	if err := run(*from, *to, *in, *out); err != nil {
		fmt.Fprintln(os.Stderr, "dlm:", err)
		os.Exit(1)
	}
}

func run(fromName, toName, inPath, outPath string) error {
	from, err := ParseFormat(fromName)
	if err != nil {
		return fmt.Errorf("-from: %w", err)
	}
	to, err := ParseFormat(toName)
	if err != nil {
		return fmt.Errorf("-to: %w", err)
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

	w := io.Writer(os.Stdout)
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	return ConvertStream(r, w, from, to)
}

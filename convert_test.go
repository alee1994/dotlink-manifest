package main

import (
	"strings"
	"testing"
)

func TestConvertArrowToJSONL(t *testing.T) {
	in := "~/.vimrc -> dotfiles/vim/vimrc\n~/.bashrc -> dotfiles/shell/bashrc\n"
	var out strings.Builder

	if err := ConvertStream(strings.NewReader(in), &out, Arrow, JSONL); err != nil {
		t.Fatal(err)
	}

	want := `{"link":"~/.vimrc","target":"dotfiles/vim/vimrc"}` + "\n" +
		`{"link":"~/.bashrc","target":"dotfiles/shell/bashrc"}` + "\n"
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestConvertJSONLToArrow(t *testing.T) {
	in := `{"link":"~/.vimrc","target":"dotfiles/vim/vimrc"}` + "\n"
	var out strings.Builder

	if err := ConvertStream(strings.NewReader(in), &out, JSONL, Arrow); err != nil {
		t.Fatal(err)
	}

	want := "~/.vimrc -> dotfiles/vim/vimrc\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

func TestConvertSkipsBlankLines(t *testing.T) {
	in := "~/.vimrc -> dotfiles/vim/vimrc\n\n   \n~/.bashrc -> dotfiles/shell/bashrc\n"
	var out strings.Builder

	if err := ConvertStream(strings.NewReader(in), &out, Arrow, Arrow); err != nil {
		t.Fatal(err)
	}

	if got := strings.Count(out.String(), "\n"); got != 2 {
		t.Errorf("got %d lines, want 2", got)
	}
}

func TestConvertBadArrowLine(t *testing.T) {
	in := "not a valid line\n"
	var out strings.Builder

	if err := ConvertStream(strings.NewReader(in), &out, Arrow, JSONL); err == nil {
		t.Fatal("expected an error for a line missing the arrow separator")
	}
}

func TestConvertArrowEscapesLiteralSeparator(t *testing.T) {
	in := `{"link":"~/weird/a -> b","target":"c -> d/target"}` + "\n"
	var arrow strings.Builder

	if err := ConvertStream(strings.NewReader(in), &arrow, JSONL, Arrow); err != nil {
		t.Fatal(err)
	}

	want := `~/weird/a\ -> b -> c\ -> d/target` + "\n"
	if arrow.String() != want {
		t.Fatalf("got %q, want %q", arrow.String(), want)
	}

	var back strings.Builder
	if err := ConvertStream(strings.NewReader(arrow.String()), &back, Arrow, JSONL); err != nil {
		t.Fatal(err)
	}
	if back.String() != in {
		t.Fatalf("round trip: got %q, want %q", back.String(), in)
	}
}

func TestConvertArrowEscapesLiteralBackslash(t *testing.T) {
	in := `{"link":"C:\\Users\\me\\.vimrc","target":"vim/vimrc"}` + "\n"
	var arrow strings.Builder

	if err := ConvertStream(strings.NewReader(in), &arrow, JSONL, Arrow); err != nil {
		t.Fatal(err)
	}

	var back strings.Builder
	if err := ConvertStream(strings.NewReader(arrow.String()), &back, Arrow, JSONL); err != nil {
		t.Fatal(err)
	}
	if back.String() != in {
		t.Fatalf("round trip: got %q, want %q", back.String(), in)
	}
}

func TestConvertArrowStrayBackslashIsError(t *testing.T) {
	in := `~/foo\bar -> target` + "\n"
	var out strings.Builder

	if err := ConvertStream(strings.NewReader(in), &out, Arrow, JSONL); err == nil {
		t.Fatal("expected an error for a stray backslash")
	}
}

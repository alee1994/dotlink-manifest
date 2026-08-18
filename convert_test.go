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

# dotlink-manifest

`dlm` converts a dotfile symlink manifest between two text formats.

## the problem

A dotfiles setup usually ends up as a pile of real symlinks in `$HOME`
pointing into a version-controlled repo, created by `stow`, `dotbot`, or a
hand-rolled loop of `ln -s`. The symlinks themselves are a bad source of
truth: they don't diff cleanly in a PR, git handles them inconsistently
across platforms (Windows checkouts, `tar` without `-h`, some backup
tools flatten them into copies), and every dotfiles tool expects its own
config shape.

`dlm` doesn't manage symlinks itself. It converts a plain list of
`link -> target` pairs between two representations so you can pick
whichever one a given step needs:

- **arrow** - one pair per line, human-readable, greppable, diffs well
  in a pull request:

  ```
  ~/.vimrc -> dotfiles/vim/vimrc
  ~/.bashrc -> dotfiles/shell/bashrc
  ```

- **jsonl** - one JSON object per line, easy to pipe into `jq` or feed
  to another script:

  ```
  {"link":"~/.vimrc","target":"dotfiles/vim/vimrc"}
  {"link":"~/.bashrc","target":"dotfiles/shell/bashrc"}
  ```

## usage

```
$ find ~/dotfiles -type l -printf '%p -> %l\n' > links.txt
$ dlm -from arrow -to jsonl -in links.txt -out links.jsonl
$ dlm -from jsonl -to arrow < links.jsonl
~/.vimrc -> dotfiles/vim/vimrc
~/.bashrc -> dotfiles/shell/bashrc
```

Both `-in` and `-out` default to stdin/stdout, so it also works as a
pipe filter:

```
$ find ~/dotfiles -type l -printf '%p -> %l\n' | dlm -from arrow -to jsonl | jq .
```

## streaming

Conversion is line-by-line: `dlm` never holds more than one entry in
memory. A manifest generated from `find /` or from a machine with tens
of thousands of managed dotfiles converts the same way a five-line one
does - constant memory, one pass over the input.

## current limitations

- No built-in filesystem scanning yet - pipe manifest generation
  through `find` (see roadmap).
- The arrow format has no escaping, so a path that itself contains the
  literal string `" -> "` will not round-trip correctly.
- Nothing here creates or removes symlinks; `dlm` only converts the
  manifest that describes them.

## license

MIT, see LICENSE.

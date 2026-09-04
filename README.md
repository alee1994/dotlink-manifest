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

`dlm` can also walk a directory tree itself instead of going through
`find`, with `-scan`:

```
$ dlm -scan ~/dotfiles -to arrow
~/dotfiles/.vimrc -> vim/vimrc
~/dotfiles/.bashrc -> shell/bashrc
```

`-scan` reports whatever `os.Readlink` returns for each link, so a
relative target (as `stow` and `dotbot` both create) comes out relative,
exactly as it's stored on disk. It can't be combined with `-from` or
`-in`, since there's no manifest to read in that mode.

The other direction, `-apply`, reads a manifest and creates the symlinks
it describes:

```
$ dlm -apply -from arrow -in links.txt
```

A `Link` path that already exists and isn't already the correct symlink
is left alone and reported as an error, rather than being overwritten -
pass `-force` to replace it instead. A link that's already correct is
skipped, so re-running `-apply` on a manifest you've already applied is
a no-op. `-apply` can't be combined with `-to`, `-out`, or `-scan`.

`-check` reads a manifest and reports entries whose target doesn't exist
on disk, without touching anything:

```
$ dlm -check -from arrow -in links.txt
~/.vimrc -> dotfiles/vim/vimrc
```

A relative `Target` is resolved against the directory containing `Link`,
the same way the filesystem resolves it when it actually follows the
symlink - not against dlm's current directory. `-check` exits non-zero
if it found at least one dangling entry, so it works as a CI step or
pre-commit hook; it works whether or not the manifest has been applied
yet, since it only asks whether `Target` exists. `-check` can be
combined with `-out` to write the dangling entries to a file instead of
stdout, but not with `-to` or `-scan`.

## streaming

Conversion is line-by-line: `dlm` never holds more than one entry in
memory. A manifest generated from `find /` or from a machine with tens
of thousands of managed dotfiles converts the same way a five-line one
does - constant memory, one pass over the input.

## escaping

A path that itself contains the literal string `" -> "` or a backslash
still round-trips through the arrow format: `dlm` writes a backslash
before an embedded separator (`\ -> `) and doubles a literal backslash
(`\\`) when it converts to arrow, and reverses both on the way back. You
only need to know this if you're hand-editing an arrow file - anything
`dlm` itself writes and reads back needs no attention from you.

## current limitations

- `-apply` only creates symlinks; there's no `dlm` command yet to remove
  the ones a manifest describes.
- jsonl has no notion of stow-style package grouping - every entry is
  flat, with no way to say "these ten links came from the same package".

## license

MIT, see LICENSE.

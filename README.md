# nvi (customized fork — not general purpose)

> **This is not a general-purpose nvi distribution.** It is a fork of
> [nvi](https://en.wikipedia.org/wiki/Nvi) (the Berkeley ex/vi reimplementation)
> that has been customized to **build on macOS** and to serve as a **reference
> implementation ("oracle") for testing** the [govi](https://github.com/goexvi-ctrl/govi)
> editor against nvi. Do not use it as a general-purpose editor or as an
> upstream-tracking nvi; the changes here exist to support that testing workflow,
> not to improve nvi for general use.

## Purpose

The [goterm](https://github.com/goexvi-ctrl/goterm) harness is a headless ANSI
terminal emulator that drives two editors on identical terminals with identical
input and diffs the rendered screen and cursor. It compares **govi** (the editor
under test) against **nvi** (the reference), so nvi must build and run reliably
on the development machine (macOS). This fork provides that reference binary.

## What is customized

- **macOS build support** — the curses/`<term.h>` and `<ctype.h>`/`isblank`
  handling and the `db.1.85` PORT symlinks are adjusted so nvi compiles under the
  macOS clang/ncurses toolchain.
- **Checked-in build scaffolding** — the generated `configure`, `config.h.in`,
  `Makefile.in` and nvi's generated headers are committed so the tree builds
  without the bootstrap autotools (e.g. a matching `aclocal`). Accordingly,
  `build.unix/Makefile` disables autotools regeneration; a bare `make` will not
  try to re-run `aclocal`/`autoconf` (which would otherwise rewrite `config.h`
  and break the macOS build). `build.unix/` itself is not tracked.
- **`GOTERM_ORACLE_PRESERVE`** — the recovery directory (`recdir`) default is
  taken from the `GOTERM_ORACLE_PRESERVE` environment variable when set, so
  conformance runs keep their recovery files out of the shared
  `/var/tmp/vi.recover`.
- **`tests/`** — a Go conformance harness (`tests/vi`, module
  `github.com/goexvi-ctrl/nvi/tests/vi`) that drives this nvi via `goterm`.

## Building

```
cd build.unix
../dist/configure   # only needed for a fresh build tree
make vi
```

The reference binary is `build.unix/vi`.

## Related repositories

- [govi](https://github.com/goexvi-ctrl/govi) — the editor under test.
- [goterm](https://github.com/goexvi-ctrl/goterm) — the headless terminal
  emulator / comparison harness.

## Upstream nvi

For the original nvi overview, directory layout, history and acknowledgements,
see the plain-text [`README`](./README).

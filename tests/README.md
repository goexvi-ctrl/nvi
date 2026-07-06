# nvi test suite

Tests for nvi/nex, written against the behavior documented in
docs/nvi.md (the Bostic "Vi/Ex Reference Manual").  The suite has
three parts:

  unit/       C unit tests that link directly against nvi objects.
              unit/regex tests the bundled Henry Spencer regex
              library (the exact code nvi searches with).
              unit/db builds the db.1.85 library's own dbtest
              driver against the same objects nvi links and runs
              the library's btree/hash/recno regression scripts.

  ex/         Functional tests for ex/nex.  Each *.t file holds test
              cases that are run in batch mode ("vi -e -s file") and
              checked against the resulting file contents, standard
              output, and exit status.

  vi/         Functional tests for full-screen vi mode.  A Go test
              harness drives the editor on a headless ANSI terminal
              emulator (the goterm package) and asserts on the
              rendered screen and cursor position.

Nothing here modifies the nvi sources or its build; the tests only
consume the tree and the objects/binary already built in build.unix.

## Prerequisites

- The nvi oracle must be built first:  build.unix/vi and the
  build.unix/*.o objects must exist (run make in build.unix).
- unit tests: a C compiler (cc).
- vi screen tests: Go, plus a checkout of goterm at ../../goterm
  relative to this directory (i.e. a sibling of the nvi tree).
  If goterm is not present these tests are skipped, not failed.

## Running

    cd tests
    make            # everything
    make unit       # C unit tests only
    make ex         # ex/nex batch tests only
    make vi         # vi screen tests only

An individual ex battery:

    sh ex/run.sh ex/t_substitute.t

Verbose mode (prints each case name):

    sh ex/run.sh -v ex/t_*.t

## The ex test format

Each ex/*.t file is line oriented:

    ### test <name>
    --- input
    <initial contents of the edited file>
    --- commands
    <ex commands fed on standard input>
    --- file
    <expected final contents of the edited file>
    --- output
    <expected standard output>
    --- exit <n>
    --- xfail <reason>
    --- noauto
    --- args <arguments>
    --- aux <name>
    <contents of an extra fixture file <name> in the scratch dir>
    ### end

"--- args" replaces the default "-e -s file.txt" invocation
arguments (whitespace-split, no quoting) so command-line handling
itself can be tested.  "--- errout" compares standard error, where
usage and startup errors appear; the runner normalizes the leading
binary path in stderr lines to "vi:" so messages compare stably.

Sections may appear in any order; all are optional except
"--- commands".  Omitted "--- file" / "--- output" sections are not
checked.  "--- exit" defaults to 0.  Unless "--- noauto" is given,
the runner appends "w!" and "q!" to the commands so the buffer is
written back for the file comparison.

Two batch-mode facts the format leans on:

- When a command fails in batch mode, nvi discards the rest of the
  script and exits 1.  So an error test usually expects exit 1, an
  unchanged file (the appended "w!" never runs), and the error text
  on standard output with a "script, <line>:" prefix.
- With all of stdin/stdout/stderr redirected, nvi emits one line of
  startup noise ("Error: stderr: <strerror text>") from the terminal
  ioctl it attempts on stderr.  The runner strips that single leading
  line before comparing output; everything else is compared exactly.

## Test failures and XFAIL policy

If a test fails, first suspect the test.  Re-read the relevant part
of docs/nvi.md and check the behavior by hand before touching
anything.  Only when the test is right and nvi is genuinely wrong is
the test marked as a known failure:

- ex tests: add "--- xfail <one-line reason, citing docs/nvi.md>".
- unit tests: use CHECK_XFAIL instead of CHECK.
- vi tests: t.Skipf with an explanation, or an explicit xfail helper.

An XFAIL that starts passing (XPASS) is reported as an error so a
stale marker gets noticed and removed.  Fixing nvi itself is out of
scope for this suite.

## Known nvi bugs (current XFAILs)

- mkexrc-source-roundtrip (ex/t_files.t): the mkexrc dump contains
  "set nosecure", which set unconditionally rejects, so an
  nvi-generated rc file can never be read back with source, contrary
  to the mkexrc documentation.
- recover-after-preserve (ex/t_misc.t) and TestSighupRecovery
  (vi/recovery_test.go): crash recovery is broken in this build.
  The recovery data file (recdir/vi.*) stays zero length for the
  whole session - it is 0 bytes at startup and never grows, even
  after an edit and an explicit preserve - so only the 430-byte
  metadata mail (recdir/recover.*) is ever written.  The buffer
  content is never persisted, so vi -r reports "No files to recover"
  and recovering by name loads an empty buffer.  This reproduces
  through a real pty (the SIGHUP test), not only in batch, and is
  independent of the local relative-recdir patch (an absolute recdir
  behaves the same).  It looks like the db/mpool layer not syncing
  the backing file on this platform; recovery is inherently
  environment-sensitive, so treat this as "broken here" rather than
  a proven logic defect in portable nvi.  The vi test reports as
  skipped with the detail and fails loudly if recovery starts
  working.
- print-then-autoprint-address-aborts (ex/t_syntax.t): in scripted
  mode only - standard input not a terminal, i.e. an ex script from
  a file or pipe - an explicit print (p) command followed by a bare
  auto-printing address (a line number, search, +, or .) smashes the
  stack and aborts on a signal (the backtrace will not unwind; a
  guard malloc turns it into SIGSEGV).  The identical keystrokes at
  an interactive ex prompt do NOT crash, verified through a pty, so
  it is invisible in ordinary interactive use and cannot be
  reproduced by hand at the : prompt.  A bare address after any
  other command is fine, and an address with an explicit command
  (2p) is fine; only the second implicit print after a prior p, with
  non-tty stdin, triggers it.  The chaos.sh fuzzer (make chaos) hits
  this repeatedly and classifies it as the known bug; it fails only
  on a crash that survives removing the print lines.
- TestScriptCommand (vi/session_test.go): the script command cannot
  allocate a pty on this platform.  ex/ex_script.c only knows the
  legacy BSD /dev/ptyXX names (the System V grantpt path is compiled
  out because configure leaves HAVE_SYS5_PTY unset), and those nodes
  fail with EAGAIN on modern macOS, so script reports "Error: pty".
- the collating-element name lookup in regex/regcomp.c (p_b_coll_elem)
  tested MEMCMP() for nonzero instead of zero, so "[[.comma.]]" resolved
  to the wrong cnames entry and an unknown name compiled instead of
  failing with REG_ECOLLATE.  This is FIXED on this branch (the
  comparison is now !MEMCMP, matching the parallel p_b_cclass lookup),
  so unit/regex/test_regex.c now asserts the corrected behavior with
  plain CHECK rather than CHECK_XFAIL.  It is the one issue the suite
  found that is fixed here rather than left as an expected failure.

## Coverage notes

The ex batteries exercise most of the cmds[] table in ex/ex_cmd.c,
plus the command-line syntax machinery (| separators, " comments,
print flags, the z window types, the substitute confirm flag in
batch, and the search, mark, and \/ address forms in t_syntax.t),
file-system edge cases (t_fs.t: missing final newline, permission
failures, control-byte round trips), size sanity (t_stress.t: a
10000-line file and a 5000-character line), and golden snapshots of
the self-describing output (t_defaults.t: set all defaults, exusage,
viusage).  ex/chaos.sh (make chaos) is a separate deterministic
crash hunt.  Deliberately untested: shell, stop, and suspend
(interactive by nature) and perl/perldo/tcl (not compiled into this
oracle).  script is covered as a known-bug XFAIL, cscope by the vi
tests (a real database is built in the fixture), and bg/fg/resize by
the vi screen tests.

The vi tests cover most of the bound keys in vikeys[] in
vi/v_cmd.c, organized by family: motions_test.go (character, word,
line, sentence, paragraph, section, mark, search, and screen
position motions), scroll_test.go (^F ^B ^D ^U ^E ^Y z),
operators_test.go (d c y p P with motions, counts, named and
numeric buffers, < > and the ! filter), charbuf_test.go (operators
with search and character-find motions, character-mode buffers,
previous-context marks), counts_test.go (counts on yanks, puts,
buffers, and the repeat command), edits_test.go (r R a A i I o O J
~ x X . u U # ^A @ &, the undo chain, and the input-mode keys ^V
^W and autoindent), input_test.go (the autoindent erase strings
^D 0^D ^^D, ^T, counted inserts, r<CR>, map! and abbreviation
expansion), modes_test.go (ZZ Q ^^ ^] ^T, vsplit and ^W),
screens_test.go (the Edit/Tag split-screen commands, ^W cycling,
bg/fg/resize/display screens, and quitting an ex-mode screen with
other screens open), startup_test.go (NEXINIT/EXINIT and
.nexrc/.exrc precedence, the exrc option), failures_test.go
(motions that fail ring the bell and leave the cursor and text
alone; arrow keys), cmdline_test.go (filec file-name completion and
the cedit command-editing window), cscope_test.go (a real cscope
database drives find/help/reset and the tag-stack interplay; skipped
if cscope is not installed), session_test.go (the second concurrent
session's read-only advisory lock, and script as a known-bug XFAIL),
and recovery_test.go (SIGHUP crash recovery, a known-bug XFAIL).
The ^L/^R redraw tests corrupt the emulated screen behind vi's back
and assert the repaint restores it.  Deliberately untested keys:
^Z (suspend) and the interrupt/quote keys ^C ^Q ^S that never
reach the key table.

Set options are covered in two places: ex/t_set.t tests the set
command itself (listing, queries, errors, the batch-mode default
changes) and every option with batch-observable behavior;
vi/options_test.go tests the options whose effect is visual (list,
number, tabstop, showmode, tildeop, remap, wrapmargin, window,
leftright, searchincr, report, paragraphs, sections, ruler,
ttywerase); backup is covered in ex/t_files.t, exrc and filec/cedit
in vi/startup_test.go and vi/cmdline_test.go.  Options
with no deterministic observable effect in these harnesses are
skipped: terminal tuning (redraw, slowopen, optimize,
w300/w1200/w9600, hardtabs, escapetime, keytime, matchtime,
timeout, flash, errorbells), showmatch (its cursor bounce lasts
matchtime tenths of a second, too timing-sensitive to sample
reliably), and environment-dependent switches (mesg, msgcat,
sourceany, shell, shellmeta, directory, windowname, octal,
noprint/print, fileencoding/inputencoding).  The lock option is
exercised indirectly by the second-session read-only test.

Startup files can only be tested in vi mode: ex batch mode never
reads them, because a redirected standard input sets the scripted
flag, which implies -s, which skips all startup information.

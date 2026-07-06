# Ex command-line syntax machinery (docs/nvi.md "Ex Description"):
# | command separators, " comments, trailing print flags, the z
# window types, substitute's confirm flag in batch mode, and the
# search and mark address forms.

### test pipe-separates-commands
--- input
one
two
three
four
five
--- commands
1d|1d
--- file
three
four
five
### end

### test pipe-under-global
# The | separator also splits the command list of a global; both
# substitutes run on each selected line.
--- input
one
two
three
four
five
--- commands
g/t/s/t/T/|s/o/0/
--- file
one
Tw0
Three
four
five
### end

### test comment-line-ignored
--- input
one
two
--- commands
" a comment line
1d
--- file
two
### end

### test comment-after-command-rejected
# Historic ex has no trailing comments; the quote is rejected by
# the command parser.
--- noauto
--- input
one
--- commands
1d " trailing comment
--- output
script, 1: Usage: [line [,line]] d[elete][flags] [buffer] [count] [flags]
--- exit 1
--- file
one
### end

### test delete-print-flag-attached
# dp deletes and prints the new current line.  The flag must be
# attached to the command name: a separated "d p" names buffer p
# instead (see delete-detached-p-is-buffer).
--- noauto
--- input
one
two
three
--- commands
2dp
q!
--- output
three
### end

### test delete-list-flag-attached
--- noauto
--- input
one
two
three
--- commands
2dl
q!
--- output
three$
### end

### test delete-hash-flag-detached
# The # flag is not a buffer name, so it works detached.
--- noauto
--- input
one
two
three
--- commands
1d #
q!
--- output
     1  two
### end

### test delete-detached-p-is-buffer
# "d p" parses p as a buffer name: nothing prints, and the deleted
# line can be put back from buffer p.
--- input
one
two
three
--- commands
2d p
$pu p
--- file
one
three
two
### end

### test z-default-window
# Plain z displays the window after the current line.
--- noauto
--- input
one
two
three
four
five
--- commands
1
z3
q!
--- output
one
two
three
four
### end

### test z-plus-window
# z+ places the addressed line at the top of the window.
--- noauto
--- input
one
two
three
four
five
--- commands
1
z+3
q!
--- output
one
one
two
three
### end

### test z-caret-window
# z^ displays the window before the addressed line.
--- noauto
--- input
one
two
three
four
five
--- commands
3
z^3
q!
--- output
three
one
two
three
### end

### test z-equals-window
# z= centers the current line between separator lines.
--- noauto
--- input
one
two
three
four
five
--- commands
3
z=3
q!
--- output
three
two
----------------------------------------
three
----------------------------------------
### end

### test substitute-confirm-in-batch
# The c flag prints each candidate with a caret under the match and
# reads the y/n answers from the script.  The prompt line has no
# trailing newline, so the next write to standard output continues
# it: the answer to "one" is yes (giving onE in the final print),
# the answer to "three" is no.
--- noauto
--- input
one
two
three
four
five
--- commands
1,3s/e/E/c
y
n
%p
q!
--- output
one
  ^[ynq]three
   ^[ynq]onE
two
three
four
five
### end

### test substitute-empty-match-global
# A pattern that can match the empty string substitutes at every
# position with the g flag.
--- noauto
--- input
one
--- commands
1s/x*/-/g
1p
q!
--- output
-o-n-e-
### end

### test mark-range-address
--- input
one
two
three
four
five
--- commands
2ka
4kb
'a,'bd
--- file
one
five
### end

### test backslash-slash-repeats-search
# \/ repeats the last search forward; with wrapscan the only match
# is found again.
--- noauto
--- input
one
two
three
--- commands
/two/
\/
q!
--- output
two
two
### end

### test backslash-slash-without-pattern
--- noauto
--- input
one
--- commands
\/
--- output
script, 1: No previous search pattern
--- exit 1
--- file
one
### end

### test print-then-autoprint-address-aborts
# Known nvi bug, but only in scripted (non-interactive) mode.  When
# standard input is NOT a terminal -- i.e. an ex script fed from a
# file or pipe, as here and as in any "vi -e -s file < script" run --
# an explicit print (p) command followed by a bare address line (a
# line number, search, +, or . that auto-prints with no command of
# its own) smashes the stack and aborts on a signal (the backtrace
# will not unwind; a guard malloc turns it into SIGSEGV).  The same
# keystrokes typed at an interactive ex prompt (stdin a tty) do NOT
# crash -- verified through a pty -- which is why it is invisible in
# ordinary use.  A bare address after any other command is fine, and
# an address with an explicit command (2p, 2d) is fine; only the
# second, implicit print after a prior p in scripted mode triggers
# it.  The output below is what the two prints should produce; the
# crash exits on the signal instead of 0.
--- xfail nvi smashes the stack on an auto-printing address after a print in scripted mode
--- noauto
--- input
one
two
three
four
five
--- commands
1p
2
q!
--- output
one
two
### end

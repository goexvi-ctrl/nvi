# The set command and the options with batch-observable behavior
# (docs/nvi.md "Set Options").  Options whose effect is visual are
# covered by the vi tests; see the coverage notes in README.md.

### test set-bare-lists-changed-options
# A bare set lists the options changed from their defaults,
# column-major.  Batch (-s) mode itself changes four: autoprint,
# prompt, and warn go off, and term is "dumb" from the harness
# environment.
--- input
x
--- commands
set nomagic
set shiftwidth=4
set
--- output
noautoprint     noprompt        term="dumb"
nomagic         shiftwidth=4    nowarn
### end

### test set-toggle-bang-not-supported
# "set option!" is a vim-ism; nvi rejects it.
--- input
x
--- commands
set ic!
--- output
script, 1: set: no ic! option: 'set all' gives all option values
--- exit 1
### end

### test set-bad-number-is-error
--- input
x
--- commands
set sw=abc
--- output
script, 1: set: sw option: abc is an illegal number
--- exit 1
### end

### test set-multiple-in-one-command
--- input
x
--- commands
set ignorecase shiftwidth=4
set ic? sw?
--- output
ignorecase      shiftwidth=4
### end

### test set-recdir-is-local
# This oracle build defaults the recovery directory to ./vi.recover
# (see the govi conformance notes in the tree); upstream nvi uses
# /var/tmp/vi.recover.
--- input
x
--- commands
set recdir?
--- output
recdir="vi.recover"
### end

### test option-number-affects-ex-print
--- input
one
two
--- commands
set number
1p
--- output
     1  one
### end

### test option-autoprint
# autoprint (off in batch mode by default) makes delete print the
# new current line.
--- input
one
two
three
--- commands
1d
set autoprint
1d
--- output
three
--- file
three
### end

### test option-iclower-lowercase-matches-any-case
# iclower: an all-lowercase RE is case insensitive...
--- input
xxx
TWO
yyy
--- commands
set iclower
/two/=
--- output
2
### end

### test option-iclower-uppercase-stays-exact
# ...but an RE containing uppercase stays case sensitive.
--- input
two
TWO
yyy
--- commands
set iclower
/TWO/=
--- output
2
### end

### test option-edcompatible-remembers-suffix-for-bare-s
# With edcompatible, the g suffix is remembered and applied to a
# subsequent bare s command.
--- input
aaa
aaa
--- commands
set edcompatible
1s/a/X/g
2
s
--- output
aaa
--- file
XXX
XXX
### end

### test option-noedcompatible-bare-s-has-no-g
# Without it, the bare s repeat substitutes the first match only.
--- input
aaa
aaa
--- commands
1s/a/X/g
2
s
--- output
aaa
--- file
XXX
Xaa
### end

### test option-readonly-blocks-write
--- noauto
--- input
one
--- commands
set readonly
w
q!
--- output
script, 2: Read-only file, not written; use ! to override
--- exit 1
--- file
one
### end

### test option-readonly-bang-overrides
--- noauto
--- input
one
--- commands
set readonly
s/one/ONE/
w!
q!
--- file
ONE
### end

### test option-writeany-existing-file
# Without writeany, writing over an existing other file needs !.
--- aux other.txt
existing
--- input
one
--- commands
w other.txt
--- output
script, 1: other.txt exists, not written; use ! to override
--- exit 1
### end

### test option-writeany-set
--- noauto
--- aux other.txt
existing
--- input
one
--- commands
set writeany
w other.txt
r other.txt
q!
--- output
--- file
one
### end

### test option-autowrite-on-next
# autowrite writes the modified file rather than failing the next
# command; wn-style movement then succeeds.
--- noauto
--- aux b.txt
bee
--- aux c.txt
cee
--- input
one
--- commands
next! b.txt c.txt
1s/bee/BEE/
set autowrite
next
1p
r b.txt
2p
q!
--- output
cee
BEE
### end

### test option-taglength-prefix-match
# With taglength=4 only the first four characters of a tag name are
# significant.
--- noauto
--- aux tags
longfunction	t1.c	1
--- aux t1.c
int longfunction;
--- input
one
--- commands
set taglength=4
tag longZZZZ
p
q!
--- output
int longfunction;
### end

### test option-tags-multiple-files
# The tags option is a whitespace-separated list of tags files,
# searched in order.
--- noauto
--- aux tags
afunc	t1.c	1
--- aux tags2
otherfunc	t2.c	1
--- aux t1.c
int afunc;
--- aux t2.c
int otherfunc;
--- input
one
--- commands
set tags=tags\ tags2
tag otherfunc
p
q!
--- output
int otherfunc;
### end

### test option-cdpath
# cd searches the cdpath directories for the target.  The test cds
# back out so the session's relative recovery directory is intact at
# exit.
--- input
one
--- commands
!mkdir -p sub/leaf
set cdpath=sub
cd leaf
!echo ok
cd ../..
--- output
ok
### end

### test option-nocdpath-cd-fails
--- input
one
--- commands
!mkdir -p sub/leaf
cd leaf
--- output
Error: script, 2: leaf: No such file or directory
--- exit 1
### end

### test option-path-for-edit
# The path option provides the search path for the edit command.
--- noauto
--- aux pfile.txt
pfile line
--- input
one
--- commands
!mkdir pdir
!mv pfile.txt pdir/
set path=pdir
e pfile.txt
1p
q!
--- output
pfile line
### end

### test option-lisp-not-implemented
# Setting lisp warns but is not an error; the script continues.
--- input
one
--- commands
set lisp
1p
--- output
script, 1: The lisp option is not implemented
one
### end

### test option-modeline-never-on
# For security, modelines cannot be enabled at all -- even the query
# form is rejected.
--- input
one
--- commands
set modeline?
--- output
script, 1: set: the modeline option may never be turned on
--- exit 1
### end

### test option-terse-is-a-noop
# terse is accepted for historic compatibility; messages do not
# change.
--- input
one
--- commands
set terse
s/zzz/x/
--- output
script, 2: No match found
--- exit 1
--- file
one
### end

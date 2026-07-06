# File-level ex commands: file, cd, source, version, z, open, !,
# wq, xit, preserve, mkexrc (docs/nvi.md "Ex Commands").

### test file-status
--- input
one
two
three
--- commands
f
--- output
file.txt: unmodified: line 3 of 3 [100%]
### end

### test file-rename
# f with an argument changes the pathname and reports it changed.
--- input
one
--- commands
f other.txt
--- output
other.txt: name changed, unmodified: line 1 of 1 [100%]
### end

### test cd-and-bad-cd
# cd to a bad directory is an error; the failure discards the rest
# of the batch script.
--- input
one
--- commands
cd /
cd /nonexistent-dir-xyz
--- output
Error: script, 2: /nonexistent-dir-xyz: No such file or directory
--- exit 1
--- file
one
### end

### test source-runs-commands
# Build a script file from buffer lines, then source it.
--- input
s/one/ONE/
one
two
--- commands
1w! script.ex
1d
so script.ex
--- file
ONE
two
### end

### test source-missing-file-is-error
--- input
one
--- commands
so nosuchscript.ex
--- exit 1
--- file
one
### end

### test version-is-silent-in-batch
# -s is the historic "silent" mode: informational messages such as
# the version banner are suppressed.
--- input
one
--- commands
version
--- output
### end

### test z-prints-window
# z prints the next count lines.
--- input
one
two
three
four
--- commands
1z3
--- output
one
two
three
### end

### test z-equals-context
# z= prints the surrounding lines with the current line set off by
# separator lines.
--- input
one
two
three
four
--- commands
3z=
--- output
one
two
----------------------------------------
three
----------------------------------------
four
### end

### test z-minus-window-before
# z- prints a window ending at the addressed line.
--- input
one
two
three
four
--- commands
4z-2
--- output
three
four
### end

### test open-not-implemented
# nvi documents its differences honestly: the historic open mode is
# not implemented.
--- input
one
--- commands
open
--- output
script, 1: The open command is not yet implemented
--- exit 1
### end

### test bang-runs-command
# ! without an address runs the command and shows its output.
--- input
one
--- commands
!echo hi there
--- output
hi there
### end

### test wq-writes-and-quits
--- noauto
--- input
one
two
--- commands
1s/one/ONE/
wq
--- file
ONE
two
### end

### test xit-writes-and-quits
--- noauto
--- input
one
two
--- commands
1s/one/ONE/
x
--- file
ONE
two
### end

### test xit-unmodified-does-not-write
# x only writes if the buffer was modified; either way it exits
# cleanly.
--- noauto
--- input
one
--- commands
x
--- file
one
### end

### test preserve-succeeds
# preserve saves the buffer to the recovery area (./vi.recover in
# this build) for later recover; in batch mode it is silent.
--- input
one
--- commands
preserve
--- output
### end

### test mkexrc-refuses-overwrite
# mkexrc without ! must not overwrite an existing file.
--- input
one
--- commands
mkexrc myrc
mkexrc myrc
--- exit 1
--- file
one
### end

### test set-nosecure-always-rejected
# The secure option may not be turned off -- nvi rejects the attempt
# even when secure is not set (see the secure entry in "Set
# Options": "Once set, the secure edit option may not be unset").
--- input
one
--- commands
set secure?
set nosecure
--- output
nosecure
script, 2: set: the secure option may not be turned off
--- exit 1
### end

### test mkexrc-source-roundtrip
# docs/nvi.md, mkexrc: "Information is written in a form which can
# later be read back in using the ex source command."  It cannot:
# the dump contains "set nosecure", which set unconditionally
# rejects, so sourcing an nvi-generated rc file always fails partway
# through and discards the remaining commands.
--- xfail mkexrc emits "set nosecure" which source cannot read back
--- input
one
--- commands
set shiftwidth=4
mkexrc myrc
set shiftwidth=8
so myrc
set shiftwidth?
--- output
shiftwidth=4
### end

### test mkexrc-existing-file-refused
--- noauto
--- input
one
--- commands
mkexrc pre.rc
mkexrc pre.rc
--- output
script, 2: pre.rc exists, not written; use ! to override
--- exit 1
### end

### test mkexrc-force-overwrites
--- noauto
--- input
one
--- commands
mkexrc pre.rc
mkexrc! pre.rc
!test -s pre.rc && echo HAVE-RC
q!
--- output
HAVE-RC
### end

### test backup-option-copies-before-write
# With the backup option set, a write first copies the file's
# current contents to the expanded backup pathname.
--- noauto
--- input
one
two
three
--- commands
set backup=%.bak
1d
w
!cat file.txt.bak
q!
--- output
one
two
three
--- file
two
three
### end

### test backup-option-leading-N-versions
# A leading N in the backup pathname asks for version numbers; each
# write creates a new numbered backup.
--- noauto
--- input
one
two
three
--- commands
set backup=N%.bak
1d
w
1d
w
!ls file.txt.bak1 file.txt.bak2
!cat file.txt.bak2
q!
--- output
file.txt.bak1
file.txt.bak2
two
three
--- file
three
### end

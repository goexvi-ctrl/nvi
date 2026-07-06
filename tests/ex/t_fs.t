# File system edge cases: files without a trailing newline,
# permission failures on files and directories, and byte-for-byte
# round-trip fidelity for control characters (docs/nvi.md "Ex
# Commands", edit and write, and the readonly discussion under "Set
# Options").
#
# Files that cannot be represented in the ASCII test format (no
# final newline, control bytes) are fabricated with !printf and
# verified with !wc/!cmp; in batch mode a plain ! command produces
# only the command's output, no echo.

### test fs-write-restores-final-newline
# A file missing its final newline reads cleanly and gains the
# newline back when written: two lines of two bytes each write as
# six bytes.
--- noauto
--- input
seed
--- commands
!printf 'ab\ncd' > nl.txt
e! nl.txt
w
!wc -c < nl.txt | tr -d ' '
q!
--- output
6
### end

### test fs-readonly-file-not-written
# An unwritable file gets a read-only session; plain write is
# refused with a pointer at the ! override.
--- noauto
--- aux ro.txt
text
--- commands
!chmod 400 ro.txt
e! ro.txt
w
--- output
script, 3: Read-only file, not written; use ! to override
--- exit 1
### end

### test fs-write-bang-permission-denied
# The ! override retries the write, which then fails in the
# operating system.
--- noauto
--- aux ro.txt
text
--- commands
!chmod 400 ro.txt
e! ro.txt
w!
--- output
Error: script, 3: ro.txt: Permission denied
--- exit 1
### end

### test fs-write-into-unwritable-directory
--- noauto
--- input
one
--- commands
!mkdir subd
!chmod 500 subd
w subd/out.txt
--- output
Error: script, 3: subd/out.txt: Permission denied
--- exit 1
### end

### test fs-control-bytes-round-trip
# Control characters (^A, tab, delete, escape) survive a read and
# write unchanged.
--- noauto
--- input
seed
--- commands
!printf 'a\001b\tc\177\nx\033y\n' > ctrl.bin
!cp ctrl.bin ctrl.orig
e! ctrl.bin
w
!cmp ctrl.bin ctrl.orig && echo SAME
q!
--- output
SAME
### end

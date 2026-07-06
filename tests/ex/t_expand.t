# File name expansion (% is the current file, # the alternate) and
# shell input/output: r !command and w !command (docs/nvi.md
# "General Editor Description", current/alternate pathname, and the
# read/write entries under "Ex Commands").

### test expand-percent-in-shell-escape
--- input
one
--- commands
!echo %
--- output
file.txt
### end

### test expand-read-percent
# Reading the file being edited warns that the read lock was
# unavailable (the session itself holds it) but reads anyway.
--- input
one
two
--- commands
r %
--- output
script, 1: file.txt: read lock was unavailable
--- file
one
two
one
two
### end

### test expand-hash-is-alternate-file
--- noauto
--- aux othr.txt
other content
--- input
one
two
--- commands
e othr.txt
e #
f
q!
--- output
file.txt: unmodified: line 2 of 2 [100%]
### end

### test expand-hash-in-shell-escape
--- noauto
--- aux othr.txt
other content
--- input
one
--- commands
e othr.txt
e #
!echo #
q!
--- output
othr.txt
### end

### test expand-escaped-percent-is-literal
--- input
one
--- commands
!echo \%
--- output
%
### end

### test read-from-command
# r !command inserts the command's output after the addressed line;
# the ! line echoes as the command runs.
--- input
one
two
--- commands
$r !echo hello
--- output
!
--- file
one
two
hello
### end

### test read-from-command-at-zero
--- input
one
--- commands
0r !echo top
--- output
!
--- file
top
one
### end

### test write-to-command
# w !command pipes the buffer to the command without writing the
# file; the buffer stays modified.
--- noauto
--- input
one
two
--- commands
w !tr a-z A-Z
q!
--- output
ONE
TWO
--- file
one
two
### end

### test write-range-to-command
--- noauto
--- input
one
two
three
--- commands
2,3w !wc -l | tr -d ' '
q!
--- output
2
--- file
one
two
three
### end

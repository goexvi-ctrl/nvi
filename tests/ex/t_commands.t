# Core ex commands (docs/nvi.md "Ex Commands"): text input, delete,
# move, copy, join, shift, yank/put, read/write, undo, edit.

### test cmd-append
--- input
one
two
--- commands
1a
new line
.
--- file
one
new line
two
### end

### test cmd-append-at-zero
# "0a" appends before the first line.
--- input
one
two
--- commands
0a
top
.
--- file
top
one
two
### end

### test cmd-insert
--- input
one
two
--- commands
2i
middle
.
--- file
one
middle
two
### end

### test cmd-change
--- input
one
two
three
--- commands
2c
TWO
also two
.
--- file
one
TWO
also two
three
### end

### test cmd-delete-range
--- input
one
two
three
four
--- commands
2,3d
--- file
one
four
### end

### test cmd-delete-count
# "1d 2" deletes two lines starting at line 1.
--- input
one
two
three
--- commands
1d 2
--- file
three
### end

### test cmd-delete-into-buffer-and-put
--- input
one
two
three
--- commands
1d a
$pu a
--- file
two
three
one
### end

### test cmd-move
--- input
one
two
three
--- commands
1m$
--- file
two
three
one
### end

### test cmd-copy-t
--- input
one
two
--- commands
1t$
--- file
one
two
one
### end

### test cmd-copy-co
--- input
one
two
--- commands
1co0
--- file
one
one
two
### end

### test cmd-join-default-spacing
# Join inserts a single space between joined lines, replacing any
# leading whitespace on the following line.
--- input
one
   two
--- commands
1,2j
--- file
one two
### end

### test cmd-join-after-period
# Join inserts two spaces after a line ending in a period.
--- input
end.
next
--- commands
1,2j
--- file
end.  next
### end

### test cmd-join-bang-preserves-whitespace
# j! joins without whitespace adjustment.
--- input
one
   two
--- commands
1,2j!
--- file
one   two
### end

### test cmd-shift-right
# With shiftwidth=8 and tabstop=8 (the defaults), > prepends a tab.
--- input
one
--- commands
1>
--- file
	one
### end

### test cmd-shift-right-sw4
# With shiftwidth=4, > prepends four spaces (less than a tabstop).
--- input
one
--- commands
set shiftwidth=4
1>
--- file
    one
### end

### test cmd-shift-left
--- input
	one
--- commands
1<
--- file
one
### end

### test cmd-shift-left-unindented-is-noop
# Shifting an unindented line left does nothing (and is not an error).
--- input
one
--- commands
1<
--- file
one
### end

### test cmd-yank-put
--- input
one
two
--- commands
1ya a
$pu a
--- file
one
two
one
### end

### test cmd-put-after-zero
--- input
one
two
--- commands
$ya b
0pu b
--- file
two
one
two
### end

### test cmd-write-part-and-read
# Write lines 1-2 to a second file, then read it back at the end.
--- input
one
two
three
--- commands
1,2w part.txt
$r part.txt
--- file
one
two
three
one
two
### end

### test cmd-write-append
# w >> appends to an existing file.
--- input
one
two
--- commands
1w! part.txt
2w >>part.txt
%d
r part.txt
--- file
one
two
### end

### test cmd-undo
--- input
one
two
--- commands
1d
u
--- file
one
two
### end

### test cmd-undo-toggles
# A second undo undoes the undo (historic toggle behavior).
--- input
one
two
--- commands
1d
u
u
--- file
two
### end

### test cmd-edit-bang-discards
# e! rereads the file, discarding unwritten changes.
--- input
one
two
--- commands
s/two/XXX/
e!
--- file
one
two
### end

### test cmd-filter-through-sort
# The ! command filters addressed lines through a shell command.
--- input
banana
apple
cherry
--- commands
1,3!sort
--- file
apple
banana
cherry
### end

### test cmd-equals-defaults-to-last-line
# "=" with no address reports the last line number.
--- input
one
two
three
--- commands
=
--- output
3
### end

### test cmd-list
# The l command displays tabs as ^I and marks end of line with $.
--- input
one	x
--- commands
1l
--- output
one^Ix$
### end

### test cmd-number
# The # / nu command prints lines with their line numbers.
--- input
one
two
--- commands
2nu
--- output
     2  two
### end

# Multiple-file editing: next, args, previous, rewind, wn
# (docs/nvi.md "Ex Commands").  The aux sections provide the second
# and third files; these tests quit with q! themselves so the
# runner's automatic w! does not rewrite an aux file.

### test next-with-new-arg-list
# next with file arguments replaces the argument list.
--- noauto
--- aux b.txt
bee line
--- aux c.txt
cee line
--- input
one
--- commands
next! b.txt c.txt
1p
q!
--- output
bee line
### end

### test args-shows-current-in-brackets
--- noauto
--- aux b.txt
bee line
--- aux c.txt
cee line
--- input
one
--- commands
next! b.txt c.txt
args
q!
--- output
[b.txt] c.txt
### end

### test next-then-rewind
--- noauto
--- aux b.txt
bee line
--- aux c.txt
cee line
--- input
one
--- commands
next! b.txt c.txt
next!
1p
rewind!
1p
q!
--- output
cee line
bee line
### end

### test previous-moves-back
--- noauto
--- aux b.txt
bee line
--- aux c.txt
cee line
--- input
one
--- commands
next! b.txt c.txt
next!
prev!
1p
q!
--- output
bee line
### end

### test next-past-last-is-error
--- noauto
--- aux b.txt
bee line
--- input
one
--- commands
next! b.txt
next!
q!
--- exit 1
### end

### test next-modified-without-bang-is-error
# Moving to the next file with an unwritten change requires !
# (or autowrite).
--- noauto
--- aux b.txt
bee line
--- aux c.txt
cee line
--- input
one
--- commands
next! b.txt c.txt
1s/bee/BEE/
next
q!
--- exit 1
### end

### test wn-writes-and-moves-on
# wn writes the current file and edits the next one.
--- noauto
--- aux b.txt
bee line
--- aux c.txt
cee line
--- input
one
--- commands
next! b.txt c.txt
1s/bee/BEE/
wn
1p
r b.txt
2p
q!
--- output
cee line
BEE line
### end

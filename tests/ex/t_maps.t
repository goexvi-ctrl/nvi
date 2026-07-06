# map, unmap, abbreviate, unabbreviate from ex (docs/nvi.md "Ex
# Commands").  Map expansion itself happens in vi mode; what ex can
# test is definition, listing, and removal.

### test map-define-and-list
--- input
one
--- commands
map x dd
map
--- output
x     dd
### end

### test map-bang-define-and-list
# map! defines an input-mode mapping; it is listed separately.
--- input
one
--- commands
map! q xyz
map!
--- output
q     xyz
### end

### test map-list-is-separate-from-map-bang
# A command-mode map does not appear in the input-mode list.  The
# input list is empty, which in batch mode prints nothing.
--- input
one
--- commands
map x dd
map!
--- output
### end

### test unmap-removes
--- input
one
--- commands
map x dd
unmap x
map
--- output
### end

### test unmap-unknown-is-error
# The failure is silent in batch mode: nonzero exit, no message.
--- input
one
--- commands
unmap zz
--- output
--- exit 1
--- file
one
### end

### test abbreviate-not-expanded-in-ex-input
# Abbreviations expand during vi text input; ex-mode text input
# leaves them alone.
--- input
one
--- commands
ab teh the
0a
teh cat
.
--- file
teh cat
one
### end

### test unabbreviate-unknown-is-error
--- input
one
--- commands
unab zzz
--- exit 1
--- file
one
### end

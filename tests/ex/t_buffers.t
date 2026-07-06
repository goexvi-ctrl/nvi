# Buffers from ex: delete/yank into named buffers, put, executing
# buffer contents with @ and * (docs/nvi.md "Ex Commands" and the
# "buffer" entry in "General Editor Description").

### test buffer-delete-and-put
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

### test buffer-yank-overwrites
# A second yank into the same lowercase buffer replaces its contents.
--- input
one
two
--- commands
1ya a
2ya a
$pu a
--- file
one
two
two
### end

### test buffer-uppercase-appends
# Yanking into the uppercase form of a buffer name appends to it.
--- input
one
two
--- commands
1ya a
2ya A
$pu a
--- file
one
two
one
two
### end

### test buffer-put-twice
# Putting does not consume the buffer.
--- input
one
two
--- commands
1ya a
$pu a
$pu a
--- file
one
two
one
one
### end

### test buffer-at-executes-commands
# @ buffer executes the buffer contents as ex commands: put a
# command into the file, delete it into buffer b, execute it.
--- input
one
two
three
--- commands
0a
2s/two/TWO/
.
1d b
@ b
--- file
one
TWO
three
### end

### test buffer-star-is-at
# * is a synonym for @.
--- input
one
two
three
--- commands
0a
2s/two/TWO/
.
1d b
* b
--- file
one
TWO
three
### end

### test buffer-at-failure-discards-pending
# When a command executed from a buffer fails, the pending buffer
# commands are discarded with a second message, and batch mode then
# discards the rest of the script.  (The reported line number counts
# only command lines: text-input lines and the closing . are not
# counted, so "@ b" here is script line 3.)
--- input
one
two
--- commands
0a
s/nomatch/x/
.
1d b
@ b
--- output
script, 3: No match found
script, 3: Ex command failed: pending commands discarded
--- exit 1
--- file
one
two
### end

### test buffer-numeric-not-addressable-from-ex
# The numeric buffers 1-9 exist for vi; in an ex command position a
# digit parses as a count, so "pu 2" is a usage error.
--- input
one
two
--- commands
1d
pu 2
--- output
script, 2: Usage: [line] pu[t] [buffer]
--- exit 1
--- file
one
two
### end

# Ex addressing (docs/nvi.md "Ex Addressing").
#
# Empirical notes encoded below:
# - In ex mode the initial current line is the last line of the file.
# - A search address (/RE/ or ?RE?) starts from the cursor position
#   in the current line, vi-style: a match in the current line after
#   column 0 resolves to the current line itself.

### test addr-absolute-line
--- input
one
two
three
--- commands
2d
--- file
one
three
### end

### test addr-dollar-is-last-line
--- input
one
two
three
--- commands
$d
--- file
one
two
### end

### test addr-initial-dot-is-last-line
--- input
one
two
three
--- commands
.=
--- output
3
--- file
one
two
three
### end

### test addr-bare-address-prints
--- input
one
two
three
--- commands
2
--- output
two
### end

### test addr-range-comma
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

### test addr-percent-is-whole-file
--- input
one
two
three
--- commands
%d
--- file
### end

### test addr-relative-offsets
# $-1 is the next-to-last line; 1+1 is line two.
--- input
one
two
three
four
--- commands
$-1=
1+1=
--- output
3
2
### end

### test addr-offset-deletes
--- input
one
two
three
four
--- commands
$-1d
--- file
one
two
four
### end

### test addr-zero-move-target
# Address 0 is valid as a destination: move line 3 above line 1.
--- input
one
two
three
--- commands
3m0
--- file
three
one
two
### end

### test addr-search-forward
--- input
one
two
three
four
--- commands
/two/d
--- file
one
three
four
### end

### test addr-search-backward
# Initial current line is the last line; ?RE? searches backward.
--- input
one
two
three
four
--- commands
?two?d
--- file
one
three
four
### end

### test addr-search-wraps
# From the last line, a forward search wraps to the top of the file
# (wrapscan is on by default).
--- input
one
two
three
--- commands
/one/=
--- output
1
### end

### test addr-search-starts-at-cursor
# nvi resolves a search address from the cursor position within the
# current line: from line 2 column 0, /Y/ matches the Y in the
# current line itself (column 1), not the next Y-line.
--- input
aX
bY
cX
dY
--- commands
2
/Y/=
--- output
bY
2
### end

### test addr-search-skips-nonmatching-dot
# Same setup, but the current line has no match, so the search moves
# on to the next matching line.
--- input
aX
bY
cX
dY
--- commands
2
/X/=
--- output
bY
3
### end

### test addr-mark
--- input
one
two
three
four
--- commands
2mark a
'ad
--- file
one
three
four
### end

### test addr-mark-k-form
--- input
one
two
three
--- commands
2k a
'ad
--- file
one
three
### end

### test addr-backwards-range-is-error
# Interactively nvi offers to swap the addresses; in batch mode the
# command fails, the rest of the script is discarded, and the editor
# exits nonzero.
--- input
one
two
three
--- commands
3,1d
--- output
script, 1: The second address is smaller than the first
--- exit 1
--- file
one
two
three
### end

### test addr-past-eof-is-error
# Addressing past end of file discards the rest of the script and
# exits nonzero; the appended w! never runs.
--- input
one
two
--- commands
9p
--- exit 1
--- file
one
two
### end

### test addr-semicolon-sets-dot
# With a semicolon the first address becomes the current line before
# the second is resolved: from line 1, /X/ finds line 3 (cX), then
# the second /X/ resolves from there.  With the file's Xs on lines
# 3 and 5, "1;/X/,/X/" is not needed -- keep it simple and verify
# dot movement with =.
--- input
aY
bY
cX
dY
eX
--- commands
1;/X/=
--- output
3
### end

### test addr-comma-does-not-set-dot
# With a comma both searches resolve from the original current line,
# so both find line 3 and the range is a single line.
--- input
aY
bY
cX
dY
eX
--- commands
1
/X/,/X/p
--- output
aY
cX
### end

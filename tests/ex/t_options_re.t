# Options that change RE interpretation (docs/nvi.md section 9 and
# the magic, extended, ignorecase, and wrapscan entries in "Set
# Options").

### test nomagic-dot-is-literal
# With nomagic, . loses its wildcard meaning: a.c matches only the
# literal string "a.c".
--- input
axc
a.c
--- commands
set nomagic
%s/a.c/X/
--- file
axc
X
### end

### test nomagic-escaped-dot-is-wildcard
# With nomagic, escaping restores the special meaning.
--- input
axc
--- commands
set nomagic
s/a\.c/X/
--- file
X
### end

### test nomagic-star-is-literal
--- input
ab*c
--- commands
set nomagic
s/b*c/X/
--- file
aX
### end

### test nomagic-bracket-is-literal
--- input
x[a]y
--- commands
set nomagic
s/[a]/X/
--- file
xXy
### end

### test nomagic-anchors-still-work
# ^ at the start and $ at the end stay special under nomagic.
--- input
aba
--- commands
set nomagic
s/^a/X/
s/a$/Y/
--- file
XbY
### end

### test nomagic-ampersand-is-literal
# With nomagic, & in the replacement is an ordinary character.
--- input
one
--- commands
set nomagic
s/one/x&y/
--- file
x&y
### end

### test nomagic-escaped-ampersand-is-match
--- input
one
--- commands
set nomagic
s/one/x\&y/
--- file
xoney
### end

### test extended-groups-and-plus
# The extended option switches to egrep-style EREs.
--- input
ababx
--- commands
set extended
s/(ab)+/X/
--- file
Xx
### end

### test extended-alternation
--- input
one
two
three
--- commands
set extended
g/one|three/d
--- file
two
### end

### test extended-question
--- input
color
colour
--- commands
set extended
%s/colou?r/X/
--- file
X
X
### end

### test ignorecase-search
--- input
one
TWO
--- commands
set ignorecase
/two/d
--- file
one
### end

### test ignorecase-substitute
--- input
One oNe
--- commands
set ignorecase
s/one/X/g
--- file
X X
### end

### test noignorecase-is-default
--- input
one
TWO
two
--- commands
/two/d
--- file
one
TWO
### end

### test nowrapscan-search-fails
# With nowrapscan a forward search from the last line cannot wrap;
# the failure discards the rest of the batch script.
--- input
one
two
--- commands
set nowrapscan
$
/one/d
--- output
two
script, 3: Reached end-of-file without finding the pattern
--- exit 1
--- file
one
two
### end

### test option-query
# "set option?" reports the current value.
--- input
x
--- commands
set shiftwidth?
--- output
shiftwidth=8
### end

### test option-abbreviation
# Abbreviated names are accepted, but the report always uses the
# full option name.
--- input
x
--- commands
set sw=4
set sw?
--- output
shiftwidth=4
### end

### test option-boolean-query
--- input
x
--- commands
set magic?
--- output
magic
### end

### test option-unknown-is-error
--- input
x
--- commands
set nosuchoption
--- exit 1
--- file
x
### end

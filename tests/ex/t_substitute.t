# The substitute command and replacement strings (docs/nvi.md
# section 9, "Regular Expressions and Replacement Strings", and the
# substitute entry under "Ex Commands").

### test subst-first-occurrence-only
--- input
aaa
--- commands
s/a/X/
--- file
Xaa
### end

### test subst-global-flag
--- input
aaa
--- commands
s/a/X/g
--- file
XXX
### end

### test subst-line-range
--- input
aaa
aaa
aaa
--- commands
1,2s/a/X/
--- file
Xaa
Xaa
aaa
### end

### test subst-count-applies-to-following-lines
# A trailing count makes the substitute act on that many lines
# starting at the addressed line.
--- input
aaa
aaa
aaa
--- commands
1s/a/X/ 2
--- file
Xaa
Xaa
aaa
### end

### test subst-print-flag
--- input
one
--- commands
s/one/two/p
--- output
two
--- file
two
### end

### test subst-alternate-delimiter
--- input
a/b
--- commands
s#a/b#c/d#
--- file
c/d
### end

### test subst-ampersand-is-match
--- input
abc
--- commands
s/b/[&]/
--- file
a[b]c
### end

### test subst-escaped-ampersand-is-literal
--- input
abc
--- commands
s/b/\&/
--- file
a&c
### end

### test subst-backreference-groups
# The docs/nvi.md section 9 example: delete the abc and def around
# the middle.
--- input
abcXYZdef
--- commands
s/abc\(.*\)def/\1/
--- file
XYZ
### end

### test subst-numbered-groups
--- input
one two
--- commands
s/\(one\) \(two\)/\2 \1/
--- file
two one
### end

### test subst-tilde-is-previous-replacement
# ~ in a replacement stands for the replacement part of the previous
# substitute.
--- input
one
xy
--- commands
1s/one/AAA/
2s/xy/x~y/
--- file
AAA
xAAAy
### end

### test subst-tilde-in-re-matches-previous-replacement
# ~ in an RE matches the replacement part of the last substitute
# (docs/nvi.md section 9 item 4).
--- input
one
xAAAy
--- commands
1s/one/AAA/
2s/~/BBB/
--- file
AAA
xBBBy
### end

### test subst-percent-reuses-replacement
# A replacement of exactly % repeats the previous replacement text.
--- input
one
two
--- commands
1s/one/AAA/
2s/two/%/
--- file
AAA
AAA
### end

### test subst-empty-re-reuses-last-re
# An empty RE means the last RE used; a search sets it too.
--- input
one
two
--- commands
/two/
s//TWO/
--- output
two
--- file
one
TWO
### end

### test subst-case-upper-one
# \u uppercases the next character of the replacement.
--- input
hello
--- commands
s/hello/\u&/
--- file
Hello
### end

### test subst-case-lower-one
--- input
HELLO
--- commands
s/HELLO/\l&/
--- file
hELLO
### end

### test subst-case-upper-string
# The docs/nvi.md example: s/abc/\U&/ replaces abc with ABC.
--- input
abc
--- commands
s/abc/\U&/
--- file
ABC
### end

### test subst-case-upper-until-E
# \U converts until \E (or \e) or the end of the replacement.
--- input
one two
--- commands
s/\(one\) \(two\)/\U\1\E \2/
--- file
ONE two
### end

### test subst-case-lower-string
--- input
ABC DEF
--- commands
s/\(ABC\) \(DEF\)/\L\1\E \2/
--- file
abc DEF
### end

### test subst-bare-s-repeats
# A bare s repeats the last substitute on the current line.
--- input
aaa
aaa
--- commands
1s/a/X/
2
s
--- output
aaa
--- file
Xaa
Xaa
### end

### test subst-ampersand-command-repeats
# The & command also repeats the last substitute.
--- input
aaa
aaa
--- commands
1s/a/X/
2&
--- file
Xaa
Xaa
### end

### test subst-split-line
# A backslash-escaped newline in the replacement splits the line.
--- input
one two three
--- commands
s/two/X\
Y/
--- file
one X
Y three
### end

### test subst-tilde-command-uses-last-re
# The ex ~ command re-substitutes using the last RE from anywhere
# (here, a search) with the last replacement; & would have reused
# the substitute's own RE instead.
--- input
one
two
--- commands
1s/one/X/
/two/
~
--- output
two
--- file
X
X
### end

### test subst-global-flag-with-count
# g and a trailing count combine: every occurrence on two lines.
--- input
aaa
aaa
aaa
--- commands
1s/a/X/g 2
--- file
XXX
XXX
aaa
### end

### test subst-no-match-is-error
--- input
one
--- commands
s/zzz/x/
--- output
script, 1: No match found
--- exit 1
--- file
one
### end

### test subst-empty-replacement
--- input
axbxc
--- commands
s/x//g
--- file
abc
### end

### test subst-anchors
--- input
aba
--- commands
s/^a/X/
s/a$/Y/
--- file
XbY
### end

### test subst-word-boundaries
# \< and \> match the beginning and end of a word (docs/nvi.md
# section 9 items 2 and 3).
--- input
foo foobar barfoo
--- commands
s/\<foo\>/X/
--- file
X foobar barfoo
### end

### test subst-word-boundary-end
--- input
foobar barfoo
--- commands
s/foo\>/X/
--- file
foobar barX
### end

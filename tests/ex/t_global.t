# The global commands: g/v (docs/nvi.md "Ex Commands", global and v).

### test global-delete
--- input
keep one
drop x
keep two
drop y
--- commands
g/drop/d
--- file
keep one
keep two
### end

### test global-default-command-is-print
--- input
one
two x
three
four x
--- commands
g/x/
--- output
two x
four x
--- file
one
two x
three
four x
### end

### test global-print
--- input
one
two x
three x
--- commands
g/x/p
--- output
two x
three x
### end

### test global-substitute-empty-re-reuses
# An empty RE in the attached substitute reuses the global's RE.
--- input
foo one
bar two
foo three
--- commands
g/foo/s//FOO/
--- file
FOO one
bar two
FOO three
### end

### test global-move-to-zero-reverses
# The classic idiom: g/^/m0 reverses the file.
--- input
one
two
three
--- commands
g/^/m0
--- file
three
two
one
### end

### test global-delete-all
--- input
one
two
three
--- commands
g/./d
--- file
### end

### test v-inverse-delete
--- input
keep x
drop
keep y
--- commands
v/keep/d
--- file
keep x
keep y
### end

### test g-bang-same-as-v
--- input
keep x
drop
keep y
--- commands
g!/keep/d
--- file
keep x
keep y
### end

### test global-copy
--- input
one x
two
--- commands
g/x/t$
--- file
one x
two
one x
### end

### test global-no-match-is-silent-noop
# Unlike substitute, a global with no matching lines is not an
# error: nothing is printed and the script continues.
--- input
one
two
--- commands
g/zzz/d
1p
--- output
one
--- file
one
two
### end

### test global-range
# A global restricted to an address range only affects those lines.
--- input
a x
b x
c x
d x
--- commands
2,3g/x/d
--- file
a x
d x
### end

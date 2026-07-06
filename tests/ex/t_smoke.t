# Sanity checks for the harness itself: if these fail, suspect the
# runner or the build, not nvi.

### test smoke-noop
--- input
one
two
--- commands
--- file
one
two
### end

### test smoke-substitute
--- input
hello
world
--- commands
%s/hello/goodbye/
--- file
goodbye
world
### end

### test smoke-print
--- input
one
two
three
--- commands
2p
--- output
two
--- file
one
two
three
### end

### test smoke-error-aborts-script
# A failing command in batch mode discards the rest of the script and
# exits 1; the appended w! never runs so the file is unchanged.
--- input
one
two
--- commands
s/nomatch/x/
s/one/ONE/
--- output
script, 1: No match found
--- exit 1
--- file
one
two
### end

### test smoke-noauto-no-write
# With noauto the buffer is never written; the change must not appear
# in the file.  The explicit q! is required or nvi would hang waiting
# for more input (stdin EOF ends the session with exit 1, so quit).
--- noauto
--- input
one
--- commands
s/one/ONE/
q!
--- file
one
### end

# Command-line invocation (the nvi manual page; docs/nvi.md
# "Startup Information").  These cases replace the default
# "-e -s file.txt" arguments with their own.

### test cli-vi-mode-needs-a-terminal
--- args file.txt
--- input
one
--- commands
q!
--- errout
ex/vi: Vi's standard input and output must be a terminal
--- exit 1
### end

### test cli-s-only-applies-to-ex
--- args -s file.txt
--- input
one
--- commands
q!
--- errout
vi: -s option is only applicable to ex.
--- exit 1
### end

### test cli-bad-flag-shows-usage
--- args -Z file.txt
--- input
one
--- commands
q!
--- errout
vi: illegal option -- Z
usage: ex [-eFRrSsv] [-c command] [-t tag] [-w size] [file ...]
usage: vi [-eFlRrSv] [-c command] [-t tag] [-w size] [file ...]
--- exit 1
### end

### test cli-F-no-longer-supported
# -F (don't copy the file) is accepted but warns and is ignored.
--- args -e -s -F file.txt
--- input
one
--- commands
1p
--- output
one
--- errout
vi: -F option no longer supported.
### end

### test cli-readonly-flag
--- args -R -e -s file.txt
--- noauto
--- input
one
--- commands
1s/one/ONE/
w
q!
--- output
script, 2: Read-only file, not written; use ! to override
--- exit 1
--- file
one
### end

### test cli-c-runs-command-after-load
# The -c command runs once the file is loaded; like ex startup, the
# current line is the last line, so the substitute needs an address.
--- args -e -s -c 1s/one/ONE/ file.txt
--- input
one
two
--- commands
1p
--- output
ONE
--- file
ONE
two
### end

### test cli-plus-form-of-c
--- args -e -s +1s/one/ONE/ file.txt
--- input
one
two
--- commands
1p
--- output
ONE
--- file
ONE
two
### end

### test cli-c-error-is-not-fatal
# A failing -c command reports with a "-c option" prefix, but the
# session still starts; the stdin script is discarded, however, and
# the file is left untouched.
--- args -e -s -c s/nomatch/x/ file.txt
--- noauto
--- input
one
--- commands
1s/one/ONE/
w!
q!
--- output
-c option, 1: No match found
--- file
one
### end

### test cli-w-sets-window
--- args -e -s -w 12 file.txt
--- input
one
--- commands
set window?
--- output
window=12
### end

### test cli-t-starts-at-tag
# With -t the session starts in the tag's file; no file argument is
# needed.
--- args -e -s -t afunc
--- noauto
--- aux tags
afunc	tt.c	1
--- aux tt.c
int afunc;
--- input
unused
--- commands
p
f
q!
--- output
int afunc;
tt.c: unmodified: line 1 of 1 [100%]
### end

### test cli-S-sets-secure
--- args -e -s -S file.txt
--- input
one
--- commands
!echo hi
--- output
script, 1: The ! command is not supported when the secure edit option is set
--- exit 1
### end

### test cli-l-lisp-warns
--- args -e -s -l file.txt
--- input
one
--- commands
1p
--- output
The lisp option is not implemented
one
### end

### test cli-multiple-files-form-arg-list
--- args -e -s file.txt second.txt
--- noauto
--- aux second.txt
second line
--- input
one
--- commands
args
next
1p
q!
--- output
[file.txt] second.txt
second line
### end

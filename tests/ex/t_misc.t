# Remaining batch-testable ex commands: display, exusage, viusage,
# recover (docs/nvi.md "Ex Commands").
#
# Not tested here because they cannot run deterministically in batch
# mode: shell, stop, suspend (interactive), script (spawns a shell
# on the session), and the screen commands bg, fg, resize, vsplit
# (vi screen mode only).

### test display-buffers
--- input
one
two
--- commands
1d a
display buffers
--- output
********** a (line mode)
one
********** default buffer (line mode)
one
--- file
two
### end

### test display-tags-stack
--- noauto
--- aux tags
afunc	target.c	1
--- aux target.c
int afunc;
--- input
one
--- commands
tag afunc
display tags
q!
--- output
 1                          target.c*    afunc
 2                          file.txt*
### end

### test display-screens-empty-in-ex
# No background screens exist in a batch ex session; the display is
# empty.
--- input
one
--- commands
display screens
--- output
### end

### test exusage-shows-command-usage
--- input
one
--- commands
exusage append
--- output
Command: append input to a line
  Usage: [line] a[ppend][!]
### end

### test viusage-shows-key-usage
--- input
one
--- commands
viusage J
--- output
  Key: J join lines
Usage: [count]J
### end

### test recover-after-preserve
# docs/nvi.md, preserve: "Save the file in a form that can later be
# recovered using the ex recover command."  It cannot: preserve
# leaves a zero-length backup in the recovery area, and recover
# (same session or a fresh one, or vi -r) then loads an empty
# buffer instead of the preserved text.
--- xfail preserve writes an empty backup; recover restores nothing
--- input
one
two
--- commands
1s/one/ONE/
preserve
recover! file.txt
%p
--- output
ONE
two
### end

### test recover-nothing-preserved-is-error
# With no preserved copy the failure is silent in batch mode:
# nonzero exit, no message.
--- input
one
--- commands
recover file.txt
--- output
--- exit 1
--- file
one
### end

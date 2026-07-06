# Golden snapshots of self-describing output: the set all option
# listing (locks in every option default under the batch
# environment: TERM=dumb, no TMPDIR, 24x80, and this build's
# patched recdir), and per-command exusage/viusage entries.
# If a default changes on purpose, regenerate the set all section
# by running the commands under env -i HOME=<dir> TERM=dumb
# LC_ALL=C PATH=/usr/bin:/bin.

### test set-all-defaults
--- noauto
--- input
one
--- commands
set all
q!
--- output
noaltwerase     noextended      mesg            report=5        term="dumb"
noautoindent    filec=""        nomodeline      noruler         noterse
noautoprint     flash           msgcat="./"     scroll=11       notildeop
noautowrite     hardtabs=0      noprint=""      nosearchincr    timeout
backup=""       noiclower       nonumber        nosecure        nottywerase
nobeautify      noignorecase    nooctal         shiftwidth=8    noverbose
cdpath=":"      keytime=6       open            noshowmatch     nowarn
cedit=""        noleftright     optimize        noshowmode      window=23
columns=80      lines=24        path=""         sidescroll=16   nowindowname
nocomment       nolisp          print=""        noslowopen      wraplen=0
noedcompatible  nolist          noprompt        nosourceany     wrapmargin=0
escapetime=1    lock            noreadonly      tabstop=8       wrapscan
noerrorbells    magic           noredraw        taglength=0     nowriteany
noexrc          matchtime=7     remap           tags="tags"
directory="/tmp"
paragraphs="IPLPPPQPP LIpplpipbp"
recdir="vi.recover"
sections="NHSHH HUnhsh"
shell="/bin/sh"
shellmeta="~{[*?$`'"\"
### end

### test exusage-single-command
--- noauto
--- input
one
--- commands
exusage d
q!
--- output
Command: delete lines from the file
  Usage: [line [,line]] d[elete][flags] [buffer] [count] [flags]
### end

### test viusage-single-key
--- noauto
--- input
one
--- commands
viusage w
q!
--- output
  Key: w move to next word
Usage: [count]w
### end

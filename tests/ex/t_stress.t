# Size sanity: a large line count and a long single line exercise
# the buffer and search paths past the one-screen cases the rest of
# the suite uses.  The oversized files are generated with shell
# commands and read with e! rather than being spelled out inline.

### test stress-ten-thousand-lines
# Build a 10000-line file with a marker near the end, then confirm
# the search finds the marker at the right line and the last line
# number is 10000.
--- noauto
--- input
seed
--- commands
!awk 'BEGIN{for(i=1;i<=10000;i++)print (i==9999)?"MARKER":i}' > big.txt
e! big.txt
/^MARKER$/=
$=
q!
--- output
9999
10000
### end

### test stress-long-line-substitute
# A single line of 5000 'a' characters ending in Z: substitute the
# trailing marker and confirm the byte length is unchanged (5000
# plus the marker plus the newline).
--- noauto
--- input
seed
--- commands
!awk 'BEGIN{for(i=0;i<5000;i++)printf "a"; printf "Z\n"}' > long.txt
e! long.txt
s/Z$/!/
w
!wc -c < long.txt | tr -d ' '
q!
--- output
5002
### end

### test stress-global-on-many-lines
# A global command touching every line of a 5000-line file: number
# each line's content, then spot-check with a couple of prints.
--- noauto
--- input
seed
--- commands
!yes x | sed 5000q > many.txt
e! many.txt
%s/x/y/
$=
5000p
q!
--- output
5000
y
### end

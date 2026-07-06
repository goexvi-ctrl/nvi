# Tags: tag, tagpop, tagtop (docs/nvi.md "Tags, Tag Stacks, and
# Cscope").  The aux sections fabricate a tags file and the target
# sources.  Tests quit with q! so the runner's automatic w! does not
# rewrite a target file.

### test tag-jumps-to-pattern
--- noauto
--- aux tags
afunc	target.c	/^int afunc/
bfunc	other.c	/^int bfunc/
--- aux target.c
int afunc(void)
{
}
--- aux other.c
static int x;
int bfunc(void)
--- input
one
--- commands
tag afunc
p
f
q!
--- output
int afunc(void)
target.c: unmodified: line 1 of 3 [33%]
### end

### test tag-line-number-form
# A tags entry may use a line number instead of a search pattern.
--- noauto
--- aux tags
second	target.c	2
--- aux target.c
first line
second line
--- input
one
--- commands
tag second
p
q!
--- output
second line
### end

### test tagpop-returns
--- noauto
--- aux tags
afunc	target.c	/^int afunc/
--- aux target.c
int afunc(void)
--- input
one
--- commands
tag afunc
tagpop
f
q!
--- output
file.txt: unmodified: line 1 of 1 [100%]
### end

### test tagtop-returns-to-start
# After chained tag jumps, tagtop pops all the way back.
--- noauto
--- aux tags
afunc	target.c	/^int afunc/
bfunc	other.c	/^int bfunc/
--- aux target.c
int afunc(void)
--- aux other.c
int bfunc(void)
--- input
one
--- commands
tag afunc
tag bfunc
tagtop
f
q!
--- output
file.txt: unmodified: line 1 of 1 [100%]
### end

### test tag-not-found-is-error
--- noauto
--- aux tags
afunc	target.c	/^int afunc/
--- aux target.c
int afunc(void)
--- input
one
--- commands
tag nosuchtag
q!
--- output
script, 1: nosuchtag: tag not found
--- exit 1
### end

### test tagpop-empty-stack-is-error
--- noauto
--- aux tags
afunc	target.c	/^int afunc/
--- aux target.c
int afunc(void)
--- input
one
--- commands
tagpop
q!
--- exit 1
### end

### test tagnext-tagprev-duplicate-entries
# With two entries for the same tag, tagnext moves to the next
# match in the group and tagprev returns.
--- noauto
--- aux tags
afunc	t1.c	/^int afunc/
afunc	t2.c	/^int afunc/
--- aux t1.c
int afunc()
body1
--- aux t2.c
int afunc()
body2
--- commands
ta afunc
f
tagnext
f
tagprev
f
q!
--- output
t1.c: unmodified: line 1 of 2 [50%]
t2.c: unmodified: line 1 of 2 [50%]
t1.c: unmodified: line 1 of 2 [50%]
### end

### test tagnext-past-last-entry
--- noauto
--- aux tags
afunc	t1.c	/^int afunc/
afunc	t2.c	/^int afunc/
--- aux t1.c
int afunc()
--- aux t2.c
int afunc()
--- commands
ta afunc
tagnext
tagnext
--- output
script, 3: Already at the last tag of this group
--- exit 1
### end

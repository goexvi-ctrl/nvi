#!/bin/sh
# run.sh - run nvi ex-mode functional test batteries.
#
# Usage: sh run.sh [-v] file.t [file.t ...]
#
# Each .t file contains test cases in the line-oriented format
# described in tests/README.md.  Every case runs in its own scratch
# directory with a minimal environment (no EXINIT, HOME pointing at
# the scratch directory) so no user startup files leak in.  The
# oracle writes its recovery files to ./vi.recover, which lands in
# the scratch directory too.
#
# Exit status: 0 if every case passed (or was an expected failure),
# 1 otherwise.

TESTDIR=$(cd "$(dirname "$0")" && pwd)
NVI=${NVI:-$(cd "$TESTDIR/../.." && pwd)/build.unix/vi}

if [ ! -x "$NVI" ]; then
	echo "run.sh: nvi binary not found or not executable: $NVI" >&2
	echo "run.sh: build it first (make in build.unix) or set NVI" >&2
	exit 1
fi

VERBOSE=0
if [ "$1" = "-v" ]; then
	VERBOSE=1
	shift
fi

if [ $# -eq 0 ]; then
	echo "usage: sh run.sh [-v] file.t ..." >&2
	exit 1
fi

WORK=$(mktemp -d "${TMPDIR:-/tmp}/nvi-ex-tests.XXXXXX") || exit 1
trap 'rm -rf "$WORK"' 0 1 2 15

npass=0 nfail=0 nxfail=0 nxpass=0

# reset_case: clear per-case state before parsing the next case.
reset_case() {
	name=
	section=
	has_input=0 has_file=0 has_output=0 has_errout=0
	want_exit=0
	xfail=
	noauto=0
	viargs=
	casedir=
}

# begin_case <name>: set up the scratch directory for a new case.
begin_case() {
	reset_case
	name=$1
	safe=$(printf '%s' "$name" | tr -c 'A-Za-z0-9_.-' '_')
	casedir=$WORK/$safe
	n=2
	while [ -e "$casedir" ]; do
		casedir=$WORK/$safe.$n
		n=$((n + 1))
	done
	mkdir "$casedir"
	: > "$casedir/input"
	: > "$casedir/commands"
	: > "$casedir/exp_file"
	: > "$casedir/exp_out"
	: > "$casedir/exp_err"
}

# run_case: execute the parsed case and report the result.
run_case() {
	[ -n "$name" ] || return 0

	cp "$casedir/input" "$casedir/file.txt"
	cp "$casedir/commands" "$casedir/cmds"
	if [ "$noauto" -eq 0 ]; then
		printf 'w!\nq!\n' >> "$casedir/cmds"
	fi

	[ -n "$viargs" ] || viargs='-e -s file.txt'
	# The subshell's own standard error is discarded so that the
	# shell's job notice ("Abort trap: 6") stays out of the runner
	# output when a known-crash xfail case dies on a signal; the
	# editor's standard error still lands in stderr.raw.
	(
		cd "$casedir" || exit 125
		# shellcheck disable=SC2086 -- word splitting is the point
		env -i HOME="$casedir" TERM=dumb LC_ALL=C \
		    PATH=/usr/bin:/bin \
		    "$NVI" $viargs \
		    < cmds > stdout.raw 2> stderr.raw
	) 2>/dev/null
	status=$?

	# Strip the line of startup noise nvi emits when stderr is not
	# a terminal (see README.md); with -c commands it is not always
	# the first line.  On stderr, normalize the binary path prefix
	# (the libtool wrapper reports its own location) so usage
	# messages compare stably.
	sed '/^Error: stderr: /d' "$casedir/stdout.raw" \
	    > "$casedir/stdout.got"
	sed 's|^/[^:]*/vi: |vi: |' "$casedir/stderr.raw" \
	    > "$casedir/stderr.got"

	why=
	if [ "$status" -ne "$want_exit" ]; then
		why="exit status $status, expected $want_exit"
	fi
	if [ -z "$why" ] && [ "$has_file" -eq 1 ] &&
	    ! cmp -s "$casedir/exp_file" "$casedir/file.txt"; then
		why="file contents differ"
	fi
	if [ -z "$why" ] && [ "$has_output" -eq 1 ] &&
	    ! cmp -s "$casedir/exp_out" "$casedir/stdout.got"; then
		why="standard output differs"
	fi
	if [ -z "$why" ] && [ "$has_errout" -eq 1 ] &&
	    ! cmp -s "$casedir/exp_err" "$casedir/stderr.got"; then
		why="standard error differs"
	fi

	if [ -n "$xfail" ]; then
		if [ -n "$why" ]; then
			nxfail=$((nxfail + 1))
			[ "$VERBOSE" -eq 1 ] &&
			    echo "XFAIL $file: $name ($xfail)"
		else
			nxpass=$((nxpass + 1))
			echo "XPASS $file: $name"
			echo "    marked xfail ($xfail) but passed;" \
			    "re-examine and remove the marker"
		fi
	elif [ -n "$why" ]; then
		nfail=$((nfail + 1))
		echo "FAIL  $file: $name"
		echo "    $why"
		case $why in
		"file contents differ")
			diff -u "$casedir/exp_file" "$casedir/file.txt" \
			    | sed 's/^/    /' ;;
		"standard output differs")
			diff -u "$casedir/exp_out" "$casedir/stdout.got" \
			    | sed 's/^/    /' ;;
		"standard error differs")
			diff -u "$casedir/exp_err" "$casedir/stderr.got" \
			    | sed 's/^/    /' ;;
		"exit status"*)
			sed 's/^/    stdout: /' "$casedir/stdout.got" ;;
		esac
	else
		npass=$((npass + 1))
		[ "$VERBOSE" -eq 1 ] && echo "PASS  $file: $name"
	fi
	name=
}

for file in "$@"; do
	if [ ! -f "$file" ]; then
		echo "run.sh: no such test file: $file" >&2
		exit 1
	fi
	reset_case
	while IFS= read -r line || [ -n "$line" ]; do
		case $line in
		'### test '*)
			run_case	# catch a missing "### end"
			begin_case "${line#'### test '}"
			;;
		'### end'*)
			run_case
			;;
		'--- input')
			[ -n "$name" ] && { section=input; has_input=1; } ;;
		'--- commands')
			[ -n "$name" ] && section=commands ;;
		'--- file')
			[ -n "$name" ] && { section=exp_file; has_file=1; } ;;
		'--- output')
			[ -n "$name" ] && { section=exp_out; has_output=1; } ;;
		'--- errout')
			[ -n "$name" ] && { section=exp_err; has_errout=1; } ;;
		'--- aux '*)
			# An extra fixture file in the scratch directory
			# (tags file, second buffer, source script, ...).
			aux=${line#'--- aux '}
			case $aux in
			*[!A-Za-z0-9._-]* | '' | file.txt | cmds | input | \
			commands | exp_file | exp_out | stdout* | stderr*)
				echo "run.sh: $file: bad aux name: $aux" >&2
				exit 1 ;;
			esac
			[ -n "$name" ] && { section=$aux; : > "$casedir/$aux"; }
			;;
		'--- exit '*)
			[ -n "$name" ] && { want_exit=${line#'--- exit '}; section=; } ;;
		'--- xfail '*)
			[ -n "$name" ] && { xfail=${line#'--- xfail '}; section=; } ;;
		'--- noauto')
			[ -n "$name" ] && { noauto=1; section=; } ;;
		'--- args '*)
			# Replace the default "-e -s file.txt" invocation
			# arguments (whitespace-split; no quoting).
			[ -n "$name" ] && { viargs=${line#'--- args '}; section=; } ;;
		*)
			if [ -n "$name" ] && [ -n "$section" ]; then
				printf '%s\n' "$line" >> "$casedir/$section"
			fi
			;;
		esac
	done < "$file"
	run_case	# catch a missing "### end" at EOF
done

total=$((npass + nfail + nxfail + nxpass))
echo "ex tests: $total run, $npass passed, $nfail failed," \
    "$nxfail expected failures, $nxpass unexpected passes"

[ "$nfail" -eq 0 ] && [ "$nxpass" -eq 0 ]

#!/bin/sh
# chaos.sh - deterministic crash hunt for nex.
#
# Generates a fixed corpus of ex command sequences from a small
# grammar with a hard-coded seed, runs each in batch mode, and flags
# any that terminate on a signal (exit status >= 128) -- a crash,
# not an ordinary command error.  The corpus is fully determined by
# the seed, so a run reproduces the same sequences every time; this
# is a regression tool, not a live fuzzer.
#
# One crash is already known (see README.md): an explicit print (p)
# command followed later by a bare address line, which auto-prints,
# corrupts the heap and aborts.  A crashing sequence is classified
# as this known bug by re-running it with the print command lines
# removed: if that stops the crash, the print/auto-print bug was the
# cause.  The script fails only on a crash that survives that
# neutralization, i.e. something new.  Fold anything new into a
# proper XFAIL case in the batteries.
#
# Not part of "make ex"; run with "make chaos".  Comparatively slow.

TESTDIR=$(cd "$(dirname "$0")" && pwd)
NVI=${NVI:-$(cd "$TESTDIR/../.." && pwd)/build.unix/vi}

if [ ! -x "$NVI" ]; then
	echo "chaos.sh: nvi binary not found: $NVI" >&2
	exit 1
fi

SEED=${CHAOS_SEED:-1729}
COUNT=${CHAOS_COUNT:-4000}

WORK=$(mktemp -d "${TMPDIR:-/tmp}/nvi-chaos.XXXXXX") || exit 1
trap 'rm -rf "$WORK"' 0 1 2 15

# The grammar: a pool of addresses (which auto-print on their own)
# and a pool of commands.  Each is an individually-valid fragment;
# crashes come from combinations and ordering.  Kept in awk below.
awk -v seed="$SEED" -v n="$COUNT" -v d="$WORK" '
BEGIN {
	na = split("1|$|.|/e/|?e?|\x27a|2,4|%", A, "|");
	nc = split("p|d|y|s/e/E/|s/x*/-/g|m0|t$|ka|>|<|z|=|j|p|/e/|\\/|&|~", C, "|");
	srand(seed);
	for (i = 0; i < n; i++) {
		len = 2 + int(rand() * 3);   # 2..4 fragments
		line = "";
		for (j = 0; j < len; j++) {
			if (j == 0 || rand() < 0.4)
				frag = A[1 + int(rand() * na)];
			else
				frag = C[1 + int(rand() * nc)];
			line = line frag "\n";
		}
		f = sprintf("%s/seq.%06d", d, i);
		printf "%s", line > f;
		close(f);
	}
}'

printf 'one\ntwo\nthree\nfour\nfive\n' > "$WORK/seed.txt"

# crashes <script-file> [neutralize]: run the sequence and return 0
# if it terminated on a signal.  With a second argument, strip lines
# that are a bare print command first (the known-bug neutralization).
crashes() {
	src=$1
	cp "$WORK/seed.txt" "$WORK/file.txt"
	if [ -n "$2" ]; then
		# Drop lines that are just a print command (optionally with a
		# leading address is NOT stripped; only the bare "p" lines the
		# grammar emits as commands).
		grep -v '^p$' "$src" > "$WORK/neut"
		printf 'q!\n' >> "$WORK/neut"
		runsrc=$WORK/neut
	else
		cp "$src" "$WORK/run"
		printf 'q!\n' >> "$WORK/run"
		runsrc=$WORK/run
	fi
	(
		cd "$WORK" || exit 125
		env -i HOME="$WORK" TERM=dumb LC_ALL=C PATH=/usr/bin:/bin \
		    "$NVI" -e -s file.txt < "$runsrc" > /dev/null 2>&1
	) 2>/dev/null
	[ "$?" -ge 128 ]
}

ncrash=0 nknown=0 nrun=0 seen=
for f in "$WORK"/seq.*; do
	[ -f "$f" ] || continue
	nrun=$((nrun + 1))
	crashes "$f" || continue

	if ! crashes "$f" neutralize; then
		nknown=$((nknown + 1))
		continue
	fi

	# Report each distinct crashing body once.
	key=$(tr '\n' '|' < "$f")
	case " $seen " in
	*" $key "*) continue ;;
	esac
	seen="$seen $key"
	ncrash=$((ncrash + 1))
	echo "NEW CRASH (survives print-line removal):"
	sed 's/^/    /' "$f"
done

echo "chaos: $nrun sequences run, $ncrash new crash(es)," \
    "$nknown known-bug crash(es)"
[ "$ncrash" -eq 0 ]

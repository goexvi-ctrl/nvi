#!/bin/sh
# run_db.sh - run the db.1.85 regression battery with the dbtest
# driver built from build.unix objects.
#
# run.test wants ./dbtest in the current directory and scribbles
# t1/t2/t3 command files there (databases go to TMPDIR), so run it
# in a scratch directory.

TESTDIR=$(cd "$(dirname "$0")" && pwd)
DBSRC=$(cd "$TESTDIR/../../../db.1.85/test" && pwd)

WORK=$(mktemp -d "${TMPDIR:-/tmp}/nvi-db-tests.XXXXXX") || exit 1
trap 'rm -rf "$WORK"' 0 1 2 15

cp "$TESTDIR/dbtest" "$WORK/dbtest" || exit 1
cd "$WORK" || exit 1
TMPDIR=$WORK sh "$DBSRC/run.test"
status=$?
if [ "$status" -eq 0 ]; then
	echo "db tests: all batteries passed"
else
	echo "db tests: FAILED (exit $status)"
fi
exit $status

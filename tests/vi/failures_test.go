// Failed vi commands (docs/nvi.md "Vi Commands"): a motion that
// cannot be satisfied fails, rings the terminal bell, and leaves
// the cursor and the text unchanged.  The failures print no message
// text on this terminal, so the assertions are the bell and that
// nothing moved.
package nvitests

import (
	"strings"
	"testing"
	"time"

	"github.com/goexvi-ctrl/goterm"
)

// settleCursor sends the keys for a command that must fail, waits
// for the screen to go quiet, and asserts that the bell rang and
// the cursor stayed at the given position.
func settleCursor(t *testing.T, term *goterm.Term,
	keys string, row, col int, what string) {
	t.Helper()
	term.ClearBell()
	send(term, keys)
	term.WaitQuiet(50*time.Millisecond, screenTimeout)
	if r, c := term.Cursor(); r != row || c != col {
		t.Fatalf("%s: cursor %d,%d, want %d,%d; screen:\n%s",
			what, r, c, row, col, strings.Join(term.Dump(), "\n"))
	}
	if term.Bell == 0 {
		t.Fatalf("%s: no bell for the failed command", what)
	}
}

func TestFailedCharacterFind(t *testing.T) {
	term := startVi(t, "abc def\n")

	send(term, "4l")
	waitCursor(t, term, 0, 4)
	settleCursor(t, term, "fZ", 0, 4, "fZ with no Z on the line")
}

func TestFailedMatchNoBracket(t *testing.T) {
	term := startVi(t, "abc def\n")

	settleCursor(t, term, "%", 0, 0, "% with no bracket on the line")
}

func TestGotoPastEndOfFileFails(t *testing.T) {
	// A G count past the end of the file is an error, not a clamp.
	term := startVi(t, "one\ntwo\nthree\n")

	settleCursor(t, term, "999G", 0, 0, "999G")
}

func TestScreenPositionCountTooLarge(t *testing.T) {
	// An H count past the file is an error too.
	term := startVi(t, "one\ntwo\nthree\n")

	settleCursor(t, term, "99H", 0, 0, "99H")
}

func TestDeleteOnEmptyLineFails(t *testing.T) {
	term := startVi(t, "one\n\nlast\n")

	send(term, "2G")
	waitCursor(t, term, 1, 0)
	settleCursor(t, term, "x", 1, 0, "x on an empty line")
	if s := term.Dump(); line(s, 1) != "" || line(s, 2) != "last" {
		t.Fatalf("x on an empty line changed the text:\n%s",
			strings.Join(s[:3], "\n"))
	}
}

func TestColumnMotionClamps(t *testing.T) {
	// Unlike G, a | count past the line end is not an error: the
	// cursor stops on the last character, without a bell.
	term := startVi(t, "abc def\n")

	term.ClearBell()
	send(term, "99|")
	waitCursor(t, term, 0, 6)
	term.WaitQuiet(50*time.Millisecond, screenTimeout)
	if term.Bell != 0 {
		t.Fatal("99| clamped but rang the bell")
	}
}

func TestSuccessfulMotionNoBell(t *testing.T) {
	term := startVi(t, "abc def\n")

	term.ClearBell()
	send(term, "w")
	waitCursor(t, term, 0, 4)
	term.WaitQuiet(50*time.Millisecond, screenTimeout)
	if term.Bell != 0 {
		t.Fatal("successful w rang the bell")
	}
}

func TestLineMotionAtTopFails(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	send(term, "$")
	waitCursor(t, term, 0, 2)
	settleCursor(t, term, "99k", 0, 2, "99k at the top of the file")
}

func TestArrowKeysMoveCursor(t *testing.T) {
	// The arrow keys arrive as the terminfo cursor-key sequences
	// (this terminal's entry advertises the ansi \E[A family) and
	// are seeded into the command key maps at startup.
	term := startVi(t, "one\ntwo\nthree\n")

	send(term, "\x1b[B") // down
	waitCursor(t, term, 1, 0)
	send(term, "\x1b[C") // right
	waitCursor(t, term, 1, 1)
	send(term, "\x1b[A") // up
	waitCursor(t, term, 0, 1)
	send(term, "\x1b[D") // left
	waitCursor(t, term, 0, 0)
}

// Scrolling and screen-repositioning commands (docs/nvi.md "Vi
// Commands"): ^F ^B ^D ^U ^E ^Y, z, and the ^L/^R redraw.  On a
// 24-row terminal the text window is 23 rows.
package nvitests

import (
	"strings"
	"testing"

	"goterm"
)

// corruptScreen scribbles over the emulated display by writing
// directly to the terminal, bypassing vi entirely -- the moral
// equivalent of line noise or a write(1) from another user.  vi's
// internal model still has the real contents, so a redraw must
// restore them.
func corruptScreen(t *testing.T, term *goterm.Term) {
	t.Helper()
	term.Write([]byte("\x1b[H\x1b[2J")) // home, clear
	for i := 0; i < rows; i++ {
		term.Write([]byte("GARBAGE GARBAGE GARBAGE\r\n"))
	}
	waitScreen(t, term, screenTimeout, "screen corruption",
		func(s []string) bool {
			return strings.Contains(s[0], "GARBAGE")
		})
}

// assertRedraw checks that a redraw key repainted the file contents
// and removed every trace of the corruption.
func assertRedraw(t *testing.T, term *goterm.Term, key, name string) {
	t.Helper()
	send(term, key)
	waitScreen(t, term, screenTimeout, name+" to repaint",
		func(s []string) bool {
			return line(s, 0) == "one" && line(s, 1) == "two" &&
				line(s, 2) == "~" &&
				!strings.Contains(strings.Join(s, "\n"), "GARBAGE")
		})
	waitCursor(t, term, 0, 0)
}

func TestRedrawControlL(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	corruptScreen(t, term)
	assertRedraw(t, term, "\x0c", "^L")
}

func TestRedrawControlR(t *testing.T) {
	// ^R is bound to the same redraw as ^L.
	term := startVi(t, "one\ntwo\n")

	corruptScreen(t, term)
	assertRedraw(t, term, "\x12", "^R")
}

func TestPageForwardBack(t *testing.T) {
	term := startVi(t, numberedLines(60))

	// ^F pages forward with a two-line overlap: l22 reaches the top.
	send(term, "\x06")
	waitScreen(t, term, screenTimeout, "^F to page forward",
		func(s []string) bool { return line(s, 0) == "l22" })

	send(term, "\x02") // ^B pages back
	waitScreen(t, term, screenTimeout, "^B to page back",
		func(s []string) bool { return line(s, 0) == "l1" })
}

func TestHalfPageScroll(t *testing.T) {
	term := startVi(t, numberedLines(60))

	// ^D scrolls down half a screen: with a 23-row window the top
	// moves by 12 lines.
	send(term, "\x04")
	waitScreen(t, term, screenTimeout, "^D to scroll down",
		func(s []string) bool { return line(s, 0) == "l13" })

	send(term, "\x15") // ^U scrolls back up
	waitScreen(t, term, screenTimeout, "^U to scroll up",
		func(s []string) bool { return line(s, 0) == "l1" })
}

func TestLineScroll(t *testing.T) {
	term := startVi(t, numberedLines(60))

	send(term, "\x05") // ^E: screen down one line, cursor stays put
	waitScreen(t, term, screenTimeout, "^E to scroll one line",
		func(s []string) bool { return line(s, 0) == "l2" })

	send(term, "\x19") // ^Y: back up
	waitScreen(t, term, screenTimeout, "^Y to scroll back",
		func(s []string) bool { return line(s, 0) == "l1" })
}

func TestZRepositioning(t *testing.T) {
	term := startVi(t, numberedLines(60))

	send(term, "30G")
	send(term, "z\r") // current line to the top
	waitScreen(t, term, screenTimeout, "z<CR> to put l30 on top",
		func(s []string) bool { return line(s, 0) == "l30" })

	send(term, "z.") // current line to the middle
	waitScreen(t, term, screenTimeout, "z. to center l30",
		func(s []string) bool {
			return line(s, (rows-1)/2) == "l30"
		})

	send(term, "z-") // current line to the bottom
	waitScreen(t, term, screenTimeout, "z- to put l30 at bottom",
		func(s []string) bool {
			return line(s, rows-2) == "l30"
		})
}

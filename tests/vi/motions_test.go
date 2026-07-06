// Cursor motion commands (docs/nvi.md "Vi Commands"): character,
// word, line, sentence, paragraph, and section motions, in-line
// character searches, marks, and screen-position motions.
package nvitests

import "testing"

func TestCharacterMotions(t *testing.T) {
	term := startVi(t, "abcdef\nxyz\n")

	send(term, "lll")
	waitCursor(t, term, 0, 3)
	send(term, "h")
	waitCursor(t, term, 0, 2)
	send(term, " ") // <space> moves right like l
	waitCursor(t, term, 0, 3)
	send(term, "\x08") // <control-H> moves left like h
	waitCursor(t, term, 0, 2)
	send(term, "3h") // a count larger than the room left clamps
	waitCursor(t, term, 0, 0)
}

func TestWordMotions(t *testing.T) {
	term := startVi(t, "one two three\n")

	send(term, "w")
	waitCursor(t, term, 0, 4)
	send(term, "w")
	waitCursor(t, term, 0, 8)
	send(term, "b")
	waitCursor(t, term, 0, 4)
	send(term, "e")
	waitCursor(t, term, 0, 6) // end of "two"
	send(term, "0")
	waitCursor(t, term, 0, 0)
	send(term, "2w")
	waitCursor(t, term, 0, 8)
}

func TestWordVsBigword(t *testing.T) {
	// "bar-baz" is three words but one bigword.
	term := startVi(t, "foo bar-baz qux\n")

	send(term, "w")
	waitCursor(t, term, 0, 4) // bar
	send(term, "w")
	waitCursor(t, term, 0, 7) // -
	send(term, "w")
	waitCursor(t, term, 0, 8) // baz
	send(term, "0W")
	waitCursor(t, term, 0, 4) // bar-baz as one bigword
	send(term, "W")
	waitCursor(t, term, 0, 12) // qux
	send(term, "B")
	waitCursor(t, term, 0, 4)
	send(term, "0E")
	waitCursor(t, term, 0, 2) // end of "foo"
	send(term, "E")
	waitCursor(t, term, 0, 10) // end of "bar-baz"
}

func TestLineStartEndMotions(t *testing.T) {
	term := startVi(t, "   abcdef\n")

	send(term, "$")
	waitCursor(t, term, 0, 8)
	send(term, "0")
	waitCursor(t, term, 0, 0)
	send(term, "^")
	waitCursor(t, term, 0, 3) // first nonblank
	send(term, "5|")
	waitCursor(t, term, 0, 4) // | is 1-based columns
}

func TestLineMotionsFirstNonblank(t *testing.T) {
	term := startVi(t, "one\n   two\nthree\n")

	send(term, "+")
	waitCursor(t, term, 1, 3) // + goes to first nonblank
	send(term, "+")
	waitCursor(t, term, 2, 0)
	send(term, "-")
	waitCursor(t, term, 1, 3)
	send(term, "1G")
	waitCursor(t, term, 0, 0)
	send(term, "\r") // <carriage-return> like +
	waitCursor(t, term, 1, 3)
	send(term, "_") // _ is first nonblank of current line
	waitCursor(t, term, 1, 3)
}

func TestGotoLine(t *testing.T) {
	term := startVi(t, numberedLines(30))

	send(term, "15G")
	waitCursor(t, term, 14, 0)
	send(term, "G") // no count: last line
	waitScreen(t, term, screenTimeout, "G to reach l30",
		func(s []string) bool {
			r, _ := term.Cursor()
			return line(s, r) == "l30"
		})
}

func TestCharacterSearchInLine(t *testing.T) {
	term := startVi(t, "abcXdefXghi\n")

	send(term, "fX")
	waitCursor(t, term, 0, 3)
	send(term, ";") // repeat
	waitCursor(t, term, 0, 7)
	send(term, ",") // reverse
	waitCursor(t, term, 0, 3)
	send(term, "0tX") // t stops short
	waitCursor(t, term, 0, 2)
	send(term, "$FX") // F searches backward
	waitCursor(t, term, 0, 7)
	send(term, "TX") // T stops one past the character, going left
	waitCursor(t, term, 0, 4)
}

func TestSentenceMotions(t *testing.T) {
	term := startVi(t, "One.  Two.  Three.\n")

	send(term, ")")
	waitCursor(t, term, 0, 6)
	send(term, ")")
	waitCursor(t, term, 0, 12)
	send(term, "(")
	waitCursor(t, term, 0, 6)
}

func TestParagraphMotions(t *testing.T) {
	term := startVi(t, "one\ntwo\n\nthree\nfour\n")

	send(term, "}")
	waitCursor(t, term, 2, 0) // the empty line
	send(term, "}")
	waitScreen(t, term, screenTimeout, "} to reach last line",
		func(s []string) bool {
			r, _ := term.Cursor()
			return line(s, r) == "four"
		})
	send(term, "{")
	waitCursor(t, term, 2, 0)
	send(term, "{")
	waitCursor(t, term, 0, 0)
}

func TestSectionMotions(t *testing.T) {
	// Default sections include the .SH nroff macro.
	term := startVi(t, "alpha\n.SH one\nbravo\n.SH two\ncharlie\n")

	send(term, "]]")
	waitCursor(t, term, 1, 0)
	send(term, "]]")
	waitCursor(t, term, 3, 0)
	send(term, "[[")
	waitCursor(t, term, 1, 0)
}

func TestMatchParen(t *testing.T) {
	term := startVi(t, "x (abc) y\n")

	// From a non-bracket character, % finds the next bracket in the
	// line and jumps to its match.
	send(term, "%")
	waitCursor(t, term, 0, 6)
	send(term, "%")
	waitCursor(t, term, 0, 2)
}

func TestMarks(t *testing.T) {
	term := startVi(t, "one\n   two\nthree\n")

	send(term, "2G$ma1G") // mark column: the 'o' of "two"
	waitCursor(t, term, 0, 0)
	send(term, "`a") // backquote: exact position
	waitCursor(t, term, 1, 5)
	send(term, "1G'a") // quote: first nonblank of marked line
	waitCursor(t, term, 1, 3)
}

func TestSearchVariants(t *testing.T) {
	term := startVi(t, "alpha\nbeta\nalpha\n")

	send(term, "/alpha\r") // forward from line 1 finds line 3
	waitCursor(t, term, 2, 0)
	send(term, "n") // next match wraps to line 1
	waitCursor(t, term, 0, 0)
	send(term, "N") // N reverses the direction
	waitCursor(t, term, 2, 0)
	send(term, "?beta\r") // backward search
	waitCursor(t, term, 1, 0)
}

func TestLineMotionAliases(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\n")

	send(term, "\x0a") // ^J down
	waitCursor(t, term, 1, 0)
	send(term, "\x0e") // ^N down
	waitCursor(t, term, 2, 0)
	send(term, "\x10") // ^P up
	waitCursor(t, term, 1, 0)
}

func TestScreenPositionMotions(t *testing.T) {
	term := startVi(t, numberedLines(30))

	send(term, "L")
	waitCursor(t, term, rows-2, 0) // bottom text row
	send(term, "M")
	waitCursor(t, term, (rows-1)/2, 0)
	send(term, "H")
	waitCursor(t, term, 0, 0)
	send(term, "2L")
	waitCursor(t, term, rows-3, 0)
	send(term, "2H")
	waitCursor(t, term, 1, 0)
}

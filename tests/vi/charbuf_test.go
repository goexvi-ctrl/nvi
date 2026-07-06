// Character-oriented buffers and operators with search motions
// (docs/nvi.md "Vi Commands" and the buffer-orientation paragraph in
// "General Editor Description").
package nvitests

import "testing"

func TestDeleteToSearch(t *testing.T) {
	term := startVi(t, "one two three\n")

	// d/pattern is a character-mode delete up to the match start.
	send(term, "d/three\r")
	waitScreen(t, term, screenTimeout, "d/three",
		func(s []string) bool { return line(s, 0) == "three" })
}

func TestChangeToSearch(t *testing.T) {
	term := startVi(t, "one two three\n")

	send(term, "c/three\rX \x1b")
	waitScreen(t, term, screenTimeout, "c/three",
		func(s []string) bool { return line(s, 0) == "X three" })
}

func TestDeleteToBackwardSearch(t *testing.T) {
	term := startVi(t, "one two three\n")

	// From the end, d?two deletes back to the match start.
	send(term, "$d?two\r")
	waitScreen(t, term, screenTimeout, "d?two",
		func(s []string) bool { return line(s, 0) == "one e" })
}

func TestDeleteWithCharacterFind(t *testing.T) {
	term := startVi(t, "abcXdefXghi\n")

	send(term, "dfX") // inclusive: deletes through the X
	waitScreen(t, term, screenTimeout, "dfX",
		func(s []string) bool { return line(s, 0) == "defXghi" })

	send(term, "dtX") // exclusive: stops before the X
	waitScreen(t, term, screenTimeout, "dtX",
		func(s []string) bool { return line(s, 0) == "Xghi" })
}

func TestDeleteRepeatFind(t *testing.T) {
	term := startVi(t, "aXbXc\n")

	send(term, "fX") // establish the find
	waitCursor(t, term, 0, 1)
	send(term, "d;") // delete through the next X
	waitScreen(t, term, screenTimeout, "d;",
		func(s []string) bool { return line(s, 0) == "ac" })
}

func TestCharacterPutTransposes(t *testing.T) {
	// The classic xp: a character-mode buffer puts after the
	// cursor, transposing two characters.
	term := startVi(t, "abc\n")

	send(term, "xp")
	waitScreen(t, term, screenTimeout, "xp",
		func(s []string) bool { return line(s, 0) == "bac" })
}

func TestCharacterYankPutMidLine(t *testing.T) {
	term := startVi(t, "one two\n")

	// yw yanks "one " character-oriented; $p inserts it after the
	// last character rather than opening a new line.
	send(term, "yw$p")
	waitScreen(t, term, screenTimeout, "yw$p",
		func(s []string) bool { return line(s, 0) == "one twoone" })
}

func TestMultilineCharacterPutSplitsLine(t *testing.T) {
	term := startVi(t, "ab\ncd\nZZ\n")

	// y/d yanks "ab\nc" -- a character-mode buffer spanning two
	// lines.  Putting it splits the target line at the cursor: the
	// buffer's first line joins the text before the split point and
	// its last line joins the rest.
	send(term, "y/d\r")
	send(term, "3G$p")
	waitScreen(t, term, screenTimeout, "multi-line character put",
		func(s []string) bool {
			return line(s, 0) == "ab" && line(s, 1) == "cd" &&
				line(s, 2) == "ZZab" && line(s, 3) == "c"
		})
}

func TestPreviousContextMarks(t *testing.T) {
	term := startVi(t, "alpha\nbeta\ngamma\ndelta\n")

	// A search sets the previous context; '' returns to it and a
	// second '' swaps back.
	send(term, "ll/delta\r")
	waitCursor(t, term, 3, 0)
	send(term, "''")
	waitCursor(t, term, 0, 0) // ' form: first nonblank of the line
	send(term, "''")
	waitCursor(t, term, 3, 0)
}

func TestPreviousContextBackquote(t *testing.T) {
	term := startVi(t, "alpha\nbeta\ngamma\ndelta\n")

	// The backquote form returns to the exact column.
	send(term, "ll/delta\r")
	waitCursor(t, term, 3, 0)
	send(term, "``")
	waitCursor(t, term, 0, 2)
}

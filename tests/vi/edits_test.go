// Editing commands (docs/nvi.md "Vi Commands"): replace, insert
// variants, join, case flip, delete characters, repeat, undo, the
// increment command, buffer macros, and input-mode editing keys.
package nvitests

import "testing"

func TestReplaceCharacter(t *testing.T) {
	term := startVi(t, "abc\n")

	send(term, "rZ")
	waitScreen(t, term, screenTimeout, "r",
		func(s []string) bool { return line(s, 0) == "Zbc" })

	send(term, "3rx") // count replaces several characters
	waitScreen(t, term, screenTimeout, "3r",
		func(s []string) bool { return line(s, 0) == "xxx" })
}

func TestReplaceMode(t *testing.T) {
	term := startVi(t, "abcdef\n")

	send(term, "RXY\x1b")
	waitScreen(t, term, screenTimeout, "R",
		func(s []string) bool { return line(s, 0) == "XYcdef" })
}

func TestInsertVariants(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, "A!\x1b") // append at end of line
	waitScreen(t, term, screenTimeout, "A",
		func(s []string) bool { return line(s, 0) == "one!" })

	send(term, "I#\x1b") // insert at first nonblank
	waitScreen(t, term, screenTimeout, "I",
		func(s []string) bool { return line(s, 0) == "#one!" })

	send(term, "la+\x1b") // a appends after the cursor
	waitScreen(t, term, screenTimeout, "a",
		func(s []string) bool { return line(s, 0) == "#o+ne!" })

	send(term, "Otop\x1b") // O opens a line above
	waitScreen(t, term, screenTimeout, "O",
		func(s []string) bool {
			return line(s, 0) == "top" && line(s, 1) == "#o+ne!"
		})
}

func TestJoinWithCount(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\n")

	send(term, "3J") // join three lines
	waitScreen(t, term, screenTimeout, "3J",
		func(s []string) bool {
			return line(s, 0) == "one two three" && line(s, 1) == "~"
		})
}

func TestCaseFlip(t *testing.T) {
	term := startVi(t, "abC\n")

	send(term, "3~")
	waitScreen(t, term, screenTimeout, "3~",
		func(s []string) bool { return line(s, 0) == "ABc" })
}

func TestDeleteChars(t *testing.T) {
	term := startVi(t, "abcdef\n")

	send(term, "3x") // x deletes at the cursor
	waitScreen(t, term, screenTimeout, "3x",
		func(s []string) bool { return line(s, 0) == "def" })

	send(term, "$X") // X deletes before the cursor
	waitScreen(t, term, screenTimeout, "X",
		func(s []string) bool { return line(s, 0) == "df" })
}

func TestDotRepeats(t *testing.T) {
	term := startVi(t, "aone\nbtwo\n")

	send(term, "x")
	waitScreen(t, term, screenTimeout, "x",
		func(s []string) bool { return line(s, 0) == "one" })

	send(term, "j0.") // . repeats the delete on the next line
	waitScreen(t, term, screenTimeout, ".",
		func(s []string) bool { return line(s, 1) == "two" })
}

func TestLineUndoU(t *testing.T) {
	// U restores the line to its state before any of the changes.
	term := startVi(t, "abcdef\n")

	send(term, "xx")
	waitScreen(t, term, screenTimeout, "two deletes",
		func(s []string) bool { return line(s, 0) == "cdef" })

	send(term, "U")
	waitScreen(t, term, screenTimeout, "U",
		func(s []string) bool { return line(s, 0) == "abcdef" })
}

func TestIncrementNumber(t *testing.T) {
	term := startVi(t, "5 apples\n")

	send(term, "#+")
	waitScreen(t, term, screenTimeout, "#+",
		func(s []string) bool { return line(s, 0) == "6 apples" })

	send(term, "#-")
	waitScreen(t, term, screenTimeout, "#-",
		func(s []string) bool { return line(s, 0) == "5 apples" })

	send(term, "3#+") // count is the amount
	waitScreen(t, term, screenTimeout, "3#+",
		func(s []string) bool { return line(s, 0) == "8 apples" })
}

func TestSearchWordControlA(t *testing.T) {
	term := startVi(t, "foo bar\nbaz foo x\n")

	send(term, "\x01") // ^A searches for the word under the cursor
	waitCursor(t, term, 1, 4)
}

func TestAtBufferMacro(t *testing.T) {
	// Delete a line of vi commands into buffer b, execute it with @.
	term := startVi(t, "3x\nabcdef\n")

	send(term, "\"bdd")
	waitScreen(t, term, screenTimeout, "command line deleted",
		func(s []string) bool { return line(s, 0) == "abcdef" })

	send(term, "@b")
	waitScreen(t, term, screenTimeout, "@b runs 3x",
		func(s []string) bool { return line(s, 0) == "def" })
}

func TestAmpersandRepeatsSubstitute(t *testing.T) {
	term := startVi(t, "one a\none b\n")

	send(term, ":1s/one/ONE/\r")
	waitScreen(t, term, screenTimeout, "substitute on line one",
		func(s []string) bool { return line(s, 0) == "ONE a" })

	send(term, "j&")
	waitScreen(t, term, screenTimeout, "& repeats",
		func(s []string) bool { return line(s, 1) == "ONE b" })
}

func TestLiteralNextInsert(t *testing.T) {
	// ^V enters the next character literally; a control character
	// displays as ^X.
	term := startVi(t, "ab\n")

	send(term, "li\x16\x01\x1b")
	waitScreen(t, term, screenTimeout, "literal ^A inserted",
		func(s []string) bool { return line(s, 0) == "a^Ab" })
}

func TestInputWordErase(t *testing.T) {
	term := startVi(t, "end\n")

	// ^W during input erases the word just typed.
	send(term, "ifoo bar\x17baz \x1b")
	waitScreen(t, term, screenTimeout, "input-mode word erase",
		func(s []string) bool { return line(s, 0) == "foo baz end" })
}

func TestAutoindent(t *testing.T) {
	term := startVi(t, "    one\n")

	send(term, ":set autoindent\r")
	send(term, "ox\x1b") // the opened line inherits the indent
	waitScreen(t, term, screenTimeout, "autoindented line",
		func(s []string) bool { return line(s, 1) == "    x" })
}

func TestUndoChainWithDot(t *testing.T) {
	// nvi extension (docs/nvi.md section 2): u toggles between undo
	// and redo, and . immediately after an undo continues undoing
	// further changes in the same direction; a u then reverses the
	// direction of the chain.
	term := startVi(t, "one\ntwo\nthree\nfour\n")

	send(term, "dddddd")
	waitScreen(t, term, screenTimeout, "three deletes",
		func(s []string) bool { return line(s, 0) == "four" })

	send(term, "u")
	waitScreen(t, term, screenTimeout, "undo the third delete",
		func(s []string) bool { return line(s, 0) == "three" })

	send(term, "..")
	waitScreen(t, term, screenTimeout, ".. continues the undo chain",
		func(s []string) bool {
			return line(s, 0) == "one" && line(s, 3) == "four"
		})

	send(term, "u")
	waitScreen(t, term, screenTimeout, "u reverses to redo",
		func(s []string) bool { return line(s, 0) == "two" })

	send(term, ".")
	waitScreen(t, term, screenTimeout, ". continues the redo chain",
		func(s []string) bool { return line(s, 0) == "three" })
}

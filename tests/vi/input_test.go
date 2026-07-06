// Text-input mode machinery (docs/nvi.md "Vi Text Input Commands"):
// the autoindent erase strings <control-D>, 0<control-D>, and
// ^<control-D>, the <control-T> indent, counted inserts, and
// replacing a character with a newline.  Input maps (map!) and
// abbreviations expand during vi text input.
package nvitests

import "testing"

func TestControlDBacktab(t *testing.T) {
	term := startVi(t, "\tone\n")

	send(term, ":set ai sw=4\r")
	// o inherits the 8-column indent; ^D while still in the
	// autoindent erases back one shiftwidth boundary.
	send(term, "o\x04yy\x1b")
	waitScreen(t, term, screenTimeout, "^D backtab to column 4",
		func(s []string) bool {
			return line(s, 0) == "        one" && line(s, 1) == "    yy"
		})
}

func TestZeroControlDKillsIndent(t *testing.T) {
	term := startVi(t, "\tone\n")

	send(term, ":set ai sw=4\r")
	// 0^D erases all the autoindent; unlike ^^D the zero indent
	// carries to the following input line.
	send(term, "o0\x04aa\rbb\x1b")
	waitScreen(t, term, screenTimeout, "0^D erases all autoindent",
		func(s []string) bool {
			return line(s, 1) == "aa" && line(s, 2) == "bb"
		})
}

func TestCaretControlDTemporary(t *testing.T) {
	term := startVi(t, "\tone\n")

	send(term, ":set ai sw=4\r")
	// ^^D erases the autoindent for the current line only (the
	// historic idiom for preprocessor lines); the next input line
	// gets the remembered indent back.
	send(term, "o^\x04#def\rnext\x1b")
	waitScreen(t, term, screenTimeout, "^^D one-line indent erase",
		func(s []string) bool {
			return line(s, 1) == "#def" && line(s, 2) == "        next"
		})
}

func TestControlTIndent(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, ":set sw=4\r")
	// ^T inserts shiftwidth whitespace even without autoindent, and
	// repeats stack.
	send(term, "i\x14abc\x1b")
	waitScreen(t, term, screenTimeout, "^T indents shiftwidth",
		func(s []string) bool { return line(s, 0) == "    abcone" })

	send(term, "o\x14\x14x\x1b")
	waitScreen(t, term, screenTimeout, "two ^T stack",
		func(s []string) bool { return line(s, 1) == "        x" })
}

func TestCountedInsert(t *testing.T) {
	term := startVi(t, "one\n")

	// A count on i repeats the entered text.
	send(term, "3iab\x1b")
	waitScreen(t, term, screenTimeout, "3iab",
		func(s []string) bool { return line(s, 0) == "abababone" })
	waitCursor(t, term, 0, 5)
}

func TestReplaceWithNewlineSplitsLine(t *testing.T) {
	term := startVi(t, "abcdef\n")

	// r<carriage-return> replaces the character with a line break.
	send(term, "3lr\r")
	waitScreen(t, term, screenTimeout, "r CR splits the line",
		func(s []string) bool {
			return line(s, 0) == "abc" && line(s, 1) == "ef"
		})
	waitCursor(t, term, 1, 0)
}

func TestInputMapExpands(t *testing.T) {
	term := startVi(t, "one\n")

	// map! applies to text input mode.
	send(term, ":map! zq ZAP\r")
	send(term, "Izq-\x1b")
	waitScreen(t, term, screenTimeout, "map! expansion in input",
		func(s []string) bool { return line(s, 0) == "ZAP-one" })
}

func TestAbbreviationExpands(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, ":ab teh the\r")
	// The abbreviation replaces the word when a non-word character
	// is entered after it.
	send(term, "Iteh \x1b")
	waitScreen(t, term, screenTimeout, "abbreviation expansion",
		func(s []string) bool { return line(s, 0) == "the one" })
}

func TestAbbreviationNeedsWordBoundary(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, ":ab teh the\r")
	// As part of a longer word the abbreviation must not fire.
	send(term, "Isteh \x1b")
	waitScreen(t, term, screenTimeout, "no expansion mid-word",
		func(s []string) bool { return line(s, 0) == "steh one" })
}

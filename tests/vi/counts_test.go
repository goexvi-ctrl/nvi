// Counts on operators, puts, and the repeat command (docs/nvi.md
// "General Editor Description", count prefixes).
package nvitests

import "testing"

func TestYankCountLines(t *testing.T) {
	term := startVi(t, "l1\nl2\nl3\nl4\n")

	// 2yy yanks two lines; P puts them above the current line.
	send(term, "2yyP")
	waitScreen(t, term, screenTimeout, "2yyP",
		func(s []string) bool {
			return line(s, 0) == "l1" && line(s, 1) == "l2" &&
				line(s, 2) == "l1" && line(s, 3) == "l2" &&
				line(s, 4) == "l3"
		})
}

func TestYankCountedMotion(t *testing.T) {
	term := startVi(t, "l1\nl2\nl3\nl4\n")

	// y2j is line oriented: the current line plus two below.
	send(term, "y2jP")
	waitScreen(t, term, screenTimeout, "y2jP",
		func(s []string) bool {
			return line(s, 0) == "l1" && line(s, 1) == "l2" &&
				line(s, 2) == "l3" && line(s, 3) == "l1" &&
				line(s, 4) == "l2" && line(s, 5) == "l3" &&
				line(s, 6) == "l4"
		})
}

func TestPutWithCount(t *testing.T) {
	term := startVi(t, "ab\ncd\n")

	// A count on p puts the buffer that many times.
	send(term, "yy3p")
	waitScreen(t, term, screenTimeout, "yy3p",
		func(s []string) bool {
			return line(s, 0) == "ab" && line(s, 1) == "ab" &&
				line(s, 2) == "ab" && line(s, 3) == "ab" &&
				line(s, 4) == "cd"
		})
}

func TestNamedBufferCountedDelete(t *testing.T) {
	term := startVi(t, "l1\nl2\nl3\nl4\n")

	// "a2dd deletes two lines into buffer a; put them at the end.
	send(term, "\"a2ddG\"ap")
	waitScreen(t, term, screenTimeout, "counted delete into buffer a",
		func(s []string) bool {
			return line(s, 0) == "l3" && line(s, 1) == "l4" &&
				line(s, 2) == "l1" && line(s, 3) == "l2"
		})
}

func TestRepeatWithNewCount(t *testing.T) {
	term := startVi(t, "l1\nl2\nl3\nl4\nl5\n")

	// A count given to . replaces the original command's count.
	send(term, "dd")
	waitScreen(t, term, screenTimeout, "dd",
		func(s []string) bool { return line(s, 0) == "l2" })
	send(term, "2.")
	waitScreen(t, term, screenTimeout, "2. repeats the delete for two lines",
		func(s []string) bool {
			return line(s, 0) == "l4" && line(s, 1) == "l5"
		})
}

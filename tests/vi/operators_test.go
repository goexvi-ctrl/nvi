// Operator commands (docs/nvi.md "Vi Commands"): delete, change,
// yank and put with motions, counts, named and numeric buffers,
// shifting, and the filter operator.
package nvitests

import "testing"

func TestDeleteWordMotions(t *testing.T) {
	term := startVi(t, "one two three\n")

	send(term, "dw") // deletes the word and trailing blanks
	waitScreen(t, term, screenTimeout, "dw",
		func(s []string) bool { return line(s, 0) == "two three" })

	send(term, "de") // deletes to end of word only
	waitScreen(t, term, screenTimeout, "de",
		func(s []string) bool { return line(s, 0) == " three" })
}

func TestDeleteToEndOfLine(t *testing.T) {
	term := startVi(t, "abcdef\n")

	send(term, "llD") // D is d$
	waitScreen(t, term, screenTimeout, "D",
		func(s []string) bool { return line(s, 0) == "ab" })
}

func TestDeleteLinewiseMotion(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\n")

	send(term, "dj") // dj is line oriented: two lines go
	waitScreen(t, term, screenTimeout, "dj",
		func(s []string) bool {
			return line(s, 0) == "three" && line(s, 1) == "~"
		})
}

func TestDeleteWithCount(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\nfour\n")

	send(term, "2dd")
	waitScreen(t, term, screenTimeout, "2dd",
		func(s []string) bool {
			return line(s, 0) == "three" && line(s, 1) == "four"
		})
}

func TestChangeWord(t *testing.T) {
	term := startVi(t, "one two\n")

	send(term, "cwXY\x1b")
	waitScreen(t, term, screenTimeout, "cw",
		func(s []string) bool { return line(s, 0) == "XY two" })
}

func TestChangeLineAndC(t *testing.T) {
	term := startVi(t, "one two\nnext\n")

	send(term, "ccnew\x1b") // cc changes the whole line
	waitScreen(t, term, screenTimeout, "cc",
		func(s []string) bool { return line(s, 0) == "new" })

	send(term, "0llCZ\x1b") // C changes to end of line
	waitScreen(t, term, screenTimeout, "C",
		func(s []string) bool { return line(s, 0) == "neZ" })
}

func TestSubstituteCommands(t *testing.T) {
	term := startVi(t, "abc\nsecond\n")

	send(term, "sZ\x1b") // s substitutes the character
	waitScreen(t, term, screenTimeout, "s",
		func(s []string) bool { return line(s, 0) == "Zbc" })

	send(term, "Swhole\x1b") // S substitutes the line
	waitScreen(t, term, screenTimeout, "S",
		func(s []string) bool {
			return line(s, 0) == "whole" && line(s, 1) == "second"
		})
}

func TestYankPut(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	send(term, "yyjp") // yank line one, put it after line two
	waitScreen(t, term, screenTimeout, "yyjp",
		func(s []string) bool {
			return line(s, 0) == "one" && line(s, 1) == "two" &&
				line(s, 2) == "one"
		})

	send(term, "1GP") // P puts above
	waitScreen(t, term, screenTimeout, "P",
		func(s []string) bool {
			return line(s, 0) == "one" && line(s, 1) == "one"
		})
}

func TestNamedBuffer(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	send(term, "\"ayyj\"ap") // yank into buffer a, put from it
	waitScreen(t, term, screenTimeout, "named buffer put",
		func(s []string) bool {
			return line(s, 1) == "two" && line(s, 2) == "one"
		})
}

func TestNumericBuffersRotate(t *testing.T) {
	// Deleted lines rotate through buffers 1-9: after deleting
	// "one" then "two", buffer 1 holds "two" and buffer 2 "one".
	term := startVi(t, "one\ntwo\nthree\n")

	send(term, "dddd")
	waitScreen(t, term, screenTimeout, "two deletes",
		func(s []string) bool { return line(s, 0) == "three" })

	send(term, "\"2p")
	waitScreen(t, term, screenTimeout, "put from buffer 2",
		func(s []string) bool {
			return line(s, 0) == "three" && line(s, 1) == "one"
		})

	send(term, "\"1p")
	waitScreen(t, term, screenTimeout, "put from buffer 1",
		func(s []string) bool { return line(s, 2) == "two" })
}

func TestShiftOperator(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	send(term, ">>") // shiftwidth=8: one tab, rendered as 8 blanks
	waitScreen(t, term, screenTimeout, ">>",
		func(s []string) bool { return line(s, 0) == "        one" })

	send(term, "<<")
	waitScreen(t, term, screenTimeout, "<<",
		func(s []string) bool { return line(s, 0) == "one" })

	send(term, ">j") // shift with a motion covers both lines
	waitScreen(t, term, screenTimeout, ">j",
		func(s []string) bool {
			return line(s, 0) == "        one" &&
				line(s, 1) == "        two"
		})
}

func TestFilterOperator(t *testing.T) {
	term := startVi(t, "banana\napple\n")

	// !j prompts on the bottom line for the command.
	send(term, "!jsort\r")
	waitScreen(t, term, screenTimeout, "filter through sort",
		func(s []string) bool {
			return line(s, 0) == "apple" && line(s, 1) == "banana"
		})
}

func TestDeleteToMark(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\nfour\n")

	send(term, "majjd'a") // delete from marked line to current
	waitScreen(t, term, screenTimeout, "d'a",
		func(s []string) bool {
			return line(s, 0) == "four" && line(s, 1) == "~"
		})
}

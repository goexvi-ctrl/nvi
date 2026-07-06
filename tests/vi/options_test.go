// Options whose effect is visual and need the screen to verify
// (docs/nvi.md "Set Options"): display options, input-mode wrapping,
// mapping behavior, window sizing, and incremental search.
package nvitests

import (
	"strings"
	"testing"
	"time"
)

func TestOptionListDisplay(t *testing.T) {
	term := startVi(t, "one\tx\n")

	// Default display expands the tab.
	waitScreen(t, term, screenTimeout, "tab expanded",
		func(s []string) bool { return line(s, 0) == "one     x" })

	send(term, ":set list\r")
	waitScreen(t, term, screenTimeout, "list mode",
		func(s []string) bool { return line(s, 0) == "one^Ix$" })

	send(term, ":set nolist\r")
	waitScreen(t, term, screenTimeout, "list mode off",
		func(s []string) bool { return line(s, 0) == "one     x" })
}

func TestOptionNumberDisplay(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	// Note the vi display format (7 columns, one space) differs
	// from ex print's number format (6 columns, two spaces).
	send(term, ":set number\r")
	waitScreen(t, term, screenTimeout, "numbered lines",
		func(s []string) bool {
			return line(s, 0) == "      1 one" &&
				line(s, 1) == "      2 two"
		})

	send(term, ":set nonumber\r")
	waitScreen(t, term, screenTimeout, "numbers off",
		func(s []string) bool { return line(s, 0) == "one" })
}

func TestOptionTabstop(t *testing.T) {
	term := startVi(t, "\tx\n")

	waitScreen(t, term, screenTimeout, "tabstop 8",
		func(s []string) bool { return line(s, 0) == "        x" })

	send(term, ":set tabstop=4\r")
	waitScreen(t, term, screenTimeout, "tabstop 4",
		func(s []string) bool { return line(s, 0) == "    x" })
}

func TestOptionShowmode(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, ":set showmode\r")
	send(term, "i")
	waitScreen(t, term, screenTimeout, "input mode indicator",
		func(s []string) bool {
			return strings.Contains(s[rows-1], "Insert")
		})
	send(term, "\x1b")
	waitScreen(t, term, screenTimeout, "command mode indicator",
		func(s []string) bool {
			return strings.Contains(s[rows-1], "Command")
		})
}

func TestOptionTildeop(t *testing.T) {
	term := startVi(t, "abc def\n")

	// Without tildeop, ~ acts on single characters; with it, ~ is
	// an operator taking a motion.
	send(term, ":set tildeop\r")
	send(term, "~w")
	waitScreen(t, term, screenTimeout, "~w flips the word",
		func(s []string) bool { return line(s, 0) == "ABC def" })
}

func TestOptionRemap(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\n")

	// q is unbound in nvi; map it to dd, then map r to q.  With
	// remap (the default), r resolves through q to dd.
	send(term, ":map q dd\r:map r q\r")
	send(term, "r")
	waitScreen(t, term, screenTimeout, "remapped delete",
		func(s []string) bool { return line(s, 0) == "two" })

	// With noremap, r maps to the unbound q and does nothing: after
	// the screen settles, no second line has been deleted.
	send(term, ":set noremap\r")
	send(term, "r")
	term.WaitQuiet(50*time.Millisecond, screenTimeout)
	if s := term.Dump(); line(s, 0) != "two" || line(s, 1) != "three" {
		t.Fatalf("noremap should make r a no-op; screen:\n%s",
			strings.Join(s, "\n"))
	}
}

func TestOptionWrapmargin(t *testing.T) {
	term := startVi(t, "x\n")

	// wrapmargin breaks input lines at a word boundary so they end
	// before the margin (80 - 40 = 40 columns here).
	send(term, ":set wrapmargin=40\r")
	send(term, "Oaaaa bbbb cccc dddd eeee ffff gggg hhhh iiii\x1b")
	waitScreen(t, term, screenTimeout, "wrapped input line",
		func(s []string) bool {
			first := line(s, 0)
			return len(first) <= 40 &&
				strings.HasPrefix(first, "aaaa") &&
				line(s, 1) != "" && line(s, 2) == "x"
		})
}

func TestOptionWindow(t *testing.T) {
	term := startVi(t, numberedLines(60))

	// window sets how many lines a screen redraw uses; ^F pages by
	// window minus the two overlap lines.
	send(term, ":set window=10\r")
	send(term, "\x06") // ^F
	waitScreen(t, term, screenTimeout, "^F with window=10",
		func(s []string) bool { return line(s, 0) == "l9" })
}

func TestControlDCountIsSticky(t *testing.T) {
	// Not an option at all: the vi ^D amount is set by a count on a
	// previous ^D/^U ("Options: None" in docs/nvi.md; the scroll
	// option is the ex scroll amount).  5^D scrolls five lines and
	// a later bare ^D reuses the five.
	term := startVi(t, numberedLines(60))

	send(term, "5\x04")
	waitScreen(t, term, screenTimeout, "5^D",
		func(s []string) bool { return line(s, 0) == "l6" })

	send(term, "\x04")
	waitScreen(t, term, screenTimeout, "bare ^D reuses the count",
		func(s []string) bool { return line(s, 0) == "l11" })
}

func TestOptionLeftright(t *testing.T) {
	long := strings.Repeat("a", 100) + "END"
	term := startVi(t, long+"\n")

	// By default long lines fold onto multiple screen rows.
	waitScreen(t, term, screenTimeout, "folded long line",
		func(s []string) bool {
			return strings.HasPrefix(line(s, 0), "aaaa") &&
				strings.Contains(line(s, 1), "END")
		})

	// With leftright the line occupies one row and the screen
	// scrolls horizontally to follow the cursor.
	send(term, ":set leftright\r")
	send(term, "$")
	waitScreen(t, term, screenTimeout, "horizontal scroll to END",
		func(s []string) bool {
			return strings.Contains(line(s, 0), "END") &&
				!strings.Contains(line(s, 1), "END")
		})
}

func TestOptionSearchincr(t *testing.T) {
	term := startVi(t, "alpha\nbeta\ngamma\n")

	send(term, ":set searchincr\r")
	send(term, "/gam") // no <CR>: the cursor moves as we type
	waitCursor(t, term, 2, 0)
	send(term, "\r")
	waitCursor(t, term, 2, 0)
}

func TestOptionReport(t *testing.T) {
	term := startVi(t, numberedLines(10))

	// report sets the threshold for change messages; the default of
	// 5 means deleting 6 lines reports.
	send(term, "6dd")
	waitScreen(t, term, screenTimeout, "deletion report",
		func(s []string) bool {
			return strings.Contains(s[rows-1], "6 lines deleted")
		})

	send(term, ":set report=2\r")
	send(term, "3dd")
	waitScreen(t, term, screenTimeout, "lowered threshold report",
		func(s []string) bool {
			return strings.Contains(s[rows-1], "3 lines deleted")
		})
}

func TestOptionParagraphsMacros(t *testing.T) {
	term := startVi(t, "one\n.PX\ntwo\n")

	// The paragraphs option lists two-character nroff macro pairs
	// that } and { treat as boundaries.
	send(term, ":set paragraphs=PX\r")
	send(term, "}")
	waitCursor(t, term, 1, 0)
}

func TestOptionSectionsMacros(t *testing.T) {
	term := startVi(t, "one\n.QY\ntwo\n")

	send(term, ":set sections=QY\r")
	send(term, "]]")
	waitCursor(t, term, 1, 0)
}

func TestOptionRuler(t *testing.T) {
	term := startVi(t, "one\ntwo three\n")

	// The ruler reports the cursor's line and column on the status
	// line.  It updates as the cursor moves.
	send(term, ":set ruler\r")
	send(term, "j$")
	waitScreen(t, term, screenTimeout, "ruler shows 2,9",
		func(s []string) bool {
			return strings.HasSuffix(strings.TrimRight(line(s, rows-1), " "),
				"2,9")
		})
}

func TestOptionTtywerase(t *testing.T) {
	// The default word erase (and altwerase) treat a run of
	// punctuation as its own word, so ^W after "foo.bar" erases only
	// "bar".  ttywerase instead erases back to whitespace, taking the
	// whole "foo.bar".
	def := startVi(t, "end\n")
	send(def, "ifoo.bar\x17\x1b")
	waitScreen(t, def, screenTimeout, "default ^W stops at the dot",
		func(s []string) bool { return line(s, 0) == "foo.end" })

	tty := startVi(t, "end\n")
	send(tty, ":set ttywerase\r")
	send(tty, "ifoo.bar\x17\x1b")
	waitScreen(t, tty, screenTimeout, "ttywerase ^W erases to whitespace",
		func(s []string) bool { return line(s, 0) == "end" })
}

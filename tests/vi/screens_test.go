// Multiple screens (docs/nvi.md section 7, "Multiple Screens", and
// the bg, fg, resize, and display entries under "Ex Commands"):
// capitalized commands split the window, <control-W> rotates through
// the foreground screens, bg/fg move screens off and onto the
// display, and resize grows a screen.
//
// Layout facts these tests lean on (24-row terminal): a split gives
// two regions of about half the window, each with its own status
// line on its bottom row; the first split puts file.txt on rows
// 0-10 with its status line on row 11, and the new screen on rows
// 12-22 with its status line on row 23.
package nvitests

import (
	"strings"
	"testing"
)

// statusOn reports whether the screen row is a status line for the
// named file ("name: unmodified: line 1" and variants).
func statusOn(s []string, row int, name string) bool {
	return strings.HasPrefix(line(s, row), name+":")
}

func TestEditSplitsScreen(t *testing.T) {
	term, _ := startViExtra(t, "one\ntwo\n",
		map[string]string{"other.txt": "bee line\n"})

	// With the cursor in the upper half, the new screen opens below
	// the old one and becomes the current screen.
	send(term, ":Edit other.txt\r")
	waitScreen(t, term, screenTimeout, "split after :Edit",
		func(s []string) bool {
			return line(s, 0) == "one" && statusOn(s, 11, "file.txt") &&
				line(s, 12) == "bee line" && statusOn(s, 23, "other.txt")
		})
	waitCursor(t, term, 12, 0)
}

func TestControlWCyclesScreens(t *testing.T) {
	term, _ := startViExtra(t, "aaa\n", map[string]string{
		"b.txt": "bbb\n", "c.txt": "ccc\n"})

	// Two splits leave three screens: file.txt on rows 0-11, b.txt
	// on 12-17, c.txt (current) on 18-23.
	send(term, ":Edit b.txt\r")
	waitScreen(t, term, screenTimeout, "second screen",
		func(s []string) bool { return line(s, 12) == "bbb" })
	send(term, ":Edit c.txt\r")
	waitScreen(t, term, screenTimeout, "third screen",
		func(s []string) bool {
			return statusOn(s, 11, "file.txt") && statusOn(s, 17, "b.txt") &&
				line(s, 18) == "ccc" && statusOn(s, 23, "c.txt")
		})
	waitCursor(t, term, 18, 0)

	// Each ^W moves to the next lower screen in the window, or wraps
	// to the first screen at the bottom.
	send(term, "\x17")
	waitCursor(t, term, 0, 0)
	send(term, "\x17")
	waitCursor(t, term, 12, 0)
	send(term, "\x17")
	waitCursor(t, term, 18, 0)
}

func TestResizeGrowsScreen(t *testing.T) {
	term, _ := startViExtra(t, "one\n",
		map[string]string{"other.txt": "bee line\n"})

	send(term, ":Edit other.txt\r")
	waitScreen(t, term, screenTimeout, "split",
		func(s []string) bool { return statusOn(s, 11, "file.txt") })

	// Growing the current (lower) screen by three rows moves the
	// boundary status line up by three.
	send(term, ":resize +3\r")
	waitScreen(t, term, screenTimeout, "resize +3",
		func(s []string) bool {
			return statusOn(s, 8, "file.txt") && line(s, 9) == "bee line" &&
				statusOn(s, 23, "other.txt")
		})
}

func TestBackgroundAndForeground(t *testing.T) {
	term, _ := startViExtra(t, "one\n",
		map[string]string{"other.txt": "bee line\n"})

	send(term, ":Edit other.txt\r")
	waitScreen(t, term, screenTimeout, "split",
		func(s []string) bool { return line(s, 12) == "bee line" })

	// bg removes the current screen from the display; its rows are
	// taken over by the neighboring screen.
	send(term, ":bg\r")
	waitScreen(t, term, screenTimeout, "bg hides the current screen",
		func(s []string) bool {
			if statusOn(s, 23, "file.txt") == false {
				return false
			}
			for _, l := range s {
				if strings.Contains(l, "bee line") {
					return false
				}
			}
			return true
		})

	// display screens lists the backgrounded screens' files.
	send(term, ":display screens\r")
	waitScreen(t, term, screenTimeout, "display screens",
		func(s []string) bool { return line(s, 23) == "other.txt" })

	// fg swaps the backgrounded screen with the current one, so
	// other.txt takes over the whole window.
	send(term, ":fg other.txt\r")
	waitScreen(t, term, screenTimeout, "fg swaps the screens",
		func(s []string) bool {
			return line(s, 0) == "bee line" && statusOn(s, 23, "other.txt")
		})
}

func TestCapitalFgForegroundsInNewScreen(t *testing.T) {
	term, _ := startViExtra(t, "one\n",
		map[string]string{"other.txt": "bee line\n"})

	send(term, ":Edit other.txt\r")
	waitScreen(t, term, screenTimeout, "split",
		func(s []string) bool { return line(s, 12) == "bee line" })
	send(term, ":bg\r")
	waitScreen(t, term, screenTimeout, "bg",
		func(s []string) bool { return statusOn(s, 23, "file.txt") })

	// Fg splits instead of swapping: both screens are displayed
	// again.
	send(term, ":Fg\r")
	waitScreen(t, term, screenTimeout, "Fg opens a new screen",
		func(s []string) bool {
			var haveOne, haveBee bool
			for _, l := range s {
				switch strings.TrimRight(l, " ") {
				case "one":
					haveOne = true
				case "bee line":
					haveBee = true
				}
			}
			return haveOne && haveBee
		})
}

func TestBackgroundOnlyScreenFails(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, ":bg\r")
	waitScreen(t, term, screenTimeout, "bg error",
		func(s []string) bool {
			return line(s, 23) ==
				"You may not background your only displayed screen."
		})
}

func TestTagInNewScreen(t *testing.T) {
	term, _ := startViExtra(t, "afunc caller\n", map[string]string{
		"tags":     "afunc\ttarget.c\t/^int afunc/\n",
		"target.c": "int afunc(void)\n{\n}\n",
	})

	// Tag performs the tag search in a new screen; the calling
	// screen stays displayed.
	send(term, ":Tag afunc\r")
	waitScreen(t, term, screenTimeout, "tag in a new screen",
		func(s []string) bool {
			return line(s, 0) == "afunc caller" &&
				statusOn(s, 11, "file.txt") &&
				line(s, 12) == "int afunc(void)" &&
				statusOn(s, 23, "target.c")
		})
	waitCursor(t, term, 12, 0)
}

// TestQuitFromExWithOtherScreens is a regression test: quitting an
// ex-mode screen while other screens exist must switch the terminal
// back to vi's screen mode for the remaining screen (historically
// nvi forgot to switch the tty out of ex mode on q in this case).
func TestQuitFromExWithOtherScreens(t *testing.T) {
	term, _ := startViExtra(t, "one\ntwo\n",
		map[string]string{"other.txt": "bee line\n"})

	send(term, ":Edit other.txt\r")
	waitScreen(t, term, screenTimeout, "split",
		func(s []string) bool { return line(s, 12) == "bee line" })

	// Q switches the current screen to ex mode; the other foreground
	// screen is backgrounded and the display leaves screen mode.
	send(term, "Q")
	waitExit(t, term, "Q to leave the vi screen")
	waitScreen(t, term, screenTimeout, "backgrounded notice",
		func(s []string) bool {
			for _, l := range s {
				if strings.Contains(l, "1 screens backgrounded") {
					return true
				}
			}
			return false
		})

	// q quits the ex screen; the remaining screen must come back as
	// a full vi screen.
	send(term, "q\n")
	waitScreen(t, term, startupTimeout, "vi screen restored after q",
		func(s []string) bool {
			return term.AltScreenActive() && line(s, 0) == "one" &&
				line(s, 1) == "two"
		})
	waitCursor(t, term, 0, 0)
}

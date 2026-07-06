// Mode switching and session commands (docs/nvi.md "Vi Commands"):
// ZZ, Q, <control-^>, tag push/pop, and split screens.
package nvitests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZZWritesAndExits(t *testing.T) {
	term, dir := startViExtra(t, "one\n", nil)

	send(term, "x")
	waitScreen(t, term, screenTimeout, "modification",
		func(s []string) bool { return line(s, 0) == "ne" })

	send(term, "ZZ")
	waitExit(t, term, "ZZ")

	data, err := os.ReadFile(filepath.Join(dir, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "ne\n" {
		t.Fatalf("file after ZZ: %q, want %q", got, "ne\n")
	}
}

func TestQSwitchesToExAndBack(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	// Q leaves the vi screen for line-oriented ex mode.
	send(term, "Q")
	waitExit(t, term, "Q to leave the vi screen")

	// Ex is live: print a line, then return to vi mode.
	send(term, "1p\n")
	send(term, "vi\n")
	waitScreen(t, term, startupTimeout, "vi command to return",
		func(s []string) bool {
			return term.AltScreenActive() && line(s, 0) == "one"
		})
}

func TestControlCaretAlternateFile(t *testing.T) {
	term, _ := startViExtra(t, "one\n",
		map[string]string{"other.txt": "bee line\n"})

	send(term, ":e other.txt\r")
	waitScreen(t, term, screenTimeout, "editing other.txt",
		func(s []string) bool { return line(s, 0) == "bee line" })

	send(term, "\x1e") // <control-^>: back to the previous file
	waitScreen(t, term, screenTimeout, "alternate file",
		func(s []string) bool {
			return line(s, 0) == "one" &&
				strings.Contains(s[rows-1], "file.txt")
		})
}

func TestTagPushAndPop(t *testing.T) {
	term, _ := startViExtra(t, "afunc caller\n", map[string]string{
		"tags":     "afunc\ttarget.c\t/^int afunc/\n",
		"target.c": "int afunc(void)\n{\n}\n",
	})

	send(term, "\x1d") // <control-]>: push tag for cursor word
	waitScreen(t, term, screenTimeout, "tag push to target.c",
		func(s []string) bool {
			return line(s, 0) == "int afunc(void)" &&
				strings.Contains(s[rows-1], "target.c")
		})

	send(term, "\x14") // <control-T>: pop back
	waitScreen(t, term, screenTimeout, "tag pop back",
		func(s []string) bool { return line(s, 0) == "afunc caller" })
}

func TestVsplitAndControlW(t *testing.T) {
	term, _ := startViExtra(t, "one\n",
		map[string]string{"other.txt": "bee line\n"})

	send(term, ":vsplit other.txt\r")
	waitScreen(t, term, screenTimeout, "vertical split",
		func(s []string) bool {
			return strings.Contains(s[0], "bee line") &&
				strings.Contains(s[0], "one")
		})
	_, startCol := term.Cursor()

	// ^W moves to the next screen; the cursor crosses to the other
	// half of the display.
	send(term, "\x17")
	waitScreen(t, term, screenTimeout, "^W to switch screens",
		func(s []string) bool {
			_, c := term.Cursor()
			return (startCol < cols/2) != (c < cols/2)
		})
}

func TestVersionCommand(t *testing.T) {
	term := startVi(t, "one\n")

	// version prints the editor's identification on the message
	// line.  The date embedded in the string changes with each
	// build, so match only the stable framing.
	send(term, ":version\r")
	waitScreen(t, term, screenTimeout, "version string",
		func(s []string) bool {
			msg := line(s, rows-1)
			return strings.HasPrefix(msg, "Version") &&
				strings.Contains(msg, "Berkeley")
		})
}

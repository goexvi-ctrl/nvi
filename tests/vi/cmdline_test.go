// Colon command-line editing (docs/nvi.md "Command Editing" and the
// cedit/filec entries under "Set Options"): file name completion
// and the command-editing window.
//
// Setting these options to a control character has its own traps:
// in an rc file a bare tab is eaten as argument whitespace, so the
// value must be backslash-escaped; on the colon line cedit can be
// entered with a literal-next escape.
package nvitests

import (
	"strings"
	"testing"
)

func TestFileCompletionUnique(t *testing.T) {
	term := startViStartup(t, map[string]string{
		".nexrc":   "set filec=\\\t\n",
		"beta.txt": "bee line\n",
	}, nil, false)

	// A unique match replaces the partial name on the colon line.
	send(term, ":e bet\x09")
	waitScreen(t, term, screenTimeout, "completion to beta.txt",
		func(s []string) bool { return line(s, rows-1) == ":e beta.txt" })

	send(term, "\r")
	waitScreen(t, term, screenTimeout, "completed file opened",
		func(s []string) bool {
			return line(s, 0) == "bee line" &&
				strings.Contains(s[rows-1], "beta.txt")
		})
}

func TestFileCompletionAmbiguous(t *testing.T) {
	term := startViStartup(t, map[string]string{
		".nexrc":     "set filec=\\\t\n",
		"alpha.txt":  "a\n",
		"alpine.txt": "b\n",
	}, nil, false)

	// An ambiguous match pops up the list of candidates and leaves
	// the partial name in place.
	send(term, ":e alp\x09")
	waitScreen(t, term, screenTimeout, "candidate list",
		func(s []string) bool {
			return strings.Contains(line(s, rows-2),
				"alpha.txt   alpine.txt") &&
				line(s, rows-1) == ":e alp"
		})
	send(term, "\x1b\x1b") // abandon the command
}

func TestCeditWindow(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\n")

	// Set cedit to escape (entered with literal-next), then escape
	// on the colon line opens the command-editing window: a split
	// showing the command in an editable buffer backed by a
	// temporary file.
	send(term, ":set cedit=\x16\x1b\r")
	send(term, ":2\x1b")
	waitScreen(t, term, screenTimeout, "command-editing window",
		func(s []string) bool {
			var haveCmd, haveTemp bool
			for _, l := range s {
				if strings.TrimRight(l, " ") == ":2" {
					haveCmd = true
				}
				if strings.Contains(l, "/vi.") &&
					strings.Contains(l, "modified") {
					haveTemp = true
				}
			}
			return haveCmd && haveTemp
		})

	// A carriage return executes the command under the cursor.
	send(term, "\r")
	waitCursor(t, term, 1, 0)
}

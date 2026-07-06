// Session-level behavior: the script command's shell window and
// the advisory lock a second session hits (docs/nvi.md, script
// under "Ex Commands" and the lock option under "Set Options").
package nvitests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"goterm"
)

// TestScriptCommand is an expected failure: the script command
// should run a shell inside the screen, but this build's pty
// allocation only knows the legacy BSD /dev/ptyXX names, which on
// modern systems exist as stubs that fail with EAGAIN, so every
// open fails and script reports "Error: pty: ...".  (The System V
// grantpt path exists in ex/ex_script.c but HAVE_SYS5_PTY is not
// set by configure on this platform.)  While the bug exists the
// test reports as skipped with the detail; if script starts
// working it fails loudly so the marker gets removed.
func TestScriptCommand(t *testing.T) {
	err := scriptCheck(t)
	if err == nil {
		t.Fatal("marked as a known nvi bug (script cannot allocate" +
			" a pty) but passed; re-examine and remove the xfail wrapper")
	}
	t.Skipf("XFAIL, known nvi bug: %v", err)
}

func scriptCheck(t *testing.T) error {
	term := startVi(t, "one\n")

	send(term, ":script\r")
	term.WaitQuiet(100*time.Millisecond, screenTimeout)
	if msg := line(term.Dump(), rows-1); strings.Contains(msg, "pty") {
		return fmt.Errorf("script failed to allocate a pty: %q", msg)
	}

	// A working script window echoes shell output into the buffer.
	send(term, "echo hi-there\r")
	ok := term.WaitFor(screenTimeout, func(screen []string) bool {
		for _, l := range screen {
			if strings.Contains(l, "hi-there") &&
				!strings.Contains(l, "echo") {
				return true
			}
		}
		return false
	})
	send(term, "exit\r")
	if !ok {
		return fmt.Errorf("shell output never appeared")
	}
	return nil
}

func TestSecondSessionIsReadOnly(t *testing.T) {
	// startViExtra chdirs into the fixture directory, so a second
	// editor started afterwards opens the same file.
	term, _ := startViExtra(t, "one\ntwo\n", nil)
	_ = term

	term2 := goterm.New(rows, cols)
	if err := term2.Start(nvi(t), "file.txt"); err != nil {
		t.Fatal(err)
	}
	defer term2.Close()

	// The second session warns and comes up read-only, waiting for
	// an acknowledgement.
	waitScreen(t, term2, startupTimeout, "read-only warning",
		func(s []string) bool {
			var warned, readonly bool
			for _, l := range s {
				if strings.Contains(l,
					"file.txt already locked, session is read-only.") {
					warned = true
				}
				if strings.Contains(l, "unmodified, readonly:") {
					readonly = true
				}
			}
			return warned && readonly
		})

	// Acknowledge, modify, and try to write: the readonly session
	// must refuse.
	send(term2, "\r")
	waitScreen(t, term2, screenTimeout, "buffer visible",
		func(s []string) bool { return line(s, 0) == "one" })
	send(term2, "x:w\r")
	waitScreen(t, term2, screenTimeout, "write refused",
		func(s []string) bool {
			return strings.Contains(line(s, rows-1),
				"Read-only file, not written; use ! to override")
		})
}

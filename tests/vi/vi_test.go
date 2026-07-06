// Functional tests for full-screen vi mode.  Each test starts
// build.unix/vi on a headless 24x80 ANSI terminal (the goterm
// package) and asserts on the rendered screen and cursor position.
//
// The assertions encode behavior from docs/nvi.md ("Vi Commands"):
// screen layout at startup, motions, simple edits, undo, search, and
// the modified-file quit protection.
package nvitests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"goterm"
)

const (
	rows = 24
	cols = 80

	startupTimeout = 10 * time.Second
	screenTimeout  = 5 * time.Second
)

var (
	nviOnce sync.Once
	nviPath string
	nviErr  error
)

// nvi returns the absolute path of the oracle binary, skipping the
// test if it has not been built.
func nvi(t *testing.T) string {
	t.Helper()
	nviOnce.Do(func() {
		nviPath, nviErr = filepath.Abs(filepath.Join("..", "..", "build.unix", "vi"))
	})
	if nviErr != nil {
		t.Fatal(nviErr)
	}
	if _, err := os.Stat(nviPath); err != nil {
		t.Skipf("nvi binary not built: %v (make in build.unix first)", err)
	}
	return nviPath
}

// unsetenvForTest removes a variable for the duration of the test so
// the user's startup files and options cannot leak in.
func unsetenvForTest(t *testing.T, key string) {
	t.Helper()
	if old, ok := os.LookupEnv(key); ok {
		os.Unsetenv(key)
		t.Cleanup(func() { os.Setenv(key, old) })
	}
}

// startVi writes content to file.txt in a fresh scratch directory,
// moves there (this build of nvi keeps its recovery files in
// ./vi.recover), and starts vi on it.
func startVi(t *testing.T, content string) *goterm.Term {
	t.Helper()
	term, _ := startViExtra(t, content, nil)
	return term
}

// startViExtra is startVi plus extra fixture files (tags files,
// second buffers, ...) in the scratch directory; it also returns the
// directory so tests can inspect files after the editor exits.
func startViExtra(t *testing.T, content string,
	extra map[string]string) (*goterm.Term, string) {
	t.Helper()

	bin := nvi(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file.txt"),
		[]byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, data := range extra {
		if err := os.WriteFile(filepath.Join(dir, name),
			[]byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("HOME", dir)
	unsetenvForTest(t, "EXINIT")
	unsetenvForTest(t, "NEXINIT")
	t.Chdir(dir)

	term := goterm.New(rows, cols)
	if err := term.Start(bin, "file.txt"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { term.Close() })

	// The bottom line shows the file name once vi is up.
	waitScreen(t, term, startupTimeout, "startup status line",
		func(s []string) bool {
			return strings.Contains(s[rows-1], "file.txt")
		})
	return term, dir
}

// waitExit waits for vi to end the session, observable as leaving
// the alternate screen.
func waitExit(t *testing.T, term *goterm.Term, what string) {
	t.Helper()
	deadline := time.Now().Add(screenTimeout)
	for term.AltScreenActive() {
		if time.Now().After(deadline) {
			t.Fatalf("%s did not exit; screen:\n%s",
				what, strings.Join(term.Dump(), "\n"))
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// numberedLines returns "l1\nl2\n...ln\n", for screen position tests.
func numberedLines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "l%d\n", i)
	}
	return b.String()
}

func waitScreen(t *testing.T, term *goterm.Term, timeout time.Duration,
	what string, pred func([]string) bool) {
	t.Helper()
	if !term.WaitFor(timeout, pred) {
		t.Fatalf("timed out waiting for %s; screen:\n%s",
			what, strings.Join(term.Dump(), "\n"))
	}
}

func waitCursor(t *testing.T, term *goterm.Term, row, col int) {
	t.Helper()
	deadline := time.Now().Add(screenTimeout)
	for {
		r, c := term.Cursor()
		if r == row && c == col {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for cursor at %d,%d; at %d,%d; screen:\n%s",
				row, col, r, c, strings.Join(term.Dump(), "\n"))
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func send(term *goterm.Term, s string) { term.Send([]byte(s)) }

func line(s []string, n int) string { return strings.TrimRight(s[n], " ") }

func TestStartupScreen(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	waitScreen(t, term, screenTimeout, "file contents",
		func(s []string) bool {
			return line(s, 0) == "one" && line(s, 1) == "two"
		})

	// Lines past the end of file display as single tilde columns
	// (docs/nvi.md "Sizing the Screen").
	s := term.Dump()
	for i := 2; i < rows-1; i++ {
		if got := line(s, i); got != "~" {
			t.Errorf("row %d: want ~, got %q", i, got)
		}
	}
	waitCursor(t, term, 0, 0)
}

func TestMotions(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\n")

	send(term, "j")
	waitCursor(t, term, 1, 0)
	send(term, "l")
	waitCursor(t, term, 1, 1)
	send(term, "j$")
	waitCursor(t, term, 2, 4) // on the last character of "three"
	send(term, "0")
	waitCursor(t, term, 2, 0)
	send(term, "1G")
	waitCursor(t, term, 0, 0)
}

func TestDeleteCharacter(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, "x")
	waitScreen(t, term, screenTimeout, "x to delete the o",
		func(s []string) bool { return line(s, 0) == "ne" })
}

func TestDeleteLineAndUndo(t *testing.T) {
	term := startVi(t, "one\ntwo\nthree\n")

	send(term, "dd")
	waitScreen(t, term, screenTimeout, "dd to delete line one",
		func(s []string) bool {
			return line(s, 0) == "two" && line(s, 1) == "three" &&
				line(s, 2) == "~"
		})

	send(term, "u")
	waitScreen(t, term, screenTimeout, "u to restore line one",
		func(s []string) bool {
			return line(s, 0) == "one" && line(s, 1) == "two"
		})
}

func TestInsert(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, "iX")
	waitScreen(t, term, screenTimeout, "inserted character",
		func(s []string) bool { return line(s, 0) == "Xone" })
	send(term, "\x1b") // <escape> ends input mode
	waitCursor(t, term, 0, 0)
}

func TestOpenLineBelow(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	send(term, "onew line\x1b")
	waitScreen(t, term, screenTimeout, "o to open a line below",
		func(s []string) bool {
			return line(s, 0) == "one" && line(s, 1) == "new line" &&
				line(s, 2) == "two"
		})
}

func TestSearchMovesCursor(t *testing.T) {
	term := startVi(t, "alpha\nbeta\ngamma\n")

	send(term, "/gamma\r")
	waitCursor(t, term, 2, 0)

	// n with wrapscan (the default) wraps back to the same match.
	send(term, "1Gn")
	waitCursor(t, term, 2, 0)
}

func TestModifiedQuitProtection(t *testing.T) {
	term := startVi(t, "one\n")

	send(term, "x")
	waitScreen(t, term, screenTimeout, "modification on screen",
		func(s []string) bool { return line(s, 0) == "ne" })

	// :q on a modified file is refused with a message.
	send(term, ":q\r")
	waitScreen(t, term, screenTimeout, "modified warning",
		func(s []string) bool {
			return strings.Contains(strings.Join(s, "\n"), "modified")
		})

	// :q! discards the change; vi leaves the alternate screen on exit.
	send(term, ":q!\r")
	waitExit(t, term, ":q!")
}

func TestControlGStatus(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	// <control-G> reports the path and position information
	// (docs/nvi.md "Vi Commands", <control-G>).
	send(term, "\x07")
	waitScreen(t, term, screenTimeout, "control-G status",
		func(s []string) bool {
			last := s[rows-1]
			return strings.Contains(last, "file.txt") &&
				strings.Contains(last, "line 1 of 2")
		})
}

func TestJoinLines(t *testing.T) {
	term := startVi(t, "one\ntwo\n")

	send(term, "J")
	waitScreen(t, term, screenTimeout, "J to join lines",
		func(s []string) bool {
			return line(s, 0) == "one two" && line(s, 1) == "~"
		})
}

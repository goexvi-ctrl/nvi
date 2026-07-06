// Startup information (docs/nvi.md "Startup Information"): the
// NEXINIT and EXINIT environment variables, the $HOME/.nexrc and
// $HOME/.exrc files, and the exrc option for local directory rc
// files.  Sources are tried in order (NEXINIT, EXINIT, .nexrc,
// .exrc) and the first one found wins.
//
// These tests run in vi mode because ex batch mode never reads
// startup information: with standard input redirected the editor
// sets the scripted flag, which implies -s.  The report option is
// the witness (default report=5).
package nvitests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goterm"
)

// startViStartup starts the editor with startup files and
// environment variables in place.  File names may contain a
// "home/" prefix; with sepHome the HOME directory is dir/home,
// separate from the working directory.
func startViStartup(t *testing.T, files map[string]string,
	env map[string]string, sepHome bool) *goterm.Term {
	t.Helper()

	bin := nvi(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file.txt"),
		[]byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	home := dir
	if sepHome {
		home = filepath.Join(dir, "home")
		if err := os.Mkdir(home, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("HOME", home)
	unsetenvForTest(t, "EXINIT")
	unsetenvForTest(t, "NEXINIT")
	for k, v := range env {
		t.Setenv(k, v)
	}
	t.Chdir(dir)

	term := goterm.New(rows, cols)
	if err := term.Start(bin, "file.txt"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { term.Close() })
	waitScreen(t, term, startupTimeout, "startup status line",
		func(s []string) bool {
			return strings.Contains(s[rows-1], "file.txt")
		})
	return term
}

// reportIs queries the report option and waits for the answer on
// the message line.
func reportIs(t *testing.T, term *goterm.Term, want string) {
	t.Helper()
	send(term, ":set report?\r")
	waitScreen(t, term, screenTimeout, "report="+want,
		func(s []string) bool { return line(s, rows-1) == "report="+want })
}

func TestNexrcRead(t *testing.T) {
	term := startViStartup(t,
		map[string]string{".nexrc": "set report=7\n"}, nil, false)
	reportIs(t, term, "7")
}

func TestNexrcPreferredOverExrc(t *testing.T) {
	term := startViStartup(t, map[string]string{
		".nexrc": "set report=7\n",
		".exrc":  "set report=6\n",
	}, nil, false)
	reportIs(t, term, "7")
}

func TestExrcReadWithoutNexrc(t *testing.T) {
	term := startViStartup(t,
		map[string]string{".exrc": "set report=6\n"}, nil, false)
	reportIs(t, term, "6")
}

func TestNexinitBeatsExinitAndRcFiles(t *testing.T) {
	term := startViStartup(t,
		map[string]string{".nexrc": "set report=7\n"},
		map[string]string{
			"NEXINIT": "set report=8",
			"EXINIT":  "set report=9",
		}, false)
	reportIs(t, term, "8")
}

func TestExinitBeatsRcFiles(t *testing.T) {
	term := startViStartup(t,
		map[string]string{".nexrc": "set report=7\n"},
		map[string]string{"EXINIT": "set report=9"}, false)
	reportIs(t, term, "9")
}

func TestExrcOptionReadsLocalRc(t *testing.T) {
	// With exrc set in the HOME startup file, a .exrc in the
	// current directory is read afterwards.
	term := startViStartup(t, map[string]string{
		"home/.nexrc": "set exrc\n",
		".exrc":       "set report=9\n",
	}, nil, true)
	reportIs(t, term, "9")
}

func TestLocalRcIgnoredByDefault(t *testing.T) {
	// Without the exrc option the local .exrc must not be read.
	term := startViStartup(t, map[string]string{
		".exrc": "set report=9\n",
	}, nil, true)
	reportIs(t, term, "5")
}

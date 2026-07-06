// Crash recovery (docs/nvi.md section 4, "Recovery"): a session
// killed by SIGHUP must leave a recovery file that vi -r restores.
// This is the path the recovery design actually promises; contrast
// with the ex :preserve command, which is a known broken XFAIL in
// ex/t_misc.t.
package nvitests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// viPID finds the running editor's process ID.  The libtool wrapper
// execs the real binary from .libs, so match on that.
func viPID(t *testing.T) int {
	t.Helper()
	out, err := exec.Command("pgrep", "-n", "-f", ".libs/vi file.txt").Output()
	if err != nil {
		t.Fatalf("pgrep found no vi process: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatalf("bad pgrep output %q: %v", out, err)
	}
	return pid
}

// exBatch runs a second editor in ex batch mode in dir and returns
// its standard output with the terminal-ioctl noise stripped.
func exBatch(t *testing.T, dir, script string, args ...string) string {
	t.Helper()
	cmd := exec.Command(nvi(t), args...)
	cmd.Dir = dir
	cmd.Env = []string{"HOME=" + dir, "TERM=dumb", "LC_ALL=C",
		"PATH=/usr/bin:/bin"}
	cmd.Stdin = strings.NewReader(script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run() // recovery listings can exit nonzero; assert on output
	var keep []string
	for _, l := range strings.Split(stdout.String(), "\n") {
		if !strings.HasPrefix(l, "Error: stderr: ") {
			keep = append(keep, l)
		}
	}
	return strings.Join(keep, "\n")
}

// TestSighupRecovery is an expected failure (the vi-harness
// equivalent of an ex XFAIL): docs/nvi.md section 4 promises that a
// hangup preserves the modified buffer for vi -r, but this build
// writes a zero-length backup (same root failure as the :preserve
// XFAIL in ex/t_misc.t), vi -r reports "No files to recover", and
// recovery restores an empty buffer.  The test asserts the
// documented behavior; while the bug exists it reports as skipped
// with the failure detail, and if recovery is ever fixed it fails
// loudly so the marker gets removed.
func TestSighupRecovery(t *testing.T) {
	err := sighupRecoveryCheck(t)
	if err == nil {
		t.Fatal("marked as a known nvi bug (broken crash recovery)" +
			" but passed; re-examine and remove the xfail wrapper")
	}
	t.Skipf("XFAIL, known nvi bug, recovery is broken: %v", err)
}

// sighupRecoveryCheck runs the documented recovery flow and returns
// nil only if every step behaves as documented.
func sighupRecoveryCheck(t *testing.T) error {
	term, dir := startViExtra(t, "one\ntwo\n", nil)

	// Make an unwritten change.
	send(term, "oadded\x1b")
	waitScreen(t, term, screenTimeout, "modification",
		func(s []string) bool {
			return line(s, 0) == "one" && line(s, 1) == "added" &&
				line(s, 2) == "two"
		})

	// The hangup a dropped connection would deliver.
	pid := viPID(t)
	if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
		t.Fatal(err)
	}
	// The editor is a child of the goterm session, which only reaps
	// it on Close, so "exited" here means gone or zombie.
	deadline := time.Now().Add(screenTimeout)
	for {
		out, err := exec.Command("ps", "-o", "stat=", "-p",
			strconv.Itoa(pid)).Output()
		if err != nil || strings.HasPrefix(strings.TrimSpace(string(out)), "Z") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("vi did not exit on SIGHUP (state %q)", out)
		}
		time.Sleep(10 * time.Millisecond)
	}
	term.Close() // reap, and release the pty

	// The recovery area must hold a metadata file and a non-empty
	// backup (the zero-length backup is the shared root failure
	// with :preserve).
	recdir := filepath.Join(dir, "vi.recover")
	entries, err := os.ReadDir(recdir)
	if err != nil {
		return fmt.Errorf("no recovery directory after SIGHUP: %v", err)
	}
	var haveMeta, haveBackup bool
	for _, e := range entries {
		info, ierr := e.Info()
		if ierr != nil {
			continue
		}
		switch {
		case strings.HasPrefix(e.Name(), "recover."):
			haveMeta = true
		case strings.HasPrefix(e.Name(), "vi."):
			if info.Size() > 0 {
				haveBackup = true
			}
		}
	}
	if !haveMeta || !haveBackup {
		return fmt.Errorf("recovery area incomplete after SIGHUP "+
			"(metadata %v, non-empty backup %v): %v",
			haveMeta, haveBackup, names(entries))
	}

	// vi -r with no file lists what is recoverable.
	listing := exBatch(t, dir, "q!\n", "-e", "-s", "-r")
	if !strings.Contains(listing, "file.txt") {
		return fmt.Errorf("vi -r listing does not mention file.txt:\n%s",
			listing)
	}

	// vi -r file.txt restores the unwritten change.
	got := exBatch(t, dir, "%p\nq!\n", "-e", "-s", "-r", "file.txt")
	if want := "one\nadded\ntwo\n"; got != want {
		return fmt.Errorf("recovered buffer %q, want %q", got, want)
	}
	return nil
}

func names(entries []os.DirEntry) []string {
	var ns []string
	for _, e := range entries {
		ns = append(ns, e.Name())
	}
	return ns
}

// Cscope support (docs/nvi.md section 8, "Tags, Tag Stacks, and
// Cscope"): connections, queries, and their tag-stack integration.
// The tests build a real cscope database over a two-file C fixture;
// they skip if cscope is not installed.
package nvitests

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/goexvi-ctrl/goterm"
)

const csMain = "int helper(int x);\n\nint main(void)\n{\n\treturn helper(1);\n}\n"
const csHelper = "int helper(int x)\n{\n\treturn x + 1;\n}\n"

// startCscope starts vi on the C fixture with a cscope database
// built alongside it, and adds the connection.
func startCscope(t *testing.T) *goterm.Term {
	t.Helper()
	if _, err := exec.LookPath("cscope"); err != nil {
		t.Skip("cscope not installed")
	}
	term, dir := startViExtra(t, csMain,
		map[string]string{"helper.c": csHelper})
	cmd := exec.Command("cscope", "-b", "-c", "helper.c", "file.txt")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cscope -b failed: %v: %s", err, out)
	}

	send(term, ":cs add cscope.out\r")
	// The add is silent; prove the connection with display.
	send(term, ":display connections\r")
	waitScreen(t, term, screenTimeout, "cscope connection listed",
		func(s []string) bool {
			return strings.HasPrefix(line(s, rows-1), " 1 . (process")
		})
	return term
}

func TestCscopeFindDefinition(t *testing.T) {
	term := startCscope(t)

	// find g: the global definition, in the other file.
	send(term, ":cs find g helper\r")
	waitScreen(t, term, screenTimeout, "definition in helper.c",
		func(s []string) bool {
			return line(s, 0) == "int helper(int x)" &&
				strings.Contains(s[rows-1], "helper.c")
		})

	// The query pushed the tag stack; ^T returns.
	send(term, "\x14")
	waitScreen(t, term, screenTimeout, "^T back to the caller",
		func(s []string) bool {
			return line(s, 0) == "int helper(int x);" &&
				strings.Contains(s[rows-1], "file.txt")
		})
}

func TestCscopeFindCallers(t *testing.T) {
	term := startCscope(t)

	// find c: the call site, on the return line of main.
	send(term, ":cs find c helper\r")
	waitScreen(t, term, screenTimeout, "caller line",
		func(s []string) bool {
			return strings.Contains(s[rows-1], "file.txt") &&
				strings.Contains(s[rows-1], "line 5")
		})
	waitCursor(t, term, 4, 8)
}

func TestCscopeHelp(t *testing.T) {
	term := startCscope(t)

	send(term, ":cs help\r")
	waitScreen(t, term, screenTimeout, "cscope help text",
		func(s []string) bool {
			var haveTitle, haveAdd bool
			for _, l := range s {
				if strings.HasPrefix(l, "cscope commands:") {
					haveTitle = true
				}
				if strings.Contains(l, "add: Add a new cscope database") {
					haveAdd = true
				}
			}
			return haveTitle && haveAdd
		})
	send(term, "\r") // dismiss the continue prompt
}

func TestCscopeReset(t *testing.T) {
	term := startCscope(t)

	// reset discards the connection; queries then fail.
	send(term, ":cs reset\r")
	send(term, ":cs find g helper\r")
	waitScreen(t, term, screenTimeout, "query after reset fails",
		func(s []string) bool {
			return strings.Contains(line(s, rows-1), "cscope")
		})
}

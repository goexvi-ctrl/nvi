module nvitests

go 1.26.1

require goterm v0.0.0

require (
	github.com/clipperhouse/uax29/v2 v2.2.0 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/pborman/ansi v1.2.0 // indirect
)

// goterm is used from a sibling checkout of the nvi tree.
replace goterm => ../../../goterm

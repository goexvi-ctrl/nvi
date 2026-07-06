# Run the nvi test suite.  See README.md.
#
# The nvi oracle (build.unix/vi and its objects) must already be built.

all: test

test: unit ex vi

unit:
	$(MAKE) -C unit/regex run
	$(MAKE) -C unit/db run

ex:
	sh ex/run.sh ex/t_*.t

# Deterministic crash hunt; not part of the default run (slow, and
# it exercises known crashes on purpose).  Fails only on a crash
# that is not the known print/auto-print bug.
chaos:
	sh ex/chaos.sh

vi:
	@if [ -d ../../goterm ] && command -v go >/dev/null 2>&1; then \
		cd vi && go test ./...; \
	else \
		echo "vi: skipped (needs go and a goterm checkout at ../../goterm)"; \
	fi

clean:
	$(MAKE) -C unit/regex clean
	$(MAKE) -C unit/db clean

.PHONY: all test unit ex vi chaos clean

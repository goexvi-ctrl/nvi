# Unit tests for nvi's bundled Spencer regex library.
#
# Links against the objects already compiled in build.unix, so the
# tests exercise exactly the code nvi runs.  Build nvi first.

TOP = ../../..
BUILD = $(TOP)/build.unix

CC = cc
CFLAGS = -g -Wall -D__REGEX_PRIVATE -I$(BUILD) -I$(TOP)/regex

REGEX_OBJS = $(BUILD)/regcomp.o $(BUILD)/regerror.o \
	$(BUILD)/regexec.o $(BUILD)/regfree.o

run: test_regex
	./test_regex

test_regex: test_regex.o
	@for o in $(REGEX_OBJS); do \
		test -f $$o || { \
		    echo "missing $$o: build nvi first (make in build.unix)"; \
		    exit 1; \
		}; \
	done
	$(CC) -o test_regex test_regex.o $(REGEX_OBJS)

test_regex.o: test_regex.c
	$(CC) $(CFLAGS) -c test_regex.c

clean:
	rm -f test_regex test_regex.o

.PHONY: run clean

# Functional tests for the bundled db.1.85 database library (btree,
# hash, and recno access methods) that nvi stores its buffers in.
#
# Builds the library's own dbtest driver (db.1.85/test/dbtest.c)
# against the objects already compiled in build.unix, then runs the
# script batteries from db.1.85/test/run.test.  Build nvi first.

TOP = ../../..
BUILD = $(TOP)/build.unix
DBTEST_SRC = $(TOP)/db.1.85/test

CC = cc
# dbtest.c is K&R-era code that modern compilers complain about at
# length; the warnings are not worth hearing about on every build.
CFLAGS = -g -Wno-deprecated-non-prototype -Wno-format \
	-Wno-macro-redefined -Wno-return-type \
	-D__DBINTERFACE_PRIVATE -DDEBUG \
	-I$(BUILD) -I$(TOP)/db.1.85/include -I$(TOP)/include

DB_OBJS = $(BUILD)/bt_close.o $(BUILD)/bt_conv.o $(BUILD)/bt_debug.o \
	$(BUILD)/bt_delete.o $(BUILD)/bt_get.o $(BUILD)/bt_open.o \
	$(BUILD)/bt_overflow.o $(BUILD)/bt_page.o $(BUILD)/bt_put.o \
	$(BUILD)/bt_search.o $(BUILD)/bt_seq.o $(BUILD)/bt_split.o \
	$(BUILD)/bt_utils.o $(BUILD)/db.o $(BUILD)/hash.o \
	$(BUILD)/hash_bigkey.o $(BUILD)/hash_buf.o $(BUILD)/hash_func.o \
	$(BUILD)/hash_log2.o $(BUILD)/hash_page.o $(BUILD)/mpool.o \
	$(BUILD)/rec_close.o $(BUILD)/rec_delete.o $(BUILD)/rec_get.o \
	$(BUILD)/rec_open.o $(BUILD)/rec_put.o $(BUILD)/rec_search.o \
	$(BUILD)/rec_seq.o $(BUILD)/rec_utils.o

run: dbtest
	sh ./run_db.sh

dbtest: dbtest.o
	@for o in $(DB_OBJS); do \
		test -f $$o || { \
		    echo "missing $$o: build nvi first (make in build.unix)"; \
		    exit 1; \
		}; \
	done
	$(CC) -o dbtest dbtest.o $(DB_OBJS)

dbtest.o: $(DBTEST_SRC)/dbtest.c
	$(CC) $(CFLAGS) -c $(DBTEST_SRC)/dbtest.c

clean:
	rm -f dbtest dbtest.o

.PHONY: run clean

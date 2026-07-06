/*
 * Unit tests for nvi's bundled Henry Spencer regex library, the
 * exact objects nvi links (build.unix/regcomp.o and friends).
 *
 * This build of nvi has USE_WIDECHAR off, so RCHAR_T is char and the
 * API matches POSIX regcomp/regexec/regerror/regfree.
 *
 * The behavior asserted here is the POSIX 1003.2 RE behavior that
 * docs/nvi.md section 9 defers to, plus this library's documented
 * extensions (regex/regex.3, regex/re_format.7): REG_PEND,
 * REG_STARTEND, REG_NOSPEC, and the [[:<:]] / [[:>:]] word
 * boundaries that nvi generates when a user writes \< and \>.
 *
 * The magic/nomagic translation (\<, ~, etc.) happens in ex above
 * this library; it is covered by the functional tests in tests/ex.
 */
#include <sys/types.h>
#include <stdarg.h>
#include <stdio.h>
#include <string.h>

#include "regex.h"

static int nrun, nfail;

static void
check(int line, int ok, const char *fmt, ...)
{
	va_list ap;

	nrun++;
	if (ok)
		return;
	nfail++;
	(void)printf("FAIL test_regex.c:%d: ", line);
	va_start(ap, fmt);
	(void)vprintf(fmt, ap);
	va_end(ap);
	(void)printf("\n");
}

#define	CHECK(cond, ...)	check(__LINE__, (cond) != 0, __VA_ARGS__)

/*
 * CHECK_XFAIL is for known library bugs: the condition states the
 * correct behavior, the check "passes" while the bug keeps it false,
 * and it fails loudly if the behavior is ever fixed so the marker
 * gets noticed and removed.
 */
static void
check_xfail(int line, int ok, const char *fmt, ...)
{
	va_list ap;

	nrun++;
	if (!ok)
		return;
	nfail++;
	(void)printf("XPASS test_regex.c:%d: ", line);
	va_start(ap, fmt);
	(void)vprintf(fmt, ap);
	va_end(ap);
	(void)printf("\n    marked as a known bug but now behaves"
	    " correctly; remove the CHECK_XFAIL\n");
}

#define	CHECK_XFAIL(cond, ...) \
	check_xfail(__LINE__, (cond) != 0, __VA_ARGS__)

/*
 * comp_err --
 *	Return the regcomp error code for a pattern (0 if it compiles).
 */
static int
comp_err(const char *pat, int cflags)
{
	regex_t re;
	int r;

	r = regcomp(&re, pat, cflags);
	if (r == 0)
		regfree(&re);
	return (r);
}

/*
 * mspan --
 *	Compile pat, run it against str, and report subexpression n's
 *	span through sop/eop (-1/-1 when there is no match).  Returns
 *	the regexec status, or the regcomp status + 1000 if the
 *	pattern does not compile.
 */
static int
mspan(const char *pat, int cflags, const char *str, int eflags,
    size_t n, regoff_t *sop, regoff_t *eop)
{
	regex_t re;
	regmatch_t pmatch[10];
	int r;
	size_t i;

	*sop = *eop = -1;
	if ((r = regcomp(&re, pat, cflags)) != 0)
		return (r + 1000);
	for (i = 0; i < 10; i++)
		pmatch[i].rm_so = pmatch[i].rm_eo = -1;
	r = regexec(&re, str, 10, pmatch, eflags);
	if (r == 0 && n <= re.re_nsub) {
		*sop = pmatch[n].rm_so;
		*eop = pmatch[n].rm_eo;
	}
	regfree(&re);
	return (r);
}

/*
 * m --
 *	True if pat matches somewhere in str.
 */
static int
m(const char *pat, int cflags, const char *str, int eflags)
{
	regoff_t so, eo;

	return (mspan(pat, cflags, str, eflags, 0, &so, &eo) == 0);
}

/*
 * span_is --
 *	True if pat's whole-match span in str is exactly [so, eo).
 */
static int
span_is(const char *pat, int cflags, const char *str,
    regoff_t so, regoff_t eo)
{
	regoff_t gso, geo;

	if (mspan(pat, cflags, str, 0, 0, &gso, &geo) != 0)
		return (0);
	return (gso == so && geo == eo);
}

static void
test_literal_and_dot(void)
{
	regoff_t so, eo;

	CHECK(m("abc", REG_BASIC, "xabcy", 0), "literal in middle");
	CHECK(!m("abc", REG_BASIC, "abd", 0), "literal non-match");
	CHECK(span_is("abc", REG_BASIC, "xabcy", 1, 4),
	    "literal match offsets");

	CHECK(m("a.c", REG_BASIC, "axc", 0), "dot matches any character");
	CHECK(!m("a.c", REG_BASIC, "ac", 0), "dot requires a character");
	CHECK(m("a.c", REG_BASIC, "a\nc", 0),
	    "dot matches newline without REG_NEWLINE");

	/* Escaped specials are literal. */
	CHECK(m("a\\.c", REG_BASIC, "a.c", 0), "escaped dot is literal");
	CHECK(!m("a\\.c", REG_BASIC, "axc", 0),
	    "escaped dot is not a wildcard");

	/* First-position * is an ordinary character in a BRE. */
	CHECK(m("*a", REG_BASIC, "*a", 0), "leading * literal in BRE");

	/* Leftmost-longest: the match starts as early as possible... */
	mspan("a*", REG_BASIC, "baaac", 0, 0, &so, &eo);
	CHECK(so == 0 && eo == 0,
	    "a* prefers leftmost (empty) match: got %ld-%ld",
	    (long)so, (long)eo);
	/* ...and is as long as possible from there. */
	CHECK(span_is("ba*", REG_BASIC, "xbaaay", 1, 5), "closure is greedy");
}

static void
test_anchors(void)
{
	CHECK(m("^ab", REG_BASIC, "abc", 0), "^ at start");
	CHECK(!m("^ab", REG_BASIC, "cab", 0), "^ rejects mid-string");
	CHECK(m("bc$", REG_BASIC, "abc", 0), "$ at end");
	CHECK(!m("bc$", REG_BASIC, "bca", 0), "$ rejects mid-string");
	CHECK(m("^$", REG_BASIC, "", 0), "^$ matches empty string");

	CHECK(!m("^a", REG_BASIC, "abc", REG_NOTBOL),
	    "REG_NOTBOL defeats ^");
	CHECK(!m("c$", REG_BASIC, "abc", REG_NOTEOL),
	    "REG_NOTEOL defeats $");

	/* In a BRE, ^ and $ are only special at the ends. */
	CHECK(m("a^b", REG_BASIC, "a^b", 0), "interior ^ literal in BRE");
	CHECK(m("a$b", REG_BASIC, "a$b", 0), "interior $ literal in BRE");
}

static void
test_brackets(void)
{
	CHECK(m("[abc]x", REG_BASIC, "bx", 0), "bracket list");
	CHECK(!m("[abc]x", REG_BASIC, "dx", 0), "bracket list rejects");
	CHECK(m("[^abc]x", REG_BASIC, "dx", 0), "negated list");
	CHECK(!m("[^abc]x", REG_BASIC, "ax", 0), "negated list rejects");
	CHECK(m("[a-z]x", REG_BASIC, "qx", 0), "range");
	CHECK(!m("[a-z]x", REG_BASIC, "Qx", 0), "range is case sensitive");
	CHECK(m("[]x]a", REG_BASIC, "]a", 0), "leading ] is literal");
	CHECK(m("[a-]b", REG_BASIC, "-b", 0), "trailing - is literal");
	CHECK(m("[[:digit:]]", REG_BASIC, "x5y", 0), "character class");
	CHECK(!m("[[:digit:]]", REG_BASIC, "xyz", 0),
	    "character class rejects");
	CHECK(m("[[:space:]]", REG_BASIC, "a\tb", 0), "space class, tab");

	/* Specials lose their meaning inside brackets. */
	CHECK(m("[.]", REG_BASIC, "a.b", 0), "dot literal in brackets");
	CHECK(!m("[.]", REG_BASIC, "axb", 0),
	    "bracket dot does not match any");
}

static void
test_bre_groups_and_backrefs(void)
{
	regex_t re;
	regoff_t so, eo;

	/* re_nsub counts escaped parentheses. */
	CHECK(regcomp(&re, "\\(a\\)\\(b\\)", REG_BASIC) == 0,
	    "two-group pattern compiles");
	CHECK(re.re_nsub == 2, "re_nsub == 2, got %lu",
	    (unsigned long)re.re_nsub);
	regfree(&re);

	/* The docs/nvi.md section 9 example. */
	mspan("abc\\(.*\\)def", REG_BASIC, "abcXYZdef", 0, 1, &so, &eo);
	CHECK(so == 3 && eo == 6,
	    "subexpression span for abc\\(.*\\)def: got %ld-%ld",
	    (long)so, (long)eo);

	CHECK(m("\\(ab\\)\\1", REG_BASIC, "abab", 0), "backreference");
	CHECK(!m("\\(ab\\)\\1", REG_BASIC, "abac", 0),
	    "backreference rejects");
	CHECK(m("\\(a*\\)b\\1", REG_BASIC, "aabaa", 0),
	    "backreference to closure");

	/* Group with closure applied. */
	CHECK(m("\\(ab\\)*c", REG_BASIC, "ababc", 0), "group closure");
	CHECK(span_is("x\\(ab\\)*", REG_BASIC, "xababy", 0, 5),
	    "group closure is greedy");
}

static void
test_bounds(void)
{
	CHECK(!m("^a\\{2,3\\}$", REG_BASIC, "a", 0), "bound rejects short");
	CHECK(m("^a\\{2,3\\}$", REG_BASIC, "aa", 0), "bound lower edge");
	CHECK(m("^a\\{2,3\\}$", REG_BASIC, "aaa", 0), "bound upper edge");
	CHECK(!m("^a\\{2,3\\}$", REG_BASIC, "aaaa", 0), "bound rejects long");
	CHECK(m("^a\\{3\\}$", REG_BASIC, "aaa", 0), "exact bound");
	CHECK(!m("^a\\{3\\}$", REG_BASIC, "aa", 0), "exact bound rejects");
	CHECK(m("^a\\{2,\\}$", REG_BASIC, "aaaaa", 0), "open bound");

	CHECK(m("^a{2,3}$", REG_EXTENDED, "aaa", 0), "ERE bound");
	CHECK(!m("^a{2,3}$", REG_EXTENDED, "aaaa", 0), "ERE bound rejects");
}

static void
test_ere(void)
{
	CHECK(m("ab|cd", REG_EXTENDED, "xcdy", 0), "alternation");
	CHECK(!m("ab|cd", REG_EXTENDED, "acbd", 0), "alternation rejects");
	CHECK(m("a+b", REG_EXTENDED, "aaab", 0), "plus");
	CHECK(!m("a+b", REG_EXTENDED, "b", 0), "plus requires one");
	CHECK(m("ab?c", REG_EXTENDED, "ac", 0), "question zero");
	CHECK(m("ab?c", REG_EXTENDED, "abc", 0), "question one");
	CHECK(m("(ab)+c", REG_EXTENDED, "ababc", 0), "group plus");
	CHECK(m("a(b|c)d", REG_EXTENDED, "acd", 0), "nested alternation");

	/* Alternation obeys leftmost-longest. */
	CHECK(span_is("ab|abc", REG_EXTENDED, "xabcy", 1, 4),
	    "leftmost-longest across alternation");

	/* In an ERE, unescaped parens group and + is special... */
	CHECK(comp_err("a+(b)", REG_EXTENDED) == 0, "ERE syntax compiles");
	/* ...while in a BRE both are ordinary characters. */
	CHECK(m("a+", REG_BASIC, "xa+y", 0), "+ literal in BRE");
	CHECK(m("(ab)", REG_BASIC, "(ab)", 0), "bare parens literal in BRE");
}

static void
test_case_and_nosub(void)
{
	regex_t re;

	CHECK(m("abc", REG_BASIC | REG_ICASE, "xABCy", 0), "REG_ICASE");
	CHECK(m("[a-z]*x", REG_BASIC | REG_ICASE, "QRSx", 0),
	    "REG_ICASE applies inside ranges");
	CHECK(!m("abc", REG_BASIC, "ABC", 0),
	    "case sensitive without REG_ICASE");

	/* REG_NOSUB compiles for match/no-match reporting only. */
	CHECK(regcomp(&re, "a\\(b\\)c", REG_BASIC | REG_NOSUB) == 0,
	    "REG_NOSUB compiles");
	CHECK(regexec(&re, "xabcy", 0, NULL, 0) == 0, "REG_NOSUB match");
	CHECK(regexec(&re, "xy", 0, NULL, 0) == REG_NOMATCH,
	    "REG_NOSUB no-match");
	regfree(&re);

	/* REG_NOSPEC: everything is literal. */
	CHECK(m("a.c*", REG_NOSPEC, "xa.c*y", 0), "REG_NOSPEC literal");
	CHECK(!m("a.c*", REG_NOSPEC, "abccc", 0),
	    "REG_NOSPEC has no specials");
}

static void
test_newline(void)
{
	/*
	 * REG_NEWLINE gives the line-oriented behavior: . and
	 * non-matching lists do not match a newline, ^/$ match after
	 * and before embedded newlines.
	 */
	CHECK(!m("a.c", REG_BASIC | REG_NEWLINE, "a\nc", 0),
	    "REG_NEWLINE: dot does not cross lines");
	CHECK(!m("a[^x]c", REG_BASIC | REG_NEWLINE, "a\nc", 0),
	    "REG_NEWLINE: negated list excludes newline");
	CHECK(m("^b", REG_BASIC | REG_NEWLINE, "a\nb", 0),
	    "REG_NEWLINE: ^ after embedded newline");
	CHECK(m("a$", REG_BASIC | REG_NEWLINE, "a\nb", 0),
	    "REG_NEWLINE: $ before embedded newline");
	CHECK(!m("^b", REG_BASIC, "a\nb", 0),
	    "without REG_NEWLINE ^ only matches at start");
}

static void
test_word_boundaries(void)
{
	/*
	 * nvi rewrites \< and \> (docs/nvi.md section 9, items 2 and
	 * 3) into this library's [[:<:]] and [[:>:]].
	 */
	CHECK(span_is("[[:<:]]foo", REG_BASIC, "a foo b", 2, 5),
	    "[[:<:]] start of word");
	CHECK(m("[[:<:]]foo", REG_BASIC, "foobar", 0),
	    "[[:<:]] at start of a longer word");
	CHECK(!m("[[:<:]]bar", REG_BASIC, "foobar", 0),
	    "[[:<:]] rejects mid-word");
	CHECK(m("foo[[:>:]]", REG_BASIC, "a foo b", 0), "[[:>:]] end of word");
	CHECK(!m("foo[[:>:]]", REG_BASIC, "foobar", 0),
	    "[[:>:]] rejects mid-word");
	CHECK(m("[[:<:]]foo[[:>:]]", REG_BASIC, "x foo y", 0),
	    "whole-word match");
	CHECK(!m("[[:<:]]foo[[:>:]]", REG_BASIC, "xfooy", 0),
	    "whole-word rejects embedded");
}

static void
test_startend_and_pend(void)
{
	regex_t re;
	regmatch_t pm;

	/*
	 * REG_STARTEND: search only [rm_so, rm_eo) of the string;
	 * reported offsets remain relative to the start of string.
	 * nvi uses this for every buffer search.
	 */
	CHECK(regcomp(&re, "abc", REG_BASIC) == 0, "STARTEND compiles");
	pm.rm_so = 2;
	pm.rm_eo = 7;
	CHECK(regexec(&re, "ababcabc", 1, &pm, REG_STARTEND) == 0,
	    "STARTEND finds match in window");
	CHECK(pm.rm_so == 2 && pm.rm_eo == 5,
	    "STARTEND offsets absolute: got %ld-%ld",
	    (long)pm.rm_so, (long)pm.rm_eo);
	pm.rm_so = 3;
	pm.rm_eo = 5;
	CHECK(regexec(&re, "ababcabc", 1, &pm, REG_STARTEND) == REG_NOMATCH,
	    "STARTEND respects window end");
	regfree(&re);

	/* REG_PEND: pattern length from re_endp, not a NUL. */
	{
		static const char pat[] = "ab*c";

		re.re_endp = pat + 4;
		CHECK(regcomp(&re, pat, REG_BASIC | REG_PEND) == 0,
		    "REG_PEND compiles");
		CHECK(regexec(&re, "xabbbcy", 0, NULL, 0) == 0,
		    "REG_PEND pattern works");
		regfree(&re);
	}
}

static void
test_errors(void)
{
	regex_t re;
	char buf[128];
	int r;

	CHECK(comp_err("[abc", REG_BASIC) == REG_EBRACK, "unclosed bracket");
	CHECK(comp_err("(ab", REG_EXTENDED) == REG_EPAREN, "unclosed paren");
	CHECK(comp_err("\\(ab", REG_BASIC) == REG_EPAREN,
	    "unclosed BRE group");
	CHECK(comp_err("ab\\", REG_BASIC) == REG_EESCAPE,
	    "trailing backslash");
	CHECK(comp_err("\\1", REG_BASIC) == REG_ESUBREG,
	    "backreference without group");
	CHECK(comp_err("a\\{1", REG_BASIC) == REG_EBRACE, "unclosed bound");
	CHECK(comp_err("a\\{4,2\\}", REG_BASIC) == REG_BADBR,
	    "reversed bound");
	CHECK(comp_err("a{4,2}", REG_EXTENDED) == REG_BADBR,
	    "reversed ERE bound");
	CHECK(comp_err("*a", REG_EXTENDED) == REG_BADRPT,
	    "ERE leading repetition");
	CHECK(comp_err("", REG_BASIC) == REG_EMPTY, "empty pattern");
	CHECK(comp_err("a|", REG_EXTENDED) == REG_EMPTY,
	    "empty alternation branch");
	CHECK(comp_err("[[:no-such-class:]]", REG_BASIC) == REG_ECTYPE,
	    "unknown character class");

	/* regerror produces a diagnostic string. */
	r = regcomp(&re, "[abc", REG_BASIC);
	CHECK(regerror(r, &re, buf, sizeof(buf)) > 0 && buf[0] != '\0',
	    "regerror fills buffer");
	CHECK(strstr(buf, "[") != NULL || strstr(buf, "bracket") != NULL,
	    "regerror message mentions brackets: got \"%s\"", buf);
}

static void
test_collating_and_many_groups(void)
{
	regex_t re;
	regmatch_t pm[12];
	size_t i;

	/*
	 * Collating elements and equivalence classes.  Single
	 * characters work.  The symbolic names (the cnames table) are
	 * broken: the name lookup in p_b_coll_elem (regcomp.c) tests
	 * MEMCMP() for nonzero instead of zero, so a known name
	 * resolves to the first same-length entry that DIFFERS from
	 * it, and an unknown name finds some entry too instead of
	 * failing with REG_ECOLLATE.
	 */
	CHECK(m("[[.a.]]", REG_BASIC, "a", 0), "[.a.] matches a");
	CHECK_XFAIL(m("[[.comma.]]", REG_BASIC, ",", 0),
	    "[.comma.] symbolic name (inverted lookup bug)");
	CHECK_XFAIL(comp_err("[[.no-such-elem.]]", REG_BASIC) ==
	    REG_ECOLLATE,
	    "unknown collating element error (inverted lookup bug)");
	CHECK(m("[[=a=]]", REG_BASIC, "a", 0), "[=a=] matches a");
	CHECK(!m("[[=a=]]", REG_BASIC, "b", 0), "[=a=] rejects b");

	/*
	 * REG_ICASE folds what a group can match, but the historic
	 * Spencer backreference comparison is exact: the backref must
	 * repeat the group's text byte for byte.
	 */
	CHECK(m("\\(ab\\)\\1", REG_BASIC | REG_ICASE, "ABAB", 0),
	    "REG_ICASE group matches uppercase");
	CHECK(!m("\\(ab\\)\\1", REG_BASIC | REG_ICASE, "abAB", 0),
	    "REG_ICASE backreference comparison stays exact");
	CHECK(!m("\\(ab\\)\\1", REG_BASIC, "abAB", 0),
	    "case-sensitive backreference rejects abAB");

	/* More than nine capturing groups. */
	CHECK(regcomp(&re, "(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)",
	    REG_EXTENDED) == 0, "ten-group ERE compiles");
	CHECK(re.re_nsub == 10, "re_nsub is 10, got %zu", re.re_nsub);
	for (i = 0; i < 12; i++)
		pm[i].rm_so = pm[i].rm_eo = -1;
	CHECK(regexec(&re, "abcdefghij", 12, pm, 0) == 0,
	    "ten-group ERE matches");
	CHECK(pm[10].rm_so == 9 && pm[10].rm_eo == 10,
	    "group 10 spans the j, got [%ld,%ld)",
	    (long)pm[10].rm_so, (long)pm[10].rm_eo);
	regfree(&re);
}

int
main(void)
{
	test_literal_and_dot();
	test_anchors();
	test_brackets();
	test_bre_groups_and_backrefs();
	test_bounds();
	test_ere();
	test_case_and_nosub();
	test_newline();
	test_word_boundaries();
	test_startend_and_pend();
	test_errors();
	test_collating_and_many_groups();

	(void)printf("regex unit tests: %d run, %d failed\n", nrun, nfail);
	return (nfail != 0);
}

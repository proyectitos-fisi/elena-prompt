package liner

import (
	"regexp"
	"unicode"

	"github.com/mattn/go-runewidth"
)

// These character classes are mostly zero width (when combined).
// A few might not be, depending on the user's font. Fixing this
// is non-trivial, given that some terminals don't support
// ANSI DSR/CPR
var zeroWidth = []*unicode.RangeTable{
	unicode.Mn,
	unicode.Me,
	unicode.Cc,
	unicode.Cf,
}

var bashColorsControlMatcher = regexp.MustCompile(
	`\x1b\[[0-9;]*[a-zA-Z]`,
)

func CleanColors(s []rune) []rune {
	return []rune(bashColorsControlMatcher.ReplaceAllString(string(s), ""))
}

// countGlyphs considers zero-width characters to be zero glyphs wide,
// and members of Chinese, Japanese, and Korean scripts to be 2 glyphs wide.
func countGlyphs(s []rune) int {
	n := 0
	for _, r := range CleanColors(s) {
		// speed up the common case
		if r < 127 {
			n++
			continue
		}

		n += runewidth.RuneWidth(r)
	}
	return n
}

func countMultiLineGlyphs(s []rune, columns int, start int) int {
	n := start
	for _, r := range CleanColors(s) {
		if r < 127 {
			n++
			continue
		}
		switch runewidth.RuneWidth(r) {
		case 0:
		case 1:
			n++
		case 2:
			n += 2
			// no room for a 2-glyphs-wide char in the ending
			// so skip a column and display it at the beginning
			if n%columns == 1 {
				n++
			}
		}
	}
	return n
}

func getPrefixGlyphs(s []rune, num int) []rune {
	p := 0
	for n := 0; n < num && p < len(s); p++ {
		if s[p] == 0x1b {
			// skip ANSI escape sequences
			for ; p < len(s) && s[p] != 'm'; p++ {
			}
			continue
		}

		// speed up the common case
		if s[p] < 127 {
			n++
			continue
		}
		if !unicode.IsOneOf(zeroWidth, s[p]) {
			n++
		}
	}
	for p < len(s) && unicode.IsOneOf(zeroWidth, s[p]) {
		p++
	}
	return s[:p]
}

func getSuffixGlyphs(s []rune, num int) []rune {
	p := len(s)
	for n := 0; n < num && p > 0; p-- {
		// speed up the common case
		if s[p-1] < 127 {
			n++
			continue
		}
		if !unicode.IsOneOf(zeroWidth, s[p-1]) {
			n++
		}
	}
	return s[p:]
}

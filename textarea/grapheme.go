package textarea

import (
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

// graphemeAt returns the rune range of the grapheme containing the rune at col.
// Keeping the public cursor and selection positions as rune offsets preserves
// the existing API while letting keyboard operations avoid stopping inside a
// user-perceived character.
func graphemeAt(runes []rune, col int) (int, int) {
	if col < 0 || col >= len(runes) {
		return col, col
	}

	text := string(runes)
	byteCol := 0
	for _, r := range runes[:col] {
		byteCol += utf8.RuneLen(r)
	}

	runeOffset := 0
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		_, end := graphemes.Positions()
		cluster := graphemes.Str()
		clusterRunes := utf8.RuneCountInString(cluster)
		if byteCol < end {
			return runeOffset, runeOffset + clusterRunes
		}
		runeOffset += clusterRunes
	}

	return col, col + 1
}

// deletesWhole reports whether Backspace should remove the grapheme as a unit.
// Emoji and variation sequences are; anything else loses one code point at a
// time so Thai and Indic marks can be fixed without erasing their consonant.
func deletesWhole(cluster []rune) bool {
	for i, r := range cluster {
		switch {
		case r == '\u200d' && i < len(cluster)-1, // ZWJ joining two emoji
			r >= 0xFE00 && r <= 0xFE0F,   // variation selectors
			r >= 0xE0100 && r <= 0xE01EF, // ideographic variation selectors
			r >= 0x1F1E6 && r <= 0x1F1FF, // regional indicators (flags)
			r >= 0x1F3FB && r <= 0x1F3FF, // skin tone modifiers
			r >= 0xE0020 && r <= 0xE007F: // tags (subdivision flags)
			return true
		}
	}
	return false
}

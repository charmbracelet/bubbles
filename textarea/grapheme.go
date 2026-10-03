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

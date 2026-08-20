package textarea

import (
	"strings"
	"testing"
)

// Benchmarks backing the numbers cited in the PR body.
//
// averagePrompt is a stand-in for a typical prompt so the "average" rows
// stay reproducible; maintainers can swap in their own estimate of an
// average prompt and the bench + PR numbers follow.

const averagePrompt = "fix the flaky approval dialog test and add a regression case for the permission scopes"

func BenchmarkWrapAverage(b *testing.B) {
	runes := []rune(averagePrompt)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = wrap(runes, 116)
	}
}

func BenchmarkWrap2k(b *testing.B) {
	runes := []rune(strings.Repeat("the quick brown fox jumps over the lazy dog ", 50)[:2000])
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = wrap(runes, 116)
	}
}

func BenchmarkLineHash20k(b *testing.B) {
	l := line{runes: []rune(strings.Repeat("the quick brown fox jumps over the lazy dog ", 500)[:20000]), width: 116}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = l.Hash()
	}
}

func BenchmarkLengthASCII(b *testing.B) {
	m := New()
	m.SetValue(strings.Repeat("the quick brown fox ", 100))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.Length()
	}
}

func BenchmarkValueMultiLine(b *testing.B) {
	m := New()
	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString("line with some words to wrap around and more text here 1234567890\n")
	}
	m.SetValue(sb.String())
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.Value()
	}
}

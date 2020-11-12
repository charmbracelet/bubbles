package list

// Suffixer is used to suffix all visible Lines.
// InitSuffixer gets called ones on the beginning of the Lines methode
// and then Suffix ones, per line to draw, to generate according suffixes.
type Suffixer interface {
	InitSuffixer(ViewPos, ScreenInfo) int
	Suffix(currentItem, currentLine int, selected bool) string
}

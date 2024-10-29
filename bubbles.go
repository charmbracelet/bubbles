package bubbles

// Styler represents an interface for styling bubbles.
type Styler[T any] interface {
	Styles(isDark bool) T
}

// StylerFunc is a function type that implements the Styler interface.
type StylerFunc[T any] func(isDark bool) T

// Styles calls the function.
// It implements the Styler interface.
func (f StylerFunc[T]) Styles(isDark bool) T {
	return f(isDark)
}

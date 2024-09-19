package picker

import "fmt"

type ListState[T any] struct {
	state     []T
	selection int
	canCycle  bool
}

func NewListState[T any](state []T, opts ...func(listState *ListState[T])) *ListState[T] {
	s := &ListState[T]{
		state: state,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *ListState[T]) GetValue() interface{} {
	return s.state[s.selection]
}

func (s *ListState[T]) GetDisplayValue() string {
	return fmt.Sprintf("%v", s.GetValue())
}

func (s *ListState[T]) Next() {
	switch {
	case s.selection < len(s.state)-1:
		s.selection++

	case s.canCycle:
		s.selection = 0
	}
}

func (s *ListState[T]) Prev() {
	switch {
	case s.selection > 0:
		s.selection--

	case s.canCycle:
		s.selection = len(s.state) - 1
	}
}

// ListState Options --------------------

func WithCycles[T any]() func(state *ListState[T]) {
	return func(s *ListState[T]) {
		s.canCycle = true
	}
}

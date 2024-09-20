package picker

type ListState[T any] struct {
	state     []T
	selection int
}

func NewListState[T any](state []T) *ListState[T] {
	return &ListState[T]{
		state: state,
	}
}

func (s *ListState[T]) GetValue() interface{} {
	return s.state[s.selection]
}

func (s *ListState[T]) Next(canCycle bool) {
	switch {
	case s.selection < len(s.state)-1:
		s.selection++

	case canCycle:
		s.selection = 0
	}
}

func (s *ListState[T]) Prev(canCycle bool) {
	switch {
	case s.selection > 0:
		s.selection--

	case canCycle:
		s.selection = len(s.state) - 1
	}
}

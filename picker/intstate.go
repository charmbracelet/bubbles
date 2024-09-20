package picker

type IntState struct {
	min       int
	max       int
	selection int
}

func NewIntState(min, max int) *IntState {
	return &IntState{
		min:       min,
		max:       max,
		selection: min,
	}
}

func (s *IntState) GetValue() interface{} {
	return s.selection
}

func (s *IntState) Next(canCycle bool) {
	switch {
	case s.selection < s.max:
		s.selection++

	case canCycle:
		s.selection = s.min
	}
}

func (s *IntState) Prev(canCycle bool) {
	switch {
	case s.selection > s.min:
		s.selection--

	case canCycle:
		s.selection = s.max
	}
}

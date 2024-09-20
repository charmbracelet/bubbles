package picker

type IntState struct {
	min       int
	max       int
	selection int
	ignoreMin bool
	ignoreMax bool
}

func NewIntState(min, max, selection int, ignoreMin, ignoreMax bool) *IntState {
	switch {
	case selection < min && !ignoreMin:
		selection = min
	case selection > max && !ignoreMax:
		selection = max
	}

	return &IntState{
		min:       min,
		max:       max,
		ignoreMin: ignoreMin,
		ignoreMax: ignoreMax,
		selection: selection,
	}
}

func (s *IntState) GetValue() interface{} {
	return s.selection
}

func (s *IntState) Next(canCycle bool) {
	switch {
	case s.ignoreMax:
		s.selection++

	case s.selection < s.max:
		s.selection++

	case canCycle:
		s.selection = s.min
	}
}

func (s *IntState) Prev(canCycle bool) {
	switch {
	case s.ignoreMin:
		s.selection--

	case s.selection > s.min:
		s.selection--

	case canCycle:
		s.selection = s.max
	}
}

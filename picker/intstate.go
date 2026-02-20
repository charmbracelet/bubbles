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

func (s *IntState) NextExists() bool {
	return s.ignoreMax ||
		s.selection < s.max
}

func (s *IntState) PrevExists() bool {
	return s.ignoreMin ||
		s.selection > s.min
}

func (s *IntState) Next(canCycle bool) {
	switch {
	case s.NextExists():
		s.selection++

	case canCycle:
		s.selection = s.min
	}
}

func (s *IntState) Prev(canCycle bool) {
	switch {
	case s.PrevExists():
		s.selection--

	case canCycle:
		s.selection = s.max
	}
}

func (s *IntState) StepForward(size int) {
	s.selection += size
	if s.selection > s.max && !s.ignoreMax {
		s.selection = s.max
	}
}

func (s *IntState) StepBackward(size int) {
	s.selection -= size
	if s.selection < s.min && !s.ignoreMin {
		s.selection = s.min
	}
}

func (s *IntState) JumpForward() {
	s.selection = s.max
}

func (s *IntState) JumpBackward() {
	s.selection = s.min
}

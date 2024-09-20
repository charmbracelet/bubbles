package picker

import "testing"

func TestNewIntState(t *testing.T) {
	tt := map[string]struct {
		min           int
		max           int
		selection     int
		ignoreMin     bool
		ignoreMax     bool
		wantSelection int
	}{
		"select min": {
			min:           0,
			max:           2,
			selection:     0,
			ignoreMin:     false,
			ignoreMax:     false,
			wantSelection: 0,
		},
		"select max": {
			min:           0,
			max:           2,
			selection:     2,
			ignoreMin:     false,
			ignoreMax:     false,
			wantSelection: 2,
		},

		"select less than min": {
			min:           0,
			max:           2,
			selection:     -10,
			ignoreMin:     false,
			ignoreMax:     false,
			wantSelection: 0,
		},
		"select less than min; ignore min": {
			min:           0,
			max:           2,
			selection:     -10,
			ignoreMin:     true,
			ignoreMax:     false,
			wantSelection: -10,
		},
		"select less than min; ignore max": {
			min:           0,
			max:           2,
			selection:     -10,
			ignoreMin:     false,
			ignoreMax:     true,
			wantSelection: 0,
		},

		"select greater than max": {
			min:           0,
			max:           2,
			selection:     10,
			ignoreMin:     false,
			ignoreMax:     false,
			wantSelection: 2,
		},
		"select greater than max; ignore max": {
			min:           0,
			max:           2,
			selection:     10,
			ignoreMin:     false,
			ignoreMax:     true,
			wantSelection: 10,
		},
		"select greater than max; ignore min": {
			min:           0,
			max:           2,
			selection:     10,
			ignoreMin:     true,
			ignoreMax:     false,
			wantSelection: 2,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got := NewIntState(tc.min, tc.max, tc.selection, tc.ignoreMin, tc.ignoreMax)

			if got.min != tc.min {
				t.Errorf("min: got %v, want %v", got.min, tc.min)
			}
			if got.max != tc.max {
				t.Errorf("max: got %v, want %v", got.max, tc.max)
			}
			if got.selection != tc.wantSelection {
				t.Errorf("selection: got %v, want %v", got.selection, tc.wantSelection)
			}
			if got.ignoreMin != tc.ignoreMin {
				t.Errorf("ignoreMin: got %v, want %v", got.ignoreMin, tc.ignoreMin)
			}
			if got.ignoreMax != tc.ignoreMax {
				t.Errorf("ignoreMax: got %v, want %v", got.ignoreMax, tc.ignoreMax)
			}
		})
	}
}

func TestIntState_GetValue(t *testing.T) {
	state := IntState{
		min:       0,
		max:       10,
		selection: 5,
		ignoreMin: false,
		ignoreMax: false,
	}
	want := 5

	got := state.GetValue()

	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestIntState_NextExists(t *testing.T) {
	tt := map[string]struct {
		state IntState
		want  bool
	}{
		"enforce max; can increment": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 9,
				ignoreMin: false,
				ignoreMax: false,
			},
			want: true,
		},
		"enforce max; cannot increment": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 10,
				ignoreMin: false,
				ignoreMax: false,
			},
			want: false,
		},

		"ignore max; can increment": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 9,
				ignoreMin: false,
				ignoreMax: true,
			},
			want: true,
		},
		"ignore max; cannot increment": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 10,
				ignoreMin: false,
				ignoreMax: true,
			},
			want: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got := tc.state.NextExists()

			if tc.want != got {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestIntState_Next(t *testing.T) {
	tt := map[string]struct {
		state         IntState
		canCycle      bool
		wantSelection int
	}{
		"ignore max; cannot increment; cannot cycle": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 10,
				ignoreMin: false,
				ignoreMax: true,
			},
			canCycle:      false,
			wantSelection: 11,
		},
		"ignore max; cannot increment; can cycle": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 10,
				ignoreMin: false,
				ignoreMax: true,
			},
			canCycle:      true,
			wantSelection: 11,
		},
		"ignore max; can increment; cannot cycle": {
			state: IntState{
				min:       0,
				max:       11,
				selection: 10,
				ignoreMin: false,
				ignoreMax: true,
			},
			canCycle:      false,
			wantSelection: 11,
		},
		"ignore max; can increment; can cycle": {
			state: IntState{
				min:       0,
				max:       11,
				selection: 10,
				ignoreMin: false,
				ignoreMax: true,
			},
			canCycle:      true,
			wantSelection: 11,
		},

		"enforce max; cannot increment; cannot cycle": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 10,
				ignoreMin: false,
				ignoreMax: false,
			},
			canCycle:      false,
			wantSelection: 10,
		},
		"enforce max; cannot increment; can cycle": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 10,
				ignoreMin: false,
				ignoreMax: false,
			},
			canCycle:      true,
			wantSelection: 0,
		},
		"enforce max; can increment; cannot cycle": {
			state: IntState{
				min:       0,
				max:       11,
				selection: 10,
				ignoreMin: false,
				ignoreMax: false,
			},
			canCycle:      false,
			wantSelection: 11,
		},
		"enforce max; can increment; can cycle": {
			state: IntState{
				min:       0,
				max:       11,
				selection: 10,
				ignoreMin: false,
				ignoreMax: false,
			},
			canCycle:      true,
			wantSelection: 11,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			tc.state.Next(tc.canCycle)

			if tc.wantSelection != tc.state.selection {
				t.Errorf("want %v, got %v", tc.wantSelection, tc.state.selection)
			}
		})
	}
}

func TestIntState_Prev(t *testing.T) {
	tt := map[string]struct {
		state         IntState
		canCycle      bool
		wantSelection int
	}{
		"ignore min; cannot decrement; cannot cycle": {
			state: IntState{
				min:       -10,
				max:       0,
				selection: -10,
				ignoreMin: true,
				ignoreMax: false,
			},
			canCycle:      false,
			wantSelection: -11,
		},
		"ignore min; cannot decrement; can cycle": {
			state: IntState{
				min:       -10,
				max:       0,
				selection: -10,
				ignoreMin: true,
				ignoreMax: false,
			},
			canCycle:      true,
			wantSelection: -11,
		},
		"ignore min; can decrement; cannot cycle": {
			state: IntState{
				min:       -11,
				max:       0,
				selection: -10,
				ignoreMin: true,
				ignoreMax: false,
			},
			canCycle:      false,
			wantSelection: -11,
		},
		"ignore min; can decrement; can cycle": {
			state: IntState{
				min:       -11,
				max:       0,
				selection: -10,
				ignoreMin: true,
				ignoreMax: false,
			},
			canCycle:      true,
			wantSelection: -11,
		},

		"enforce min; cannot decrement; cannot cycle": {
			state: IntState{
				min:       -10,
				max:       0,
				selection: -10,
				ignoreMin: false,
				ignoreMax: false,
			},
			canCycle:      false,
			wantSelection: -10,
		},
		"enforce min; cannot decrement; can cycle": {
			state: IntState{
				min:       -10,
				max:       0,
				selection: -10,
				ignoreMin: false,
				ignoreMax: false,
			},
			canCycle:      true,
			wantSelection: 0,
		},
		"enforce min; can decrement; cannot cycle": {
			state: IntState{
				min:       -11,
				max:       0,
				selection: -10,
				ignoreMin: false,
				ignoreMax: false,
			},
			canCycle:      false,
			wantSelection: -11,
		},
		"enforce min; can decrement; can cycle": {
			state: IntState{
				min:       -11,
				max:       0,
				selection: -10,
				ignoreMin: false,
				ignoreMax: false,
			},
			canCycle:      true,
			wantSelection: -11,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			tc.state.Prev(tc.canCycle)

			if tc.wantSelection != tc.state.selection {
				t.Errorf("want %v, got %v", tc.wantSelection, tc.state.selection)
			}
		})
	}
}

func TestIntState_JumpForward(t *testing.T) {
	tt := map[string]struct {
		state IntState
		want  int
	}{
		"from min": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 0,
			},
			want: 10,
		},
		"from middle": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 5,
			},
			want: 10,
		},
		"from max": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 10,
			},
			want: 10,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			tc.state.JumpForward()

			if tc.want != tc.state.selection {
				t.Errorf("want %v, got %v", tc.want, tc.state.selection)
			}
		})
	}
}

func TestIntState_JumpBackward(t *testing.T) {
	tt := map[string]struct {
		state IntState
		want  int
	}{
		"from min": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 0,
			},
			want: 0,
		},
		"from middle": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 5,
			},
			want: 0,
		},
		"from max": {
			state: IntState{
				min:       0,
				max:       10,
				selection: 10,
			},
			want: 0,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			tc.state.JumpBackward()

			if tc.want != tc.state.selection {
				t.Errorf("want %v, got %v", tc.want, tc.state.selection)
			}
		})
	}
}

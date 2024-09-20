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

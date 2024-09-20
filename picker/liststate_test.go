package picker

import "testing"

func TestNewListState(t *testing.T) {
	want := ListState[string]{
		state: []string{"One", "Two", "Three"},
	}

	got := NewListState([]string{"One", "Two", "Three"})

	for i := range got.state {
		if got.state[i] != want.state[i] {
			t.Errorf("state[%d]: want %v, got %v", i, want.state[i], got.state[i])
		}
	}

	if got.selection != 0 {
		t.Errorf("selection: want 0, got %v", got.selection)
	}
}

func TestListState_GetValue(t *testing.T) {
	state := ListState[string]{
		state:     []string{"One", "Two", "Three"},
		selection: 1,
	}
	want := "Two"

	got := state.GetValue()

	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestListState_Next(t *testing.T) {
	tt := map[string]struct {
		state         ListState[string]
		canCycle      bool
		wantSelection int
	}{
		"can increment; can cycle": {
			state: ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 1,
			},
			canCycle:      true,
			wantSelection: 2,
		},
		"can increment; cannot cycle": {
			state: ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 1,
			},
			canCycle:      false,
			wantSelection: 2,
		},
		"cannot increment; can cycle": {
			state: ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 2,
			},
			canCycle:      true,
			wantSelection: 0,
		},
		"cannot increment; cannot cycle": {
			state: ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 2,
			},
			canCycle:      false,
			wantSelection: 2,
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

func TestListState_Prev(t *testing.T) {
	tt := map[string]struct {
		state         ListState[string]
		canCycle      bool
		wantSelection int
	}{
		"can decrement; can cycle": {
			state: ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 1,
			},
			canCycle:      true,
			wantSelection: 0,
		},
		"can decrement; cannot cycle": {
			state: ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 1,
			},
			canCycle:      false,
			wantSelection: 0,
		},
		"cannot decrement; can cycle": {
			state: ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 0,
			},
			canCycle:      true,
			wantSelection: 2,
		},
		"cannot decrement; cannot cycle": {
			state: ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 0,
			},
			canCycle:      false,
			wantSelection: 0,
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

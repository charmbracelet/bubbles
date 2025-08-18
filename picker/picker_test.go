package picker

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tt := map[string]struct {
		state    State
		opts     []func(*Model)
		wantFunc func() Model
	}{
		"default": {
			state: NewListState([]string{"One", "Two", "Three"}, 0),
			opts:  nil,
			wantFunc: func() Model {
				return Model{
					State:          NewListState([]string{"One", "Two", "Three"}, 0),
					ShowIndicators: true,
					CanCycle:       false,
					CanJump:        false,
					StepSize:       10,
					DisplayFunc: func(v interface{}) string {
						return fmt.Sprintf("%v", v)
					},
					Keys:   DefaultKeyMap(),
					Styles: DefaultStyles(),
				}
			},
		},
		"WithKeys": {
			state: NewListState([]string{"One", "Two", "Three"}, 0),
			opts: []func(*Model){
				WithKeys(&KeyMap{
					Next: key.NewBinding(key.WithKeys("test", "key")),
				}),
			},
			wantFunc: func() Model {
				return Model{
					State:          NewListState([]string{"One", "Two", "Three"}, 0),
					ShowIndicators: true,
					CanCycle:       false,
					CanJump:        false,
					StepSize:       10,
					DisplayFunc: func(v interface{}) string {
						return fmt.Sprintf("%v", v)
					},
					Keys: &KeyMap{
						Next: key.NewBinding(key.WithKeys("test", "key")),
					},
					Styles: DefaultStyles(),
				}
			},
		},
		"WithoutIndicators": {
			state: NewListState([]string{"One", "Two", "Three"}, 0),
			opts: []func(*Model){
				WithoutIndicators(),
			},
			wantFunc: func() Model {
				return Model{
					State:          NewListState([]string{"One", "Two", "Three"}, 0),
					ShowIndicators: false,
					CanCycle:       false,
					CanJump:        false,
					StepSize:       10,
					DisplayFunc: func(v interface{}) string {
						return fmt.Sprintf("%v", v)
					},
					Keys:   DefaultKeyMap(),
					Styles: DefaultStyles(),
				}
			},
		},
		"WithCycles": {
			state: NewListState([]string{"One", "Two", "Three"}, 0),
			opts: []func(*Model){
				WithCycles(),
			},
			wantFunc: func() Model {
				return Model{
					State:          NewListState([]string{"One", "Two", "Three"}, 0),
					ShowIndicators: true,
					CanCycle:       true,
					CanJump:        false,
					StepSize:       10,
					DisplayFunc: func(v interface{}) string {
						return fmt.Sprintf("%v", v)
					},
					Keys:   DefaultKeyMap(),
					Styles: DefaultStyles(),
				}
			},
		},
		"WithDisplayFunc": {
			state: NewListState([]string{"One", "Two", "Three"}, 0),
			opts: []func(*Model){
				WithDisplayFunc(func(_ interface{}) string {
					return fmt.Sprint("test")
				}),
			},
			wantFunc: func() Model {
				return Model{
					State:          NewListState([]string{"One", "Two", "Three"}, 0),
					ShowIndicators: true,
					CanCycle:       false,
					CanJump:        false,
					StepSize:       10,
					DisplayFunc: func(_ interface{}) string {
						return fmt.Sprint("test")
					},
					Keys:   DefaultKeyMap(),
					Styles: DefaultStyles(),
				}
			},
		},
		"WithStyles": {
			state: NewListState([]string{"One", "Two", "Three"}, 0),
			opts: []func(*Model){
				WithStyles(Styles{
					Selection: lipgloss.NewStyle().Width(555).Height(-555),
				}),
			},
			wantFunc: func() Model {
				return Model{
					State:          NewListState([]string{"One", "Two", "Three"}, 0),
					ShowIndicators: true,
					CanCycle:       false,
					CanJump:        false,
					StepSize:       10,
					DisplayFunc: func(v interface{}) string {
						return fmt.Sprintf("%v", v)
					},
					Keys: DefaultKeyMap(),
					Styles: Styles{
						Selection: lipgloss.NewStyle().Width(555).Height(-555),
					},
				}
			},
		},
		"WithJumping": {
			state: NewListState([]string{"One", "Two", "Three"}, 0),
			opts: []func(*Model){
				WithJumping(),
			},
			wantFunc: func() Model {
				return Model{
					State:          NewListState([]string{"One", "Two", "Three"}, 0),
					ShowIndicators: true,
					CanCycle:       false,
					CanJump:        true,
					StepSize:       10,
					DisplayFunc: func(v interface{}) string {
						return fmt.Sprintf("%v", v)
					},
					Keys:   DefaultKeyMap(),
					Styles: DefaultStyles(),
				}
			},
		},
		"WithStepping": {
			state: NewListState([]string{"One", "Two", "Three"}, 0),
			opts: []func(*Model){
				WithStepSize(2),
			},
			wantFunc: func() Model {
				return Model{
					State:          NewListState([]string{"One", "Two", "Three"}, 0),
					ShowIndicators: true,
					CanCycle:       false,
					CanJump:        false,
					StepSize:       2,
					DisplayFunc: func(v interface{}) string {
						return fmt.Sprintf("%v", v)
					},
					Keys:   DefaultKeyMap(),
					Styles: DefaultStyles(),
				}
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			want := tc.wantFunc()
			got := New(tc.state, tc.opts...)

			if !reflect.DeepEqual(got.State, want.State) {
				t.Errorf("State: \ngot: \n%v \nwant: \n%v", got, want)
			}

			if got.ShowIndicators != want.ShowIndicators {
				t.Errorf("ShowIndicators: \ngot: \n%v \nwant: \n%v", got, want)
			}

			if got.CanCycle != want.CanCycle {
				t.Errorf("CanCycle: \ngot: \n%v \nwant: \n%v", got, want)
			}

			if got.CanJump != want.CanJump {
				t.Errorf("CanJump: \ngot: \n%v \nwant: \n%v", got, want)
			}

			if got.StepSize != want.StepSize {
				t.Errorf("StepSize: \ngot: \n%v \nwant: \n%v", got, want)
			}

			if got.DisplayFunc == nil {
				t.Errorf("DisplayFunc: \ngot: \n%v \nwant: \n%v", got, want)
			} else if got.GetDisplayValue() != want.GetDisplayValue() {
				t.Errorf("GetDisplayValue: \ngot: \n%v \nwant: \n%v", got, want)
			}

			if !reflect.DeepEqual(got.Keys, want.Keys) {
				t.Errorf("Keys: \ngot: \n%v \nwant: \n%v", got, want)
			}

			if !reflect.DeepEqual(got.Styles, want.Styles) {
				t.Errorf("Styles: \ngot: \n%v \nwant: \n%v", got, want)
			}
		})
	}
}

func TestModel_View(t *testing.T) {
	model := New(
		&ListState[string]{
			state:     []string{"One", "Two", "Three"},
			selection: 1,
		},
	)
	want := heredoc.Doc(`
		< Two >`,
	)

	got := model.View()

	if want != got {
		t.Errorf("View: \ngot: \n%q\nwant: \n%q", got, want)
	}
}

func TestModel_GetValue(t *testing.T) {
	tt := map[string]struct {
		state State
		want  interface{}
	}{
		"min": {
			state: &ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 0,
			},
			want: "One",
		},
		"middle": {
			state: &ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 1,
			},
			want: "Two",
		},
		"end": {
			state: &ListState[string]{
				state:     []string{"One", "Two", "Three"},
				selection: 2,
			},
			want: "Three",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			model := Model{
				State: tc.state,
				DisplayFunc: func(v interface{}) string {
					return fmt.Sprintf("%v", v)
				},
			}

			got := model.GetDisplayValue()

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("\ngot: \n%v \nwant: \n%v", got, tc.want)
			}
		})
	}
}

func TestGetIndicator(t *testing.T) {
	tt := map[string]struct {
		styles  IndicatorStyles
		enabled bool
		want    string
	}{
		"enabled": {
			styles: IndicatorStyles{
				Value:    "test",
				Enabled:  lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderTop(true),
				Disabled: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true),
			},
			enabled: true,
			want: heredoc.Doc(`
				────
				test`,
			),
		},
		"disabled": {
			styles: IndicatorStyles{
				Value:    "test",
				Enabled:  lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderTop(true),
				Disabled: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true),
			},
			enabled: false,
			want: heredoc.Doc(`
				test
				────`,
			),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got := getIndicator(tc.styles, tc.enabled)

			if got != tc.want {
				t.Errorf("\ngot: \n%q \nwant: \n%q", got, tc.want)
			}
		})
	}
}

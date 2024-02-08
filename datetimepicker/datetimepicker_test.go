package datetimepicker

import (
	"testing"
	"time"
	"strings"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	picker := New()
	view := picker.View()

	if picker.Pos != Date {
		t.Errorf("Expected default position to be Date, got %v", picker.Pos)
	}
	if !strings.Contains(view, ">") {
		t.Log(view)
		t.Error("datetimepicker did not render the prompt")
	}
}


func TestUpdate(t *testing.T) {
	picker := New()

	testCases := []struct {
		name           string
		keyMsgs        []tea.Msg
		expectedPos    PositionType
		expectedDate   time.Time
		pickerType     PickerType
		intializeDate  time.Time
	}{
		// Test key bindings.
		{
			name: "Left key press",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}},
			},
			expectedPos: Date,
			pickerType: DateTime,
		},
		{
			name: "Right key press",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
			},
			expectedPos: Month,
			pickerType: DateTime,
		},
		{
			name: "Forward key press for DateOnly picker",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
			},
			expectedPos: Year,
			pickerType: DateOnly,
		},
		{
			name: "Backward key press for TimeOnly picker",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyLeft, Alt: false, Runes: []rune{}},
			},
			expectedPos: Hour,
			pickerType: TimeOnly,
		},
		// Test Increment/Decrement.
		{
			name: "Increment Test (Up key)",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyUp, Alt: false, Runes: []rune{}},
			},
			expectedDate: picker.Date.AddDate(0, 0, 1),
			pickerType: DateTime,
		},
		{
			name: "Decrement Test : Decrement the month by 1",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}},
			},
			expectedDate: time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC),
			pickerType: DateTime,
			expectedPos: Month,
			intializeDate: time.Date(2024, time.February, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "Avoid negative years (by decrementing month)",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}},
			},
			expectedDate: time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC),
			pickerType: DateTime,
			intializeDate: time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC),
			expectedPos: Month,
		},
		{
			name: "Avoid negative years (by decrementing Date)",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}},
			},
			expectedDate: time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC),
			pickerType: DateTime,
			intializeDate: time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC),
			expectedPos: Date,
		},
		{
			name: "Avoid negative years (by decrementing Year)",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}},
			},
			expectedDate: time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC),
			pickerType: DateTime,
			intializeDate: time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC),
			expectedPos: Year,
		},
		{
			name: "Avoid negative years (by decrementing Hour)",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}},
			},
			expectedDate: time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC),
			pickerType: DateTime,
			intializeDate: time.Date(0, time.January, 1, 0, 59, 0, 0, time.UTC),
			expectedPos: Hour,
		},
		{
			name: "Avoid negative years (by decrementing Minute)",
			keyMsgs: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyRight, Alt: false, Runes: []rune{}},
				tea.KeyMsg{Type: tea.KeyDown, Alt: false, Runes: []rune{}},
			},
			expectedDate: time.Date(0, time.January, 1, 23, 59, 0, 0, time.UTC),
			pickerType: DateTime,
			intializeDate: time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedPos: Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := picker
			p.SetPickerType(tc.pickerType)
			if !tc.intializeDate.IsZero() {
				p.SetValue(tc.intializeDate)
			}
			for _, msg := range tc.keyMsgs {
				pModel, _ := p.Update(msg)
				p = pModel.(Model)
			}

			if p.Pos != tc.expectedPos {
				t.Errorf("Expected position %v after %s, got %v", tc.expectedPos, tc.name, p.Pos)
			}

			if !tc.expectedDate.IsZero() && p.Date != tc.expectedDate {
				t.Errorf("Expected date %v after %s, got %v", tc.expectedDate, tc.name, p.Date)
			}
		})
	}
}

func TestSetValue(t *testing.T) {
	picker := New()

	// Set date value and check if it's correctly set.
	newDate := time.Date(2024, time.February, 1, 12, 0, 0, 0, time.UTC)
	picker.SetValue(newDate)
	if !picker.Date.Equal(newDate) {
		t.Error("Expected date value to be set to", newDate)
	}
}

func TestSetTimeFormat(t *testing.T) {
	picker := New()

	// Set time format to 24-hour and check if it's correctly set.
	picker.SetTimeFormat(Hour24)
	if picker.TimeFormat != Hour24 {
		t.Error("Expected time format to be set to 24-hour")
	}

	// Should auto handle if timeFormat is out of defined enum.
	picker.SetTimeFormat(-1)
	if picker.TimeFormat != Hour12 {
		t.Error("Expected time format to be set to 12-hour")
	}

	picker.SetTimeFormat(2)
	if picker.TimeFormat != Hour24 {
		t.Error("Expected time format to be set to 12-hour")
	}

}

func TestSetPickerType(t *testing.T) {
	picker := New()

	// Test 1: Set picker type to TimeOnly and check if it's correctly set.
	picker.SetPickerType(TimeOnly)
	if picker.PickerType != TimeOnly {
		t.Error("Expected picker type to be set to TimeOnly")
	}
	if picker.Pos != Hour {
		t.Error("Expected Pos to be set to Hour")
	}

	// Test 2: 
	picker.SetPickerType(DateOnly)
	if picker.PickerType != DateOnly {
		t.Error("Expected picker type to be set to DateOnly")
	}
	if picker.Pos != Date {
		t.Error("Expected Pos to be set to Date")
	}
 
	// Test 3: 
	picker.SetPickerType(DateTime)
	if picker.PickerType != DateTime {
		t.Error("Expected picker type to be set to DateTime")
	}
	if picker.Pos != Date {
		t.Error("Expected Pos to be set to Date")
	}
}

func TestValue(t *testing.T) {
	picker := New()

	// Test value formatting for different picker types and time formats.
	// Test 1: (DateTime pickerType).
	inputTime := time.Date(2024, time.February, 1, 12, 0, 0, 0, time.UTC)
	picker.SetValue(inputTime)

	if err := validateValue(picker, DateTime, inputTime); err != nil {
		t.Error(err)
	}

	// Test 2: (DateOnly pickerType).
	picker.SetPickerType(DateOnly)

	if err := validateValue(picker, DateOnly, inputTime); err != nil {
		t.Error(err)
	}

	// Test 3: (TimeOnly pickerType).
	picker.SetPickerType(TimeOnly)

	if err := validateValue(picker, TimeOnly, inputTime); err != nil {
		t.Error(err)
	}
}

func validateValue(m Model, pickerType PickerType, inputTime time.Time) error {
	expectedValue := ""
	if pickerType == DateTime {
		expectedValue = inputTime.Format("02 January 2006 03:04 PM")
	} else if pickerType == DateOnly {
		expectedValue = inputTime.Format("02 January 2006")
	} else {
		// TimeOnly.
		if m.TimeFormat == Hour12 {
			expectedValue =  inputTime.Format("03:04 PM")
		} else { // Hour24.
			expectedValue = inputTime.Format("15:04")
		}
	}

	if val := m.Value(); val != expectedValue {
		return fmt.Errorf("Expected value %s, got %s", expectedValue, val)
	}
	return nil
}

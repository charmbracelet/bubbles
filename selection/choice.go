package selection

import (
	"fmt"
	"reflect"
)

// Choice represents a single choice. This type used as an input
// for the selection prompt, for filtering and as a result value.
// The index is populated by the prompt itself and is exported
// to be accessed when filtering.
type Choice struct {
	Index  int
	String string
	Value  interface{}
}

// NewChoice creates a new choice for a given input and chooses
// a suitable string representation. The index is left at 0 to
// be populated by the selection prompt later on.
func NewChoice(item interface{}) *Choice {
	choice := &Choice{Index: 0, Value: item}

	switch i := item.(type) {
	case Choice:
		choice.String = i.String
		choice.Value = i.Value
	case *Choice:
		choice.String = i.String
		choice.Value = i.Value
	case string:
		choice.String = i
	case fmt.Stringer:
		choice.String = i.String()
	default:
		choice.String = fmt.Sprintf("%#v", i)
	}

	return choice
}

// StringChoices converts a string slice to a slice of choices.
func StringChoices(choiceStrings []string) []*Choice {
	choices := make([]*Choice, 0, len(choiceStrings))

	for _, c := range choiceStrings {
		choices = append(choices, NewChoice(c))
	}

	return choices
}

// StringerChoices converts a slice of Stringers to a slice of choices.
func StringerChoices(choiceStrings []fmt.Stringer) []*Choice {
	choices := make([]*Choice, 0, len(choiceStrings))

	for _, c := range choiceStrings {
		choices = append(choices, NewChoice(c))
	}

	return choices
}

// SliceChoices converts a slice of anything to a slice of choices.
// SliceChoices panics if the input is not a slice.
func SliceChoices(sliceChoices interface{}) []*Choice {
	switch reflect.TypeOf(sliceChoices).Kind() {
	case reflect.Slice:
		slice := reflect.ValueOf(sliceChoices)

		choices := make([]*Choice, 0, slice.Len())

		for i := 0; i < slice.Len(); i++ {
			value := slice.Index(i).Interface()
			choice := NewChoice(value)

			choices = append(choices, choice)
		}

		return choices
	default:
		panic("SliceChoices argument is not a slice")
	}
}

package textinput

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func Test_CurrentSuggestion(t *testing.T) {
	textinput := New()
	textinput.ShowSuggestions = true

	suggestion := textinput.CurrentSuggestion()
	expected := ""
	if suggestion != expected {
		t.Fatalf("Error: expected no current suggestion but was %s", suggestion)
	}

	textinput.SetSuggestions([]string{"test1", "test2", "test3"})
	suggestion = textinput.CurrentSuggestion()
	expected = ""
	if suggestion != expected {
		t.Fatalf("Error: expected no current suggestion but was %s", suggestion)
	}

	textinput.SetValue("test")
	textinput.updateSuggestions()
	textinput.nextSuggestion()
	suggestion = textinput.CurrentSuggestion()
	expected = "test2"
	if suggestion != expected {
		t.Fatalf("Error: expected first suggestion but was %s", suggestion)
	}
}

func Test_SlicingOutsideCap(t *testing.T) {
	textinput := New()
	textinput.Placeholder = "作業ディレクトリを指定してください"
	textinput.Width = 32
	textinput.View()
}

func ExampleValidateFunc() {
	creditCardNumber := New()
	creditCardNumber.Placeholder = "4505 **** **** 1234"
	creditCardNumber.Focus()
	creditCardNumber.CharLimit = 20
	creditCardNumber.Width = 30
	creditCardNumber.Prompt = ""
	// The anonymous function is a valid function for ValidateFunc.
	creditCardNumber.Validate = func(s string) error {
		// Credit Card Number should a string less than 20 digits
		// It should include 16 integers and 3 spaces
		if len(s) > 16+3 {
			return fmt.Errorf("CCN is too long")
		}

		if len(s) == 0 || len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
			return fmt.Errorf("CCN is invalid")
		}

		// The last digit should be a number unless it is a multiple of 4 in which
		// case it should be a space
		if len(s)%5 == 0 && s[len(s)-1] != ' ' {
			return fmt.Errorf("CCN must separate groups with spaces")
		}

		// The remaining digits should be integers
		c := strings.ReplaceAll(s, " ", "")
		_, err := strconv.ParseInt(c, 10, 64)

		return err
	}
}

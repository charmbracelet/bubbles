package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/tea"
	"github.com/charmbracelet/teaparty/input"
)

func main() {
	if err := tea.NewProgram(
		initialize,
		update,
		view,
		subscriptions,
	).Start(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}

type Model struct {
	index         int
	nameInput     input.Model
	nickNameInput input.Model
	emailInput    input.Model
}

func initialize() (tea.Model, tea.Cmd) {
	name := input.DefaultModel()
	name.Placeholder = "Name"
	name.Focus()

	nickName := input.DefaultModel()
	nickName.Placeholder = "Nickname"

	email := input.DefaultModel()
	email.Placeholder = "Email"

	return Model{0, name, nickName, email}, nil
}

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, ok := model.(Model)
	if !ok {
		panic("could not perform assertion on model")
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit

		// Cycle between inputs
		case "tab":
			fallthrough
		case "shift+tab":
			fallthrough
		case "enter":
			fallthrough
		case "up":
			fallthrough
		case "down":
			inputs := []input.Model{
				m.nameInput,
				m.nickNameInput,
				m.emailInput,
			}

			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.index--
			} else {
				m.index++
			}

			if m.index > len(inputs)-1 {
				m.index = 0
			} else if m.index < 0 {
				m.index = len(inputs) - 1
			}

			for i := 0; i < len(inputs); i++ {
				if i == m.index {
					inputs[i].Focus()
					continue
				}
				inputs[i].Blur()
			}

			m.nameInput = inputs[0]
			m.nickNameInput = inputs[1]
			m.emailInput = inputs[2]

			return m, nil

		default:
			m.nameInput, _ = input.Update(msg, m.nameInput)
			m.nickNameInput, _ = input.Update(msg, m.nickNameInput)
			m.emailInput, _ = input.Update(msg, m.emailInput)
			return m, nil
		}

	default:
		m.nameInput, _ = input.Update(msg, m.nameInput)
		m.nickNameInput, _ = input.Update(msg, m.nickNameInput)
		m.emailInput, _ = input.Update(msg, m.emailInput)
		return m, nil
	}
}

func subscriptions(model tea.Model) tea.Subs {
	return tea.Subs{
		"blink": func(model tea.Model) tea.Msg {
			m, _ := model.(Model)
			return input.Blink(m.nameInput)
		},
	}
}

func view(model tea.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "[error] could not perform assertion on model"
	}

	s := "\n"

	inputs := []string{
		input.View(m.nameInput),
		input.View(m.nickNameInput),
		input.View(m.emailInput),
	}

	for i := 0; i < len(inputs); i++ {
		s += inputs[i]
		if i < len(inputs)-1 {
			s += "\n"
		}
	}

	s += "\n"

	return s
}

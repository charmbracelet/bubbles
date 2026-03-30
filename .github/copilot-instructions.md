# Copilot Instructions for charmbracelet/bubbles

## Project Overview

Bubbles is a Go TUI component library for Bubble Tea applications. Module path: `charm.land/bubbles/v2`. Each component is a separate Go package in its own directory.

## Component Architecture

Every component follows the Elm Architecture:

1. **Model** -- a struct holding component state
2. **Update(msg tea.Msg) (Model, tea.Cmd)** -- processes messages, returns updated model
3. **View() string** -- renders the component to a string using lipgloss styling

Important: `Update` returns the concrete `Model` type, not `tea.Model`. This enables value-type embedding in parent models without type assertions.

### Constructor Pattern

Newer components use functional options:
```go
type Option func(*Model)

func New(opts ...Option) Model {
    // apply defaults, then options
}
```

### KeyMap Pattern

Every interactive component defines:
```go
type KeyMap struct {
    ActionName key.Binding
    // ...
}

func DefaultKeyMap() KeyMap { ... }
```

Key bindings use `key.NewBinding(key.WithKeys(...), key.WithHelp(...))`. If the KeyMap implements `help.KeyMap` (ShortHelp/FullHelp), it integrates with the help bubble.

### Styles Pattern

```go
type Styles struct {
    Header   lipgloss.Style
    Selected lipgloss.Style
    // ...
}

func DefaultStyles() Styles { ... }
```

## Import Paths

Always use v2 charm.land paths:
```go
import (
    "charm.land/bubbles/v2/viewport"
    "charm.land/bubbles/v2/key"
    tea "charm.land/bubbletea/v2"
    "charm.land/lipgloss/v2"
)
```

## Coding Standards

- Format with `gofumpt` and `goimports`
- All comments end with a period (godot linter)
- All exported types and functions need doc comments
- Errors from external packages must be wrapped (wrapcheck linter)
- Switch on enum types must be exhaustive (exhaustive linter)
- Use `//nolint:mnd` only for magic numbers in spinner frame definitions

## Testing

- Use table-driven tests with `map[string]struct{}` and named subtests
- Use golden file testing for View() output: `github.com/charmbracelet/x/exp/golden`
- Golden files in `<component>/testdata/<TestName>/<subtest>.golden`
- Update golden files: `go test ./... -update`

## Component Composition

Components embed each other as value types:
- table embeds viewport + help
- list embeds paginator + spinner + textinput + help
- textarea embeds cursor + viewport
- textinput embeds cursor

## Key Interfaces

- `list.Item` -- requires `FilterValue() string`
- `list.ItemDelegate` -- controls item rendering (Render, Height, Spacing, Update)
- `help.KeyMap` -- enables auto-generated help (ShortHelp, FullHelp)

## Animated Component ID Pattern

Components with timers (spinner, progress, cursor, stopwatch, timer, filepicker) use atomic ID counters to ensure tick messages reach only the correct instance:
```go
var lastID int64
func nextID() int { return int(atomic.AddInt64(&lastID, 1)) }
```

## Init() Is Optional

Only implement `Init() tea.Cmd` if the component needs an initial command (e.g., filepicker reads directory, timer starts ticking). Most components omit it.

## Internal Packages

- `internal/memoization/` -- render cache for textarea
- `internal/runeutil/` -- rune width and sanitization utilities

Not importable externally.

# Bubbles - Claude Code Instructions

## What This Project Is

Bubbles is a Go component library for building terminal user interfaces with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Each component lives in its own package and follows the Elm Architecture (Model-Update-View). This is v2, using the `charm.land/bubbles/v2` module path.

## Build and Test

```bash
# Run all tests
go test ./...

# Run tests for a specific component
go test ./textarea/...

# Update golden test fixtures after intentional rendering changes
go test ./... -update

# Lint (requires golangci-lint v2.9+)
golangci-lint run

# Or use Taskfile shortcuts
task test
task lint
```

## Architecture: The Component Pattern

Every component in this repo follows the same pattern. Learn it once, apply everywhere.

### The Model struct

Each component defines a `Model` struct in its package. Public fields are the user-facing API. Private fields hold internal state.

```go
type Model struct {
    KeyMap KeyMap           // User-configurable keybindings
    Style  lipgloss.Style   // User-configurable styling
    // ... public config fields ...

    cursor int              // internal state
    lines  []string         // internal state
}
```

### Constructor: New()

Newer components use functional options:
```go
func New(opts ...Option) Model { ... }
// where Option is: type Option func(*Model)
```

Older components use simple constructors or struct literals.

### Update returns concrete type, NOT tea.Model

This is critical. Components return `(Model, tea.Cmd)`, not `(tea.Model, tea.Cmd)`. This lets parent models embed components as value types without type assertions:

```go
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { ... }
```

### Not all components implement Init()

Only components that need an initial command implement `Init() tea.Cmd`. These are: filepicker, viewport, progress, timer, stopwatch. The rest do not have Init.

### KeyMap and help integration

Components define a `KeyMap` struct with `key.Binding` fields and a `DefaultKeyMap()` constructor. If the KeyMap implements `help.KeyMap` (ShortHelp/FullHelp methods), it integrates with the help bubble automatically.

### Styles

Components define a `Styles` struct with `lipgloss.Style` fields and a `DefaultStyles()` constructor. Some `DefaultStyles()` accept `isDark bool` for theme awareness.

## Component Composition

Components compose by embedding. For example:
- `table` embeds `viewport` and `help`
- `list` embeds `paginator`, `spinner`, `textinput`, and `help`
- `textarea` embeds `cursor` and `viewport`
- `textinput` embeds `cursor`

## Key Interfaces

- **`list.Item`**: anything in a list must implement `FilterValue() string`
- **`list.ItemDelegate`**: controls rendering and input handling for list items (Render, Height, Spacing, Update)
- **`help.KeyMap`**: `ShortHelp() []key.Binding` + `FullHelp() [][]key.Binding` -- enables auto-generated help text

## ID Pattern for Animated Components

Components with timers/animations (spinner, progress, cursor, stopwatch, timer, filepicker) use an atomic counter (`lastID`/`nextID()`) to generate unique IDs. This ensures tick messages are only processed by the component instance that sent them, preventing cross-talk when multiple instances run concurrently.

## Testing Conventions

- **Table-driven tests** with `map[string]struct{}` and named subtests
- **Golden file tests** using `github.com/charmbracelet/x/exp/golden` for View() output
- Golden files live in `<component>/testdata/<TestName>/<subtest>.golden`
- Run `go test ./... -update` to regenerate golden files after intentional changes
- Tests construct models, pipe messages through `Update()`, then check `View()` output or model state

## Linting Rules

Uses golangci-lint v2 with strict settings. Key rules to follow:
- `godot`: all comments must end with a period.
- `gofumpt` + `goimports`: strict formatting beyond standard gofmt.
- `wrapcheck`: errors from external packages must be wrapped.
- `exhaustive`: switch statements on enum types must be exhaustive.
- `gosec`: no security issues in code.
- `//nolint:mnd` is acceptable for magic numbers in spinner frame definitions.

## Import Paths (v2)

Always use `charm.land/` paths, not `github.com/charmbracelet/`:
```go
import (
    "charm.land/bubbles/v2/viewport"
    tea "charm.land/bubbletea/v2"
    "charm.land/lipgloss/v2"
)
```

## Internal Packages

- `internal/memoization/` -- generic cache used by textarea for render optimization
- `internal/runeutil/` -- rune sanitization and width calculation

These are not importable by external consumers.

## File Structure Convention

Each component is a directory with:
- `<component>.go` -- main implementation
- `<component>_test.go` -- tests
- `testdata/` -- golden files (if applicable)
- Some components split into multiple files (list has `list.go`, `defaultitem.go`, `keys.go`, `style.go`)

## What NOT To Do

- Do not make components implement `tea.Model` interface directly. They return concrete `Model` types.
- Do not add `Init()` unless the component genuinely needs to fire an initial command.
- Do not use `github.com/charmbracelet/bubbles` import paths -- this is v2 with `charm.land/` paths.
- Do not skip the `DefaultKeyMap()` / `DefaultStyles()` pattern when adding configurable bindings or styles.
- Do not use global mutable state except for the atomic ID counter pattern.

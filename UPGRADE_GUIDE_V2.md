# Upgrading to Bubbles v2

This guide covers every breaking change when migrating from Bubbles v1 (`github.com/charmbracelet/bubbles`) to Bubbles v2 (`charm.land/bubbles/v2`). It is written for both humans and LLM-assisted migration tools.

> **Companion upgrades required.** Bubbles v2 requires Bubble Tea v2 and Lip Gloss v2. Upgrade all three together:
>
> ```sh
> go get charm.land/bubbletea/v2
> go get charm.land/bubbles/v2
> go get charm.land/lipgloss/v2
> ```

---

## Table of Contents

1. [Import Paths](#1-import-paths)
2. [Global Patterns](#2-global-patterns)
3. [Per-Component Migration](#3-per-component-migration)
   - [Cursor](#cursor)
   - [Filepicker](#filepicker)
   - [Help](#help)
   - [List](#list)
   - [Paginator](#paginator)
   - [Progress](#progress)
   - [Spinner](#spinner)
   - [Stopwatch](#stopwatch)
   - [Table](#table)
   - [Textarea](#textarea)
   - [Textinput](#textinput)
   - [Timer](#timer)
   - [Viewport](#viewport)
4. [Light and Dark Styles](#4-light-and-dark-styles)
5. [Removed Symbols Reference](#5-removed-symbols-reference)

---

## 1. Import Paths

Replace all `github.com/charmbracelet/bubbles` imports with `charm.land/bubbles/v2`:

```go
// Before
import (
    "github.com/charmbracelet/bubbles/cursor"
    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/paginator"
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/runeutil"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/stopwatch"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/timer"
    "github.com/charmbracelet/bubbles/viewport"
)

// After
import (
    "charm.land/bubbles/v2/cursor"
    "charm.land/bubbles/v2/help"
    "charm.land/bubbles/v2/key"
    "charm.land/bubbles/v2/list"
    "charm.land/bubbles/v2/paginator"
    "charm.land/bubbles/v2/progress"
    "charm.land/bubbles/v2/spinner"
    "charm.land/bubbles/v2/stopwatch"
    "charm.land/bubbles/v2/table"
    "charm.land/bubbles/v2/textarea"
    "charm.land/bubbles/v2/textinput"
    "charm.land/bubbles/v2/timer"
    "charm.land/bubbles/v2/viewport"
)
```

> **Note:** The `runeutil` and `memoization` packages are now internal and no longer importable.

**Search-and-replace pattern:**

```
github.com/charmbracelet/bubbles/  →  charm.land/bubbles/v2/
github.com/charmbracelet/bubbles   →  charm.land/bubbles/v2
```

---

## 2. Global Patterns

These patterns repeat across multiple components. Address them first for the broadest impact.

### 2a. `tea.KeyMsg` → `tea.KeyPressMsg`

Bubble Tea v2 renames `tea.KeyMsg` to `tea.KeyPressMsg`. All Bubbles that handle key events have been updated. Update your own `Update` functions:

```go
// Before
case tea.KeyMsg:

// After
case tea.KeyPressMsg:
```

### 2b. Exported Width/Height Fields → Getter/Setter Methods

Many components replaced exported `Width` and `Height` fields with methods. The general pattern:

```go
// Before
m.Width = 40
m.Height = 20
fmt.Println(m.Width, m.Height)

// After
m.SetWidth(40)
m.SetHeight(20)
fmt.Println(m.Width(), m.Height())
```

**Affected components:** `filepicker`, `help`, `progress`, `table`, `textinput`, `viewport`.

### 2c. `DefaultKeyMap` Variables → Functions

Global mutable `DefaultKeyMap` variables are now functions returning fresh values:

```go
// Before
km := textinput.DefaultKeyMap
km.Paste.SetEnabled(false)

// After
km := textinput.DefaultKeyMap()
km.Paste.SetEnabled(false)
```

**Affected components:** `paginator`, `textarea`, `textinput`.

### 2d. `AdaptiveColor` → `LightDark` with `isDark bool`

Lip Gloss v2 removes `AdaptiveColor`. Style functions that previously auto-adapted now require an explicit `isDark bool` parameter. See [Section 4](#4-light-and-dark-styles) for the full pattern.

### 2e. Removed `NewModel` Aliases

All `NewModel` variables (deprecated aliases for `New`) have been removed. Use `New` directly.

**Affected components:** `help`, `list`, `paginator`, `spinner`, `textinput`.

---

## 3. Per-Component Migration

### Cursor

| v1 | v2 |
|----|-----|
| `model.Blink` | `model.IsBlinked` |
| `model.BlinkCmd()` | `model.Blink()` |

### Filepicker

| v1 | v2 |
|----|-----|
| `DefaultStylesWithRenderer(r)` | `DefaultStyles()` |
| `model.Height = 10` | `model.SetHeight(10)` |
| `_ = model.Height` | `_ = model.Height()` |

### Help

| v1 | v2 |
|----|-----|
| `model.Width = 80` | `model.SetWidth(80)` |
| `_ = model.Width` | `_ = model.Width()` |
| `NewModel()` | `New()` |

New functions:
- `DefaultStyles(isDark bool) Styles`
- `DefaultDarkStyles() Styles`
- `DefaultLightStyles() Styles`

Apply styles explicitly:

```go
// Before
h := help.New()
// Colors auto-adapted to terminal background

// After
h := help.New()
h.Styles = help.DefaultStyles(isDark)
```

### List

| v1 | v2 |
|----|-----|
| `DefaultStyles()` | `DefaultStyles(isDark)` |
| `NewDefaultItemStyles()` | `NewDefaultItemStyles(isDark)` |
| `styles.FilterPrompt` | `styles.Filter.Focused.Prompt` / `styles.Filter.Blurred.Prompt` |
| `styles.FilterCursor` | `styles.Filter.Cursor` |
| `NewModel(...)` | `New(...)` |

The `Styles.FilterPrompt` and `Styles.FilterCursor` fields have been consolidated into `Styles.Filter`, which is a `textinput.Styles` struct.

### Paginator

| v1 | v2 |
|----|-----|
| `DefaultKeyMap` (var) | `DefaultKeyMap()` (func) |
| `model.UsePgUpPgDownKeys` | Removed — customize `KeyMap` directly |
| `model.UseLeftRightKeys` | Removed — customize `KeyMap` directly |
| `model.UseUpDownKeys` | Removed — customize `KeyMap` directly |
| `model.UseHLKeys` | Removed — customize `KeyMap` directly |
| `model.UseJKKeys` | Removed — customize `KeyMap` directly |
| `NewModel(...)` | `New(...)` |

### Progress

This component has the most extensive changes.

#### Width

```go
// Before
p.Width = 40
fmt.Println(p.Width)

// After
p.SetWidth(40)
fmt.Println(p.Width())
```

#### Colors

Color types changed from `string` to `image/color.Color`:

```go
// Before
p.FullColor = "#FF0000"
p.EmptyColor = "#333333"

// After
p.FullColor = lipgloss.Color("#FF0000")
p.EmptyColor = lipgloss.Color("#333333")
```

#### Gradient/Blend Options

```go
// Before
progress.New(progress.WithGradient("#5A56E0", "#EE6FF8"))
progress.New(progress.WithDefaultGradient())
progress.New(progress.WithScaledGradient("#5A56E0", "#EE6FF8"))
progress.New(progress.WithDefaultScaledGradient())
progress.New(progress.WithSolidFill("#7571F9"))

// After
progress.New(progress.WithColors(lipgloss.Color("#5A56E0"), lipgloss.Color("#EE6FF8")))
progress.New(progress.WithDefaultBlend())
progress.New(progress.WithColors(lipgloss.Color("#5A56E0"), lipgloss.Color("#EE6FF8")), progress.WithScaled(true))
progress.New(progress.WithDefaultBlend(), progress.WithScaled(true))
progress.New(progress.WithColors(lipgloss.Color("#7571F9")))
```

| v1 | v2 |
|----|-----|
| `WithGradient(a, b string)` | `WithColors(colors ...color.Color)` |
| `WithDefaultGradient()` | `WithDefaultBlend()` |
| `WithScaledGradient(a, b string)` | `WithColors(...) + WithScaled(true)` |
| `WithDefaultScaledGradient()` | `WithDefaultBlend() + WithScaled(true)` |
| `WithSolidFill(string)` | `WithColors(color)` (single color) |
| `WithColorProfile(termenv.Profile)` | Removed (automatic) |
| `Update() (tea.Model, tea.Cmd)` | `Update() (Model, tea.Cmd)` |

New options:
- `WithColorFunc(func(total, current float64) color.Color)` — dynamic per-cell coloring
- `WithScaled(bool)` — scale blend to filled portion

### Spinner

| v1 | v2 |
|----|-----|
| `NewModel()` | `New()` |
| `spinner.Tick()` (package func) | `model.Tick()` (method) |

### Stopwatch

```go
// Before
sw := stopwatch.NewWithInterval(500 * time.Millisecond)

// After
sw := stopwatch.New(stopwatch.WithInterval(500 * time.Millisecond))
```

| v1 | v2 |
|----|-----|
| `NewWithInterval(d)` | `New(WithInterval(d))` |

### Table

| v1 | v2 |
|----|-----|
| `model.viewport.Width` | `model.Width()` / `model.SetWidth(w)` |
| `model.viewport.Height` | `model.Height()` / `model.SetHeight(h)` |

The table already had `SetWidth`/`SetHeight`/`Width()`/`Height()` in v1, but internally these now use viewport getter/setters.

### Textarea

#### KeyMap

```go
// Before
km := textarea.DefaultKeyMap
// After
km := textarea.DefaultKeyMap()
```

New key bindings added: `PageUp`, `PageDown`.

#### Styles

The styling system has been restructured:

```go
// Before
ta := textarea.New()
ta.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
ta.BlurredStyle.Base = lipgloss.NewStyle().Border(lipgloss.HiddenBorder())

// After
ta := textarea.New()
// Styles are now nested under a Styles struct
// Access via Styles.Focused and Styles.Blurred (type StyleState)
```

| v1 | v2 |
|----|-----|
| `textarea.Style` (type) | `textarea.StyleState` (type) |
| `model.FocusedStyle` | `model.Styles.Focused` |
| `model.BlurredStyle` | `model.Styles.Blurred` |
| `DefaultStyles() (focused, blurred Style)` | `DefaultStyles(isDark bool) Styles` |

#### Cursor

```go
// Before
ta.Cursor                           // cursor.Model (virtual cursor)
ta.SetCursor(col)                   // set cursor column

// After
ta.Cursor()                         // func() *tea.Cursor (real cursor)
ta.SetCursorColumn(col)             // renamed for clarity
ta.VirtualCursor                    // bool: true = virtual, false = real
ta.Styles.Cursor                    // CursorStyle for cursor appearance
```

New additions:
- `Column()` — returns current cursor column (0-indexed)
- `ScrollYOffset()` — returns vertical scroll offset
- `ScrollPosition()` — returns scroll position
- `MoveToBeginning()` / `MoveToEnd()` — navigate to start/end

### Textinput

#### KeyMap

```go
// Before
km := textinput.DefaultKeyMap
// After
km := textinput.DefaultKeyMap()
```

#### Width

```go
// Before
ti.Width = 40
// After
ti.SetWidth(40)
```

#### Styles

Individual style fields have moved into a `Styles` struct:

```go
// Before
ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
ti.TextStyle = lipgloss.NewStyle()
ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
ti.CompletionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// After
s := textinput.DefaultStyles(isDark)
s.Focused.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
s.Focused.Text = lipgloss.NewStyle()
s.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
s.Focused.Suggestion = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
ti.SetStyles(s)
```

| v1 Field | v2 Location |
|----------|-------------|
| `Model.PromptStyle` | `StyleState.Prompt` |
| `Model.TextStyle` | `StyleState.Text` |
| `Model.PlaceholderStyle` | `StyleState.Placeholder` |
| `Model.CompletionStyle` | `StyleState.Suggestion` |
| `Model.CursorStyle` | `Styles.Cursor` |
| `Model.Cursor` (cursor.Model) | `Model.Cursor()` (func → *tea.Cursor) |

New:
- `Model.Styles()` / `Model.SetStyles(Styles)` — get/set styles
- `Model.VirtualCursor()` / `Model.SetVirtualCursor(bool)` — toggle cursor mode

### Timer

```go
// Before
t := timer.NewWithInterval(30*time.Second, 100*time.Millisecond)
t := timer.New(30 * time.Second)

// After
t := timer.New(30*time.Second, timer.WithInterval(100*time.Millisecond))
t := timer.New(30 * time.Second)
```

| v1 | v2 |
|----|-----|
| `NewWithInterval(timeout, interval)` | `New(timeout, WithInterval(interval))` |

### Viewport

This component has the most new features alongside its breaking changes.

#### Constructor

```go
// Before
vp := viewport.New(80, 24)

// After
vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(24))
// or
vp := viewport.New()
vp.SetWidth(80)
vp.SetHeight(24)
```

#### Width, Height, YOffset

```go
// Before
vp.Width = 80
vp.Height = 24
vp.YOffset = 5
fmt.Println(vp.Width, vp.Height, vp.YOffset)

// After
vp.SetWidth(80)
vp.SetHeight(24)
vp.SetYOffset(5)
fmt.Println(vp.Width(), vp.Height(), vp.YOffset())
```

#### Removed

- `HighPerformanceRendering` — removed entirely (deprecated in Bubble Tea v2)

#### New Features (non-breaking)

These are additions you can adopt incrementally:

- **Soft wrapping:** `vp.SoftWrap = true`
- **Left gutter** for line numbers:
  ```go
  vp.LeftGutterFunc = func(info viewport.GutterContext) string {
      if info.Soft { return "     │ " }
      if info.Index >= info.TotalLines { return "   ~ │ " }
      return fmt.Sprintf("%4d │ ", info.Index+1)
  }
  ```
- **Highlighting:**
  ```go
  vp.SetHighlights(regexp.MustCompile("pattern").FindAllStringIndex(vp.GetContent(), -1))
  vp.HighlightNext()
  vp.HighlightPrevious()
  vp.ClearHighlights()
  ```
- **`SetContentLines([]string)`** — set lines directly with virtual soft-wrap support
- **`GetContent() string`** — retrieve content
- **`FillHeight bool`** — fill viewport with empty lines
- **`StyleLineFunc func(int) lipgloss.Style`** — per-line styling
- **Horizontal scrolling** with left/right arrow keys
- **Horizontal mouse wheel scrolling**

---

## 4. Light and Dark Styles

Lip Gloss v2 removes `AdaptiveColor`, so Bubbles no longer auto-detect terminal background. You must explicitly choose light or dark styles.

### Recommended: Query via Bubble Tea

```go
func (m model) Init() tea.Cmd {
    return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.BackgroundColorMsg:
        isDark := msg.IsDark()
        m.help.Styles = help.DefaultStyles(isDark)
        m.list.Styles = list.DefaultStyles(isDark)
        // ... apply to other components
    }
    return m, nil
}
```

This is required when using [Wish](https://github.com/charmbracelet/wish) to detect the client's background.

### Quick: Use `compat` Package

```go
import "charm.land/lipgloss/v2/compat"

var isDark = compat.HasDarkBackground()

func main() {
    h := help.New()
    h.Styles = help.DefaultStyles(isDark)
}
```

> **Warning:** The `compat` approach uses blocking I/O outside Bubble Tea's event loop and will not detect remote client backgrounds over SSH.

### Manual

```go
h.Styles = help.DefaultDarkStyles()   // force dark
h.Styles = help.DefaultLightStyles()  // force light
```

---

## 5. Removed Symbols Reference

Quick-reference table of all removed symbols and their replacements:

| Package | Removed | Replacement |
|---------|---------|-------------|
| `cursor` | `Model.Blink` | `Model.IsBlinked` |
| `cursor` | `Model.BlinkCmd()` | `Model.Blink()` |
| `filepicker` | `DefaultStylesWithRenderer(r)` | `DefaultStyles()` |
| `filepicker` | `Model.Height` (field) | `Model.SetHeight()` / `Model.Height()` |
| `help` | `NewModel` | `New()` |
| `help` | `Model.Width` (field) | `Model.SetWidth()` / `Model.Width()` |
| `list` | `NewModel` | `New()` |
| `list` | `DefaultStyles()` | `DefaultStyles(isDark)` |
| `list` | `NewDefaultItemStyles()` | `NewDefaultItemStyles(isDark)` |
| `list` | `Styles.FilterPrompt` | `Styles.Filter` (`textinput.Styles`) |
| `list` | `Styles.FilterCursor` | `Styles.Filter.Cursor` |
| `paginator` | `DefaultKeyMap` (var) | `DefaultKeyMap()` (func) |
| `paginator` | `NewModel` | `New()` |
| `paginator` | `UsePgUpPgDownKeys` etc. | Customize `KeyMap` directly |
| `progress` | `WithGradient(a, b)` | `WithColors(colors...)` |
| `progress` | `WithDefaultGradient()` | `WithDefaultBlend()` |
| `progress` | `WithScaledGradient(a, b)` | `WithColors(...) + WithScaled(true)` |
| `progress` | `WithDefaultScaledGradient()` | `WithDefaultBlend() + WithScaled(true)` |
| `progress` | `WithSolidFill(string)` | `WithColors(color)` |
| `progress` | `WithColorProfile(p)` | Removed (automatic) |
| `progress` | `Model.Width` (field) | `Model.SetWidth()` / `Model.Width()` |
| `spinner` | `NewModel` | `New()` |
| `spinner` | `Tick()` (package func) | `Model.Tick()` |
| `stopwatch` | `NewWithInterval(d)` | `New(WithInterval(d))` |
| `table` | `Model.Width` (field) | `Model.SetWidth()` / `Model.Width()` |
| `table` | `Model.Height` (field) | `Model.SetHeight()` / `Model.Height()` |
| `textarea` | `DefaultKeyMap` (var) | `DefaultKeyMap()` (func) |
| `textarea` | `Style` (type) | `StyleState` (type) |
| `textarea` | `Model.FocusedStyle` | `Model.Styles.Focused` |
| `textarea` | `Model.BlurredStyle` | `Model.Styles.Blurred` |
| `textarea` | `Model.SetCursor(col)` | `Model.SetCursorColumn(col)` |
| `textarea` | `DefaultStyles()` | `DefaultStyles(isDark)` |
| `textinput` | `DefaultKeyMap` (var) | `DefaultKeyMap()` (func) |
| `textinput` | `NewModel` | `New()` |
| `textinput` | `Model.Width` (field) | `Model.SetWidth()` / `Model.Width()` |
| `textinput` | `Model.PromptStyle` | `StyleState.Prompt` |
| `textinput` | `Model.TextStyle` | `StyleState.Text` |
| `textinput` | `Model.PlaceholderStyle` | `StyleState.Placeholder` |
| `textinput` | `Model.CompletionStyle` | `StyleState.Suggestion` |
| `textinput` | `Model.CursorStyle` | `Styles.Cursor` |
| `textinput` | `Model.Cursor` (cursor.Model) | `Model.Cursor()` (func → *tea.Cursor) |
| `timer` | `NewWithInterval(t, i)` | `New(t, WithInterval(i))` |
| `viewport` | `New(w, h int)` | `New(...Option)` |
| `viewport` | `Model.Width` (field) | `Model.SetWidth()` / `Model.Width()` |
| `viewport` | `Model.Height` (field) | `Model.SetHeight()` / `Model.Height()` |
| `viewport` | `Model.YOffset` (field) | `Model.SetYOffset()` / `Model.YOffset()` |
| `viewport` | `HighPerformanceRendering` | Removed |
| `runeutil` | Entire package | Moved to `internal/runeutil` (not importable) |

---

Part of [Charm](https://charm.land).

<a href="https://charm.land/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

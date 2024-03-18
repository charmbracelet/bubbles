Bubbles
=======

<p>
  <img src="https://stuff.charm.sh/bubbles/bubbles-github.png" width="233" alt="The Bubbles Logo">
</p>

[![Latest Release](https://img.shields.io/github/release/charmbracelet/bubbles.svg)](https://github.com/charmbracelet/bubbles/releases)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/charmbracelet/bubbles)
[![Build Status](https://github.com/charmbracelet/bubbles/workflows/build/badge.svg)](https://github.com/charmbracelet/bubbles/actions)
[![Go ReportCard](https://goreportcard.com/badge/charmbracelet/bubbles)](https://goreportcard.com/report/charmbracelet/bubbles)
<a href="https://www.phorm.ai/query?projectId=a0e324b6-b706-4546-b951-6671ea60c13f"><img src="https://img.shields.io/badge/Phorm-Ask_AI-%23F2777A.svg?&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNSIgaGVpZ2h0PSI0IiBmaWxsPSJub25lIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogIDxwYXRoIGQ9Ik00LjQzIDEuODgyYTEuNDQgMS40NCAwIDAgMS0uMDk4LjQyNmMtLjA1LjEyMy0uMTE1LjIzLS4xOTIuMzIyLS4wNzUuMDktLjE2LjE2NS0uMjU1LjIyNmExLjM1MyAxLjM1MyAwIDAgMS0uNTk1LjIxMmMtLjA5OS4wMTItLjE5Mi4wMTQtLjI3OS4wMDZsLTEuNTkzLS4xNHYtLjQwNmgxLjY1OGMuMDkuMDAxLjE3LS4xNjkuMjQ2LS4xOTFhLjYwMy42MDMgMCAwIDAgLjItLjEwNi41MjkuNTI5IDAgMCAwIC4xMzgtLjE3LjY1NC42NTQgMCAwIDAgLjA2NS0uMjRsLjAyOC0uMzJhLjkzLjkzIDAgMCAwLS4wMzYtLjI0OS41NjcuNTY3IDAgMCAwLS4xMDMtLjIuNTAyLjUwMiAwIDAgMC0uMTY4LS4xMzguNjA4LjYwOCAwIDAgMC0uMjQtLjA2N0wyLjQzNy43MjkgMS42MjUuNjcxYS4zMjIuMzIyIDAgMCAwLS4yMzIuMDU4LjM3NS4zNzUgMCAwIDAtLjExNi4yMzJsLS4xMTYgMS40NS0uMDU4LjY5Ny0uMDU4Ljc1NEwuNzA1IDRsLS4zNTctLjA3OUwuNjAyLjkwNkMuNjE3LjcyNi42NjMuNTc0LjczOS40NTRhLjk1OC45NTggMCAwIDEgLjI3NC0uMjg1Ljk3MS45NzEgMCAwIDEgLjMzNy0uMTRjLjExOS0uMDI2LjIyNy0uMDM0LjMyNS0uMDI2TDMuMjMyLjE2Yy4xNTkuMDE0LjMzNi4wMy40NTkuMDgyYTEuMTczIDEuMTczIDAgMCAxIC41NDUuNDQ3Yy4wNi4wOTQuMTA5LjE5Mi4xNDQuMjkzYTEuMzkyIDEuMzkyIDAgMCAxIC4wNzguNThsLS4wMjkuMzJaIiBmaWxsPSIjRjI3NzdBIi8+CiAgPHBhdGggZD0iTTQuMDgyIDIuMDA3YTEuNDU1IDEuNDU1IDAgMCAxLS4wOTguNDI3Yy0uMDUuMTI0LS4xMTQuMjMyLS4xOTIuMzI0YTEuMTMgMS4xMyAwIDAgMS0uMjU0LjIyNyAxLjM1MyAxLjM1MyAwIDAgMS0uNTk1LjIxNGMtLjEuMDEyLS4xOTMuMDE0LS4yOC4wMDZsLTEuNTYtLjEwOC4wMzQtLjQwNi4wMy0uMzQ4IDEuNTU5LjE1NGMuMDkgMCAuMTczLS4wMS4yNDgtLjAzM2EuNjAzLjYwMyAwIDAgMCAuMi0uMTA2LjUzMi41MzIgMCAwIDAgLjEzOS0uMTcyLjY2LjY2IDAgMCAwIC4wNjQtLjI0MWwuMDI5LS4zMjFhLjk0Ljk0IDAgMCAwLS4wMzYtLjI1LjU3LjU3IDAgMCAwLS4xMDMtLjIwMi41MDIuNTAyIDAgMCAwLS4xNjgtLjEzOC42MDUuNjA1IDAgMCAwLS4yNC0uMDY3TDEuMjczLjgyN2MtLjA5NC0uMDA4LS4xNjguMDEtLjIyMS4wNTUtLjA1My4wNDUtLjA4NC4xMTQtLjA5Mi4yMDZMLjcwNSA0IDAgMy45MzhsLjI1NS0yLjkxMUExLjAxIDEuMDEgMCAwIDEgLjM5My41NzIuOTYyLjk2MiAwIDAgMSAuNjY2LjI4NmEuOTcuOTcgMCAwIDEgLjMzOC0uMTRDMS4xMjIuMTIgMS4yMy4xMSAxLjMyOC4xMTlsMS41OTMuMTRjLjE2LjAxNC4zLjA0Ny40MjMuMWExLjE3IDEuMTcgMCAwIDEgLjU0NS40NDhjLjA2MS4wOTUuMTA5LjE5My4xNDQuMjk1YTEuNDA2IDEuNDA2IDAgMCAxIC4wNzcuNTgzbC0uMDI4LjMyMloiIGZpbGw9IndoaXRlIi8+CiAgPHBhdGggZD0iTTQuMDgyIDIuMDA3YTEuNDU1IDEuNDU1IDAgMCAxLS4wOTguNDI3Yy0uMDUuMTI0LS4xMTQuMjMyLS4xOTIuMzI0YTEuMTMgMS4xMyAwIDAgMS0uMjU0LjIyNyAxLjM1MyAxLjM1MyAwIDAgMS0uNTk1LjIxNGMtLjEuMDEyLS4xOTMuMDE0LS4yOC4wMDZsLTEuNTYtLjEwOC4wMzQtLjQwNi4wMy0uMzQ4IDEuNTU5LjE1NGMuMDkgMCAuMTczLS4wMS4yNDgtLjAzM2EuNjAzLjYwMyAwIDAgMCAuMi0uMTA2LjUzMi41MzIgMCAwIDAgLjEzOS0uMTcyLjY2LjY2IDAgMCAwIC4wNjQtLjI0MWwuMDI5LS4zMjFhLjk0Ljk0IDAgMCAwLS4wMzYtLjI1LjU3LjU3IDAgMCAwLS4xMDMtLjIwMi41MDIuNTAyIDAgMCAwLS4xNjgtLjEzOC42MDUuNjA1IDAgMCAwLS4yNC0uMDY3TDEuMjczLjgyN2MtLjA5NC0uMDA4LS4xNjguMDEtLjIyMS4wNTUtLjA1My4wNDUtLjA4NC4xMTQtLjA5Mi4yMDZMLjcwNSA0IDAgMy45MzhsLjI1NS0yLjkxMUExLjAxIDEuMDEgMCAwIDEgLjM5My41NzIuOTYyLjk2MiAwIDAgMSAuNjY2LjI4NmEuOTcuOTcgMCAwIDEgLjMzOC0uMTRDMS4xMjIuMTIgMS4yMy4xMSAxLjMyOC4xMTlsMS41OTMuMTRjLjE2LjAxNC4zLjA0Ny40MjMuMWExLjE3IDEuMTcgMCAwIDEgLjU0NS40NDhjLjA2MS4wOTUuMTA5LjE5My4xNDQuMjk1YTEuNDA2IDEuNDA2IDAgMCAxIC4wNzcuNTgzbC0uMDI4LjMyMloiIGZpbGw9IndoaXRlIi8+Cjwvc3ZnPgo=" alt="phorm.ai"></a>

Some components for [Bubble Tea](https://github.com/charmbracelet/bubbletea)
applications. These components are used in production in [Glow][glow],
[Charm][charm] and [many other applications][otherstuff].

[glow]: https://github.com/charmbracelet/glow
[charm]: https://github.com/charmbracelet/charm
[otherstuff]: https://github.com/charmbracelet/bubbletea/#bubble-tea-in-the-wild


## Spinner

<img src="https://stuff.charm.sh/bubbles-examples/spinner.gif" width="400" alt="Spinner Example">

A spinner, useful for indicating that some kind an operation is happening.
There are a couple default ones, but you can also pass your own ”frames.”

* [Example code, basic spinner](https://github.com/charmbracelet/bubbletea/tree/master/examples/spinner/main.go)
* [Example code, various spinners](https://github.com/charmbracelet/bubbletea/tree/master/examples/spinners/main.go)


## Text Input

<img src="https://stuff.charm.sh/bubbles-examples/textinput.gif" width="400" alt="Text Input Example">

A text input field, akin to an `<input type="text">` in HTML. Supports unicode,
pasting, in-place scrolling when the value exceeds the width of the element and
the common, and many customization options.

* [Example code, one field](https://github.com/charmbracelet/bubbletea/tree/master/examples/textinput/main.go)
* [Example code, many fields](https://github.com/charmbracelet/bubbletea/tree/master/examples/textinputs/main.go)

## Text Area

<img src="https://stuff.charm.sh/bubbles-examples/textarea.gif" width="400" alt="Text Area Example">

A text area field, akin to an `<textarea />` in HTML. Allows for input that
spans multiple lines. Supports unicode, pasting, vertical scrolling when the
value exceeds the width and height of the element,  and many customization
options.

* [Example code, chat input](https://github.com/charmbracelet/bubbletea/tree/master/examples/chat/main.go)
* [Example code, story time input](https://github.com/charmbracelet/bubbletea/tree/master/examples/textarea/main.go)

## Table

<img src="https://stuff.charm.sh/bubbles-examples/table.gif" width="400" alt="Table Example">

A component for displaying and navigating tabular data (columns and rows).
Supports vertical scrolling and many customization options.

* [Example code, countries and populations](https://github.com/charmbracelet/bubbletea/tree/master/examples/table/main.go)

## Progress

<img src="https://stuff.charm.sh/bubbles-examples/progress.gif" width="800" alt="Progressbar Example">

A simple, customizable progress meter, with optional animation via
[Harmonica][harmonica]. Supports solid and gradient fills. The empty and filled
runes can be set to whatever you'd like. The percentage readout is customizable
and can also be omitted entirely.

* [Animated example](https://github.com/charmbracelet/bubbletea/blob/master/examples/progress-animated/main.go)
* [Static example](https://github.com/charmbracelet/bubbletea/blob/master/examples/progress-static/main.go)

[harmonica]: https://github.com/charmbracelet/harmonica


## Paginator

<img src="https://stuff.charm.sh/bubbles-examples/pagination.gif" width="200" alt="Paginator Example">

A component for handling pagination logic and optionally drawing pagination UI.
Supports "dot-style" pagination (similar to what you might see on iOS) and
numeric page numbering, but you could also just use this component for the
logic and visualize pagination however you like.

* [Example code](https://github.com/charmbracelet/bubbletea/blob/master/examples/paginator/main.go)


## Viewport

<img src="https://stuff.charm.sh/bubbles-examples/viewport.gif" width="600" alt="Viewport Example">

A viewport for vertically scrolling content. Optionally includes standard
pager keybindings and mouse wheel support. A high performance mode is available
for applications which make use of the alternate screen buffer.

* [Example code](https://github.com/charmbracelet/bubbletea/tree/master/examples/pager/main.go)

This component is well complemented with [Reflow][reflow] for ANSI-aware
indenting and text wrapping.

[reflow]: https://github.com/muesli/reflow


## List

<img src="https://stuff.charm.sh/bubbles-examples/list.gif" width="600" alt="List Example">

A customizable, batteries-included component for browsing a set of items.
Features pagination, fuzzy filtering, auto-generated help, an activity spinner,
and status messages, all of which can be enabled and disabled as needed.
Extrapolated from [Glow][glow].

* [Example code, default list](https://github.com/charmbracelet/bubbletea/tree/master/examples/list-default/main.go)
* [Example code, simple list](https://github.com/charmbracelet/bubbletea/tree/master/examples/list-simple/main.go)
* [Example code, all features](https://github.com/charmbracelet/bubbletea/tree/master/examples/list-fancy/main.go)

## File Picker

<img src="https://vhs.charm.sh/vhs-yET2HNiJNEbyqaVfYuLnY.gif" width="600" alt="File picker example">

A customizable component for picking a file from the file system. Navigate
through directories and select files, optionally limit to certain file
extensions.

* [Example code](https://github.com/charmbracelet/bubbletea/tree/master/examples/file-picker/main.go)

## Timer

A simple, flexible component for counting down. The update frequency and output
can be customized as you like.

<img src="https://stuff.charm.sh/bubbles-examples/timer.gif" width="400" alt="Timer example">

* [Example code](https://github.com/charmbracelet/bubbletea/blob/master/examples/timer/main.go)


## Stopwatch

<img src="https://stuff.charm.sh/bubbles-examples/stopwatch.gif" width="400" alt="Stopwatch example">

A simple, flexible component for counting up. The update frequency and output
can be customized as you see fit.

* [Example code](https://github.com/charmbracelet/bubbletea/blob/master/examples/stopwatch/main.go)


## Help

<img src="https://stuff.charm.sh/bubbles-examples/help.gif" width="500" alt="Help Example">

A customizable horizontal mini help view that automatically generates itself
from your keybindings. It features single and multi-line modes, which the user
can optionally toggle between. It will truncate gracefully if the terminal is
too wide for the content.

* [Example code](https://github.com/charmbracelet/bubbletea/blob/master/examples/help/main.go)


## Key

A non-visual component for managing keybindings. It’s useful for allowing users
to remap keybindings as well as generating help views corresponding to your
keybindings.

```go
type KeyMap struct {
    Up key.Binding
    Down key.Binding
}

var DefaultKeyMap = KeyMap{
    Up: key.NewBinding(
        key.WithKeys("k", "up"),        // actual keybindings
        key.WithHelp("↑/k", "move up"), // corresponding help text
    ),
    Down: key.NewBinding(
        key.WithKeys("j", "down"),
        key.WithHelp("↓/j", "move down"),
    ),
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, DefaultKeyMap.Up):
            // The user pressed up
        case key.Matches(msg, DefaultKeyMap.Down):
            // The user pressed down
        }
    }
    return m, nil
}
```


## Additional Bubbles

<!-- in alphabetical order by author -->

* [76creates/stickers](https://github.com/76creates/stickers): Responsive
  flexbox and table components.
* [calyptia/go-bubble-table](https://github.com/calyptia/go-bubble-table): An
  interactive, customizable, scrollable table component.
* [erikgeiser/promptkit](https://github.com/erikgeiser/promptkit): A collection
  of common prompts for cases like selection, text input, and confirmation.
  Each prompt comes with sensible defaults, remappable keybindings, any many
  customization options.
* [evertras/bubble-table](https://github.com/Evertras/bubble-table): Interactive,
  customizable, paginated tables.
* [knipferrc/teacup](https://github.com/knipferrc/teacup): Various handy
  bubbles and utilities for building Bubble Tea applications.
* [mritd/bubbles](https://github.com/mritd/bubbles): Some general-purpose
  bubbles. Inputs with validation, menu selection, a modified progressbar, and
  so on.
* [treilik/bubbleboxer](https://github.com/treilik/bubbleboxer): Layout
  multiple bubbles side-by-side in a layout-tree.
* [treilik/bubblelister](https://github.com/treilik/bubblelister): An alternate
  list that is scrollable without pagination and has the ability to contain
  other bubbles as list items.

If you’ve built a Bubble you think should be listed here, please create a Pull
Request. Please note that for a project to be included, it must meet the
following requirements:
- The README has a demo GIF.
- The README clearly states the purpose of the bubble along with an example on how to use it. 
- The bubble must *always* be in a working state on its `main` branch.

Thank you!

## Feedback

We’d love to hear your thoughts on this project. Feel free to drop us a note!

* [Twitter](https://twitter.com/charmcli)
* [The Fediverse](https://mastodon.social/@charmcli)
* [Discord](https://charm.sh/chat)

## License

[MIT](https://github.com/charmbracelet/bubbletea/raw/master/LICENSE)

***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source

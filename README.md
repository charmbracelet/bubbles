# Bubbles

<img src="https://github.com/user-attachments/assets/b89fa46e-d451-4b33-a009-c68d4765520f" width="350" />

[![Latest Release](https://img.shields.io/github/release/charmbracelet/bubbles.svg)](https://github.com/charmbracelet/bubbles/releases)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/charmbracelet/bubbles)
[![Build Status](https://github.com/charmbracelet/bubbles/workflows/build/badge.svg)](https://github.com/charmbracelet/bubbles/actions)
[![Go ReportCard](https://goreportcard.com/badge/charmbracelet/bubbles)](https://goreportcard.com/report/charmbracelet/bubbles)

Primitives for [Bubble Tea](https://github.com/charmbracelet/bubbletea)
applications. These components are used in production in [Crush][crush], and [many other applications][otherstuff].

> [!TIP]
>
> Upgrading from v1? Check out the [upgrade guide](./UPGRADE_GUIDE_V2.md), or
> point your LLM at it and let it go to town.

[crush]: https://github.com/charmbracelet/crush
[otherstuff]: https://github.com/charmbracelet/bubbletea/#bubble-tea-in-the-wild

## Spinner

<img src="https://stuff.charm.sh/bubbles-examples/spinner.gif" width="400" alt="Spinner Example">

A spinner, useful for indicating that some kind an operation is happening.
There are a couple default ones, but you can also pass your own ”frames.”

- [Example code, basic spinner](https://github.com/charmbracelet/bubbletea/blob/main/examples/spinner/main.go)
- [Example code, various spinners](https://github.com/charmbracelet/bubbletea/blob/main/examples/spinners/main.go)

## Text Input

<img src="https://stuff.charm.sh/bubbles-examples/textinput.gif" width="400" alt="Text Input Example">

A text input field, akin to an `<input type="text">` in HTML. Supports unicode,
pasting, in-place scrolling when the value exceeds the width of the element and
the common, and many customization options.

- [Example code, one field](https://github.com/charmbracelet/bubbletea/blob/main/examples/textinput/main.go)
- [Example code, many fields](https://github.com/charmbracelet/bubbletea/blob/main/examples/textinputs/main.go)

## Text Area

<img src="https://stuff.charm.sh/bubbles-examples/textarea.gif" width="400" alt="Text Area Example">

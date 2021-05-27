Bubbles
=======

<p>
  <img src="https://stuff.charm.sh/bubbles/bubbles-github.png" width="233" alt="The Bubbles Logo">
</p>

[![Latest Release](https://img.shields.io/github/release/charmbracelet/bubbles.svg)](https://github.com/charmbracelet/bubbles/releases)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/charmbracelet/bubbles)
[![Build Status](https://github.com/charmbracelet/bubbles/workflows/build/badge.svg)](https://github.com/charmbracelet/bubbles/actions)
[![Go ReportCard](https://goreportcard.com/badge/charmbracelet/bubbles)](https://goreportcard.com/report/charmbracelet/bubbles)

Some components for [Bubble Tea](https://github.com/charmbracelet/bubbletea) applications.

These components are used in production in [Glow][glow] and [Charm][charm].

[glow]: https://github.com/charmbracelet/glow
[charm]: https://github.com/charmbracelet/charm


## Spinner

<img src="https://stuff.charm.sh/bubbles-examples/spinner.gif" width="400" alt="Spinner Example">

A spinner, useful for indicating that some kind an operation is happening.
There are a couple default ones, but you can also pass your own ”frames.”

* [Example code](https://github.com/charmbracelet/tea/tree/master/examples/spinner/main.go)


## Text Input

<img src="https://stuff.charm.sh/bubbles-examples/textinput.gif" width="400" alt="Text Input Example">

A text input field, akin to an `<input type="text">` in HTML. Supports unicode,
pasting, in-place scrolling when the value exceeds the width of the element and
the common, and many customization options.

* [Example code, one field](https://github.com/charmbracelet/tea/tree/master/examples/textinput/main.go)
* [Example code, many fields](https://github.com/charmbracelet/tea/tree/master/examples/textinputs/main.go)


## Progress

<img src="https://stuff.charm.sh/bubbles-examples/progress.gif" width="800" alt="Progressbar Example">

A simple, customizable progress meter. Supports solid and gradient fills. The
empty and filled runes can be set to whatever you'd like. The percentage readout
is customizable and can also be omitted entirely.

* [Example code](https://github.com/charmbracelet/bubbletea/blob/master/examples/progress/main.go)


## Paginator

<img src="https://stuff.charm.sh/bubbles-examples/pagination.gif" width="200" alt="Paginator Example">

A component for handling pagination logic and optionally drawing pagination UI.
Supports "dot-style" pagination (similar to what you might see on iOS) and
numeric page numbering, but you could also just use this component for the
logic and visualize pagination however you like.

This component is used in [Glow][glow] to browse documents and [Charm][charm] to
browse SSH keys.


## Viewport

<img src="https://stuff.charm.sh/bubbles-examples/viewport.gif" width="600" alt="Viewport Example">

A viewport for vertically scrolling content. Optionally includes standard
pager keybindings and mouse wheel support. A high performance mode is available
for applications which make use of the alternate screen buffer.

* [Example code](https://github.com/charmbracelet/tea/tree/master/examples/pager/main.go)

This component is well complimented with [Reflow][reflow] for ANSI-aware
indenting and text wrapping.

[reflow]: https://github.com/muesli/reflow


## License

[MIT](https://github.com/charmbracelet/teaparty/raw/master/LICENSE)


***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge-unrounded.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source

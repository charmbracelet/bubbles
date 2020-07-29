# Bubbles

Some components for [Bubble Tea](https://github.com/charmbraclet/bubbletea):

* Spinner
* Text Input
* Paginator
* Viewport

[glow]: https://github.com/charmbraclet/glow
[charm]: https://github.com/charmbraclet/charm

These components are used in production in [Glow][glow] and [Charm][charm].


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
* [Example code, many fields](https://github.com/charmbracelet/tea/tree/master/examples/textinput/main.go)


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
for applications which make use of the alterate screen buffer.

* [Example code](https://github.com/charmbracelet/tea/tree/master/examples/pager/main.go)

This compoent is well complimented with [Reflow][reflow] for ANSI-aware
indenting and text wrapping.

[reflow]: https://github.com/muesli/reflow


## License

[MIT](https://github.com/charmbracelet/teaparty/raw/master/LICENSE)


***

A [Charm](https://charm.sh) project.

<img alt="the Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400">

Charm热爱开源!

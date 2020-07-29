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

A spinner, useful for indicating that some kind an operation is happening.
There are a couple default ones, but you can also pass your own ”frames.”

* [Example](https://github.com/charmbracelet/tea/tree/master/examples/spinner)


## Text Input

A text input field, akin to an `<input type="text">` in HTML. Supports unicode,
pasting, in-place scrolling when the value exceeds the width of the element and
the common, and many customization options.

An example of the text field

* [Example, one field](https://github.com/charmbracelet/tea/tree/master/examples/textinput)
* [Example, many fields](https://github.com/charmbracelet/tea/tree/master/examples/textinput)


## Paginator

A component for handling pagination logic and optionally drawing pagination UI.

This component is used in [Glow][glow] to browse documents and [Charm][charm] to
browse SSH keys.


## Viewport

A viewport for vertically scrolling content. Optionally includes standard
pager keybindings and mouse wheel support. A high performance mode is available
for applications which make use of the alterate screen buffer.

* [Example](https://github.com/charmbracelet/tea/tree/master/examples/pager)

This compoent is well complimented with [Reflow][reflow] for ANSI-aware
indenting and text wrapping.

[reflow]: https://github.com/muesli/reflow


## License

[MIT](https://github.com/charmbracelet/teaparty/raw/master/LICENSE)


***

A [Charm](https://charm.sh) project.

<img alt="the Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400">

Charm热爱开源!

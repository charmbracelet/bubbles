Output dropdown.gif

Set Shell "bash"
Set FontSize 14
Set Width 640
Set Height 420
Set Theme "Catppuccin Mocha"
Set Padding 20

Type "go run ./examples/dropdown"
Enter
Sleep 1s

# Open the first dropdown
Type " "
Sleep 500ms

# Navigate down twice and select "Bubbles"
Type "jj"
Sleep 300ms
Type " "
Sleep 500ms

# Reload options at runtime
Type "r"
Sleep 500ms

# Open again and pick "Rust"
Type " "
Sleep 300ms
Type "j"
Sleep 300ms
Enter
Sleep 500ms

# Tab to the disabled dropdown (no effect on input)
Type " "
Sleep 300ms
Tab
Sleep 500ms

# Tab to empty dropdown
Tab
Sleep 500ms

# Quit
Type "q"

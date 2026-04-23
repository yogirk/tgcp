Output demo.gif

Require go

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 800
Set Padding 20
Set Theme "Catppuccin Macchiato"

# Setup: ensure tgcp is built
Hide
Type "go build -o tgcp ./cmd/tgcp"
Enter
Sleep 2s
Show

Type "./tgcp --project cloudside-academy"
Enter
Sleep 3s

# Go to Overview
Type "over"
Sleep 1s
Enter
Sleep 4s

# Show command palette and switch to gce
Type ":"
Sleep 1s
Type "gce:"
Sleep 1s
Enter
Sleep 4s

# Navigate down
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Open instance detail
Enter
Sleep 3s

# Go back
Type "q"
Sleep 1s

# Show command palette and switch to bq
Type ":"
Sleep 1s
Type "bigquery"
Sleep 1s
Enter
Sleep 8s

# Open dataset
Type "j"
Sleep 500ms
Enter
Sleep 8s

# Open table
Type "j"
Sleep 500ms
Enter
Sleep 6s

# Quit out
Type "q"
Sleep 1s
Type "q"
Sleep 1s
Type "q"
Sleep 1s

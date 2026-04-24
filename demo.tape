Output demo.gif

Require go

Set Shell "bash"
Set FontSize 14
Set Width 1400
Set Height 800
Set Padding 20
Set Theme "GitHub Dark"
Set TypingSpeed 60ms

# Build silently, launch in demo mode (no GCP auth required)
Hide
Type "go build -o tgcp ./cmd/tgcp"
Enter
Sleep 2s
Show

Type "./tgcp --demo"
Enter
Sleep 6s

# Filter mode — cyan FILTER badge, fuzzy match on the home menu
Type "/"
Sleep 500ms
Type "bq"
Sleep 2s
Escape
Sleep 800ms

# Jump to Compute Engine via command palette
Type ":"
Sleep 600ms
Type "gce"
Sleep 1s
Enter
Sleep 4s

# Drill into an instance — uses vim j keys (more reliable than arrow escapes)
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Enter
Sleep 4s

# Back to the list
Type "q"
Sleep 1s

# Jump to Bigtable — empty-state hero shot
Type ":"
Sleep 600ms
Type "bigtable"
Sleep 1s
Enter
Sleep 4s

# Back to home
Type "q"
Sleep 1s

# Open help overlay (banner shows inside the dialog)
Type "?"
Sleep 4s
Escape
Sleep 500ms

# Closing frame — let the ASCII banner hold the stage on the home screen
Sleep 4s

# Exit cleanly (single q from home closes the app)
Type "q"
Sleep 1s

Output demo.gif

Require go

Set Shell "bash"
Set FontSize 14
Set Width 1400
Set Height 800
Set Padding 20
Set Theme "GitHub Dark"
Set TypingSpeed 60ms

# Build silently, then launch in demo mode — no GCP auth required
Hide
Type "go build -o tgcp ./cmd/tgcp"
Enter
Sleep 2s
Show

Type "./tgcp --demo"
Enter
Sleep 6s

# Filter mode — cyan FILTER badge, fuzzy match
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

# Drill into an instance — shows detail card
Down
Sleep 400ms
Down
Sleep 400ms
Enter
Sleep 4s

# Back to the list
Type "q"
Sleep 1s

# Jump to Bigtable — showcases the empty-state playful copy
Type ":"
Sleep 600ms
Type "bigtable"
Sleep 1s
Enter
Sleep 4s

# Back out, open the help overlay
Type "q"
Sleep 1s
Type "?"
Sleep 4s
Escape
Sleep 1s

# Return to home — let the ASCII banner hold the closing frame
Type "q"
Sleep 500ms
Type "q"
Sleep 4s

# Exit cleanly
Type "q"
Sleep 1s

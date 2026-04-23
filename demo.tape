Output demo.gif

Require go

Set Shell "bash"
Set FontSize 14
Set Width 1400
Set Height 800
Set Padding 20
Set Theme "GitHub Dark"

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

# Show command palette and switch to Cloud Storage
Type ":"
Sleep 1s
Type "gcs"
Sleep 1s
Enter
Sleep 3s

# Navigate down in bucket list
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Open bucket to browse objects
Enter
Sleep 3s

# Scroll through objects
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Quit out
Type "q"
Sleep 1s
Type "q"
Sleep 1s
Type "q"
Sleep 1s

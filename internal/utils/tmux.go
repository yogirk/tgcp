package utils

import "os"

// IsTmux checks if the application is running inside a tmux session
func IsTmux() bool {
	return os.Getenv("TMUX") != ""
}

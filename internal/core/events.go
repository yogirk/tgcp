package core

import "time"

// StatusMsg updates the status bar message
type StatusMsg struct {
	Message string
	IsError bool
}

// LastUpdatedMsg updates the "Last Updated" timestamp
type LastUpdatedMsg time.Time

// SwitchToLogsMsg requests a context switch to the logging service
type SwitchToLogsMsg struct {
	Filter string
}

// ToastType defines the visual style of a toast notification
type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
)

// ToastMsg triggers a toast notification in the UI
type ToastMsg struct {
	Message  string
	Type     ToastType
	Duration time.Duration // 0 means use default (3 seconds)
}

// LoadingMsg signals loading state changes to MainModel
// Services emit this when they start/stop loading data
type LoadingMsg struct {
	IsLoading bool
	Message   string // Optional custom message (empty = use playful messages)
}

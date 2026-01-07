package gce

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Action messages
type actionResultMsg struct {
	err error
	msg string
}

// StartInstanceCmd triggers the start operation
func (s *Service) StartInstanceCmd(instance Instance) tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return actionResultMsg{err: fmt.Errorf("client not initialized")}
		}
		err := s.client.StartInstance(s.projectID, instance.Zone, instance.Name)
		if err != nil {
			return actionResultMsg{err: err}
		}
		return actionResultMsg{msg: fmt.Sprintf("Starting instance %s...", instance.Name)}
	}
}

// StopInstanceCmd triggers the stop operation
func (s *Service) StopInstanceCmd(instance Instance) tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return actionResultMsg{err: fmt.Errorf("client not initialized")}
		}
		err := s.client.StopInstance(s.projectID, instance.Zone, instance.Name)
		if err != nil {
			return actionResultMsg{err: err}
		}
		return actionResultMsg{msg: fmt.Sprintf("Stopping instance %s...", instance.Name)}
	}
}

// SSHCmd constructs the gcloud ssh command
// Note: In Bubbletea, interactive exec requires Tea.Exec, but for simple MVP
// we might just print the command or try to run it detached.
// For a terminal app, we usually pause the TUI, run SSH, then resume.
func (s *Service) SSHCmd(instance Instance) tea.Cmd {
	cmd := exec.Command("gcloud", "compute", "ssh", instance.Name, "--zone", instance.Zone, "--project", s.projectID)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("SSH failed: %w", err)}
		}
		return actionResultMsg{msg: "SSH session ended"}
	})
}

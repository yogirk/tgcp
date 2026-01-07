package gce

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rk/tgcp/internal/utils"
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
func (s *Service) SSHCmd(instance Instance) tea.Cmd {
	// Build base arguments
	args := []string{"compute", "ssh", instance.Name, "--zone", instance.Zone, "--project", s.projectID}

	// Auto-detect IAP: If no external IP, use IAP tunnel
	if instance.ExternalIP == "" {
		args = append(args, "--tunnel-through-iap")
	}

	// Check for Tmux
	if utils.IsTmux() {
		return func() tea.Msg {
			// Construct the full command string for tmux
			fullCmd := fmt.Sprintf("gcloud %s", strings.Join(args, " "))

			// Tmux split-window command
			// -h for horizontal split
			cmd := exec.Command("tmux", "split-window", "-h", fullCmd)

			if err := cmd.Run(); err != nil {
				return actionResultMsg{err: fmt.Errorf("Tmux split failed: %w", err)}
			}
			return actionResultMsg{msg: "Opened SSH in new pane"}
		}
	}

	// Standard Full Screen SSH
	cmd := exec.Command("gcloud", args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return actionResultMsg{err: fmt.Errorf("SSH failed: %w", err)}
		}
		return actionResultMsg{msg: "SSH session ended"}
	})
}

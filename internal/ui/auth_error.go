package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
)

func renderAuthError(m MainModel) string {
	errTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red
		Bold(true).
		Render("⚠️  Authentication Error ⚠️")

	errMsg := ""
	if m.AuthState.Error != nil {
		errMsg = m.AuthState.Error.Error()
	}

	instruction := `
 TGCP requires Application Default Credentials (ADC) to authenticate with Google Cloud.

 How to fix this:
 
 1. Install gcloud CLI: https://cloud.google.com/sdk/docs/install
 2. Login and set up ADC:
    $ gcloud auth application-default login

 3. Alternatively, set GOOGLE_APPLICATION_CREDENTIALS environment variable:
    $ export GOOGLE_APPLICATION_CREDENTIALS="/path/to/key.json"

 Press 'q' to quit.
`

	box := styles.BoxStyle.Copy().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(60).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				errTitle,
				"\n",
				styles.ErrorStyle.Render(errMsg),
				"\n",
				styles.BaseStyle.Render(instruction),
			),
		)

	return lipgloss.Place(
		m.Width, m.Height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

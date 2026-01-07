package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Google Colors
	colorBlue   = lipgloss.Color("#4285F4")
	colorRed    = lipgloss.Color("#DB4437")
	colorYellow = lipgloss.Color("#F4B400")
	colorGreen  = lipgloss.Color("#0F9D58")

	// Individual Letter Styles
	styleT = lipgloss.NewStyle().Foreground(colorBlue)
	styleG = lipgloss.NewStyle().Foreground(colorRed)
	styleC = lipgloss.NewStyle().Foreground(colorYellow)
	styleP = lipgloss.NewStyle().Foreground(colorGreen)
)

// GetBanner returns the colored ASCII art banner
func GetBanner() string {
	// Individual letters broken down from the original banner
	// Original:
	// ████████╗ ██████╗  ██████╗██████╗
	// ╚══██╔══╝██╔════╝ ██╔════╝██╔══██╗
	//    ██║   ██║  ███╗██║     ██████╔╝
	//    ██║   ██║   ██║██║     ██╔═══╝
	//    ██║   ╚██████╔╝╚██████╗██║
	//    ╚═╝    ╚═════╝  ╚═════╝╚═╝

	letterT := `████████╗
╚══██╔══╝
   ██║   
   ██║   
   ██║   
   ╚═╝   `

	letterG := ` ██████╗ 
██╔════╝ 
██║  ███╗
██║   ██║
╚██████╔╝
 ╚═════╝ `

	letterC := ` ██████╗
 ██╔════╝
 ██║     
 ██║     
 ╚██████╗
  ╚═════╝`

	letterP := `██████╗
██╔══██╗
██████╔╝
██╔═══╝ 
██║     
╚═╝     `

	// Join them horizontally
	// We need to split them into lines first because simple string concatenation
	// won't work for horizontal joining of multiline strings without Lipgloss's help
	// or manual line-by-line zipping.
	// Thankfully, Lipgloss JoinHorizontal handles rendered blocks nicely.

	blockT := styleT.Render(letterT)
	blockG := styleG.Render(letterG)
	blockC := styleC.Render(letterC)
	blockP := styleP.Render(letterP)

	return lipgloss.JoinHorizontal(lipgloss.Bottom, blockT, blockG, blockC, blockP)
}

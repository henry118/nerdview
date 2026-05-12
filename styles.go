package main

import "github.com/charmbracelet/lipgloss"

var (
	styleHeader = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Bold(true)

	styleHeaderNS = lipgloss.NewStyle().
			Foreground(lipgloss.Color("115")).
			Bold(true)

	styleTabActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("63")).
			Bold(true).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Padding(0, 1)

	styleHelpBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236"))

	styleHelpKey = lipgloss.NewStyle().
			Foreground(lipgloss.Color("115")).
			Background(lipgloss.Color("236")).
			Bold(true)

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Background(lipgloss.Color("236")).
			Bold(true)

	styleNSList = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	styleNSListTitle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("115")).
				Bold(true).
				Padding(0, 1)

	styleNSListItem = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250")).
				Padding(0, 1)

	styleNSListSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("232")).
				Background(lipgloss.Color("115")).
				Bold(true).
				Padding(0, 1)
)

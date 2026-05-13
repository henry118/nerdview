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

	styleStatsLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236"))

	styleStatsPID = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236"))

	styleStatsCPU = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Background(lipgloss.Color("236")).
			Bold(true)

	styleStatsVMS = lipgloss.NewStyle().
			Foreground(lipgloss.Color("177")).
			Background(lipgloss.Color("236"))

	styleStatsRSS = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")).
			Background(lipgloss.Color("236")).
			Bold(true)

	styleStatsThreads = lipgloss.NewStyle().
				Foreground(lipgloss.Color("75")).
				Background(lipgloss.Color("236"))

	styleStatsUptime = lipgloss.NewStyle().
				Foreground(lipgloss.Color("223")).
				Background(lipgloss.Color("236"))

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

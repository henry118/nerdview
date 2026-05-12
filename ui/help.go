package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	helpBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236"))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("115")).
			Background(lipgloss.Color("236")).
			Bold(true)
)

func HelpView(width int) string {
	parts := []string{
		helpKeyStyle.Render("Tab") + helpBarStyle.Render(":resource  "),
		helpKeyStyle.Render("n") + helpBarStyle.Render(":namespace  "),
		helpKeyStyle.Render("Enter") + helpBarStyle.Render(":detail  "),
		helpKeyStyle.Render("q") + helpBarStyle.Render(":quit"),
	}
	text := strings.Join(parts, "")
	textWidth := lipgloss.Width(text)
	if width > textWidth {
		text += helpBarStyle.Render(strings.Repeat(" ", width-textWidth))
	}
	return text
}

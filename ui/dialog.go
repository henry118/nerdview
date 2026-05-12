package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	dialogBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	dialogTitleBarStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("63")).
				Padding(0, 1).
				Align(lipgloss.Center)

	dialogFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Align(lipgloss.Right)
)

type DialogModel struct {
	Title    string
	viewport viewport.Model
	width    int
	height   int
}

func NewDialog(width, height int) DialogModel {
	w := min(width-10, 70)
	h := min(height-10, 18)
	if w < 30 {
		w = 30
	}
	if h < 5 {
		h = 5
	}
	vp := viewport.New(w, h)
	return DialogModel{
		viewport: vp,
		width:    w,
		height:   h,
	}
}

func (d *DialogModel) SetContent(title, body string) {
	d.Title = title
	d.viewport.SetContent(body)
	d.viewport.GotoTop()
}

func (d *DialogModel) SetSize(width, height int) {
	w := min(width-10, 70)
	h := min(height-10, 18)
	if w < 30 {
		w = 30
	}
	if h < 5 {
		h = 5
	}
	d.width = w
	d.height = h
	d.viewport.Width = w
	d.viewport.Height = h
}

func (d DialogModel) Update(msg tea.Msg) (DialogModel, tea.Cmd) {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

func (d DialogModel) View() string {
	titleBar := dialogTitleBarStyle.Width(d.width).Render(d.Title)
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Render(strings.Repeat("─", d.width))
	content := d.viewport.View()
	footer := dialogFooterStyle.Width(d.width).Render("Esc: close │ j/k: scroll")

	inner := lipgloss.JoinVertical(lipgloss.Left, titleBar, separator, content, separator, footer)
	return dialogBoxStyle.Render(inner)
}

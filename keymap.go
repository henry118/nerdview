package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	NextResource    key.Binding
	PrevResource    key.Binding
	SelectNS        key.Binding
	SelectSnapshotter key.Binding
	Enter           key.Binding
	Escape          key.Binding
	Quit            key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	NextResource: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("Tab", "next resource"),
	),
	PrevResource: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("S-Tab", "prev resource"),
	),
	SelectNS: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "namespace"),
	),
	SelectSnapshotter: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "snapshotter"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "detail"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "close"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHelpView_ContainsKeys(t *testing.T) {
	view := HelpView(100)

	keys := []string{"Tab", "←/→", "n", "s", "Space", "Enter", "Esc"}
	for _, key := range keys {
		if !strings.Contains(view, key) {
			t.Errorf("HelpView should contain %q", key)
		}
	}
}

func TestHelpView_PadsToWidth(t *testing.T) {
	view := HelpView(120)
	viewWidth := lipgloss.Width(view)

	if viewWidth < 120 {
		t.Errorf("HelpView width = %d, should pad to at least 120", viewWidth)
	}
}

func TestHelpView_NarrowTerminal(t *testing.T) {
	view := HelpView(20)
	// Should not panic or produce empty output
	if view == "" {
		t.Error("HelpView should produce output even for narrow terminals")
	}
}

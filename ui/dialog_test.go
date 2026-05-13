package ui

import (
	"strings"
	"testing"
)

func TestNewDialog(t *testing.T) {
	d := NewDialog(80, 24)
	if d.width != 0 && d.height != 0 {
		// Before SetContent, dimensions are from viewport init
	}
	if d.Title != "" {
		t.Error("New dialog should have empty title")
	}
}

func TestDialogSetContent_AutoSize(t *testing.T) {
	d := NewDialog(80, 24)
	d.SetSize(80, 24)
	d.SetContent("Test Title", "Line 1\nLine 2\nLine 3")

	if d.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", d.Title, "Test Title")
	}
	// Height should be 3 (number of lines)
	if d.height != 3 {
		t.Errorf("Height = %d, want 3 (lines of content)", d.height)
	}
}

func TestDialogSetContent_LargeContent(t *testing.T) {
	d := NewDialog(80, 24)
	d.SetSize(80, 24)

	// Content exceeding terminal height
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line content here"
	}
	d.SetContent("Big", strings.Join(lines, "\n"))

	// Max height should be termH - 8 = 16
	if d.height > 16 {
		t.Errorf("Height = %d, should be capped at 16", d.height)
	}
}

func TestDialogSetContent_WideContent(t *testing.T) {
	d := NewDialog(80, 24)
	d.SetSize(80, 24)

	longLine := strings.Repeat("x", 200)
	d.SetContent("Wide", longLine)

	// Max width should be termW - 6 = 74
	if d.width > 74 {
		t.Errorf("Width = %d, should be capped at 74", d.width)
	}
}

func TestDialogView_ContainsElements(t *testing.T) {
	d := NewDialog(80, 24)
	d.SetSize(80, 24)
	d.SetContent("My Title", "Body text here")

	view := d.View()

	if !strings.Contains(view, "My Title") {
		t.Error("View should contain title")
	}
	if !strings.Contains(view, "Esc: close") {
		t.Error("View should contain footer hint")
	}
	if !strings.Contains(view, "─") {
		t.Error("View should contain separator")
	}
}

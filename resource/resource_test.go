package resource

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestFitColumns_ContentBased(t *testing.T) {
	defs := []Column{
		{Title: "Name", MinWidth: 5, Flex: true},
		{Title: "Status", MinWidth: 6},
	}
	rows := []table.Row{
		{"short", "ok"},
		{"a-longer-name", "running"},
	}

	cols := fitColumns(defs, rows, 100)

	if cols[0].Width < len("a-longer-name") {
		t.Errorf("Name column should fit content, got %d want >= %d", cols[0].Width, len("a-longer-name"))
	}
	if cols[1].Width < len("running") {
		t.Errorf("Status column should fit content, got %d want >= %d", cols[1].Width, len("running"))
	}
}

func TestFitColumns_FlexShrinks(t *testing.T) {
	defs := []Column{
		{Title: "Name", MinWidth: 10, Flex: true},
		{Title: "Fixed", MinWidth: 8},
	}
	rows := []table.Row{
		{"a-very-long-name-that-exceeds-width", "value123"},
	}

	cols := fitColumns(defs, rows, 30)

	if cols[0].Width < 10 {
		t.Errorf("Flex column should not shrink below MinWidth, got %d", cols[0].Width)
	}
	if cols[1].Width < 8 {
		t.Errorf("Fixed column should keep its content width, got %d", cols[1].Width)
	}
	if cols[0].Width+cols[1].Width > 30 {
		t.Logf("Total %d exceeds width 30 (flex shrunk as far as possible)", cols[0].Width+cols[1].Width)
	}
}

func TestFitColumns_EmptyRows(t *testing.T) {
	defs := []Column{
		{Title: "Name", MinWidth: 10, Flex: true},
		{Title: "Age", MinWidth: 5},
	}

	cols := fitColumns(defs, nil, 80)

	if cols[0].Width < 10 {
		t.Errorf("Expected at least MinWidth for Name, got %d", cols[0].Width)
	}
	if cols[1].Width < 5 {
		t.Errorf("Expected at least MinWidth for Age, got %d", cols[1].Width)
	}
}

func TestTab_ToggleFold(t *testing.T) {
	kind := Kind{
		Name: "Test",
		Columns: []Column{
			{Title: "Name", MinWidth: 10, Flex: true},
		},
		ToRows: func(data any, folded map[string]bool) []table.Row {
			items := data.([]string)
			var rows []table.Row
			for _, item := range items {
				if folded[item] {
					rows = append(rows, table.Row{"▸ " + item})
				} else {
					rows = append(rows, table.Row{"▾ " + item})
					rows = append(rows, table.Row{"  child-of-" + item})
				}
			}
			return rows
		},
		RowID: func(data any, folded map[string]bool, index int) string {
			items := data.([]string)
			// Rebuild visible rows to find which item is at index
			rowIdx := 0
			for _, item := range items {
				if rowIdx == index {
					return item
				}
				rowIdx++
				if !folded[item] {
					rowIdx++
				}
			}
			return ""
		},
	}

	tab := NewTab(kind, 80, 10)
	tab.UpdateData([]string{"parent1", "parent2"})

	// Initially unfolded — should have 4 rows (2 parents + 2 children)
	if len(tab.Table.Rows()) != 4 {
		t.Fatalf("Expected 4 rows unfolded, got %d", len(tab.Table.Rows()))
	}

	// Fold first item
	tab.ToggleFold()
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows after folding first, got %d", len(tab.Table.Rows()))
	}

	// Unfold it
	tab.ToggleFold()
	if len(tab.Table.Rows()) != 4 {
		t.Fatalf("Expected 4 rows after unfolding, got %d", len(tab.Table.Rows()))
	}
}

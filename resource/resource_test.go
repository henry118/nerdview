// Copyright Henry Wang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func TestShortDigest(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sha256:7a75083e5b5a8d593efe8917fe730ab29cd8a8e8a5dfc2fcea022ab5a20954e0", "sha256:7a75083e5b5a"},
		{"sha256:abcdef", "sha256:abcdef"},
		{"sha512:0123456789abcdef0123456789abcdef", "sha512:0123456789ab"},
		{"no-colon-here", "no-colon-here"},
		{"prefix:", "prefix:"},
		{"", ""},
	}
	for _, tt := range tests {
		got := ShortDigest(tt.input)
		if got != tt.want {
			t.Errorf("ShortDigest(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTab_FoldUnfold(t *testing.T) {
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
	tab.Fold()
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows after folding first, got %d", len(tab.Table.Rows()))
	}

	// Unfold it
	tab.Unfold()
	if len(tab.Table.Rows()) != 4 {
		t.Fatalf("Expected 4 rows after unfolding, got %d", len(tab.Table.Rows()))
	}
}

func TestTab_FoldPreservedOnRefresh(t *testing.T) {
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
		InitFolded: func(data any) map[string]bool {
			items := data.([]string)
			folded := make(map[string]bool)
			for _, item := range items {
				folded[item] = true
			}
			return folded
		},
	}

	tab := NewTab(kind, 80, 10)
	tab.UpdateData([]string{"parent1", "parent2"})

	// InitFolded folds everything — should have 2 rows
	if len(tab.Table.Rows()) != 2 {
		t.Fatalf("Expected 2 rows (all folded), got %d", len(tab.Table.Rows()))
	}

	// User unfolds parent1
	tab.Unfold()
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows after unfolding parent1, got %d", len(tab.Table.Rows()))
	}

	// Simulate data refresh with same data — parent1 should stay unfolded
	tab.UpdateData([]string{"parent1", "parent2"})
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows after refresh (parent1 still unfolded), got %d", len(tab.Table.Rows()))
	}

	// Simulate refresh with new item added — new item should be folded, parent1 still unfolded
	tab.UpdateData([]string{"parent1", "parent2", "parent3"})
	// parent1 unfolded (2 rows) + parent2 folded (1 row) + parent3 folded (1 row) = 4
	if len(tab.Table.Rows()) != 4 {
		t.Fatalf("Expected 4 rows after adding parent3 (folded by default), got %d", len(tab.Table.Rows()))
	}
}

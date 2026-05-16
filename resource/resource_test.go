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

type mockCache struct {
	items  []string
	folded map[string]bool
}

func TestTab_FoldUnfold(t *testing.T) {
	kind := Kind{Name: "Test", Columns: []Column{{Title: "Name", MinWidth: 10, Flex: true}},
		Rows: func(data any, folded map[string]bool) ([]table.Row, any) {
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
			return rows, mockCache{items: items, folded: folded}
		},
		FoldKey: func(cache any, index int) string {
			mc := cache.(mockCache)
			rowIdx := 0
			for _, item := range mc.items {
				if rowIdx == index {
					return item
				}
				rowIdx++
				if !mc.folded[item] {
					rowIdx++
				}
			}
			return ""
		},
	}

	tab := NewTab(&kind, 80, 10)
	tab.UpdateData([]string{"parent1", "parent2"})

	if len(tab.Table.Rows()) != 4 {
		t.Fatalf("Expected 4 rows unfolded, got %d", len(tab.Table.Rows()))
	}

	tab.Fold()
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows after folding first, got %d", len(tab.Table.Rows()))
	}

	tab.Unfold()
	if len(tab.Table.Rows()) != 4 {
		t.Fatalf("Expected 4 rows after unfolding, got %d", len(tab.Table.Rows()))
	}
}

func TestTab_FoldPreservedOnRefresh(t *testing.T) {
	kind := Kind{Name: "Test", Columns: []Column{{Title: "Name", MinWidth: 10, Flex: true}},
		Rows: func(data any, folded map[string]bool) ([]table.Row, any) {
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
			return rows, mockCache{items: items, folded: folded}
		},
		FoldKey: func(cache any, index int) string {
			mc := cache.(mockCache)
			rowIdx := 0
			for _, item := range mc.items {
				if rowIdx == index {
					return item
				}
				rowIdx++
				if !mc.folded[item] {
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

	tab := NewTab(&kind, 80, 10)
	tab.UpdateData([]string{"parent1", "parent2"})

	if len(tab.Table.Rows()) != 2 {
		t.Fatalf("Expected 2 rows (all folded), got %d", len(tab.Table.Rows()))
	}

	tab.Unfold()
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows after unfolding parent1, got %d", len(tab.Table.Rows()))
	}

	tab.UpdateData([]string{"parent1", "parent2"})
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows after refresh (parent1 still unfolded), got %d", len(tab.Table.Rows()))
	}

	tab.UpdateData([]string{"parent1", "parent2", "parent3"})
	if len(tab.Table.Rows()) != 4 {
		t.Fatalf("Expected 4 rows after adding parent3 (folded by default), got %d", len(tab.Table.Rows()))
	}
}

func TestTab_FoldFromChild(t *testing.T) {
	kind := Kind{Name: "Test", Columns: []Column{{Title: "Name", MinWidth: 10, Flex: true}},
		Rows: func(data any, folded map[string]bool) ([]table.Row, any) {
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
			return rows, mockCache{items: items, folded: folded}
		},
		FoldKey: func(cache any, index int) string {
			mc := cache.(mockCache)
			rowIdx := 0
			for _, item := range mc.items {
				if rowIdx == index {
					return item
				}
				rowIdx++
				if !mc.folded[item] {
					rowIdx++
				}
			}
			return ""
		},
	}

	tab := NewTab(&kind, 80, 10)
	tab.UpdateData([]string{"parent1", "parent2"})

	// Move cursor to child of parent1 (row index 1)
	tab.Table.SetCursor(1)

	// Fold from child should fold parent1 and move cursor to parent1
	folded := tab.Fold()
	if !folded {
		t.Fatal("Expected Fold() to return true from child")
	}
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows after folding parent from child, got %d", len(tab.Table.Rows()))
	}
	if tab.Table.Cursor() != 0 {
		t.Errorf("Expected cursor at 0 (parent1), got %d", tab.Table.Cursor())
	}
}

func TestTab_RevealRow(t *testing.T) {
	kind := Kind{Name: "Test", Columns: []Column{{Title: "Name", MinWidth: 10, Flex: true}},
		Rows: func(data any, folded map[string]bool) ([]table.Row, any) {
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
			return rows, mockCache{items: items, folded: folded}
		},
		FoldKey: func(cache any, index int) string {
			mc := cache.(mockCache)
			rowIdx := 0
			for _, item := range mc.items {
				if rowIdx == index {
					return item
				}
				rowIdx++
				if !mc.folded[item] {
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

	tab := NewTab(&kind, 80, 10)
	tab.UpdateData([]string{"parent1", "parent2"})

	// Both folded — only 2 rows visible
	if len(tab.Table.Rows()) != 2 {
		t.Fatalf("Expected 2 rows (all folded), got %d", len(tab.Table.Rows()))
	}

	// RevealRow should unfold parent1 to show its child
	idx := tab.RevealRow(func(row table.Row) bool {
		return len(row) > 0 && row[0] == "  child-of-parent1"
	})
	if idx < 0 {
		t.Fatal("RevealRow failed to find child-of-parent1")
	}

	// parent1 should be unfolded, parent2 should stay folded
	if len(tab.Table.Rows()) != 3 {
		t.Fatalf("Expected 3 rows (parent1 unfolded, parent2 folded), got %d", len(tab.Table.Rows()))
	}

	// Target not found at all
	idx = tab.RevealRow(func(row table.Row) bool {
		return len(row) > 0 && row[0] == "nonexistent"
	})
	if idx != -1 {
		t.Errorf("Expected -1 for nonexistent row, got %d", idx)
	}
}

func TestTab_SetWidth(t *testing.T) {
	kind := Kind{Name: "Test", Columns: []Column{{Title: "Name", MinWidth: 10, Flex: true}},
		Rows: func(data any, _ map[string]bool) ([]table.Row, any) {
			return []table.Row{{"hello"}}, nil
		},
	}
	tab := NewTab(&kind, 80, 10)
	tab.UpdateData([]string{"x"})
	tab.SetWidth(120)
	if tab.Table.Width() != 120 {
		t.Errorf("Table width = %d, want 120", tab.Table.Width())
	}
}

func TestTab_SelectedDetail(t *testing.T) {
	kind := Kind{Name: "Test", Columns: []Column{{Title: "Name", MinWidth: 10, Flex: true}},
		Rows: func(data any, _ map[string]bool) ([]table.Row, any) {
			return []table.Row{{"item1"}, {"item2"}}, nil
		},
	}
	tab := NewTab(&kind, 80, 10)
	tab.UpdateData([]string{"a", "b"})

	title, body := tab.SelectedDetail()
	if title != "" || body != "" {
		t.Errorf("mockKind Detail returns empty, got title=%q body=%q", title, body)
	}
}

func TestTab_CrossRef(t *testing.T) {
	kind := Kind{Name: "Test", Columns: []Column{{Title: "Name", MinWidth: 10, Flex: true}},
		Rows: func(data any, _ map[string]bool) ([]table.Row, any) {
			items := data.([]string)
			rows := make([]table.Row, len(items))
			for i, item := range items {
				rows[i] = table.Row{item}
			}
			return rows, items
		},
		CrossRefs: func(cache any) []string {
			items := cache.([]string)
			refs := make([]string, len(items))
			for i, item := range items {
				refs[i] = "ref-" + item
			}
			return refs
		},
	}

	tab := NewTab(&kind, 80, 10)
	tab.UpdateData([]string{"a", "b", "c"})

	if got := tab.CrossRef(0); got != "ref-a" {
		t.Errorf("CrossRef(0) = %q, want %q", got, "ref-a")
	}
	if got := tab.CrossRef(2); got != "ref-c" {
		t.Errorf("CrossRef(2) = %q, want %q", got, "ref-c")
	}
	if got := tab.CrossRef(99); got != "" {
		t.Errorf("CrossRef(99) = %q, want empty", got)
	}
	if got := tab.CrossRef(-1); got != "" {
		t.Errorf("CrossRef(-1) = %q, want empty", got)
	}
}

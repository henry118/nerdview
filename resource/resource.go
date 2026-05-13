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

// Package resource defines the abstraction for displaying containerd resources
// (images, containers, tasks, snapshots, events) in foldable tree tables.
package resource

import "github.com/charmbracelet/bubbles/table"

// Column defines a table column with optional flex behavior for dynamic sizing.
type Column struct {
	Title    string
	MinWidth int
	Flex     bool // Flex columns shrink when total width exceeds terminal width.
}

// Kind describes a resource type and how to render it. Each resource (images,
// containers, etc.) defines a Kind with column definitions and callbacks for
// converting raw data into table rows and detail views.
type Kind struct {
	Name    string
	Columns []Column
	// ToRows converts raw data into table rows, respecting the current fold state.
	ToRows func(data any, folded map[string]bool) []table.Row
	// RowID returns a unique fold key for the row at the given index.
	// Return empty string for non-foldable rows. If nil, folding is disabled.
	RowID func(data any, folded map[string]bool, index int) string
	// InitFolded returns the default fold state for newly loaded data.
	// Keys map to true (folded) or false (unfolded). If nil, nothing is folded.
	InitFolded func(data any) map[string]bool
	// ToDetail returns a title and formatted body for the detail dialog.
	ToDetail func(data any, folded map[string]bool, index int) (title string, body string)
}

// Tab wraps a table model with its Kind, raw data, and fold state.
type Tab struct {
	Kind    Kind
	Table   table.Model
	RawData any
	Folded  map[string]bool
	width   int
}

// NewTab creates a Tab for the given Kind with initial dimensions.
func NewTab(kind Kind, width, height int) Tab {
	cols := fitColumns(kind.Columns, nil, width)
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(height),
		table.WithWidth(width),
	)
	s := table.DefaultStyles()
	t.SetStyles(s)

	return Tab{
		Kind:   kind,
		Table:  t,
		Folded: make(map[string]bool),
		width:  width,
	}
}

// SetWidth updates the table and column widths for the current terminal width.
func (t *Tab) SetWidth(width int) {
	t.width = width
	t.Table.SetWidth(width)
	t.recalcColumns()
}

// UpdateData replaces the tab's data and refreshes rows. New foldable nodes
// are folded by default; user-toggled fold state is preserved across refreshes.
func (t *Tab) UpdateData(data any) {
	t.RawData = data
	if t.Kind.InitFolded != nil {
		// Merge: fold new nodes that aren't already tracked
		defaults := t.Kind.InitFolded(data)
		for k, v := range defaults {
			if _, exists := t.Folded[k]; !exists {
				t.Folded[k] = v
			}
		}
	}
	t.refreshRows()
}

func (t *Tab) refreshRows() {
	rows := t.Kind.ToRows(t.RawData, t.Folded)
	t.Table.SetRows(rows)
	t.recalcColumns()
}

// ToggleFold folds or unfolds the currently selected row.
func (t *Tab) ToggleFold() {
	if t.Kind.RowID == nil || t.RawData == nil {
		return
	}
	idx := t.Table.Cursor()
	id := t.Kind.RowID(t.RawData, t.Folded, idx)
	if id == "" {
		return
	}
	t.Folded[id] = !t.Folded[id]
	t.refreshRows()
}

// CanFold reports whether this tab supports folding.
func (t *Tab) CanFold() bool {
	return t.Kind.RowID != nil
}

func (t *Tab) recalcColumns() {
	rows := t.Table.Rows()
	cols := fitColumns(t.Kind.Columns, rows, t.width)
	t.Table.SetColumns(cols)
}

// SelectedDetail returns the title and body for the currently selected row's detail view.
func (t *Tab) SelectedDetail() (string, string) {
	if t.RawData == nil {
		return "", ""
	}
	idx := t.Table.Cursor()
	return t.Kind.ToDetail(t.RawData, t.Folded, idx)
}

func fitColumns(defs []Column, rows []table.Row, totalWidth int) []table.Column {
	cols := make([]table.Column, len(defs))

	widths := make([]int, len(defs))
	for i, d := range defs {
		widths[i] = len(d.Title)
		if widths[i] < d.MinWidth {
			widths[i] = d.MinWidth
		}
	}

	for _, row := range rows {
		for i := range defs {
			if i < len(row) {
				cellLen := len(row[i])
				if cellLen > widths[i] {
					widths[i] = cellLen
				}
			}
		}
	}

	total := 0
	for _, w := range widths {
		total += w
	}
	if total > totalWidth {
		overflow := total - totalWidth
		for i, d := range defs {
			if d.Flex && overflow > 0 {
				shrink := widths[i] - d.MinWidth
				if shrink > overflow {
					shrink = overflow
				}
				if shrink > 0 {
					widths[i] -= shrink
					overflow -= shrink
				}
			}
		}
	}

	for i, d := range defs {
		cols[i] = table.Column{Title: d.Title, Width: widths[i]}
	}
	return cols
}

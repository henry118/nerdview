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

import (
	"fmt"
	"maps"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/henry118/nerdview/ui"
)

// FormatBytes formats a byte count into a human-readable string (B/K/M/G).
func FormatBytes(b uint64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

// ShortDigest truncates "algo:hex" to show only the first 12 hex chars.
// Strings without a colon are returned unchanged.
func ShortDigest(s string) string {
	if idx := strings.Index(s, ":"); idx >= 0 {
		hex := s[idx+1:]
		if len(hex) > 12 {
			return s[:idx+1] + hex[:12]
		}
	}
	return s
}

// Column defines a table column with optional flex behavior for dynamic sizing.
type Column struct {
	Title    string
	MinWidth int
	Flex     bool // Flex columns shrink proportionally when total exceeds terminal width.
}

// Kind defines how a resource type (images, containers, etc.) is displayed
// and interacted with in the TUI. Each resource implements this interface.
type Kind interface {
	Name() string
	Columns() []Column
	// Rows converts raw data into table rows, respecting the current fold state.
	Rows(data any, folded map[string]bool) []table.Row
	// FoldKey returns the fold key for the row at index, or "" if not foldable.
	FoldKey(data any, folded map[string]bool, index int) string
	// InitFolded returns the default fold state for newly loaded data.
	InitFolded(data any) map[string]bool
	// Detail returns a title and formatted body for the detail popup.
	Detail(data any, folded map[string]bool, index int) (string, string)
	// CrossRefs returns navigation references for all visible rows.
	CrossRefs(data any, folded map[string]bool) []string
}

// Tab wraps a table model with its Kind, raw data, and fold state.
// It manages the lifecycle of rendering rows, column sizing, and fold operations.
type Tab struct {
	Kind      Kind            // Resource type defining how data is rendered.
	Table     table.Model     // Underlying bubbletea table component.
	RawData   any             // Raw data slice (type-asserted by Kind methods).
	Folded    map[string]bool // Fold state keyed by fold key; true means collapsed.
	crossRefs []string        // Cached cross-references for navigation.
	width     int             // Current terminal width for column sizing.
}

// NewTab creates a Tab for the given Kind with initial dimensions.
func NewTab(kind Kind, width, height int) Tab {
	cols := fitColumns(kind.Columns(), nil, width)
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(height),
		table.WithWidth(width),
	)
	s := table.DefaultStyles()
	s.Selected = ui.StyleTableSelected
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

// UpdateData replaces the tab's data and refreshes rows. User-toggled fold
// state is preserved across refreshes; new foldable nodes use InitFolded defaults.
func (t *Tab) UpdateData(data any) {
	t.RawData = data
	defaults := t.Kind.InitFolded(data)
	for k, v := range defaults {
		if _, exists := t.Folded[k]; !exists {
			t.Folded[k] = v
		}
	}
	t.refreshRows()
}

func (t *Tab) refreshRows() {
	rows := t.Kind.Rows(t.RawData, t.Folded)
	t.Table.SetRows(rows)
	t.recalcColumns()
	t.buildCrossRefs()
}

func (t *Tab) buildCrossRefs() {
	t.crossRefs = t.Kind.CrossRefs(t.RawData, t.Folded)
}

// Unfold expands the current row if it is a folded node. Returns true if it unfolded.
func (t *Tab) Unfold() bool {
	if t.RawData == nil {
		return false
	}
	idx := t.Table.Cursor()
	id := t.Kind.FoldKey(t.RawData, t.Folded, idx)
	if id == "" || !t.Folded[id] {
		return false
	}
	t.Folded[id] = false
	t.refreshRows()
	return true
}

// Fold collapses the current node or its nearest foldable ancestor.
// Moves the cursor to the folded parent. Returns true if it folded.
func (t *Tab) Fold() bool {
	if t.RawData == nil {
		return false
	}
	idx := t.Table.Cursor()

	id := t.Kind.FoldKey(t.RawData, t.Folded, idx)
	if id != "" && !t.Folded[id] {
		t.Folded[id] = true
		t.refreshRows()
		return true
	}

	for i := idx - 1; i >= 0; i-- {
		parentID := t.Kind.FoldKey(t.RawData, t.Folded, i)
		if parentID != "" && !t.Folded[parentID] {
			t.Folded[parentID] = true
			t.refreshRows()
			t.Table.SetCursor(i)
			return true
		}
	}
	return false
}

// RevealRow unfolds only the ancestor that hides a row matching the predicate.
// Returns the row index if found, or -1.
func (t *Tab) RevealRow(match func(row table.Row) bool) int {
	rows := t.Table.Rows()
	for i, row := range rows {
		if match(row) {
			return i
		}
	}

	savedFolded := make(map[string]bool, len(t.Folded))
	maps.Copy(savedFolded, t.Folded)

	for id := range t.Folded {
		t.Folded[id] = false
	}
	t.refreshRows()

	rows = t.Table.Rows()
	targetIdx := -1
	for i, row := range rows {
		if match(row) {
			targetIdx = i
			break
		}
	}

	if targetIdx < 0 {
		t.Folded = savedFolded
		t.refreshRows()
		return -1
	}

	ancestorID := ""
	for i := targetIdx - 1; i >= 0; i-- {
		id := t.Kind.FoldKey(t.RawData, t.Folded, i)
		if id != "" {
			ancestorID = id
			break
		}
	}

	t.Folded = savedFolded
	if ancestorID != "" {
		t.Folded[ancestorID] = false
	}
	t.refreshRows()

	rows = t.Table.Rows()
	for i, row := range rows {
		if match(row) {
			return i
		}
	}
	return -1
}

func (t *Tab) recalcColumns() {
	rows := t.Table.Rows()
	cols := fitColumns(t.Kind.Columns(), rows, t.width)
	t.Table.SetColumns(cols)
}

// CrossRef returns the cached cross-reference for the row at index.
func (t *Tab) CrossRef(index int) string {
	if index >= 0 && index < len(t.crossRefs) {
		return t.crossRefs[index]
	}
	return ""
}

// SelectedDetail returns the title and body for the currently selected row's detail popup.
func (t *Tab) SelectedDetail() (string, string) {
	if t.RawData == nil {
		return "", ""
	}
	idx := t.Table.Cursor()
	return t.Kind.Detail(t.RawData, t.Folded, idx)
}

func fitColumns(defs []Column, rows []table.Row, totalWidth int) []table.Column {
	cols := make([]table.Column, len(defs))

	cellPadding := 2 * len(defs)
	availableWidth := max(totalWidth-cellPadding, len(defs))

	widths := make([]int, len(defs))
	for i, d := range defs {
		widths[i] = max(len(d.Title), d.MinWidth)
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
	if total > availableWidth {
		overflow := total - availableWidth
		for overflow > 0 {
			var flexCount int
			for i, d := range defs {
				if d.Flex && widths[i] > d.MinWidth {
					flexCount++
				}
			}
			if flexCount == 0 {
				break
			}
			perCol := overflow / flexCount
			if perCol == 0 {
				perCol = 1
			}
			shrunk := 0
			for i, d := range defs {
				if d.Flex && widths[i] > d.MinWidth {
					shrink := min(perCol, widths[i]-d.MinWidth)
					widths[i] -= shrink
					shrunk += shrink
				}
			}
			overflow -= shrunk
		}
	}

	for i, d := range defs {
		cols[i] = table.Column{Title: d.Title, Width: widths[i]}
	}
	return cols
}

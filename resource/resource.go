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

	"charm.land/bubbles/v2/table"
	"github.com/henry118/nerdview/logging"
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
	// Flex columns shrink proportionally when total exceeds terminal width.
	Flex bool
}

// Kind defines how a resource type is displayed. Rows produces an opaque cache
// token that callers pass back to FoldKey, Detail, and CrossRefs. Nil function
// fields are safe to leave unset — Tab methods handle nil gracefully.
type Kind struct {
	// Display name for the tab header.
	Name string
	// Table column definitions.
	Columns []Column
	// Builds display rows from data and returns an opaque cache for subsequent lookups.
	// Required.
	Rows func(data any, folded map[string]bool) ([]table.Row, any)
	// Returns the fold key for the row at index, or "" if not foldable.
	// Nil if the resource has no tree structure.
	FoldKey func(cache any, index int) string
	// Returns the default fold state for newly loaded data.
	// Nil if no nodes should be folded by default.
	InitFolded func(data any) map[string]bool
	// Returns a title and formatted body for the detail popup.
	// Nil if detail view is not supported.
	Detail func(cache any, index int) (string, string)
	// Returns navigation references for all visible rows.
	// Nil if cross-tab navigation is not supported.
	CrossRefs func(cache any) []string
	// Returns a title and formatted spec body for the spec popup.
	// Nil if spec view is not supported.
	Spec func(cache any, index int) (string, string)
}

// Tab owns all mutable runtime state for a displayed resource: data, fold map,
// cached tree result, and cross-references. It delegates rendering and
// interpretation logic to its Kind.
type Tab struct {
	// Resource type defining display logic.
	kind *Kind
	// Underlying bubbletea table component.
	Table table.Model
	// Opaque cache token produced by Kind.Rows.
	cache any
	// Raw data slice from containerd.
	rawData any
	// Fold state keyed by fold key; true means collapsed.
	folded map[string]bool
	// Cached cross-references for navigation.
	crossRefs []string
	// Current terminal width for column sizing.
	width int
}

// NewTab creates a Tab for the given Kind with initial dimensions.
func NewTab(kind *Kind, width, height int) *Tab {
	cols := fitColumns(kind.Columns, nil, width)
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(height),
		table.WithWidth(width),
	)
	s := table.DefaultStyles()
	s.Header = ui.StyleTableHeader
	s.Selected = ui.StyleTableSelected
	t.SetStyles(s)

	return &Tab{
		kind:   kind,
		Table:  t,
		folded: make(map[string]bool),
		width:  width,
	}
}

// Name returns the display name for this tab.
func (t *Tab) Name() string {
	return t.kind.Name
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
	t.rawData = data
	if t.kind.InitFolded != nil {
		defaults := t.kind.InitFolded(data)
		for k, v := range defaults {
			if _, exists := t.folded[k]; !exists {
				t.folded[k] = v
			}
		}
	}
	t.refreshRows()
}

func (t *Tab) refreshRows() {
	rows, cache := t.kind.Rows(t.rawData, t.folded)
	t.cache = cache
	t.Table.SetRows(rows)
	t.recalcColumns()
	t.buildCrossRefs()
}

func (t *Tab) buildCrossRefs() {
	if t.kind.CrossRefs != nil {
		t.crossRefs = t.kind.CrossRefs(t.cache)
	} else {
		t.crossRefs = nil
	}
}

// Clear resets the tab to an empty state.
func (t *Tab) Clear() {
	t.rawData = nil
	t.cache = nil
	t.crossRefs = nil
	t.folded = make(map[string]bool)
	t.Table.SetRows(nil)
}

// Unfold expands the current row if it is a folded node. Returns true if it unfolded.
func (t *Tab) Unfold() bool {
	if t.rawData == nil || t.kind.FoldKey == nil {
		return false
	}
	idx := t.Table.Cursor()
	id := t.kind.FoldKey(t.cache, idx)
	if id == "" || !t.folded[id] {
		return false
	}
	t.folded[id] = false
	t.refreshRows()
	return true
}

// Fold collapses the current node or its nearest foldable ancestor.
// Moves the cursor to the folded parent. Returns true if it folded.
func (t *Tab) Fold() bool {
	if t.rawData == nil || t.kind.FoldKey == nil {
		return false
	}
	idx := t.Table.Cursor()

	id := t.kind.FoldKey(t.cache, idx)
	if id != "" && !t.folded[id] {
		t.folded[id] = true
		t.refreshRows()
		return true
	}

	for i := idx - 1; i >= 0; i-- {
		parentID := t.kind.FoldKey(t.cache, i)
		if parentID != "" && !t.folded[parentID] {
			t.folded[parentID] = true
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

	if t.kind.FoldKey == nil {
		return -1
	}

	savedFolded := make(map[string]bool, len(t.folded))
	maps.Copy(savedFolded, t.folded)

	for id := range t.folded {
		t.folded[id] = false
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
		t.folded = savedFolded
		t.refreshRows()
		return -1
	}

	ancestorID := ""
	for i := targetIdx - 1; i >= 0; i-- {
		id := t.kind.FoldKey(t.cache, i)
		if id != "" {
			ancestorID = id
			break
		}
	}

	t.folded = savedFolded
	if ancestorID != "" {
		logging.Debug("RevealRow: unfolding ancestor %s", ancestorID)
		t.folded[ancestorID] = false
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
	cols := fitColumns(t.kind.Columns, rows, t.width)
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
	if t.rawData == nil || t.kind.Detail == nil {
		return "", ""
	}
	idx := t.Table.Cursor()
	return t.kind.Detail(t.cache, idx)
}

// HasSpec reports whether this tab's Kind supports the spec popup.
func (t *Tab) HasSpec() bool {
	return t.kind.Spec != nil
}

// Spec returns the runtime spec for the currently selected row.
// Returns empty strings if the Kind does not support specs.
func (t *Tab) Spec() (string, string) {
	if t.kind.Spec != nil {
		return t.kind.Spec(t.cache, t.Table.Cursor())
	}
	return "", ""
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

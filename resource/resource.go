package resource

import "github.com/charmbracelet/bubbles/table"

type Column struct {
	Title    string
	MinWidth int
	Flex     bool
}

type Kind struct {
	Name     string
	Columns  []Column
	ToRows   func(data any, folded map[string]bool) []table.Row
	// RowID returns a unique fold key for the row at index.
	// If nil, folding is not supported for this kind.
	RowID func(data any, folded map[string]bool, index int) string
	// InitFolded returns the default folded set for new data.
	// If nil, nothing is folded by default.
	InitFolded func(data any) map[string]bool
	ToDetail   func(data any, folded map[string]bool, index int) (title string, body string)
}

type Tab struct {
	Kind    Kind
	Table   table.Model
	RawData any
	Folded  map[string]bool
	width   int
}

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

func (t *Tab) SetWidth(width int) {
	t.width = width
	t.Table.SetWidth(width)
	t.recalcColumns()
}

func (t *Tab) UpdateData(data any) {
	firstLoad := t.RawData == nil
	t.RawData = data
	if firstLoad && t.Kind.InitFolded != nil {
		t.Folded = t.Kind.InitFolded(data)
	}
	t.refreshRows()
}

func (t *Tab) refreshRows() {
	rows := t.Kind.ToRows(t.RawData, t.Folded)
	t.Table.SetRows(rows)
	t.recalcColumns()
}

func (t *Tab) ToggleFold() {
	if t.Kind.RowID == nil || t.RawData == nil {
		return
	}
	idx := t.Table.Cursor()
	id := t.Kind.RowID(t.RawData, t.Folded, idx)
	if id == "" {
		return
	}
	if t.Folded[id] {
		delete(t.Folded, id)
	} else {
		t.Folded[id] = true
	}
	t.refreshRows()
}

func (t *Tab) CanFold() bool {
	return t.Kind.RowID != nil
}

func (t *Tab) recalcColumns() {
	rows := t.Table.Rows()
	cols := fitColumns(t.Kind.Columns, rows, t.width)
	t.Table.SetColumns(cols)
}

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

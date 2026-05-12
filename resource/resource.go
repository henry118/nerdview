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
	ToRows   func(data any) []table.Row
	ToDetail func(data any, index int) (title string, body string)
}

type Tab struct {
	Kind    Kind
	Table   table.Model
	RawData any
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
		Kind:  kind,
		Table: t,
		width: width,
	}
}

func (t *Tab) SetWidth(width int) {
	t.width = width
	t.Table.SetWidth(width)
	t.recalcColumns()
}

func (t *Tab) UpdateData(data any) {
	t.RawData = data
	rows := t.Kind.ToRows(data)
	t.Table.SetRows(rows)
	t.recalcColumns()
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
	return t.Kind.ToDetail(t.RawData, idx)
}

// fitColumns sizes each column just wide enough to show its content.
// If the total exceeds terminal width, flex columns get shrunk back to MinWidth.
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

	// If total exceeds available width, shrink flex columns
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

package resource

import "github.com/charmbracelet/bubbles/table"

type Kind struct {
	Name     string
	Columns  []table.Column
	ToRows   func(data any) []table.Row
	ToDetail func(data any, index int) (title string, body string)
}

type Tab struct {
	Kind    Kind
	Table   table.Model
	RawData any
}

func NewTab(kind Kind, width, height int) Tab {
	t := table.New(
		table.WithColumns(kind.Columns),
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
	}
}

func (t *Tab) UpdateData(data any) {
	t.RawData = data
	rows := t.Kind.ToRows(data)
	t.Table.SetRows(rows)
}

func (t *Tab) SelectedDetail() (string, string) {
	if t.RawData == nil {
		return "", ""
	}
	idx := t.Table.Cursor()
	return t.Kind.ToDetail(t.RawData, idx)
}

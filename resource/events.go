package resource

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

type Event struct {
	Timestamp time.Time
	Namespace string
	Topic     string
}

var EventKind = Kind{
	Name: "Events",
	Columns: []Column{
		{Title: "Time", MinWidth: 14},
		{Title: "Namespace", MinWidth: 14},
		{Title: "Topic", MinWidth: 20, Flex: true},
	},
	ToRows: func(data any) []table.Row {
		evts, ok := data.([]Event)
		if !ok || len(evts) == 0 {
			return nil
		}
		rows := make([]table.Row, len(evts))
		for i, e := range evts {
			rows[i] = table.Row{
				e.Timestamp.Format("15:04:05.000"),
				e.Namespace,
				e.Topic,
			}
		}
		return rows
	},
	ToDetail: func(data any, index int) (string, string) {
		evts, ok := data.([]Event)
		if !ok || index < 0 || index >= len(evts) {
			return "", ""
		}
		e := evts[index]
		var b strings.Builder
		fmt.Fprintf(&b, "Timestamp:  %s\n", e.Timestamp.Format(time.RFC3339Nano))
		fmt.Fprintf(&b, "Namespace:  %s\n", e.Namespace)
		fmt.Fprintf(&b, "Topic:      %s\n", e.Topic)
		return e.Topic, b.String()
	},
}

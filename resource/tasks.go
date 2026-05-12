package resource

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tasktypes "github.com/containerd/containerd/api/types/task"
)

var TaskKind = Kind{
	Name: "Tasks",
	Columns: []table.Column{
		{Title: "Container ID", Width: 20},
		{Title: "PID", Width: 8},
		{Title: "Status", Width: 12},
	},
	ToRows: func(data any) []table.Row {
		procs, ok := data.([]*tasktypes.Process)
		if !ok || len(procs) == 0 {
			return nil
		}
		rows := make([]table.Row, len(procs))
		for i, p := range procs {
			id := p.ID
			if len(id) > 19 {
				id = id[:19]
			}
			rows[i] = table.Row{
				id,
				fmt.Sprintf("%d", p.Pid),
				p.Status.String(),
			}
		}
		return rows
	},
	ToDetail: func(data any, index int) (string, string) {
		procs, ok := data.([]*tasktypes.Process)
		if !ok || index < 0 || index >= len(procs) {
			return "", ""
		}
		p := procs[index]
		var b strings.Builder
		fmt.Fprintf(&b, "Container ID: %s\n", p.ID)
		fmt.Fprintf(&b, "PID:          %d\n", p.Pid)
		fmt.Fprintf(&b, "Status:       %s\n", p.Status)
		fmt.Fprintf(&b, "Terminal:     %t\n", p.Terminal)
		fmt.Fprintf(&b, "Exit Status:  %d\n", p.ExitStatus)
		if p.ExitedAt != nil {
			fmt.Fprintf(&b, "Exited At:    %s\n", p.ExitedAt.AsTime().Format("2006-01-02 15:04:05"))
		}
		return p.ID, b.String()
	},
}

package resource

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/henry118/nerdtui/ctr"
)

var TaskKind = Kind{
	Name: "Tasks",
	Columns: []Column{
		{Title: "Container ID", MinWidth: 12, Flex: true},
		{Title: "PID", MinWidth: 8},
		{Title: "Status", MinWidth: 12},
	},
	ToRows: func(data any, folded map[string]bool) []table.Row {
		tasks, ok := data.([]ctr.TaskInfo)
		if !ok || len(tasks) == 0 {
			return nil
		}
		rows := make([]table.Row, len(tasks))
		for i, t := range tasks {
			rows[i] = table.Row{
				t.Process.ID,
				fmt.Sprintf("%d", t.Process.Pid),
				t.Process.Status.String(),
			}
		}
		return rows
	},
	ToDetail: func(data any, folded map[string]bool, index int) (string, string) {
		tasks, ok := data.([]ctr.TaskInfo)
		if !ok || index < 0 || index >= len(tasks) {
			return "", ""
		}
		t := tasks[index]
		p := t.Process
		var b strings.Builder
		fmt.Fprintf(&b, "Container ID: %s\n", p.ID)
		fmt.Fprintf(&b, "PID:          %d\n", p.Pid)
		fmt.Fprintf(&b, "Status:       %s\n", p.Status)

		if p.Status == tasktypes.Status_STOPPED {
			fmt.Fprintf(&b, "Exit Status:  %d\n", p.ExitStatus)
			if p.ExitedAt != nil {
				fmt.Fprintf(&b, "Exited At:    %s\n", p.ExitedAt.AsTime().Format("2006-01-02 15:04:05"))
			}
		}

		if t.BundlePath != "" {
			fmt.Fprintf(&b, "Bundle:       %s\n", t.BundlePath)
		}

		if t.Spec != nil {
			fmt.Fprintf(&b, "\n--- Runtime Spec ---\n")
			if t.Spec.Root != nil {
				fmt.Fprintf(&b, "RootFS:       %s\n", t.Spec.Root.Path)
				fmt.Fprintf(&b, "Readonly:     %t\n", t.Spec.Root.Readonly)
			}
			if t.Spec.Process != nil {
				fmt.Fprintf(&b, "Cwd:          %s\n", t.Spec.Process.Cwd)
				if len(t.Spec.Process.Args) > 0 {
					fmt.Fprintf(&b, "Args:         %s\n", strings.Join(t.Spec.Process.Args, " "))
				}
				fmt.Fprintf(&b, "Terminal:     %t\n", t.Spec.Process.Terminal)
				if len(t.Spec.Process.Env) > 0 {
					fmt.Fprintf(&b, "Env:\n")
					for _, e := range t.Spec.Process.Env {
						fmt.Fprintf(&b, "  %s\n", e)
					}
				}
			}
			if t.Spec.Hostname != "" {
				fmt.Fprintf(&b, "Hostname:     %s\n", t.Spec.Hostname)
			}
			if len(t.Spec.Mounts) > 0 {
				fmt.Fprintf(&b, "Mounts:\n")
				for _, m := range t.Spec.Mounts {
					fmt.Fprintf(&b, "  %s -> %s (%s)\n", m.Source, m.Destination, m.Type)
				}
			}
		}
		return p.ID, b.String()
	},
}

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
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/henry118/nerdview/ctr"
)

// TaskContainerRef returns the container ID for the task row at the given index.
func TaskContainerRef(data any, _ map[string]bool, index int) string {
	tasks, ok := data.([]ctr.TaskInfo)
	if !ok || index < 0 || index >= len(tasks) {
		return ""
	}
	return tasks[index].ContainerID
}

var TaskKind = Kind{
	Name: "Tasks",
	Columns: []Column{
		{Title: "ID", MinWidth: 8, Flex: true},
		{Title: "Container", MinWidth: 8, Flex: true},
		{Title: "Type", MinWidth: 4},
		{Title: "PID", MinWidth: 5},
		{Title: "Status", MinWidth: 7},
		{Title: "Cmdline", MinWidth: 10, Flex: true},
		{Title: "Started", MinWidth: 19, Flex: true},
	},
	ToRows: func(data any, folded map[string]bool) []table.Row {
		tasks, ok := data.([]ctr.TaskInfo)
		if !ok || len(tasks) == 0 {
			return nil
		}
		rows := make([]table.Row, len(tasks))
		for i, t := range tasks {
			id := t.ContainerID
			typ := "init"
			if t.ExecID != "" {
				id = t.ExecID
				typ = "exec"
			}
			rows[i] = table.Row{
				id,
				t.ContainerID,
				typ,
				fmt.Sprintf("%d", t.Process.Pid),
				t.Process.Status.String(),
				t.Cmdline,
				t.StartedAt,
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

		typ := "init"
		if t.ExecID != "" {
			typ = "exec"
		}

		detail := fmt.Sprintf("ID:           %s\nContainer:    %s\nType:         %s\nPID:          %d\nStatus:       %s\n",
			taskID(t), t.ContainerID, typ, p.Pid, p.Status)

		if t.Cmdline != "" {
			detail += fmt.Sprintf("Cmdline:      %s\n", t.Cmdline)
		}

		if t.StartedAt != "" {
			detail += fmt.Sprintf("Started:      %s\n", t.StartedAt)
		}

		if p.Status == tasktypes.Status_STOPPED {
			detail += fmt.Sprintf("Exit Status:  %d\n", p.ExitStatus)
			if p.ExitedAt != nil {
				detail += fmt.Sprintf("Exited At:    %s\n", p.ExitedAt.AsTime().Format("2006-01-02 15:04:05"))
			}
		}

		if t.BundlePath != "" {
			detail += fmt.Sprintf("Bundle:       %s\n", t.BundlePath)
		}

		if namespaces := ctr.ProcessNamespaces(p.Pid); len(namespaces) > 0 {
			first := true
			for _, target := range namespaces {
				if first {
					detail += fmt.Sprintf("Namespaces:   %s\n", target)
					first = false
				} else {
					detail += fmt.Sprintf("              %s\n", target)
				}
			}
		}

		if t.Spec != nil {
			data, err := json.MarshalIndent(t.Spec, "", "  ")
			if err == nil {
				detail += "\n--- Runtime Spec ---\n" + string(data) + "\n"
			}
		}
		return taskID(t), detail
	},
}

func taskID(t ctr.TaskInfo) string {
	if t.ExecID != "" {
		return t.ExecID
	}
	return t.ContainerID
}

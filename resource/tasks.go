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
	return tasks[index].Process.ID
}

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
		detail := fmt.Sprintf("Container ID: %s\nPID:          %d\nStatus:       %s\n",
			p.ID, p.Pid, p.Status)

		if p.Status == tasktypes.Status_STOPPED {
			detail += fmt.Sprintf("Exit Status:  %d\n", p.ExitStatus)
			if p.ExitedAt != nil {
				detail += fmt.Sprintf("Exited At:    %s\n", p.ExitedAt.AsTime().Format("2006-01-02 15:04:05"))
			}
		}

		if t.BundlePath != "" {
			detail += fmt.Sprintf("Bundle:       %s\n", t.BundlePath)
		}

		if t.Spec != nil {
			data, err := json.MarshalIndent(t.Spec, "", "  ")
			if err == nil {
				detail += "\n--- Runtime Spec ---\n" + string(data) + "\n"
			}
		}
		return p.ID, detail
	},
}

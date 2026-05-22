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
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/henry118/nerdview/ctr"
)

var taskTreeSpec = TreeSpec[ctr.TaskInfo]{
	ID: func(t ctr.TaskInfo) string {
		return taskID(t)
	},
	ParentID: func(t ctr.TaskInfo) string {
		if t.ExecID != "" {
			return t.ContainerID
		}
		return ""
	},
	Foldable: func(t ctr.TaskInfo, hasChildren bool) bool {
		return t.ExecID == "" && hasChildren
	},
	Sort: func(a, b ctr.TaskInfo) bool {
		return taskID(a) < taskID(b)
	},
	Row: func(t ctr.TaskInfo, _ bool) table.Row {
		id := t.ContainerID
		typ := "init"
		if t.ExecID != "" {
			id = t.ExecID
			typ = "exec"
		}
		return table.Row{
			id,
			t.ContainerID,
			typ,
			fmt.Sprintf("%d", t.Process.Pid),
			t.Process.Status.String(),
			t.Cmdline,
			t.StartedAt,
		}
	},
}

var TaskKind = Kind{
	Name: "Tasks",
	Columns: []Column{
		{Title: "ID", MinWidth: 8, Flex: true},
		{Title: "CONTAINER", MinWidth: 8, Flex: true},
		{Title: "TYPE", MinWidth: 4},
		{Title: "PID", MinWidth: 5},
		{Title: "STATUS", MinWidth: 7},
		{Title: "CMDLINE", MinWidth: 10, Flex: true},
		{Title: "STARTED", MinWidth: 19, Flex: true},
	},
	Rows: func(data any, folded map[string]bool) ([]table.Row, any) {
		tasks, ok := data.([]ctr.TaskInfo)
		if !ok || len(tasks) == 0 {
			return nil, nil
		}
		result := BuildTree(taskTreeSpec, tasks, folded)
		return result.Rows, result
	},
	FoldKey: func(cache any, index int) string {
		result, ok := cache.(BuildResult[ctr.TaskInfo])
		if !ok || index < 0 || index >= len(result.Nodes) {
			return ""
		}
		node := result.Nodes[index]
		if node.HasChildren && node.Item.ExecID == "" {
			return node.ID
		}
		return ""
	},
	InitFolded: func(data any) map[string]bool {
		tasks, ok := data.([]ctr.TaskInfo)
		if !ok || len(tasks) == 0 {
			return nil
		}
		folded := make(map[string]bool)
		DefaultFoldState(taskTreeSpec, tasks, folded)
		return folded
	},
	Detail: func(cache any, index int) (string, string) {
		result, ok := cache.(BuildResult[ctr.TaskInfo])
		if !ok || index < 0 || index >= len(result.Nodes) {
			return "", ""
		}
		return formatTaskDetail(result.Nodes[index].Item)
	},
	CrossRefs: func(cache any) []string {
		result, ok := cache.(BuildResult[ctr.TaskInfo])
		if !ok || len(result.Nodes) == 0 {
			return nil
		}
		refs := make([]string, len(result.Nodes))
		for i, node := range result.Nodes {
			refs[i] = node.Item.ContainerID
		}
		return refs
	},
}

func formatTaskDetail(t ctr.TaskInfo) (string, string) {
	p := t.Process
	typ := "init"
	if t.ExecID != "" {
		typ = "exec"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "ID:           %s\n", taskID(t))
	fmt.Fprintf(&b, "Container:    %s\n", t.ContainerID)
	fmt.Fprintf(&b, "Type:         %s\n", typ)
	fmt.Fprintf(&b, "PID:          %d\n", p.Pid)
	fmt.Fprintf(&b, "Status:       %s\n", p.Status)

	if t.Cmdline != "" {
		fmt.Fprintf(&b, "Cmdline:      %s\n", t.Cmdline)
	}
	if t.StartedAt != "" {
		fmt.Fprintf(&b, "Started:      %s\n", t.StartedAt)
	}
	if p.Status == tasktypes.Status_STOPPED {
		fmt.Fprintf(&b, "Exit Status:  %d\n", p.ExitStatus)
		if p.ExitedAt != nil {
			fmt.Fprintf(&b, "Exited At:    %s\n", p.ExitedAt.AsTime().Format("2006-01-02 15:04:05"))
		}
	}
	if t.Root != "" {
		fmt.Fprintf(&b, "Root:         %s\n", t.Root)
	}
	if t.Cwd != "" {
		fmt.Fprintf(&b, "Cwd:          %s\n", t.Cwd)
	}
	if t.BundlePath != "" {
		fmt.Fprintf(&b, "Bundle:       %s\n", t.BundlePath)
	}
	if len(t.Cgroups) > 0 {
		for i, cg := range t.Cgroups {
			if i == 0 {
				fmt.Fprintf(&b, "Cgroup:       %s\n", cg)
			} else {
				fmt.Fprintf(&b, "              %s\n", cg)
			}
		}
	}
	if len(t.Namespaces) > 0 {
		first := true
		for _, target := range t.Namespaces {
			if first {
				fmt.Fprintf(&b, "Namespaces:   %s\n", target)
				first = false
			} else {
				fmt.Fprintf(&b, "              %s\n", target)
			}
		}
	}
	return taskID(t), b.String()
}

func taskID(t ctr.TaskInfo) string {
	if t.ExecID != "" {
		return t.ExecID
	}
	return t.ContainerID
}

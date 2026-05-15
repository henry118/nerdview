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

type taskKind struct{}

var TaskKind Kind = taskKind{}

func (taskKind) Name() string { return "Tasks" }

func (taskKind) Columns() []Column {
	return []Column{
		{Title: "ID", MinWidth: 8, Flex: true},
		{Title: "Container", MinWidth: 8, Flex: true},
		{Title: "Type", MinWidth: 4},
		{Title: "PID", MinWidth: 5},
		{Title: "Status", MinWidth: 7},
		{Title: "Cmdline", MinWidth: 10, Flex: true},
		{Title: "Started", MinWidth: 19, Flex: true},
	}
}

func (taskKind) Rows(data any, _ map[string]bool) []table.Row {
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
}

func (taskKind) FoldKey(_ any, _ map[string]bool, _ int) string { return "" }

func (taskKind) InitFolded(_ any) map[string]bool { return nil }

func (taskKind) Detail(data any, _ map[string]bool, index int) (string, string) {
	tasks, ok := data.([]ctr.TaskInfo)
	if !ok || index < 0 || index >= len(tasks) {
		return "", ""
	}
	return formatTaskDetail(tasks[index])
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
	if root := ctr.ProcessRoot(p.Pid); root != "" {
		fmt.Fprintf(&b, "Root:         %s\n", root)
	}
	if cwd := ctr.ProcessCwd(p.Pid); cwd != "" {
		fmt.Fprintf(&b, "Cwd:          %s\n", cwd)
	}
	if t.BundlePath != "" {
		fmt.Fprintf(&b, "Bundle:       %s\n", t.BundlePath)
	}
	if cgroups := ctr.ProcessCgroup(p.Pid); len(cgroups) > 0 {
		for i, cg := range cgroups {
			if i == 0 {
				fmt.Fprintf(&b, "Cgroup:       %s\n", cg)
			} else {
				fmt.Fprintf(&b, "              %s\n", cg)
			}
		}
	}
	if namespaces := ctr.ProcessNamespaces(p.Pid); len(namespaces) > 0 {
		first := true
		for _, target := range namespaces {
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

func (taskKind) CrossRefs(data any, _ map[string]bool) []string {
	tasks, ok := data.([]ctr.TaskInfo)
	if !ok {
		return nil
	}
	refs := make([]string, len(tasks))
	for i, t := range tasks {
		refs[i] = t.ContainerID
	}
	return refs
}

func taskID(t ctr.TaskInfo) string {
	if t.ExecID != "" {
		return t.ExecID
	}
	return t.ContainerID
}

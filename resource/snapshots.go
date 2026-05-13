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
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/containerd/containerd/v2/core/snapshots"
)

var SnapshotKind = Kind{
	Name: "Snapshots",
	Columns: []Column{
		{Title: "Name", MinWidth: 20, Flex: true},
		{Title: "Kind", MinWidth: 10},
		{Title: "Created", MinWidth: 20},
	},
	ToRows: func(data any, folded map[string]bool) []table.Row {
		infos, ok := data.([]snapshots.Info)
		if !ok || len(infos) == 0 {
			return nil
		}
		return buildSnapshotTree(infos, folded)
	},
	RowID: func(data any, folded map[string]bool, index int) string {
		infos, ok := data.([]snapshots.Info)
		if !ok || index < 0 {
			return ""
		}
		rows := buildSnapshotTree(infos, folded)
		if index >= len(rows) {
			return ""
		}
		name := stripTreePrefix(rows[index][0])
		// Only root snapshots (no parent) are foldable
		for _, info := range infos {
			if info.Name == name && info.Parent == "" {
				children := buildChildrenMap(infos)
				if len(children[name]) > 0 {
					return name
				}
			}
		}
		return ""
	},
	InitFolded: func(data any) map[string]bool {
		infos, ok := data.([]snapshots.Info)
		if !ok {
			return nil
		}
		children := buildChildrenMap(infos)
		folded := make(map[string]bool)
		// Only fold root snapshots
		for _, info := range infos {
			if info.Parent == "" && len(children[info.Name]) > 0 {
				folded[info.Name] = true
			}
		}
		return folded
	},
	ToDetail: func(data any, folded map[string]bool, index int) (string, string) {
		infos, ok := data.([]snapshots.Info)
		if !ok || index < 0 {
			return "", ""
		}
		rows := buildSnapshotTree(infos, folded)
		if index >= len(rows) {
			return "", ""
		}
		name := stripTreePrefix(rows[index][0])
		for _, info := range infos {
			if info.Name == name {
				return formatSnapshotDetail(info)
			}
		}
		return "", ""
	},
}

func buildChildrenMap(infos []snapshots.Info) map[string][]string {
	children := make(map[string][]string)
	for _, info := range infos {
		if info.Parent != "" {
			children[info.Parent] = append(children[info.Parent], info.Name)
		}
	}
	return children
}

func buildSnapshotTree(infos []snapshots.Info, folded map[string]bool) []table.Row {
	byName := make(map[string]snapshots.Info, len(infos))
	children := make(map[string][]string)
	var roots []string

	for _, info := range infos {
		byName[info.Name] = info
		if info.Parent == "" {
			roots = append(roots, info.Name)
		} else {
			children[info.Parent] = append(children[info.Parent], info.Name)
		}
	}

	sort.Strings(roots)
	for k := range children {
		sort.Strings(children[k])
	}

	var rows []table.Row
	for _, root := range roots {
		rows = appendTreeRows(rows, root, "", true, false, byName, children, folded)
	}
	return rows
}

func appendTreeRows(rows []table.Row, name, prefix string, isRoot bool, isLast bool, byName map[string]snapshots.Info, children map[string][]string, folded map[string]bool) []table.Row {
	info := byName[name]
	hasChildren := len(children[name]) > 0
	isFolded := isRoot && hasChildren && folded[name]

	var displayName string
	var childPrefix string

	foldIcon := ""
	if isRoot && hasChildren {
		if isFolded {
			foldIcon = "▸ "
		} else {
			foldIcon = "▾ "
		}
	}

	if isRoot {
		displayName = foldIcon + name
		childPrefix = ""
	} else {
		connector := "├─ "
		if isLast {
			connector = "└─ "
		}
		displayName = prefix + connector + name
		if isLast {
			childPrefix = prefix + "   "
		} else {
			childPrefix = prefix + "│  "
		}
	}

	rows = append(rows, table.Row{
		displayName,
		info.Kind.String(),
		info.Created.Format("2006-01-02 15:04:05"),
	})

	if isFolded {
		return rows
	}

	kids := children[name]
	for i, child := range kids {
		childIsLast := i == len(kids)-1
		rows = appendTreeRows(rows, child, childPrefix, false, childIsLast, byName, children, folded)
	}
	return rows
}

func stripTreePrefix(s string) string {
	prefixes := []string{"▸ ", "▾ ", "├─ ", "└─ ", "│  ", "   "}
	for {
		matched := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(s, prefix) {
				s = s[len(prefix):]
				matched = true
			}
		}
		if !matched {
			break
		}
	}
	return s
}

func formatSnapshotDetail(info snapshots.Info) (string, string) {
	var b strings.Builder
	fmt.Fprintf(&b, "Name:    %s\n", info.Name)
	fmt.Fprintf(&b, "Kind:    %s\n", info.Kind)
	fmt.Fprintf(&b, "Parent:  %s\n", info.Parent)
	fmt.Fprintf(&b, "Created: %s\n", info.Created.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "Updated: %s\n", info.Updated.Format("2006-01-02 15:04:05"))
	if len(info.Labels) > 0 {
		fmt.Fprintf(&b, "Labels:\n")
		for k, v := range info.Labels {
			fmt.Fprintf(&b, "  %s: %s\n", k, v)
		}
	}
	return info.Name, b.String()
}

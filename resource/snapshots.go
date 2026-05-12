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
	ToRows: func(data any) []table.Row {
		infos, ok := data.([]snapshots.Info)
		if !ok || len(infos) == 0 {
			return nil
		}
		return buildSnapshotTree(infos)
	},
	ToDetail: func(data any, index int) (string, string) {
		infos, ok := data.([]snapshots.Info)
		if !ok || index < 0 || index >= len(infos) {
			return "", ""
		}
		// Find the info that corresponds to the tree row at this index
		rows := buildSnapshotTree(infos)
		if index >= len(rows) {
			return "", ""
		}
		name := rows[index][0]
		// Strip tree prefixes to get actual name
		actualName := stripTreePrefix(name)
		for _, info := range infos {
			if info.Name == actualName {
				return formatSnapshotDetail(info)
			}
		}
		return "", ""
	},
}

func buildSnapshotTree(infos []snapshots.Info) []table.Row {
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
		rows = appendTreeRows(rows, root, "", true, false, byName, children)
	}
	return rows
}

func appendTreeRows(rows []table.Row, name, prefix string, isRoot bool, isLast bool, byName map[string]snapshots.Info, children map[string][]string) []table.Row {
	info := byName[name]

	var displayName string
	var childPrefix string
	if isRoot {
		displayName = name
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

	kids := children[name]
	for i, child := range kids {
		childIsLast := i == len(kids)-1
		rows = appendTreeRows(rows, child, childPrefix, false, childIsLast, byName, children)
	}
	return rows
}

func stripTreePrefix(s string) string {
	// Remove tree drawing characters to get the actual snapshot name
	for _, prefix := range []string{"├─ ", "└─ ", "│  ", "   "} {
		for strings.HasPrefix(s, prefix) {
			s = s[len(prefix):]
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

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
	"github.com/containerd/containerd/v2/core/snapshots"
)

var snapshotTreeSpec = TreeSpec[snapshots.Info]{
	ID:       func(info snapshots.Info) string { return info.Name },
	ParentID: func(info snapshots.Info) string { return info.Parent },
	Foldable: func(info snapshots.Info, hasChildren bool) bool {
		return hasChildren && info.Parent == ""
	},
	Sort: func(a, b snapshots.Info) bool { return a.Name < b.Name },
	Row: func(info snapshots.Info, _ bool) table.Row {
		return table.Row{
			ShortDigest(info.Name),
			info.Kind.String(),
			info.Created.Format("2006-01-02 15:04:05"),
		}
	},
}

type snapshotKind struct{}

var SnapshotKind Kind = snapshotKind{}

func (snapshotKind) Name() string { return "Snapshots" }

func (snapshotKind) Columns() []Column {
	return []Column{
		{Title: "NAME", MinWidth: 20, Flex: true},
		{Title: "KIND", MinWidth: 10},
		{Title: "CREATED", MinWidth: 20},
	}
}

func (snapshotKind) Rows(data any, folded map[string]bool) []table.Row {
	infos, ok := data.([]snapshots.Info)
	if !ok || len(infos) == 0 {
		return nil
	}
	return BuildTree(snapshotTreeSpec, infos, folded).Rows
}

func (snapshotKind) FoldKey(data any, folded map[string]bool, index int) string {
	infos, ok := data.([]snapshots.Info)
	if !ok || index < 0 {
		return ""
	}
	result := BuildTree(snapshotTreeSpec, infos, folded)
	if index >= len(result.Nodes) {
		return ""
	}
	node := result.Nodes[index]
	if node.HasChildren && nodeByID(infos, node.ID).Parent == "" {
		return node.ID
	}
	return ""
}

func (snapshotKind) InitFolded(data any) map[string]bool {
	infos, ok := data.([]snapshots.Info)
	if !ok {
		return nil
	}
	folded := make(map[string]bool)
	DefaultFoldState(snapshotTreeSpec, infos, folded)
	return folded
}

func (snapshotKind) Detail(data any, folded map[string]bool, index int) (string, string) {
	infos, ok := data.([]snapshots.Info)
	if !ok || index < 0 {
		return "", ""
	}
	result := BuildTree(snapshotTreeSpec, infos, folded)
	if index >= len(result.Nodes) {
		return "", ""
	}
	return formatSnapshotDetail(result.Nodes[index].Item)
}

func (snapshotKind) CrossRefs(_ any, _ map[string]bool) []string {
	return nil
}

func nodeByID(infos []snapshots.Info, id string) snapshots.Info {
	for _, info := range infos {
		if info.Name == id {
			return info
		}
	}
	return snapshots.Info{}
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

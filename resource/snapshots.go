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

	"charm.land/bubbles/v2/table"
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

type snapshotCache struct {
	result BuildResult[snapshots.Info]
	infos  []snapshots.Info
}

var SnapshotKind = Kind{
	Name: "Snapshots",
	Columns: []Column{
		{Title: "NAME", MinWidth: 20, Flex: true},
		{Title: "KIND", MinWidth: 10},
		{Title: "CREATED", MinWidth: 20},
	},
	Rows: func(data any, folded map[string]bool) ([]table.Row, any) {
		infos, ok := data.([]snapshots.Info)
		if !ok || len(infos) == 0 {
			return nil, nil
		}
		result := BuildTree(snapshotTreeSpec, infos, folded)
		return result.Rows, snapshotCache{result: result, infos: infos}
	},
	FoldKey: func(cache any, index int) string {
		sc, ok := cache.(snapshotCache)
		if !ok || index < 0 || index >= len(sc.result.Nodes) {
			return ""
		}
		node := sc.result.Nodes[index]
		if node.HasChildren && nodeByID(sc.infos, node.ID).Parent == "" {
			return node.ID
		}
		return ""
	},
	InitFolded: func(data any) map[string]bool {
		infos, ok := data.([]snapshots.Info)
		if !ok {
			return nil
		}
		folded := make(map[string]bool)
		DefaultFoldState(snapshotTreeSpec, infos, folded)
		return folded
	},
	Detail: func(cache any, index int) (string, string) {
		sc, ok := cache.(snapshotCache)
		if !ok || index < 0 || index >= len(sc.result.Nodes) {
			return "", ""
		}
		return formatSnapshotDetail(sc.result.Nodes[index].Item)
	},
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
		keys := make([]string, 0, len(info.Labels))
		for k := range info.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "  %s: %s\n", k, info.Labels[k])
		}
	}
	return info.Name, b.String()
}

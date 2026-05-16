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
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/henry118/nerdview/ctr"
)

var containerTreeSpec = TreeSpec[ctr.ContainerInfo]{
	ID: func(info ctr.ContainerInfo) string { return info.Container.ID },
	ParentID: func(info ctr.ContainerInfo) string {
		c := info.Container
		if !info.IsSandbox && c.SandboxID != "" && c.SandboxID != c.ID {
			return c.SandboxID
		}
		return ""
	},
	Foldable: func(info ctr.ContainerInfo, hasChildren bool) bool {
		return info.IsSandbox && hasChildren
	},
	Sort: func(a, b ctr.ContainerInfo) bool { return a.Container.ID < b.Container.ID },
	Row: func(info ctr.ContainerInfo, _ bool) table.Row {
		c := info.Container
		typ := "container"
		if info.IsSandbox {
			typ = "sandbox"
		}
		return table.Row{
			c.ID,
			typ,
			c.Image,
			c.Runtime.Name,
			c.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	},
}

var ContainerKind = Kind{
	Name: "Containers",
	Columns: []Column{
		{Title: "ID", MinWidth: 12, Flex: true},
		{Title: "TYPE", MinWidth: 8},
		{Title: "IMAGE", MinWidth: 20, Flex: true},
		{Title: "RUNTIME", MinWidth: 16},
		{Title: "CREATED", MinWidth: 20},
	},
	Rows: func(data any, folded map[string]bool) ([]table.Row, any) {
		infos, ok := data.([]ctr.ContainerInfo)
		if !ok || len(infos) == 0 {
			return nil, nil
		}
		result := BuildTree(containerTreeSpec, infos, folded)
		return result.Rows, result
	},
	FoldKey: func(cache any, index int) string {
		result, ok := cache.(BuildResult[ctr.ContainerInfo])
		if !ok || index < 0 || index >= len(result.Nodes) {
			return ""
		}
		node := result.Nodes[index]
		if node.HasChildren && node.Item.IsSandbox {
			return node.ID
		}
		return ""
	},
	Detail: func(cache any, index int) (string, string) {
		result, ok := cache.(BuildResult[ctr.ContainerInfo])
		if !ok || index < 0 || index >= len(result.Nodes) {
			return "", ""
		}
		return formatContainerDetail(result.Nodes[index].Item)
	},
	CrossRefs: func(cache any) []string {
		result, ok := cache.(BuildResult[ctr.ContainerInfo])
		if !ok || len(result.Nodes) == 0 {
			return nil
		}
		refs := make([]string, len(result.Nodes))
		for i, node := range result.Nodes {
			refs[i] = node.Item.Container.SnapshotKey
		}
		return refs
	},
	Spec: func(cache any, index int) (string, string) {
		result, ok := cache.(BuildResult[ctr.ContainerInfo])
		if !ok || index < 0 || index >= len(result.Nodes) {
			return "", ""
		}
		item := result.Nodes[index].Item
		if item.Spec == nil {
			return "", ""
		}
		specJSON, err := json.MarshalIndent(item.Spec, "", "  ")
		if err != nil {
			return "", ""
		}
		return item.Container.ID, string(specJSON)
	},
}

func formatContainerDetail(info ctr.ContainerInfo) (string, string) {
	c := info.Container
	var b strings.Builder
	fmt.Fprintf(&b, "ID:          %s\n", c.ID)
	if info.IsSandbox {
		fmt.Fprintf(&b, "Type:        sandbox\n")
	} else {
		fmt.Fprintf(&b, "Type:        container\n")
	}
	fmt.Fprintf(&b, "Image:       %s\n", c.Image)
	fmt.Fprintf(&b, "Runtime:     %s\n", c.Runtime.Name)
	fmt.Fprintf(&b, "Snapshotter: %s\n", c.Snapshotter)
	fmt.Fprintf(&b, "SnapshotKey: %s\n", c.SnapshotKey)
	fmt.Fprintf(&b, "Created:     %s\n", c.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "Updated:     %s\n", c.UpdatedAt.Format("2006-01-02 15:04:05"))
	if c.SandboxID != "" {
		fmt.Fprintf(&b, "SandboxID:   %s\n", c.SandboxID)
	}
	if len(c.Labels) > 0 {
		fmt.Fprintf(&b, "Labels:\n")
		for k, v := range c.Labels {
			fmt.Fprintf(&b, "  %s: %s\n", k, v)
		}
	}
	return c.ID, b.String()
}

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
	"github.com/henry118/nerdview/ctr"
)

var ContainerKind = Kind{
	Name: "Containers",
	Columns: []Column{
		{Title: "ID", MinWidth: 12, Flex: true},
		{Title: "Type", MinWidth: 8},
		{Title: "Image", MinWidth: 20, Flex: true},
		{Title: "Runtime", MinWidth: 16},
		{Title: "Created", MinWidth: 20},
	},
	ToRows: func(data any, folded map[string]bool) []table.Row {
		infos, ok := data.([]ctr.ContainerInfo)
		if !ok || len(infos) == 0 {
			return nil
		}
		return buildContainerTree(infos, folded)
	},
	RowID: func(data any, folded map[string]bool, index int) string {
		infos, ok := data.([]ctr.ContainerInfo)
		if !ok || index < 0 {
			return ""
		}
		rows := buildContainerTree(infos, folded)
		if index >= len(rows) {
			return ""
		}
		id := stripContainerPrefix(rows[index][0])
		// Only sandboxes with children are foldable
		children := buildSandboxChildren(infos)
		if len(children[id]) > 0 {
			return id
		}
		return ""
	},
	ToDetail: func(data any, folded map[string]bool, index int) (string, string) {
		infos, ok := data.([]ctr.ContainerInfo)
		if !ok || index < 0 {
			return "", ""
		}
		rows := buildContainerTree(infos, folded)
		if index >= len(rows) {
			return "", ""
		}
		id := stripContainerPrefix(rows[index][0])
		for _, info := range infos {
			if info.Container.ID == id {
				return formatContainerDetail(info)
			}
		}
		return "", ""
	},
}

func buildSandboxChildren(infos []ctr.ContainerInfo) map[string][]string {
	children := make(map[string][]string)
	for _, info := range infos {
		c := info.Container
		// A non-sandbox container that has a SandboxID is a child of that sandbox
		if !info.IsSandbox && c.SandboxID != "" && c.SandboxID != c.ID {
			children[c.SandboxID] = append(children[c.SandboxID], c.ID)
		}
	}
	for k := range children {
		sort.Strings(children[k])
	}
	return children
}

func buildContainerTree(infos []ctr.ContainerInfo, folded map[string]bool) []table.Row {
	byID := make(map[string]ctr.ContainerInfo, len(infos))
	for _, info := range infos {
		byID[info.Container.ID] = info
	}

	sandboxChildren := buildSandboxChildren(infos)

	// Categorize
	var sandboxIDs []string
	var standalone []string
	for _, info := range infos {
		c := info.Container
		if info.IsSandbox {
			sandboxIDs = append(sandboxIDs, c.ID)
		} else if c.SandboxID == "" || c.SandboxID == c.ID {
			standalone = append(standalone, c.ID)
		}
		// containers with SandboxID pointing elsewhere are children, rendered under their sandbox
	}
	sort.Strings(sandboxIDs)
	sort.Strings(standalone)

	var rows []table.Row

	// Render sandboxes with their children
	for _, id := range sandboxIDs {
		info := byID[id]
		c := info.Container
		hasChildren := len(sandboxChildren[id]) > 0
		isFolded := hasChildren && folded[id]

		foldIcon := ""
		if hasChildren {
			if isFolded {
				foldIcon = "▸ "
			} else {
				foldIcon = "▾ "
			}
		}

		rows = append(rows, table.Row{
			foldIcon + c.ID,
			"sandbox",
			c.Image,
			c.Runtime.Name,
			c.CreatedAt.Format("2006-01-02 15:04:05"),
		})

		if isFolded {
			continue
		}

		kids := sandboxChildren[id]
		for i, childID := range kids {
			child := byID[childID]
			connector := "├─ "
			if i == len(kids)-1 {
				connector = "└─ "
			}
			rows = append(rows, table.Row{
				connector + child.Container.ID,
				"container",
				child.Container.Image,
				child.Container.Runtime.Name,
				child.Container.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	// Render standalone containers
	for _, id := range standalone {
		info := byID[id]
		c := info.Container
		rows = append(rows, table.Row{
			c.ID,
			"container",
			c.Image,
			c.Runtime.Name,
			c.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return rows
}

func stripContainerPrefix(s string) string {
	prefixes := []string{"▸ ", "▾ ", "├─ ", "└─ "}
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

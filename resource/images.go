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
	"github.com/henry118/nerdview/ctr"
)

var ImageKind = Kind{
	Name: "Images",
	Columns: []Column{
		{Title: "Name", MinWidth: 20, Flex: true},
		{Title: "Type", MinWidth: 12},
		{Title: "Digest", MinWidth: 20},
		{Title: "Layers", MinWidth: 6},
		{Title: "Size", MinWidth: 10},
	},
	ToRows: func(data any, folded map[string]bool) []table.Row {
		trees, ok := data.([]ctr.ImageTree)
		if !ok || len(trees) == 0 {
			return nil
		}
		var rows []table.Row
		for _, tree := range trees {
			rows = appendImageRows(rows, tree, folded)
		}
		return rows
	},
	RowID: func(data any, folded map[string]bool, index int) string {
		trees, ok := data.([]ctr.ImageTree)
		if !ok || index < 0 {
			return ""
		}
		nodes := flattenVisibleNodes(trees, folded)
		if index >= len(nodes) {
			return ""
		}
		n := nodes[index]
		if n.hasChildren {
			return n.node.Desc.Digest.String()
		}
		return ""
	},
	InitFolded: func(data any) map[string]bool {
		trees, ok := data.([]ctr.ImageTree)
		if !ok {
			return nil
		}
		folded := make(map[string]bool)
		for _, tree := range trees {
			collectFoldable(folded, tree)
		}
		return folded
	},
	ToDetail: func(data any, folded map[string]bool, index int) (string, string) {
		trees, ok := data.([]ctr.ImageTree)
		if !ok || index < 0 {
			return "", ""
		}
		nodes := flattenVisibleNodes(trees, folded)
		if index >= len(nodes) {
			return "", ""
		}
		return formatImageDetail(nodes[index].node)
	},
}

type visibleNode struct {
	node        ctr.ImageTree
	depth       int
	hasChildren bool
}

// ImageSnapshotRef returns the snapshot key for the image row at the given index.
// Only returns a key for parent nodes (manifests/indices), not leaf nodes (configs/layers).
func ImageSnapshotRef(data any, folded map[string]bool, index int) string {
	trees, ok := data.([]ctr.ImageTree)
	if !ok || index < 0 {
		return ""
	}
	nodes := flattenVisibleNodes(trees, folded)
	if index >= len(nodes) {
		return ""
	}
	n := nodes[index]
	if !n.hasChildren {
		return ""
	}
	return n.node.SnapshotKey
}

func collectFoldable(folded map[string]bool, node ctr.ImageTree) {
	if len(node.Children) > 0 {
		folded[node.Desc.Digest.String()] = true
		for _, child := range node.Children {
			collectFoldable(folded, child)
		}
	}
}

func flattenVisibleNodes(trees []ctr.ImageTree, folded map[string]bool) []visibleNode {
	var nodes []visibleNode
	for _, tree := range trees {
		nodes = flattenTree(nodes, tree, 0, folded)
	}
	return nodes
}

func flattenTree(nodes []visibleNode, node ctr.ImageTree, depth int, folded map[string]bool) []visibleNode {
	hasChildren := len(node.Children) > 0
	nodes = append(nodes, visibleNode{node: node, depth: depth, hasChildren: hasChildren})
	if hasChildren && folded != nil && folded[node.Desc.Digest.String()] {
		return nodes
	}
	for _, child := range node.Children {
		if child.Name == "" {
			child.Name = descLabel(child)
		}
		nodes = flattenTree(nodes, child, depth+1, folded)
	}
	return nodes
}

func appendImageRows(rows []table.Row, tree ctr.ImageTree, folded map[string]bool) []table.Row {
	type entry struct {
		node   ctr.ImageTree
		prefix string
		isLast bool
		isRoot bool
	}

	// BFS-like iteration using a stack to maintain order
	var stack []entry
	stack = append(stack, entry{node: tree, isRoot: true})

	for len(stack) > 0 {
		e := stack[0]
		stack = stack[1:]

		node := e.node
		isFolded := folded[node.Desc.Digest.String()] && len(node.Children) > 0

		var displayName string
		if e.isRoot {
			foldIcon := ""
			if len(node.Children) > 0 {
				if isFolded {
					foldIcon = IconFolded
				} else {
					foldIcon = IconUnfolded
				}
			}
			displayName = foldIcon + node.Name
		} else {
			connector := ConnMid
			if e.isLast {
				connector = ConnLast
			}
			foldIcon := ""
			if len(node.Children) > 0 {
				if isFolded {
					foldIcon = IconFolded
				} else {
					foldIcon = IconUnfolded
				}
			}
			displayName = e.prefix + connector + foldIcon + node.Name
		}

		digest := ShortDigest(node.Desc.Digest.String())
		size := node.Desc.Size
		layers := ""
		if len(node.Children) > 0 {
			size = totalSize(node)
			layers = fmt.Sprintf("%d", countLayers(node))
		}
		rows = append(rows, table.Row{
			displayName,
			shortMediaType(node.Desc.MediaType),
			digest,
			layers,
			formatSize(size),
		})

		if isFolded {
			continue
		}

		// Queue children
		var childPrefix string
		if e.isRoot {
			childPrefix = ""
		} else if e.isLast {
			childPrefix = e.prefix + ConnBlank
		} else {
			childPrefix = e.prefix + ConnPipe
		}

		// Insert children at the front of stack to maintain DFS order
		var childEntries []entry
		for i, child := range node.Children {
			if child.Name == "" {
				child.Name = descLabel(child)
			}
			childEntries = append(childEntries, entry{
				node:   child,
				prefix: childPrefix,
				isLast: i == len(node.Children)-1,
			})
		}
		stack = append(childEntries, stack...)
	}
	return rows
}


func countLayers(node ctr.ImageTree) int {
	count := 0
	for _, child := range node.Children {
		mt := shortMediaType(child.Desc.MediaType)
		if strings.HasPrefix(mt, "layer") {
			count++
		} else {
			count += countLayers(child)
		}
	}
	return count
}

func totalSize(node ctr.ImageTree) int64 {
	size := node.Desc.Size
	for _, child := range node.Children {
		size += totalSize(child)
	}
	return size
}

func descLabel(node ctr.ImageTree) string {
	if node.Desc.Platform != nil {
		p := node.Desc.Platform
		label := p.OS + "/" + p.Architecture
		if p.Variant != "" {
			label += "/" + p.Variant
		}
		return label
	}
	return shortMediaType(node.Desc.MediaType)
}

func shortMediaType(mt string) string {
	switch mt {
	case "application/vnd.oci.image.index.v1+json":
		return "index"
	case "application/vnd.oci.image.manifest.v1+json":
		return "manifest"
	case "application/vnd.oci.image.config.v1+json":
		return "config"
	case "application/vnd.oci.image.layer.v1.tar+gzip":
		return "layer/gzip"
	case "application/vnd.oci.image.layer.v1.tar+zstd":
		return "layer/zstd"
	case "application/vnd.oci.image.layer.v1.tar":
		return "layer"
	case "application/vnd.docker.distribution.manifest.v2+json":
		return "manifest"
	case "application/vnd.docker.distribution.manifest.list.v2+json":
		return "index"
	case "application/vnd.docker.image.rootfs.diff.tar.gzip":
		return "layer/gzip"
	case "application/vnd.docker.container.image.v1+json":
		return "config"
	default:
		if idx := strings.LastIndex(mt, "."); idx >= 0 {
			return mt[idx+1:]
		}
		return mt
	}
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func formatImageDetail(node ctr.ImageTree) (string, string) {
	var b strings.Builder
	if node.Name != "" {
		fmt.Fprintf(&b, "Name:       %s\n", node.Name)
	}
	fmt.Fprintf(&b, "MediaType:  %s\n", node.Desc.MediaType)
	fmt.Fprintf(&b, "Digest:     %s\n", node.Desc.Digest)
	fmt.Fprintf(&b, "Size:       %d\n", node.Desc.Size)
	if node.Desc.Platform != nil {
		p := node.Desc.Platform
		fmt.Fprintf(&b, "Platform:   %s/%s", p.OS, p.Architecture)
		if p.Variant != "" {
			fmt.Fprintf(&b, "/%s", p.Variant)
		}
		fmt.Fprintf(&b, "\n")
	}
	if len(node.Desc.Annotations) > 0 {
		fmt.Fprintf(&b, "Annotations:\n")
		for k, v := range node.Desc.Annotations {
			fmt.Fprintf(&b, "  %s: %s\n", k, v)
		}
	}
	if len(node.Labels) > 0 {
		fmt.Fprintf(&b, "Labels:\n")
		for k, v := range node.Labels {
			fmt.Fprintf(&b, "  %s: %s\n", k, v)
		}
	}
	title := node.Name
	if title == "" {
		title = node.Desc.Digest.String()
	}
	return title, b.String()
}

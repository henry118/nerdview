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

var imageTreeSpec = TreeSpec[ctr.ImageTree]{
	ID: func(node ctr.ImageTree) string { return node.Desc.Digest.String() },
	Children: func(node ctr.ImageTree) []ctr.ImageTree {
		children := make([]ctr.ImageTree, len(node.Children))
		for i, child := range node.Children {
			if child.Name == "" {
				child.Name = descLabel(child)
			}
			children[i] = child
		}
		return children
	},
	ToRow: func(node ctr.ImageTree, hasChildren bool) table.Row {
		size := node.Desc.Size
		layers := ""
		if hasChildren {
			size = totalSize(node)
			layers = fmt.Sprintf("%d", countLayers(node))
		}
		return table.Row{
			node.Name,
			shortMediaType(node.Desc.MediaType),
			ShortDigest(node.Desc.Digest.String()),
			layers,
			FormatBytes(uint64(size)),
		}
	},
}

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
		return BuildTree(imageTreeSpec, trees, folded).Rows
	},
	RowID: func(data any, folded map[string]bool, index int) string {
		trees, ok := data.([]ctr.ImageTree)
		if !ok || index < 0 {
			return ""
		}
		result := BuildTree(imageTreeSpec, trees, folded)
		if index >= len(result.Nodes) {
			return ""
		}
		node := result.Nodes[index]
		if node.HasChildren {
			return node.ID
		}
		return ""
	},
	InitFolded: func(data any) map[string]bool {
		trees, ok := data.([]ctr.ImageTree)
		if !ok {
			return nil
		}
		folded := make(map[string]bool)
		DefaultFoldState(imageTreeSpec, trees, folded)
		return folded
	},
	ToDetail: func(data any, folded map[string]bool, index int) (string, string) {
		trees, ok := data.([]ctr.ImageTree)
		if !ok || index < 0 {
			return "", ""
		}
		result := BuildTree(imageTreeSpec, trees, folded)
		if index >= len(result.Nodes) {
			return "", ""
		}
		return formatImageDetail(result.Nodes[index].Item)
	},
	GoToRef: func(data any, folded map[string]bool, index int) string {
		trees, ok := data.([]ctr.ImageTree)
		if !ok || index < 0 {
			return ""
		}
		result := BuildTree(imageTreeSpec, trees, folded)
		if index >= len(result.Nodes) {
			return ""
		}
		node := result.Nodes[index]
		if !node.HasChildren {
			return ""
		}
		return node.Item.SnapshotKey
	},
}

// ImageSnapshotRef returns the snapshot key for the image row at the given index.
func ImageSnapshotRef(data any, folded map[string]bool, index int) string {
	trees, ok := data.([]ctr.ImageTree)
	if !ok || index < 0 {
		return ""
	}
	result := BuildTree(imageTreeSpec, trees, folded)
	if index >= len(result.Nodes) {
		return ""
	}
	node := result.Nodes[index]
	if !node.HasChildren {
		return ""
	}
	return node.Item.SnapshotKey
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

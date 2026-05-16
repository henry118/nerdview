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
	ctdimages "github.com/containerd/containerd/v2/core/images"
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
	Row: func(node ctr.ImageTree, hasChildren bool) table.Row {
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
		{Title: "NAME", MinWidth: 20, Flex: true},
		{Title: "TYPE", MinWidth: 12},
		{Title: "DIGEST", MinWidth: 20},
		{Title: "LAYERS", MinWidth: 6},
		{Title: "SIZE", MinWidth: 10},
	},
	Rows: func(data any, folded map[string]bool) ([]table.Row, any) {
		trees := toSortedImages(data)
		if trees == nil {
			return nil, nil
		}
		result := BuildTree(imageTreeSpec, trees, folded)
		return result.Rows, result
	},
	FoldKey: func(cache any, index int) string {
		result, ok := cache.(BuildResult[ctr.ImageTree])
		if !ok || index < 0 || index >= len(result.Nodes) {
			return ""
		}
		node := result.Nodes[index]
		if node.HasChildren {
			return node.ID
		}
		return ""
	},
	InitFolded: func(data any) map[string]bool {
		trees := toSortedImages(data)
		if trees == nil {
			return nil
		}
		folded := make(map[string]bool)
		DefaultFoldState(imageTreeSpec, trees, folded)
		return folded
	},
	Detail: func(cache any, index int) (string, string) {
		result, ok := cache.(BuildResult[ctr.ImageTree])
		if !ok || index < 0 || index >= len(result.Nodes) {
			return "", ""
		}
		return formatImageDetail(result.Nodes[index].Item)
	},
	CrossRefs: func(cache any) []string {
		result, ok := cache.(BuildResult[ctr.ImageTree])
		if !ok || len(result.Nodes) == 0 {
			return nil
		}
		refs := make([]string, len(result.Nodes))
		for i, node := range result.Nodes {
			if node.HasChildren {
				refs[i] = node.Item.SnapshotKey
			}
		}
		return refs
	},
}

func toSortedImages(data any) []ctr.ImageTree {
	trees, ok := data.([]ctr.ImageTree)
	if !ok || len(trees) == 0 {
		return nil
	}
	sorted := make([]ctr.ImageTree, len(trees))
	copy(sorted, trees)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Desc.Digest.String() < sorted[j].Desc.Digest.String()
	})
	return sorted
}

func countLayers(node ctr.ImageTree) int {
	count := 0
	for _, child := range node.Children {
		if ctdimages.IsLayerType(child.Desc.MediaType) {
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
	switch {
	case ctdimages.IsIndexType(mt):
		return "index"
	case ctdimages.IsManifestType(mt):
		return "manifest"
	case ctdimages.IsConfigType(mt):
		return "config"
	case ctdimages.IsLayerType(mt):
		if i := strings.LastIndexByte(mt, '+'); i >= 0 {
			return "layer/" + mt[i+1:]
		}
		if i := strings.LastIndex(mt, ".tar."); i >= 0 {
			return "layer/" + mt[i+5:]
		}
		return "layer"
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

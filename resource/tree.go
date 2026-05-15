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
	"sort"

	"github.com/charmbracelet/bubbles/table"
)

// Tree display characters for hierarchical views.
const (
	IconFolded   = "▸ "
	IconUnfolded = "▾ "
	ConnMid      = "├─ "
	ConnLast     = "└─ "
	ConnPipe     = "│  "
	ConnBlank    = "   "
)

// TreeNode represents one visible node in a rendered tree.
type TreeNode[T any] struct {
	Item        T
	ID          string
	HasChildren bool
}

// TreeSpec defines how to derive tree structure from items.
type TreeSpec[T any] struct {
	// ID returns a unique identifier for the item.
	ID func(item T) string

	// ParentID returns the parent's ID, or "" if root.
	// Ignored if Children is set.
	ParentID func(item T) string

	// Children returns child items directly. When set, ParentID is ignored.
	Children func(item T) []T

	// Foldable returns whether this node supports fold/unfold.
	// If nil, any node with children is foldable.
	Foldable func(item T, hasChildren bool) bool

	// Sort orders siblings. If nil, insertion order is preserved.
	Sort func(a, b T) bool

	// ToRow converts an item to column values. Column 0 gets the tree prefix prepended.
	ToRow func(item T, hasChildren bool) table.Row
}

// BuildResult holds the output of tree building.
type BuildResult[T any] struct {
	Rows  []table.Row
	Nodes []TreeNode[T]
}

// BuildTree constructs visible tree rows from items, respecting fold state.
func BuildTree[T any](spec TreeSpec[T], items []T, folded map[string]bool) BuildResult[T] {
	if len(items) == 0 {
		return BuildResult[T]{}
	}

	var result BuildResult[T]

	if spec.Children != nil {
		for _, item := range items {
			result = buildChildrenMode(spec, item, "", true, false, folded, result)
		}
	} else {
		byID := make(map[string]T, len(items))
		children := make(map[string][]T)
		var roots []T

		for _, item := range items {
			id := spec.ID(item)
			byID[id] = item
			parentID := spec.ParentID(item)
			if parentID == "" {
				roots = append(roots, item)
			} else {
				children[parentID] = append(children[parentID], item)
			}
		}

		if spec.Sort != nil {
			sort.Slice(roots, func(i, j int) bool { return spec.Sort(roots[i], roots[j]) })
			for k := range children {
				kids := children[k]
				sort.Slice(kids, func(i, j int) bool { return spec.Sort(kids[i], kids[j]) })
			}
		}

		for i, root := range roots {
			isLast := i == len(roots)-1
			result = buildParentIDMode(spec, root, "", true, isLast, children, folded, result)
		}
	}

	return result
}

// DefaultFoldState returns a fold map with all foldable nodes marked as folded.
func DefaultFoldState[T any](spec TreeSpec[T], items []T, folded map[string]bool) {
	if spec.Children != nil {
		defaultFoldChildren(spec, items, folded)
	} else {
		children := make(map[string][]T)
		for _, item := range items {
			parentID := spec.ParentID(item)
			if parentID != "" {
				children[parentID] = append(children[parentID], item)
			}
		}
		for _, item := range items {
			id := spec.ID(item)
			hasChildren := len(children[id]) > 0
			if isFoldable(spec, item, hasChildren) {
				folded[id] = true
			}
		}
	}
}

func defaultFoldChildren[T any](spec TreeSpec[T], items []T, folded map[string]bool) {
	for _, item := range items {
		children := spec.Children(item)
		hasChildren := len(children) > 0
		if isFoldable(spec, item, hasChildren) {
			folded[spec.ID(item)] = true
		}
		if hasChildren {
			defaultFoldChildren(spec, children, folded)
		}
	}
}

func isFoldable[T any](spec TreeSpec[T], item T, hasChildren bool) bool {
	if spec.Foldable != nil {
		return spec.Foldable(item, hasChildren)
	}
	return hasChildren
}

func buildParentIDMode[T any](spec TreeSpec[T], item T, prefix string, isRoot, isLast bool, children map[string][]T, folded map[string]bool, result BuildResult[T]) BuildResult[T] {
	id := spec.ID(item)
	kids := children[id]
	hasChildren := len(kids) > 0
	foldable := isFoldable(spec, item, hasChildren)
	isFolded := foldable && folded[id]

	displayPrefix, childPrefix := renderPrefixes(isRoot, isLast, foldable, isFolded, prefix)

	row := spec.ToRow(item, hasChildren)
	row[0] = displayPrefix + row[0]

	result.Rows = append(result.Rows, row)
	result.Nodes = append(result.Nodes, TreeNode[T]{Item: item, ID: id, HasChildren: hasChildren})

	if isFolded {
		return result
	}

	for i, child := range kids {
		childIsLast := i == len(kids)-1
		result = buildParentIDMode(spec, child, childPrefix, false, childIsLast, children, folded, result)
	}
	return result
}

func buildChildrenMode[T any](spec TreeSpec[T], item T, prefix string, isRoot, isLast bool, folded map[string]bool, result BuildResult[T]) BuildResult[T] {
	id := spec.ID(item)
	kids := spec.Children(item)
	hasChildren := len(kids) > 0
	foldable := isFoldable(spec, item, hasChildren)
	isFolded := foldable && folded[id]

	displayPrefix, childPrefix := renderPrefixes(isRoot, isLast, foldable, isFolded, prefix)

	row := spec.ToRow(item, hasChildren)
	row[0] = displayPrefix + row[0]

	result.Rows = append(result.Rows, row)
	result.Nodes = append(result.Nodes, TreeNode[T]{Item: item, ID: id, HasChildren: hasChildren})

	if isFolded {
		return result
	}

	for i, child := range kids {
		childIsLast := i == len(kids)-1
		result = buildChildrenMode(spec, child, childPrefix, false, childIsLast, folded, result)
	}
	return result
}

func renderPrefixes(isRoot, isLast, foldable, isFolded bool, prefix string) (displayPrefix, childPrefix string) {
	foldIcon := ""
	if foldable {
		if isFolded {
			foldIcon = IconFolded
		} else {
			foldIcon = IconUnfolded
		}
	}

	if isRoot {
		displayPrefix = foldIcon
		childPrefix = ""
	} else {
		connector := ConnMid
		if isLast {
			connector = ConnLast
		}
		displayPrefix = prefix + connector + foldIcon
		if isLast {
			childPrefix = prefix + ConnBlank
		} else {
			childPrefix = prefix + ConnPipe
		}
	}
	return
}

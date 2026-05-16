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
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

type testItem struct {
	id       string
	parentID string
	value    string
}

func testSpec() TreeSpec[testItem] {
	return TreeSpec[testItem]{
		ID:       func(item testItem) string { return item.id },
		ParentID: func(item testItem) string { return item.parentID },
		Sort:     func(a, b testItem) bool { return a.id < b.id },
		Row: func(item testItem, _ bool) table.Row {
			return table.Row{item.id, item.value}
		},
	}
}

func TestBuildTree_ParentIDMode(t *testing.T) {
	items := []testItem{
		{id: "root1", parentID: "", value: "r1"},
		{id: "child1", parentID: "root1", value: "c1"},
		{id: "child2", parentID: "root1", value: "c2"},
		{id: "root2", parentID: "", value: "r2"},
	}

	result := BuildTree(testSpec(), items, nil)

	if len(result.Rows) != 4 {
		t.Fatalf("Expected 4 rows, got %d", len(result.Rows))
	}
	if result.Rows[0][0] != IconUnfolded+"root1" {
		t.Errorf("Row 0 = %q, want %q", result.Rows[0][0], IconUnfolded+"root1")
	}
	if result.Rows[1][0] != ConnMid+"child1" {
		t.Errorf("Row 1 = %q, want %q", result.Rows[1][0], ConnMid+"child1")
	}
	if result.Rows[2][0] != ConnLast+"child2" {
		t.Errorf("Row 2 = %q, want %q", result.Rows[2][0], ConnLast+"child2")
	}
	if result.Rows[3][0] != "root2" {
		t.Errorf("Row 3 = %q, want %q", result.Rows[3][0], "root2")
	}
}

func TestBuildTree_Folded(t *testing.T) {
	items := []testItem{
		{id: "root1", parentID: "", value: "r1"},
		{id: "child1", parentID: "root1", value: "c1"},
		{id: "root2", parentID: "", value: "r2"},
	}

	folded := map[string]bool{"root1": true}
	result := BuildTree(testSpec(), items, folded)

	if len(result.Rows) != 2 {
		t.Fatalf("Expected 2 rows (root1 folded), got %d", len(result.Rows))
	}
	if result.Rows[0][0] != IconFolded+"root1" {
		t.Errorf("Row 0 = %q, want %q", result.Rows[0][0], IconFolded+"root1")
	}
	if result.Rows[1][0] != "root2" {
		t.Errorf("Row 1 = %q, want %q", result.Rows[1][0], "root2")
	}
}

func TestBuildTree_NodeLookup(t *testing.T) {
	items := []testItem{
		{id: "root1", parentID: "", value: "r1"},
		{id: "child1", parentID: "root1", value: "c1"},
	}

	result := BuildTree(testSpec(), items, nil)

	if result.Nodes[0].ID != "root1" {
		t.Errorf("Node 0 ID = %q, want %q", result.Nodes[0].ID, "root1")
	}
	if !result.Nodes[0].HasChildren {
		t.Error("Node 0 should have children")
	}
	if result.Nodes[1].ID != "child1" {
		t.Errorf("Node 1 ID = %q, want %q", result.Nodes[1].ID, "child1")
	}
}

type testTreeItem struct {
	name     string
	children []testTreeItem
}

func TestBuildTree_ChildrenMode(t *testing.T) {
	items := []testTreeItem{
		{
			name: "parent",
			children: []testTreeItem{
				{name: "child1"},
				{name: "child2"},
			},
		},
		{name: "standalone"},
	}

	spec := TreeSpec[testTreeItem]{
		ID:       func(item testTreeItem) string { return item.name },
		Children: func(item testTreeItem) []testTreeItem { return item.children },
		Row: func(item testTreeItem, _ bool) table.Row {
			return table.Row{item.name}
		},
	}

	result := BuildTree(spec, items, nil)

	if len(result.Rows) != 4 {
		t.Fatalf("Expected 4 rows, got %d", len(result.Rows))
	}
	if result.Rows[0][0] != IconUnfolded+"parent" {
		t.Errorf("Row 0 = %q, want %q", result.Rows[0][0], IconUnfolded+"parent")
	}
	if result.Rows[1][0] != ConnMid+"child1" {
		t.Errorf("Row 1 = %q, want %q", result.Rows[1][0], ConnMid+"child1")
	}
	if result.Rows[2][0] != ConnLast+"child2" {
		t.Errorf("Row 2 = %q, want %q", result.Rows[2][0], ConnLast+"child2")
	}
	if result.Rows[3][0] != "standalone" {
		t.Errorf("Row 3 = %q, want %q", result.Rows[3][0], "standalone")
	}
}

func TestBuildTree_CustomFoldable(t *testing.T) {
	items := []testItem{
		{id: "root1", parentID: "", value: "foldable"},
		{id: "child1", parentID: "root1", value: "c1"},
		{id: "root2", parentID: "", value: "not-foldable"},
		{id: "child2", parentID: "root2", value: "c2"},
	}

	spec := testSpec()
	spec.Foldable = func(item testItem, hasChildren bool) bool {
		return hasChildren && item.value == "foldable"
	}

	result := BuildTree(spec, items, nil)

	// root1 has fold icon (foldable)
	if result.Rows[0][0] != IconUnfolded+"root1" {
		t.Errorf("Row 0 = %q, want fold icon", result.Rows[0][0])
	}
	// root2 has children but is not foldable, so no fold icon
	if result.Rows[2][0] != "root2" {
		t.Errorf("Row 2 = %q, want no fold icon", result.Rows[2][0])
	}
}

func TestDefaultFoldState(t *testing.T) {
	items := []testItem{
		{id: "root1", parentID: "", value: "r1"},
		{id: "child1", parentID: "root1", value: "c1"},
		{id: "root2", parentID: "", value: "r2"},
	}

	folded := make(map[string]bool)
	DefaultFoldState(testSpec(), items, folded)

	if !folded["root1"] {
		t.Error("root1 should be folded by default (has children)")
	}
	if folded["root2"] {
		t.Error("root2 should not be folded (no children)")
	}
}

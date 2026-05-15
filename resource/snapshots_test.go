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
	"time"

	"github.com/containerd/containerd/v2/core/snapshots"
)

func testSnapshots() []snapshots.Info {
	now := time.Now()
	return []snapshots.Info{
		{Name: "layer1", Parent: "", Kind: snapshots.KindCommitted, Created: now},
		{Name: "layer2", Parent: "layer1", Kind: snapshots.KindCommitted, Created: now},
		{Name: "layer3", Parent: "layer2", Kind: snapshots.KindCommitted, Created: now},
		{Name: "active1", Parent: "layer3", Kind: snapshots.KindActive, Created: now},
		{Name: "rootB", Parent: "", Kind: snapshots.KindCommitted, Created: now},
		{Name: "childB", Parent: "rootB", Kind: snapshots.KindActive, Created: now},
	}
}

func TestSnapshotKindToRows_Unfolded(t *testing.T) {
	data := testSnapshots()
	rows := SnapshotKind.Rows(data, nil)

	// All 6 snapshots should be visible
	if len(rows) != 6 {
		t.Fatalf("Expected 6 rows, got %d", len(rows))
	}

	// First row should be root with fold icon
	name := rows[0][0]
	if name != "▾ layer1" {
		t.Errorf("First row = %q, want %q", name, "▾ layer1")
	}

	// Second row should be child with connector
	name = rows[1][0]
	if name != "└─ layer2" {
		t.Errorf("Second row = %q, want %q", name, "└─ layer2")
	}
}

func TestSnapshotKindToRows_Folded(t *testing.T) {
	data := testSnapshots()
	folded := map[string]bool{"layer1": true, "rootB": true}

	rows := SnapshotKind.Rows(data, folded)

	// Only root nodes visible: layer1 + rootB = 2
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows with all folded, got %d", len(rows))
	}

	if rows[0][0] != "▸ layer1" {
		t.Errorf("Folded root = %q, want %q", rows[0][0], "▸ layer1")
	}
}

func TestSnapshotKindInitFolded(t *testing.T) {
	data := testSnapshots()
	folded := SnapshotKind.InitFolded(data)

	if !folded["layer1"] {
		t.Error("layer1 (root with children) should be folded")
	}
	if !folded["rootB"] {
		t.Error("rootB (root with children) should be folded")
	}
	if folded["layer2"] {
		t.Error("layer2 (non-root) should NOT be folded")
	}
}

func TestSnapshotKindRowID(t *testing.T) {
	data := testSnapshots()
	folded := map[string]bool{}

	// Index 0 is layer1 (root with children)
	id := SnapshotKind.FoldKey(data, folded, 0)
	if id != "layer1" {
		t.Errorf("RowID index 0 = %q, want %q", id, "layer1")
	}

	// Index 1 is layer2 (non-root, should not be foldable)
	id = SnapshotKind.FoldKey(data, folded, 1)
	if id != "" {
		t.Errorf("RowID index 1 (non-root) = %q, want empty", id)
	}
}

func TestSnapshotNodeAtIndex(t *testing.T) {
	infos := testSnapshots()
	folded := map[string]bool{}

	result := BuildTree(snapshotTreeSpec, infos, folded)

	// Unfolded order: layer1, layer2, layer3, active1, rootB, childB
	if result.Nodes[0].ID != "layer1" {
		t.Errorf("index 0 = %q, want %q", result.Nodes[0].ID, "layer1")
	}
	if result.Nodes[1].ID != "layer2" {
		t.Errorf("index 1 = %q, want %q", result.Nodes[1].ID, "layer2")
	}
	if result.Nodes[3].ID != "active1" {
		t.Errorf("index 3 = %q, want %q", result.Nodes[3].ID, "active1")
	}
	if result.Nodes[4].ID != "rootB" {
		t.Errorf("index 4 = %q, want %q", result.Nodes[4].ID, "rootB")
	}

	// Folded: layer1 folded hides its children
	foldedMap := map[string]bool{"layer1": true}
	foldedResult := BuildTree(snapshotTreeSpec, infos, foldedMap)
	if foldedResult.Nodes[0].ID != "layer1" {
		t.Errorf("folded index 0 = %q, want %q", foldedResult.Nodes[0].ID, "layer1")
	}
	if foldedResult.Nodes[1].ID != "rootB" {
		t.Errorf("folded index 1 = %q, want %q", foldedResult.Nodes[1].ID, "rootB")
	}
}

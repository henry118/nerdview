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
	"strings"
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

func TestSnapshotKindRows_Unfolded(t *testing.T) {
	data := testSnapshots()
	rows, _ := SnapshotKind.Rows(data, nil)

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

func TestSnapshotKindRows_Folded(t *testing.T) {
	data := testSnapshots()
	folded := map[string]bool{"layer1": true, "rootB": true}

	rows, _ := SnapshotKind.Rows(data, folded)

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

func TestSnapshotKindFoldKey(t *testing.T) {
	data := testSnapshots()
	folded := map[string]bool{}

	_, cache := SnapshotKind.Rows(data, folded)
	id := SnapshotKind.FoldKey(cache, 0)
	if id != "layer1" {
		t.Errorf("FoldKey index 0 = %q, want %q", id, "layer1")
	}

	id = SnapshotKind.FoldKey(cache, 1)
	if id != "" {
		t.Errorf("FoldKey index 1 (non-root) = %q, want empty", id)
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

func TestSnapshotKindDetail(t *testing.T) {
	data := testSnapshots()
	_, cache := SnapshotKind.Rows(data, nil)

	title, body := SnapshotKind.Detail(cache, 0)
	if title != "layer1" {
		t.Errorf("Title = %q, want %q", title, "layer1")
	}
	if !strings.Contains(body, "Name:    layer1") {
		t.Error("Should contain name")
	}
	if !strings.Contains(body, "Kind:    Committed") {
		t.Error("Should contain kind")
	}
}

func TestSnapshotKindNameAndColumns(t *testing.T) {
	if SnapshotKind.Name != "Snapshots" {
		t.Errorf("Name = %q, want %q", SnapshotKind.Name, "Snapshots")
	}
	cols := SnapshotKind.Columns
	if len(cols) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(cols))
	}
}

func TestSnapshotKindDetail_WithLabels(t *testing.T) {
	data := []snapshots.Info{
		{
			Name:   "labeled-snap",
			Kind:   snapshots.KindActive,
			Labels: map[string]string{"containerd.io/gc.root": "true"},
		},
	}

	_, cache := SnapshotKind.Rows(data, nil)
	_, body := SnapshotKind.Detail(cache, 0)
	if !strings.Contains(body, "Labels:") {
		t.Error("Should show labels section")
	}
	if !strings.Contains(body, "containerd.io/gc.root: true") {
		t.Error("Should show label content")
	}
}

func TestSnapshotKind_NilData(t *testing.T) {
	rows, cache := SnapshotKind.Rows(nil, nil)
	if rows != nil {
		t.Error("Rows(nil) should be nil")
	}
	if id := SnapshotKind.FoldKey(cache, 0); id != "" {
		t.Error("FoldKey(nil) should be empty")
	}
	if folded := SnapshotKind.InitFolded(nil); folded != nil {
		t.Error("InitFolded(nil) should be nil")
	}
	if _, body := SnapshotKind.Detail(cache, 0); body != "" {
		t.Error("Detail(nil) should be empty")
	}
}

func TestSnapshotKindCrossRefs(t *testing.T) {
	if SnapshotKind.CrossRefs != nil {
		t.Error("Snapshots should have no cross refs func")
	}
}

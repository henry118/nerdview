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

	"github.com/containerd/containerd/v2/core/containers"
	"github.com/henry118/nerdview/ctr"
)

func testContainers() []ctr.ContainerInfo {
	now := time.Now()
	return []ctr.ContainerInfo{
		{
			Container: containers.Container{
				ID:          "sandbox-abc",
				Image:       "registry.k8s.io/pause:3.9",
				Runtime:     containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt:   now,
				SandboxID:   "sandbox-abc",
				SnapshotKey: "sha256:sandbox-snap",
			},
			IsSandbox: true,
		},
		{
			Container: containers.Container{
				ID:          "app-container-1",
				Image:       "docker.io/library/nginx:latest",
				Runtime:     containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt:   now,
				SandboxID:   "sandbox-abc",
				SnapshotKey: "sha256:app1-snap",
			},
			IsSandbox: false,
		},
		{
			Container: containers.Container{
				ID:          "app-container-2",
				Image:       "docker.io/library/redis:7",
				Runtime:     containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt:   now,
				SandboxID:   "sandbox-abc",
				SnapshotKey: "sha256:app2-snap",
			},
			IsSandbox: false,
		},
		{
			Container: containers.Container{
				ID:        "standalone-ctr",
				Image:     "docker.io/library/alpine:3.19",
				Runtime:   containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt: now,
			},
			IsSandbox: false,
		},
	}
}

func TestContainerTreeNodes(t *testing.T) {
	data := testContainers()
	result := BuildTree(containerTreeSpec, data, nil)

	if len(result.Nodes) != 4 {
		t.Fatalf("Expected 4 nodes, got %d", len(result.Nodes))
	}
	if result.Nodes[0].ID != "sandbox-abc" || !result.Nodes[0].HasChildren {
		t.Errorf("Node 0: ID=%q HasChildren=%v, want sandbox-abc with children", result.Nodes[0].ID, result.Nodes[0].HasChildren)
	}
	if result.Nodes[1].ID != "app-container-1" {
		t.Errorf("Node 1: ID=%q, want app-container-1", result.Nodes[1].ID)
	}
	if result.Nodes[2].ID != "app-container-2" {
		t.Errorf("Node 2: ID=%q, want app-container-2", result.Nodes[2].ID)
	}
	if result.Nodes[3].ID != "standalone-ctr" {
		t.Errorf("Node 3: ID=%q, want standalone-ctr", result.Nodes[3].ID)
	}
}

func TestContainerKindToRows_Unfolded(t *testing.T) {
	data := testContainers()
	rows := ContainerKind.Rows(data, nil)

	// sandbox (1) + 2 children + standalone (1) = 4
	if len(rows) != 4 {
		t.Fatalf("Expected 4 rows, got %d", len(rows))
	}

	// Sandbox row with fold icon
	if rows[0][0] != "▾ sandbox-abc" {
		t.Errorf("First row = %q, want %q", rows[0][0], "▾ sandbox-abc")
	}
	if rows[0][1] != "sandbox" {
		t.Errorf("First row type = %q, want %q", rows[0][1], "sandbox")
	}

	// Child rows with connectors
	if rows[1][0] != "├─ app-container-1" {
		t.Errorf("Child 1 = %q, want %q", rows[1][0], "├─ app-container-1")
	}
	if rows[2][0] != "└─ app-container-2" {
		t.Errorf("Child 2 = %q, want %q", rows[2][0], "└─ app-container-2")
	}

	// Standalone
	if rows[3][0] != "standalone-ctr" {
		t.Errorf("Standalone = %q, want %q", rows[3][0], "standalone-ctr")
	}
	if rows[3][1] != "container" {
		t.Errorf("Standalone type = %q, want %q", rows[3][1], "container")
	}
}

func TestContainerKindToRows_Folded(t *testing.T) {
	data := testContainers()
	folded := map[string]bool{"sandbox-abc": true}
	rows := ContainerKind.Rows(data, folded)

	// sandbox folded (1) + standalone (1) = 2
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows folded, got %d", len(rows))
	}
	if rows[0][0] != "▸ sandbox-abc" {
		t.Errorf("Folded sandbox = %q, want %q", rows[0][0], "▸ sandbox-abc")
	}
}

func TestContainerKindRowID(t *testing.T) {
	data := testContainers()
	folded := map[string]bool{}

	// Index 0 is sandbox with children
	id := ContainerKind.FoldKey(data, folded, 0)
	if id != "sandbox-abc" {
		t.Errorf("RowID index 0 = %q, want %q", id, "sandbox-abc")
	}

	// Index 3 is standalone (no children)
	id = ContainerKind.FoldKey(data, folded, 3)
	if id != "" {
		t.Errorf("RowID index 3 (standalone) = %q, want empty", id)
	}
}

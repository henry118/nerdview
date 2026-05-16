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

	"github.com/containerd/containerd/v2/core/containers"
	"github.com/henry118/nerdview/ctr"
	specs "github.com/opencontainers/runtime-spec/specs-go"
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
	if rows[3][0] != "  standalone-ctr" {
		t.Errorf("Standalone = %q, want %q", rows[3][0], "  standalone-ctr")
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

func TestContainerKindDetail(t *testing.T) {
	data := testContainers()

	title, body := ContainerKind.Detail(data, nil, 0)
	if title != "sandbox-abc" {
		t.Errorf("Title = %q, want %q", title, "sandbox-abc")
	}
	if !strings.Contains(body, "Type:        sandbox") {
		t.Error("Should show type as sandbox")
	}
	if !strings.Contains(body, "SnapshotKey: sha256:sandbox-snap") {
		t.Error("Should show snapshot key")
	}
}

func TestContainerKindNameAndColumns(t *testing.T) {
	if ContainerKind.Name() != "Containers" {
		t.Errorf("Name = %q, want %q", ContainerKind.Name(), "Containers")
	}
	cols := ContainerKind.Columns()
	if len(cols) != 5 {
		t.Errorf("Expected 5 columns, got %d", len(cols))
	}
}

func TestContainerKindDetail_WithLabels(t *testing.T) {
	data := []ctr.ContainerInfo{
		{
			Container: containers.Container{
				ID:        "labeled-ctr",
				Image:     "nginx",
				Runtime:   containers.RuntimeInfo{Name: "runc"},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				SandboxID: "labeled-ctr",
				Labels:    map[string]string{"env": "prod", "app": "web"},
			},
			IsSandbox: true,
		},
	}

	_, body := ContainerKind.Detail(data, nil, 0)
	if !strings.Contains(body, "SandboxID:   labeled-ctr") {
		t.Error("Should contain SandboxID")
	}
	if !strings.Contains(body, "env: prod") {
		t.Error("Should contain labels")
	}
	if !strings.Contains(body, "Type:        sandbox") {
		t.Error("Should show type as sandbox")
	}
}

func TestContainerSpec(t *testing.T) {
	// No spec
	data := testContainers()
	title, body := ContainerSpec(data, nil, 0)
	if title != "" || body != "" {
		t.Errorf("Expected empty for container without spec, got title=%q", title)
	}

	// Out of bounds
	title, body = ContainerSpec(data, nil, 99)
	if title != "" || body != "" {
		t.Error("Expected empty for out of bounds")
	}

	// Nil data
	title, body = ContainerSpec(nil, nil, 0)
	if title != "" || body != "" {
		t.Error("Expected empty for nil data")
	}
}

func TestContainerKindInitFolded(t *testing.T) {
	data := testContainers()
	folded := ContainerKind.InitFolded(data)
	if folded != nil {
		t.Errorf("Containers InitFolded should be nil, got %v", folded)
	}
}

func TestContainerSpec_WithSpec(t *testing.T) {
	data := []ctr.ContainerInfo{
		{
			Container: containers.Container{
				ID:        "spec-ctr",
				Image:     "nginx",
				Runtime:   containers.RuntimeInfo{Name: "runc"},
				CreatedAt: time.Now(),
			},
			Spec: &specs.Spec{
				Hostname: "test-host",
				Root:     &specs.Root{Path: "/rootfs"},
			},
		},
	}
	title, body := ContainerSpec(data, nil, 0)
	if title != "spec-ctr" {
		t.Errorf("Title = %q, want %q", title, "spec-ctr")
	}
	if !strings.Contains(body, "test-host") {
		t.Error("Should contain hostname from spec JSON")
	}
}

func TestContainerKind_NilData(t *testing.T) {
	if rows := ContainerKind.Rows(nil, nil); rows != nil {
		t.Error("Rows(nil) should be nil")
	}
	if id := ContainerKind.FoldKey(nil, nil, 0); id != "" {
		t.Error("FoldKey(nil) should be empty")
	}
	if _, body := ContainerKind.Detail(nil, nil, 0); body != "" {
		t.Error("Detail(nil) should be empty")
	}
	if refs := ContainerKind.CrossRefs(nil, nil); refs != nil {
		t.Error("CrossRefs(nil) should be nil")
	}
}

func TestContainerKindCrossRefs(t *testing.T) {
	data := testContainers()
	refs := ContainerKind.CrossRefs(data, nil)

	if len(refs) != 4 {
		t.Fatalf("Expected 4 refs, got %d", len(refs))
	}
	if refs[0] != "sha256:sandbox-snap" {
		t.Errorf("CrossRef[0] = %q, want %q", refs[0], "sha256:sandbox-snap")
	}
	if refs[3] != "" {
		t.Errorf("CrossRef[3] = %q, want empty (standalone has no snapshot key)", refs[3])
	}
}

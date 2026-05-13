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
	"github.com/henry118/nerdtui/ctr"
)

func testContainers() []ctr.ContainerInfo {
	now := time.Now()
	return []ctr.ContainerInfo{
		{
			Container: containers.Container{
				ID:        "sandbox-abc",
				Image:     "registry.k8s.io/pause:3.9",
				Runtime:   containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt: now,
				SandboxID: "sandbox-abc",
			},
			IsSandbox: true,
		},
		{
			Container: containers.Container{
				ID:        "app-container-1",
				Image:     "docker.io/library/nginx:latest",
				Runtime:   containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt: now,
				SandboxID: "sandbox-abc",
			},
			IsSandbox: false,
		},
		{
			Container: containers.Container{
				ID:        "app-container-2",
				Image:     "docker.io/library/redis:7",
				Runtime:   containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt: now,
				SandboxID: "sandbox-abc",
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

func TestBuildSandboxChildren(t *testing.T) {
	data := testContainers()
	children := buildSandboxChildren(data)

	kids := children["sandbox-abc"]
	if len(kids) != 2 {
		t.Fatalf("sandbox-abc should have 2 children, got %d", len(kids))
	}
	if kids[0] != "app-container-1" || kids[1] != "app-container-2" {
		t.Errorf("Children = %v, want [app-container-1 app-container-2]", kids)
	}
}

func TestContainerKindToRows_Unfolded(t *testing.T) {
	data := testContainers()
	rows := ContainerKind.ToRows(data, nil)

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
	rows := ContainerKind.ToRows(data, folded)

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
	id := ContainerKind.RowID(data, folded, 0)
	if id != "sandbox-abc" {
		t.Errorf("RowID index 0 = %q, want %q", id, "sandbox-abc")
	}

	// Index 3 is standalone (no children)
	id = ContainerKind.RowID(data, folded, 3)
	if id != "" {
		t.Errorf("RowID index 3 (standalone) = %q, want empty", id)
	}
}

func TestStripContainerPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sandbox-abc", "sandbox-abc"},
		{"▸ sandbox-abc", "sandbox-abc"},
		{"▾ sandbox-abc", "sandbox-abc"},
		{"├─ app-container-1", "app-container-1"},
		{"└─ app-container-2", "app-container-2"},
	}
	for _, tt := range tests {
		got := stripContainerPrefix(tt.input)
		if got != tt.want {
			t.Errorf("stripContainerPrefix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

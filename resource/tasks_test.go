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

	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/henry118/nerdview/ctr"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTaskContainerRef(t *testing.T) {
	data := []ctr.TaskInfo{
		{ContainerID: "container-1", Process: &tasktypes.Process{ID: "container-1", Pid: 1234, Status: tasktypes.Status_RUNNING}},
		{ContainerID: "container-2", Process: &tasktypes.Process{ID: "container-2", Pid: 5678, Status: tasktypes.Status_STOPPED}},
	}

	if got := TaskContainerRef(data, nil, 0); got != "container-1" {
		t.Errorf("TaskContainerRef(0) = %q, want %q", got, "container-1")
	}
	if got := TaskContainerRef(data, nil, 1); got != "container-2" {
		t.Errorf("TaskContainerRef(1) = %q, want %q", got, "container-2")
	}
	if got := TaskContainerRef(data, nil, 99); got != "" {
		t.Errorf("TaskContainerRef(99) = %q, want empty", got)
	}
	if got := TaskContainerRef(nil, nil, 0); got != "" {
		t.Errorf("TaskContainerRef(nil) = %q, want empty", got)
	}
}

func TestTaskKindToRows(t *testing.T) {
	data := []ctr.TaskInfo{
		{
			ContainerID: "container-1",
			Process: &tasktypes.Process{
				ID:     "container-1",
				Pid:    1234,
				Status: tasktypes.Status_RUNNING,
			},
			StartedAt: "2025-01-01 10:00:00",
		},
		{
			ContainerID: "container-2",
			Process: &tasktypes.Process{
				ID:     "exec-shell",
				Pid:    5678,
				Status: tasktypes.Status_RUNNING,
			},
			ExecID: "exec-shell",
		},
	}

	rows := TaskKind.ToRows(data, nil)

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}
	// Init task: ID = container ID, Type = init
	if rows[0][0] != "container-1" {
		t.Errorf("Row 0 ID = %q, want %q", rows[0][0], "container-1")
	}
	if rows[0][1] != "container-1" {
		t.Errorf("Row 0 Container = %q, want %q", rows[0][1], "container-1")
	}
	if rows[0][2] != "init" {
		t.Errorf("Row 0 Type = %q, want %q", rows[0][2], "init")
	}
	if rows[0][3] != "1234" {
		t.Errorf("Row 0 PID = %q, want %q", rows[0][3], "1234")
	}
	if rows[0][4] != "RUNNING" {
		t.Errorf("Row 0 Status = %q, want %q", rows[0][4], "RUNNING")
	}
	if rows[0][6] != "2025-01-01 10:00:00" {
		t.Errorf("Row 0 Started = %q, want %q", rows[0][6], "2025-01-01 10:00:00")
	}
	// Exec task: ID = exec ID, Container = container ID, Type = exec
	if rows[1][0] != "exec-shell" {
		t.Errorf("Row 1 ID = %q, want %q", rows[1][0], "exec-shell")
	}
	if rows[1][1] != "container-2" {
		t.Errorf("Row 1 Container = %q, want %q", rows[1][1], "container-2")
	}
	if rows[1][2] != "exec" {
		t.Errorf("Row 1 Type = %q, want %q", rows[1][2], "exec")
	}
}

func TestTaskKindToDetail_Running(t *testing.T) {
	data := []ctr.TaskInfo{
		{
			ContainerID: "my-container",
			Process: &tasktypes.Process{
				ID:         "my-container",
				Pid:        1234,
				Status:     tasktypes.Status_RUNNING,
				ExitStatus: 0,
			},
			BundlePath: "/run/containerd/io.containerd.runtime.v2.task/default/my-container",
			StartedAt:  "2025-01-01 10:00:00",
			Spec: &specs.Spec{
				Root: &specs.Root{Path: "/var/lib/containerd/rootfs", Readonly: true},
				Process: &specs.Process{
					Args: []string{"/bin/sh", "-c", "echo hello"},
					Cwd:  "/app",
					User: specs.User{UID: 1000, GID: 1000},
				},
				Hostname: "my-host",
			},
		},
	}

	title, body := TaskKind.ToDetail(data, nil, 0)

	if title != "my-container" {
		t.Errorf("Title = %q, want %q", title, "my-container")
	}
	if !strings.Contains(body, "Type:         init") {
		t.Error("Should show type as init")
	}
	if !strings.Contains(body, "Started:      2025-01-01 10:00:00") {
		t.Error("Should show started timestamp")
	}
	if strings.Contains(body, "Exit Status") {
		t.Error("Running task should NOT show exit info")
	}
	if !strings.Contains(body, "Bundle:") {
		t.Error("Should contain bundle path")
	}
	if !strings.Contains(body, "Runtime Spec") {
		t.Error("Should contain runtime spec section")
	}
	if !strings.Contains(body, "/var/lib/containerd/rootfs") {
		t.Error("Should contain rootfs path in JSON")
	}
	if !strings.Contains(body, "echo hello") {
		t.Error("Should contain process args in JSON")
	}
	if !strings.Contains(body, "my-host") {
		t.Error("Should contain hostname in JSON")
	}
}

func TestTaskKindToDetail_Exec(t *testing.T) {
	data := []ctr.TaskInfo{
		{
			ContainerID: "my-container",
			ExecID:      "my-exec",
			Process: &tasktypes.Process{
				ID:     "my-exec",
				Pid:    9999,
				Status: tasktypes.Status_RUNNING,
			},
		},
	}

	title, body := TaskKind.ToDetail(data, nil, 0)

	if title != "my-exec" {
		t.Errorf("Title = %q, want %q", title, "my-exec")
	}
	if !strings.Contains(body, "Type:         exec") {
		t.Error("Should show type as exec")
	}
	if !strings.Contains(body, "Container:    my-container") {
		t.Error("Should show container ID")
	}
}

func TestTaskKindToDetail_Stopped(t *testing.T) {
	data := []ctr.TaskInfo{
		{
			ContainerID: "exited-ctr",
			Process: &tasktypes.Process{
				ID:         "exited-ctr",
				Pid:        0,
				Status:     tasktypes.Status_STOPPED,
				ExitStatus: 137,
				ExitedAt:   timestamppb.Now(),
			},
		},
	}

	_, body := TaskKind.ToDetail(data, nil, 0)

	if !strings.Contains(body, "Exit Status:  137") {
		t.Error("Stopped task should show exit status")
	}
	if !strings.Contains(body, "Exited At:") {
		t.Error("Stopped task should show exited_at time")
	}
}

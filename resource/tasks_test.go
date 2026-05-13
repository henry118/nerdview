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
	"github.com/henry118/nerdtui/ctr"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTaskKindToRows(t *testing.T) {
	data := []ctr.TaskInfo{
		{
			Process: &tasktypes.Process{
				ID:     "container-1",
				Pid:    1234,
				Status: tasktypes.Status_RUNNING,
			},
		},
		{
			Process: &tasktypes.Process{
				ID:     "container-2",
				Pid:    5678,
				Status: tasktypes.Status_STOPPED,
			},
		},
	}

	rows := TaskKind.ToRows(data, nil)

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "container-1" {
		t.Errorf("Row 0 ID = %q, want %q", rows[0][0], "container-1")
	}
	if rows[0][1] != "1234" {
		t.Errorf("Row 0 PID = %q, want %q", rows[0][1], "1234")
	}
	if rows[0][2] != "RUNNING" {
		t.Errorf("Row 0 Status = %q, want %q", rows[0][2], "RUNNING")
	}
}

func TestTaskKindToDetail_Running(t *testing.T) {
	data := []ctr.TaskInfo{
		{
			Process: &tasktypes.Process{
				ID:         "my-container",
				Pid:        1234,
				Status:     tasktypes.Status_RUNNING,
				ExitStatus: 0,
			},
			BundlePath: "/run/containerd/io.containerd.runtime.v2.task/default/my-container",
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
	if strings.Contains(body, "Exit Status") {
		t.Error("Running task should NOT show exit info")
	}
	if !strings.Contains(body, "Bundle:") {
		t.Error("Should contain bundle path")
	}
	if !strings.Contains(body, "/var/lib/containerd/rootfs") {
		t.Error("Should contain rootfs path")
	}
	if !strings.Contains(body, "/bin/sh -c echo hello") {
		t.Error("Should contain process args")
	}
	if !strings.Contains(body, "uid=1000 gid=1000") {
		t.Error("Should contain user info")
	}
	if !strings.Contains(body, "my-host") {
		t.Error("Should contain hostname")
	}
}

func TestTaskKindToDetail_Stopped(t *testing.T) {
	data := []ctr.TaskInfo{
		{
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

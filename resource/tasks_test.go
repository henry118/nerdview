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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTaskKindRows(t *testing.T) {
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

	rows := TaskKind.Rows(data, nil)

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

func TestTaskKindDetail_Running(t *testing.T) {
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
		},
	}

	title, body := TaskKind.Detail(data, nil, 0)

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
}

func TestTaskKindDetail_Exec(t *testing.T) {
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

	title, body := TaskKind.Detail(data, nil, 0)

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

func TestTaskKindDetail_Stopped(t *testing.T) {
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

	_, body := TaskKind.Detail(data, nil, 0)

	if !strings.Contains(body, "Exit Status:  137") {
		t.Error("Stopped task should show exit status")
	}
	if !strings.Contains(body, "Exited At:") {
		t.Error("Stopped task should show exited_at time")
	}
}

func TestTaskKindNameAndColumns(t *testing.T) {
	if TaskKind.Name() != "Tasks" {
		t.Errorf("Name = %q, want %q", TaskKind.Name(), "Tasks")
	}
	cols := TaskKind.Columns()
	if len(cols) != 7 {
		t.Errorf("Expected 7 columns, got %d", len(cols))
	}
}

func TestTaskKindCrossRefs(t *testing.T) {
	data := []ctr.TaskInfo{
		{ContainerID: "ctr-1", Process: &tasktypes.Process{ID: "ctr-1", Pid: 1}},
		{ContainerID: "ctr-2", Process: &tasktypes.Process{ID: "exec-1", Pid: 2}, ExecID: "exec-1"},
	}
	refs := TaskKind.CrossRefs(data, nil)
	if len(refs) != 2 {
		t.Fatalf("Expected 2 refs, got %d", len(refs))
	}
	if refs[0] != "ctr-1" {
		t.Errorf("CrossRef[0] = %q, want %q", refs[0], "ctr-1")
	}
	if refs[1] != "ctr-2" {
		t.Errorf("CrossRef[1] = %q, want %q", refs[1], "ctr-2")
	}
}

func TestTaskKind_NilData(t *testing.T) {
	if rows := TaskKind.Rows(nil, nil); rows != nil {
		t.Error("Rows(nil) should be nil")
	}
	if _, body := TaskKind.Detail(nil, nil, 0); body != "" {
		t.Error("Detail(nil) should be empty")
	}
	if refs := TaskKind.CrossRefs(nil, nil); refs != nil {
		t.Error("CrossRefs(nil) should be nil")
	}
}

func TestTaskKindFoldKeyAndInitFolded(t *testing.T) {
	data := []ctr.TaskInfo{
		{ContainerID: "ctr-1", Process: &tasktypes.Process{ID: "ctr-1", Pid: 1}},
	}
	if got := TaskKind.FoldKey(data, nil, 0); got != "" {
		t.Errorf("Tasks should not be foldable, got %q", got)
	}
	if got := TaskKind.InitFolded(data); got != nil {
		t.Errorf("Tasks InitFolded should be nil, got %v", got)
	}
}

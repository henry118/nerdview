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
)

func TestEventKindRows(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 45, 123000000, time.UTC)
	data := []Event{
		{Timestamp: now, Namespace: "default", Topic: "/images/create"},
		{Timestamp: now.Add(time.Second), Namespace: "k8s.io", Topic: "/containers/delete"},
	}

	rows := EventKind.Rows(data, nil)

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "10:30:45.123" {
		t.Errorf("Row 0 time = %q, want %q", rows[0][0], "10:30:45.123")
	}
	if rows[0][1] != "default" {
		t.Errorf("Row 0 namespace = %q, want %q", rows[0][1], "default")
	}
	if rows[0][2] != "/images/create" {
		t.Errorf("Row 0 topic = %q, want %q", rows[0][2], "/images/create")
	}
}

func TestEventKindDetail(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 45, 0, time.UTC)
	data := []Event{
		{Timestamp: now, Namespace: "default", Topic: "/tasks/exit"},
	}

	title, body := EventKind.Detail(data, nil, 0)

	if title != "/tasks/exit" {
		t.Errorf("Title = %q, want %q", title, "/tasks/exit")
	}
	if !strings.Contains(body, "default") {
		t.Error("Body should contain namespace")
	}
	if !strings.Contains(body, "/tasks/exit") {
		t.Error("Body should contain topic")
	}
}

func TestEventKindDetail_WithPayload(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 45, 0, time.UTC)

	type fakePayload struct {
		Digest string `json:"digest"`
		Size   int64  `json:"size"`
	}

	data := []Event{
		{
			Timestamp: now,
			Namespace: "default",
			Topic:     "/content/create",
			Payload:   &fakePayload{Digest: "sha256:abc123", Size: 4096},
		},
	}

	_, body := EventKind.Detail(data, nil, 0)

	if !strings.Contains(body, "--- Payload ---") {
		t.Error("Body should contain payload section")
	}
	if !strings.Contains(body, "sha256:abc123") {
		t.Error("Body should contain digest from payload")
	}
	if !strings.Contains(body, "4096") {
		t.Error("Body should contain size from payload")
	}
}

func TestEventKindDetail_NilPayload(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 45, 0, time.UTC)
	data := []Event{
		{Timestamp: now, Namespace: "default", Topic: "/images/create", Payload: nil},
	}

	_, body := EventKind.Detail(data, nil, 0)

	if strings.Contains(body, "Payload") {
		t.Error("Body should NOT contain payload section when payload is nil")
	}
}

func TestEventKindNameAndColumns(t *testing.T) {
	if EventKind.Name() != "Events" {
		t.Errorf("Name = %q, want %q", EventKind.Name(), "Events")
	}
	cols := EventKind.Columns()
	if len(cols) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(cols))
	}
}

func TestEventKind_NilAndEdgeCases(t *testing.T) {
	if rows := EventKind.Rows(nil, nil); rows != nil {
		t.Error("Rows(nil) should be nil")
	}
	if id := EventKind.FoldKey(nil, nil, 0); id != "" {
		t.Error("FoldKey should always be empty")
	}
	if folded := EventKind.InitFolded(nil); folded != nil {
		t.Error("InitFolded should always be nil")
	}
	if refs := EventKind.CrossRefs(nil, nil); refs != nil {
		t.Error("CrossRefs should always be nil")
	}
	if _, body := EventKind.Detail(nil, nil, 0); body != "" {
		t.Error("Detail(nil) should be empty")
	}
}

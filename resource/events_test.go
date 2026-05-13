package resource

import (
	"strings"
	"testing"
	"time"
)

func TestEventKindToRows(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 45, 123000000, time.UTC)
	data := []Event{
		{Timestamp: now, Namespace: "default", Topic: "/images/create"},
		{Timestamp: now.Add(time.Second), Namespace: "k8s.io", Topic: "/containers/delete"},
	}

	rows := EventKind.ToRows(data, nil)

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

func TestEventKindToDetail(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 45, 0, time.UTC)
	data := []Event{
		{Timestamp: now, Namespace: "default", Topic: "/tasks/exit"},
	}

	title, body := EventKind.ToDetail(data, nil, 0)

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

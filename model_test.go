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

package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henry118/nerdview/ctr"
	"github.com/henry118/nerdview/resource"
)

func testModel() model {
	m := model{
		namespaces: []string{"default", "k8s.io"},
		activeNS:   0,
		resources: []*resource.Tab{
			ptab(resource.NewTab(resource.ImageKind, 80, 10)),
			ptab(resource.NewTab(resource.ContainerKind, 80, 10)),
			ptab(resource.NewTab(resource.TaskKind, 80, 10)),
			ptab(resource.NewTab(resource.SnapshotKind, 80, 10)),
			ptab(resource.NewTab(resource.EventKind, 80, 10)),
		},
	}
	return m
}

func TestEventsFilteredByNamespace(t *testing.T) {
	m := testModel()

	// Event matching active namespace ("default") should be recorded
	msg := ctr.EventMsg{
		Namespace: "default",
		Topic:     "/images/create",
		Timestamp: time.Now(),
	}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if len(m.events) != 1 {
		t.Fatalf("Expected 1 event for matching namespace, got %d", len(m.events))
	}
	if m.events[0].Topic != "/images/create" {
		t.Errorf("Event topic = %q, want %q", m.events[0].Topic, "/images/create")
	}

	// Event from different namespace ("k8s.io") should NOT be recorded
	msg2 := ctr.EventMsg{
		Namespace: "k8s.io",
		Topic:     "/containers/create",
		Timestamp: time.Now(),
	}
	newModel, _ = m.Update(msg2)
	m = newModel.(model)

	if len(m.events) != 1 {
		t.Fatalf("Expected still 1 event after non-matching namespace event, got %d", len(m.events))
	}
}

func TestEventsClearedOnNamespaceSwitch(t *testing.T) {
	m := testModel()
	m.width = 80
	m.height = 24

	// Add an event
	m.events = []resource.Event{
		{Timestamp: time.Now(), Namespace: "default", Topic: "/images/create"},
		{Timestamp: time.Now(), Namespace: "default", Topic: "/tasks/start"},
	}
	m.resources[4].UpdateData(m.events)

	if len(m.events) != 2 {
		t.Fatalf("Expected 2 events before switch, got %d", len(m.events))
	}

	// Simulate namespace switch: set mode and cursor, then press Enter
	m.mode = modeNSSelect
	m.nsCursor = 1

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(enterKey)
	m = newModel.(model)

	if len(m.events) != 0 {
		t.Fatalf("Expected 0 events after namespace switch, got %d", len(m.events))
	}
	if m.activeNS != 1 {
		t.Errorf("activeNS = %d, want 1", m.activeNS)
	}
}

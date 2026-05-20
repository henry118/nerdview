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
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/henry118/nerdview/ctr"
	"github.com/henry118/nerdview/resource"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func testModel() model {
	m := model{
		ctx:        context.Background(),
		namespaces: []string{"default", "k8s.io"},
		activeNS:   0,
		resources: []*resource.Tab{
			resource.NewTab(&resource.ImageKind, 80, 10),
			resource.NewTab(&resource.SnapshotKind, 80, 10),
			resource.NewTab(&resource.ContainerKind, 80, 10),
			resource.NewTab(&resource.TaskKind, 80, 10),
			resource.NewTab(&resource.EventKind, 80, 10),
		},
		dirtyTabs: make(map[int]bool),
	}
	return m
}

func TestGoToContainerToSnapshot(t *testing.T) {
	m := testModel()
	m.width = 200
	m.height = 24
	for _, tab := range m.resources {
		tab.SetWidth(m.width)
		tab.Table.SetHeight(20)
	}

	snapshotKey := "sha256:7a75083e5b5a8d593efe8917fe730ab29cd8a8e8a5dfc2fcea022ab5a20954e0"

	// Load a container with a snapshot key
	containers := []ctr.ContainerInfo{
		{
			Container: containers.Container{
				ID:          "test-container",
				Image:       "nginx:latest",
				Runtime:     containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt:   time.Now(),
				SnapshotKey: snapshotKey,
			},
			IsSandbox: false,
		},
	}
	m.resources[tabContainers].UpdateData(containers)

	// Load snapshots including one matching the key
	snaps := []snapshots.Info{
		{Name: snapshotKey, Kind: snapshots.KindActive, Created: time.Now()},
	}
	m.resources[tabSnapshots].UpdateData(snaps)

	// Navigate to containers tab and press GoTo
	m.activeRes = tabContainers
	m.resources[tabContainers].Table.Focus()

	goToKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := m.Update(goToKey)
	m = newModel.(model)

	if m.activeRes != tabSnapshots {
		t.Fatalf("Expected to jump to snapshots tab (%d), got %d", tabSnapshots, m.activeRes)
	}

	// Verify cursor landed on the matching snapshot row
	rows := m.resources[tabSnapshots].Table.Rows()
	cursor := m.resources[tabSnapshots].Table.Cursor()
	if cursor >= len(rows) {
		t.Fatalf("Cursor %d out of range (rows=%d)", cursor, len(rows))
	}
	// The displayed name should contain the short digest
	shortKey := resource.ShortDigest(snapshotKey)
	if !strings.Contains(rows[cursor][0], shortKey) {
		t.Errorf("Cursor row = %q, does not contain short key %q", rows[cursor][0], shortKey)
	}

	// GoBack should return to containers tab
	backKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	newModel, _ = m.Update(backKey)
	m = newModel.(model)

	if m.activeRes != tabContainers {
		t.Errorf("Expected to return to containers tab (%d), got %d", tabContainers, m.activeRes)
	}
}

func TestGoToTaskToContainer(t *testing.T) {
	m := testModel()
	m.width = 200
	m.height = 24
	for _, tab := range m.resources {
		tab.SetWidth(m.width)
		tab.Table.SetHeight(20)
	}

	containerID := "7a75083e5b5a8d593efe8917fe730ab29cd8a8e8a5dfc2fcea022ab5a20954e0"

	ctrs := []ctr.ContainerInfo{
		{
			Container: containers.Container{
				ID:        containerID,
				Image:     "nginx:latest",
				Runtime:   containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt: time.Now(),
			},
		},
	}
	m.resources[tabContainers].UpdateData(ctrs)

	tasks := []ctr.TaskInfo{
		{
			ContainerID: containerID,
			Process:     &tasktypes.Process{ID: containerID, Pid: 1234, Status: tasktypes.Status_RUNNING},
		},
	}
	m.resources[tabTasks].UpdateData(tasks)

	m.activeRes = tabTasks
	m.resources[tabTasks].Table.Focus()

	goToKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := m.Update(goToKey)
	m = newModel.(model)

	if m.activeRes != tabContainers {
		t.Fatalf("Expected to jump to containers tab (%d), got %d", tabContainers, m.activeRes)
	}

	rows := m.resources[tabContainers].Table.Rows()
	cursor := m.resources[tabContainers].Table.Cursor()
	if cursor >= len(rows) {
		t.Fatalf("Cursor %d out of range (rows=%d)", cursor, len(rows))
	}
	if !strings.Contains(rows[cursor][0], containerID) {
		t.Errorf("Cursor row = %q, does not contain container ID", rows[cursor][0])
	}

	// GoBack
	backKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	newModel, _ = m.Update(backKey)
	m = newModel.(model)
	if m.activeRes != tabTasks {
		t.Errorf("Expected to return to tasks tab (%d), got %d", tabTasks, m.activeRes)
	}
}

func TestGoToImageToSnapshot(t *testing.T) {
	m := testModel()
	m.width = 200
	m.height = 24
	for _, tab := range m.resources {
		tab.SetWidth(m.width)
		tab.Table.SetHeight(20)
	}

	snapshotKey := "sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	images := []ctr.ImageTree{
		{
			Name:        "docker.io/library/nginx:latest",
			Desc:        ocispec.Descriptor{MediaType: "application/vnd.oci.image.manifest.v1+json", Digest: digest.FromString("nginx"), Size: 1024},
			SnapshotKey: snapshotKey,
			Children: []ctr.ImageTree{
				{Desc: ocispec.Descriptor{MediaType: "application/vnd.oci.image.config.v1+json", Digest: digest.FromString("config"), Size: 512}},
			},
		},
	}
	m.resources[tabImages].UpdateData(images)

	snaps := []snapshots.Info{
		{Name: snapshotKey, Kind: snapshots.KindCommitted, Created: time.Now()},
	}
	m.resources[tabSnapshots].UpdateData(snaps)

	m.activeRes = tabImages
	m.resources[tabImages].Table.Focus()

	goToKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := m.Update(goToKey)
	m = newModel.(model)

	if m.activeRes != tabSnapshots {
		t.Fatalf("Expected to jump to snapshots tab (%d), got %d", tabSnapshots, m.activeRes)
	}

	rows := m.resources[tabSnapshots].Table.Rows()
	cursor := m.resources[tabSnapshots].Table.Cursor()
	if cursor >= len(rows) {
		t.Fatalf("Cursor %d out of range (rows=%d)", cursor, len(rows))
	}
	shortKey := resource.ShortDigest(snapshotKey)
	if !strings.Contains(rows[cursor][0], shortKey) {
		t.Errorf("Cursor row = %q, does not contain short key %q", rows[cursor][0], shortKey)
	}

	// GoBack
	backKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	newModel, _ = m.Update(backKey)
	m = newModel.(model)
	if m.activeRes != tabImages {
		t.Errorf("Expected to return to images tab (%d), got %d", tabImages, m.activeRes)
	}
}

func TestGoToUnfoldsHiddenSnapshot(t *testing.T) {
	m := testModel()
	m.width = 200
	m.height = 24
	for _, tab := range m.resources {
		tab.SetWidth(m.width)
		tab.Table.SetHeight(20)
	}

	rootSnap := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa0000"
	childSnap := "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1111"

	// Container references the child snapshot
	ctrs := []ctr.ContainerInfo{
		{
			Container: containers.Container{
				ID:          "my-container",
				Image:       "nginx:latest",
				Runtime:     containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt:   time.Now(),
				SnapshotKey: childSnap,
			},
		},
	}
	m.resources[tabContainers].UpdateData(ctrs)

	// Snapshots: root with child underneath (child is hidden when root is folded)
	snaps := []snapshots.Info{
		{Name: rootSnap, Parent: "", Kind: snapshots.KindCommitted, Created: time.Now()},
		{Name: childSnap, Parent: rootSnap, Kind: snapshots.KindActive, Created: time.Now()},
	}
	m.resources[tabSnapshots].UpdateData(snaps)

	// Verify the child is hidden (root is folded by default via InitFolded)
	rows := m.resources[tabSnapshots].Table.Rows()
	shortChild := resource.ShortDigest(childSnap)
	found := false
	for _, row := range rows {
		if len(row) > 0 && strings.Contains(row[0], shortChild) {
			found = true
		}
	}
	if found {
		t.Fatal("Child snapshot should be hidden (folded) before navigation")
	}

	// Navigate from container to snapshot
	m.activeRes = tabContainers
	m.resources[tabContainers].Table.Focus()

	goToKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := m.Update(goToKey)
	m = newModel.(model)

	if m.activeRes != tabSnapshots {
		t.Fatalf("Expected to jump to snapshots tab (%d), got %d", tabSnapshots, m.activeRes)
	}

	// Verify the child is now visible and cursor is on it
	rows = m.resources[tabSnapshots].Table.Rows()
	cursor := m.resources[tabSnapshots].Table.Cursor()
	if cursor >= len(rows) {
		t.Fatalf("Cursor %d out of range (rows=%d)", cursor, len(rows))
	}
	if !strings.Contains(rows[cursor][0], shortChild) {
		t.Errorf("Cursor row = %q, does not contain child snapshot %q", rows[cursor][0], shortChild)
	}
}

func TestGoToUnfoldsOnlyTargetAncestor(t *testing.T) {
	m := testModel()
	m.width = 200
	m.height = 24
	for _, tab := range m.resources {
		tab.SetWidth(m.width)
		tab.Table.SetHeight(20)
	}

	rootA := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	childA := "sha256:aaaa111111111111111111111111111111111111111111111111111111111111"
	rootB := "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	childB := "sha256:bbbb111111111111111111111111111111111111111111111111111111111111"

	// Container references childA
	ctrs := []ctr.ContainerInfo{
		{
			Container: containers.Container{
				ID:          "my-container",
				Image:       "nginx:latest",
				Runtime:     containers.RuntimeInfo{Name: "io.containerd.runc.v2"},
				CreatedAt:   time.Now(),
				SnapshotKey: childA,
			},
		},
	}
	m.resources[tabContainers].UpdateData(ctrs)

	// Two snapshot trees, both folded by default
	snaps := []snapshots.Info{
		{Name: rootA, Parent: "", Kind: snapshots.KindCommitted, Created: time.Now()},
		{Name: childA, Parent: rootA, Kind: snapshots.KindActive, Created: time.Now()},
		{Name: rootB, Parent: "", Kind: snapshots.KindCommitted, Created: time.Now()},
		{Name: childB, Parent: rootB, Kind: snapshots.KindActive, Created: time.Now()},
	}
	m.resources[tabSnapshots].UpdateData(snaps)

	// Verify both roots are folded (only 2 visible rows)
	rows := m.resources[tabSnapshots].Table.Rows()
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows (both folded), got %d", len(rows))
	}

	// Navigate to snapshot from container
	m.activeRes = tabContainers
	m.resources[tabContainers].Table.Focus()
	goToKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := m.Update(goToKey)
	m = newModel.(model)

	// rootA should be unfolded (childA visible), rootB should stay folded
	rows = m.resources[tabSnapshots].Table.Rows()
	shortChildA := resource.ShortDigest(childA)
	shortChildB := resource.ShortDigest(childB)

	foundChildA := false
	foundChildB := false
	for _, row := range rows {
		if len(row) > 0 {
			if strings.Contains(row[0], shortChildA) {
				foundChildA = true
			}
			if strings.Contains(row[0], shortChildB) {
				foundChildB = true
			}
		}
	}
	if !foundChildA {
		t.Error("childA should be visible (its ancestor was unfolded)")
	}
	if foundChildB {
		t.Error("childB should still be hidden (unrelated tree should stay folded)")
	}
}

func TestTabSwitchForward(t *testing.T) {
	m := testModel()
	m.width = 80
	m.height = 24

	if m.activeRes != tabImages {
		t.Fatalf("Expected initial tab to be images (%d), got %d", tabImages, m.activeRes)
	}

	tabKey := tea.KeyMsg{Type: tea.KeyTab}

	newModel, _ := m.Update(tabKey)
	m = newModel.(model)
	if m.activeRes != tabSnapshots {
		t.Errorf("After 1st Tab: expected %d, got %d", tabSnapshots, m.activeRes)
	}

	newModel, _ = m.Update(tabKey)
	m = newModel.(model)
	if m.activeRes != tabContainers {
		t.Errorf("After 2nd Tab: expected %d, got %d", tabContainers, m.activeRes)
	}

	newModel, _ = m.Update(tabKey)
	m = newModel.(model)
	if m.activeRes != tabTasks {
		t.Errorf("After 3rd Tab: expected %d, got %d", tabTasks, m.activeRes)
	}

	newModel, _ = m.Update(tabKey)
	m = newModel.(model)
	if m.activeRes != tabEvents {
		t.Errorf("After 4th Tab: expected %d, got %d", tabEvents, m.activeRes)
	}

	// Wraps around
	newModel, _ = m.Update(tabKey)
	m = newModel.(model)
	if m.activeRes != tabImages {
		t.Errorf("After 5th Tab (wrap): expected %d, got %d", tabImages, m.activeRes)
	}
}

func TestTabSwitchBackward(t *testing.T) {
	m := testModel()
	m.width = 80
	m.height = 24

	shiftTabKey := tea.KeyMsg{Type: tea.KeyShiftTab}

	newModel, _ := m.Update(shiftTabKey)
	m = newModel.(model)
	if m.activeRes != tabEvents {
		t.Errorf("After 1st Shift+Tab: expected %d, got %d", tabEvents, m.activeRes)
	}

	newModel, _ = m.Update(shiftTabKey)
	m = newModel.(model)
	if m.activeRes != tabTasks {
		t.Errorf("After 2nd Shift+Tab: expected %d, got %d", tabTasks, m.activeRes)
	}
}

func TestDebounceCoalescesMultipleEvents(t *testing.T) {
	m := testModel()

	// Send 5 container events rapidly
	for range 5 {
		msg := ctr.EventMsg{
			Namespace: "default",
			Topic:     "/containers/create",
			Timestamp: time.Now(),
		}
		newModel, _ := m.Update(msg)
		m = newModel.(model)
	}

	// Only container tab should be dirty
	if !m.dirtyTabs[tabContainers] {
		t.Error("tabContainers should be dirty after container events")
	}
	if m.dirtyTabs[tabImages] {
		t.Error("tabImages should NOT be dirty")
	}

	// debounceGen should reflect the latest event
	if m.debounceGen != 5 {
		t.Errorf("debounceGen = %d, want 5", m.debounceGen)
	}

	// Stale debounce msg (gen=1) should be ignored
	staleMsg := debounceMsg{gen: 1}
	newModel, _ := m.Update(staleMsg)
	m = newModel.(model)
	if !m.dirtyTabs[tabContainers] {
		t.Error("Stale debounce should NOT clear dirty tabs")
	}

	// Current gen debounce msg should clear dirty tabs
	currentMsg := debounceMsg{gen: m.debounceGen}
	newModel, _ = m.Update(currentMsg)
	m = newModel.(model)
	if len(m.dirtyTabs) != 0 {
		t.Errorf("After debounce fires, dirtyTabs should be empty, got %v", m.dirtyTabs)
	}
}

func TestDebounceMixedEventTypes(t *testing.T) {
	m := testModel()

	// Mix of event types
	topics := []string{"/containers/create", "/tasks/start", "/containers/delete", "/images/pull"}
	for _, topic := range topics {
		msg := ctr.EventMsg{Namespace: "default", Topic: topic, Timestamp: time.Now()}
		newModel, _ := m.Update(msg)
		m = newModel.(model)
	}

	// Three tab types should be dirty
	if !m.dirtyTabs[tabContainers] {
		t.Error("tabContainers should be dirty")
	}
	if !m.dirtyTabs[tabTasks] {
		t.Error("tabTasks should be dirty")
	}
	if !m.dirtyTabs[tabImages] {
		t.Error("tabImages should be dirty")
	}
	if m.dirtyTabs[tabSnapshots] {
		t.Error("tabSnapshots should NOT be dirty")
	}
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

func TestConnectionStatus_StartsDisconnected(t *testing.T) {
	m := testModel()
	if m.connected {
		t.Error("model should start disconnected")
	}
}

func TestConnectionStatus_ConnectedOnDataLoad(t *testing.T) {
	m := testModel()

	msg := resourcesLoadedMsg{
		namespace: "default",
		images:    nil,
	}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if !m.connected {
		t.Error("should be connected after resourcesLoadedMsg")
	}
}

func TestConnectionStatus_DisconnectedOnEventErr(t *testing.T) {
	m := testModel()
	m.connected = true

	msg := ctr.EventErrMsg{Err: fmt.Errorf("stream closed")}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if m.connected {
		t.Error("should be disconnected after EventErrMsg")
	}
}

func TestConnectionStatus_DisconnectedOnUnavailable(t *testing.T) {
	m := testModel()
	m.connected = true

	msg := errorMsg{err: fmt.Errorf("some other error")}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if !m.connected {
		t.Error("non-unavailable error should not disconnect")
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"generic error", fmt.Errorf("timeout"), false},
		{"connection closing", fmt.Errorf("connection is closing"), true},
		{"transport closing", fmt.Errorf("transport is closing"), true},
		{"wrapped connection closing", fmt.Errorf("rpc: %s", "connection is closing"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isConnectionError(tt.err); got != tt.want {
				t.Errorf("isConnectionError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestDisconnect_ResetsState(t *testing.T) {
	m := testModel()
	m.connected = true
	m.snapshotters = []string{"overlayfs", "native"}
	m.daemonStats = ctr.DaemonStats{PID: 1234, CPUPct: 5.0, RSS: 1024, VMS: 2048, Threads: 10}

	// Load data into tabs
	m.resources[tabImages].UpdateData([]ctr.ImageTree{
		{Name: "nginx:latest"},
	})
	m.resources[tabContainers].UpdateData([]ctr.ContainerInfo{
		{Container: containers.Container{ID: "ctr1", Runtime: containers.RuntimeInfo{Name: "runc"}, CreatedAt: time.Now()}},
	})
	m.resources[tabTasks].UpdateData([]ctr.TaskInfo{
		{ContainerID: "ctr1", Process: &tasktypes.Process{ID: "ctr1", Pid: 100, Status: tasktypes.Status_RUNNING}},
	})

	// Send connection error
	msg := errorMsg{err: fmt.Errorf("transport is closing")}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if m.connected {
		t.Error("should be disconnected")
	}
	if m.err != nil {
		t.Error("err should be nil for connection errors")
	}
	if len(m.namespaces) != 1 {
		t.Errorf("namespaces should be reduced to 1, got %d", len(m.namespaces))
	}
	if m.namespaces[0] != "default" {
		t.Errorf("remaining namespace should be 'default', got %q", m.namespaces[0])
	}
	if m.activeNS != 0 {
		t.Errorf("activeNS should be 0, got %d", m.activeNS)
	}
	if m.snapshotters != nil {
		t.Errorf("snapshotters should be nil, got %v", m.snapshotters)
	}
	if m.daemonStats.PID != 0 {
		t.Errorf("daemonStats.PID should be 0, got %d", m.daemonStats.PID)
	}
	for i, tab := range m.resources {
		if len(tab.Table.Rows()) != 0 {
			t.Errorf("tab %d should have 0 rows, got %d", i, len(tab.Table.Rows()))
		}
	}
}

func TestDisconnect_PreservesActiveNamespace(t *testing.T) {
	m := testModel()
	m.connected = true
	m.namespaces = []string{"default", "k8s.io", "test"}
	m.activeNS = 2

	msg := errorMsg{err: fmt.Errorf("connection is closing")}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if len(m.namespaces) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(m.namespaces))
	}
	if m.namespaces[0] != "test" {
		t.Errorf("preserved namespace should be 'test', got %q", m.namespaces[0])
	}
	if m.activeNS != 0 {
		t.Errorf("activeNS should be reset to 0, got %d", m.activeNS)
	}
}

func TestDisconnect_NonConnectionError_KeepsState(t *testing.T) {
	m := testModel()
	m.connected = true
	m.snapshotters = []string{"overlayfs"}
	m.daemonStats = ctr.DaemonStats{PID: 999}
	m.resources[tabContainers].UpdateData([]ctr.ContainerInfo{
		{Container: containers.Container{ID: "ctr1", Runtime: containers.RuntimeInfo{Name: "runc"}, CreatedAt: time.Now()}},
	})

	msg := errorMsg{err: fmt.Errorf("permission denied")}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if !m.connected {
		t.Error("should remain connected for non-connection errors")
	}
	if m.err == nil || m.err.Error() != "permission denied" {
		t.Errorf("err should be 'permission denied', got %v", m.err)
	}
	if m.snapshotters == nil {
		t.Error("snapshotters should not be cleared")
	}
	if m.daemonStats.PID != 999 {
		t.Error("daemonStats should not be cleared")
	}
	if len(m.resources[tabContainers].Table.Rows()) == 0 {
		t.Error("tab rows should not be cleared")
	}
}

func TestReconnectedMsg_RestoresState(t *testing.T) {
	m := testModel()
	m.connected = false
	m.err = fmt.Errorf("old error")

	msg := ctr.ReconnectedMsg{}
	newModel, cmd := m.Update(msg)
	m = newModel.(model)

	if !m.connected {
		t.Error("should be connected after ReconnectedMsg")
	}
	if m.err != nil {
		t.Errorf("err should be nil after reconnect, got %v", m.err)
	}
	if cmd == nil {
		t.Error("should return commands to reload data after reconnect")
	}
}

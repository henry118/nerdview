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
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/containerd/containerd/v2/core/snapshots"

	"github.com/henry118/nerdview/ctr"
	"github.com/henry118/nerdview/logging"
	"github.com/henry118/nerdview/resource"
	"github.com/henry118/nerdview/ui"
)

// maxEvents caps the number of events retained in the events tab.
const maxEvents = 200

// Tab indices matching the order in newModel().
const (
	tabImages     = 0
	tabSnapshots  = 1
	tabContainers = 2
	tabTasks      = 3
	tabEvents     = 4
)

// viewMode tracks which UI overlay is active.
type viewMode int

const (
	modeNormal            viewMode = iota // Default table navigation.
	modeDialog                            // Detail popup is open.
	modeNSSelect                          // Namespace selector overlay.
	modeSnapshotterSelect                 // Snapshotter selector overlay.
)

// navState stores the position before a cross-tab navigation for "go back".
type navState struct {
	tabIndex int
	cursor   int
}

// model is the top-level bubbletea model for the TUI.
type model struct {
	client       *ctr.Client      // Containerd gRPC client.
	namespaces   []string         // Available namespaces.
	activeNS     int              // Index into namespaces.
	snapshotter  string           // Active snapshotter name.
	snapshotters []string         // Available snapshotters.
	snCursor     int              // Cursor for snapshotter selector.
	resources    []*resource.Tab  // One tab per resource type.
	activeRes    int              // Index of the active tab.
	events       []resource.Event // Buffered events for the events tab.
	dialog       ui.DialogModel   // Detail/spec popup.
	mode         viewMode         // Current UI mode.
	nsCursor     int              // Cursor for namespace selector.
	daemonPID    int              // Containerd daemon PID.
	daemonStats  ctr.DaemonStats  // Latest daemon resource stats.
	navHistory   []navState       // Stack for cross-tab back navigation.
	width        int              // Terminal width.
	height       int              // Terminal height.
	err          error            // Last error to display in status bar.
}

// newModel creates the initial model with default state.
func newModel(client *ctr.Client, namespace string) model {
	return model{
		client:      client,
		namespaces:  []string{namespace},
		snapshotter: "overlayfs",
		resources: []*resource.Tab{
			new(resource.NewTab(resource.ImageKind, 80, 10)),
			new(resource.NewTab(resource.SnapshotKind, 80, 10)),
			new(resource.NewTab(resource.ContainerKind, 80, 10)),
			new(resource.NewTab(resource.TaskKind, 80, 10)),
			new(resource.NewTab(resource.EventKind, 80, 10)),
		},
		dialog: ui.NewDialog(80, 24),
	}
}

// Init starts background data loading, event subscription, and periodic refresh.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadNamespaces(m.client),
		loadSnapshotters(m.client),
		loadResources(m.client, m.namespaces[m.activeNS], m.snapshotter),
		ctr.WaitForEvent(m.client),
		initDaemonStats(m.client),
		tickCmd(),
		statsTickCmd(),
	)
}

// Update handles all incoming messages (events, key presses, data loads).
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// overhead: stats (1) + tab bar (1) + help bar (1) = 3
		tableHeight := max(m.height-3, 3)
		for _, tab := range m.resources {
			tab.SetWidth(m.width)
			tab.Table.SetHeight(tableHeight)
		}
		m.dialog.SetSize(m.width, m.height)
		return m, nil

	case namespacesLoadedMsg:
		if len(msg.namespaces) == 0 {
			return m, nil
		}
		currentNS := m.namespaces[m.activeNS]
		m.namespaces = msg.namespaces
		m.activeNS = 0
		for i, ns := range m.namespaces {
			if ns == currentNS {
				m.activeNS = i
				break
			}
		}
		return m, nil

	case snapshottersLoadedMsg:
		m.snapshotters = msg.snapshotters
		return m, nil

	case resourcesLoadedMsg:
		if msg.namespace == m.namespaces[m.activeNS] {
			m.resources[tabImages].UpdateData(msg.images)
			m.resources[tabSnapshots].UpdateData(msg.snapshots)
			m.resources[tabContainers].UpdateData(msg.containers)
			m.resources[tabTasks].UpdateData(msg.tasks)
		}
		return m, nil

	case ctr.EventMsg:
		var cmds []tea.Cmd
		cmds = append(cmds, ctr.WaitForEvent(m.client))
		if msg.Namespace == m.namespaces[m.activeNS] {
			cmds = append(cmds, refreshResource(m.client, m.namespaces[m.activeNS], m.snapshotter, msg.Topic))
			m.events = append([]resource.Event{{
				Timestamp: msg.Timestamp,
				Namespace: msg.Namespace,
				Topic:     msg.Topic,
				Payload:   msg.Event,
			}}, m.events...)
			if len(m.events) > maxEvents {
				m.events = m.events[:maxEvents]
			}
			m.resources[tabEvents].UpdateData(m.events)
		}
		return m, tea.Batch(cmds...)

	case ctr.EventErrMsg:
		m.err = msg.Err
		return m, nil

	case daemonStatsMsg:
		m.daemonStats = msg.stats
		m.daemonPID = msg.stats.PID
		return m, nil

	case statsTickMsg:
		return m, tea.Batch(
			refreshDaemonStats(m.client, m.daemonPID),
			statsTickCmd(),
		)

	case imagesRefreshedMsg:
		m.resources[tabImages].UpdateData([]ctr.ImageTree(msg))
		return m, nil

	case snapshotsRefreshedMsg:
		m.resources[tabSnapshots].UpdateData([]snapshots.Info(msg))
		return m, nil

	case containersRefreshedMsg:
		m.resources[tabContainers].UpdateData([]ctr.ContainerInfo(msg))
		return m, nil

	case tasksRefreshedMsg:
		m.resources[tabTasks].UpdateData([]ctr.TaskInfo(msg))
		return m, nil

	case errorMsg:
		m.err = msg.err
		logging.Error("ui error: %v", msg.err)
		return m, nil

	case tickMsg:
		return m, tea.Batch(
			loadResources(m.client, m.namespaces[m.activeNS], m.snapshotter),
			tickCmd(),
		)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey dispatches key events to the active mode handler.
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeDialog:
		return m.handleDialogKey(msg)
	case modeNSSelect:
		return m.handleNSSelectKey(msg)
	case modeSnapshotterSelect:
		return m.handleSnapshotterSelectKey(msg)
	default:
		return m.handleNormalKey(msg)
	}
}

// handleDialogKey handles keys when the detail popup is open.
func (m model) handleDialogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = modeNormal
		return m, nil
	default:
		var cmd tea.Cmd
		m.dialog, cmd = m.dialog.Update(msg)
		return m, cmd
	}
}

// handleNSSelectKey handles keys in the namespace selector overlay.
func (m model) handleNSSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = modeNormal
		return m, nil
	case key.Matches(msg, keys.Up):
		if m.nsCursor > 0 {
			m.nsCursor--
		}
		return m, nil
	case key.Matches(msg, keys.Down):
		if m.nsCursor < len(m.namespaces)-1 {
			m.nsCursor++
		}
		return m, nil
	case key.Matches(msg, keys.Enter):
		m.activeNS = m.nsCursor
		m.mode = modeNormal
		m.events = nil
		m.resources[tabEvents].UpdateData(m.events)
		logging.Info("switched to namespace: %s", m.namespaces[m.activeNS])
		return m, loadResources(m.client, m.namespaces[m.activeNS], m.snapshotter)
	}
	return m, nil
}

// handleSnapshotterSelectKey handles keys in the snapshotter selector overlay.
func (m model) handleSnapshotterSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = modeNormal
		return m, nil
	case key.Matches(msg, keys.Up):
		if m.snCursor > 0 {
			m.snCursor--
		}
		return m, nil
	case key.Matches(msg, keys.Down):
		if m.snCursor < len(m.snapshotters)-1 {
			m.snCursor++
		}
		return m, nil
	case key.Matches(msg, keys.Enter):
		m.snapshotter = m.snapshotters[m.snCursor]
		m.mode = modeNormal
		logging.Info("switched to snapshotter: %s", m.snapshotter)
		return m, loadResources(m.client, m.namespaces[m.activeNS], m.snapshotter)
	}
	return m, nil
}

// handleNormalKey handles keys in the default table navigation mode.
func (m model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit), key.Matches(msg, keys.Escape):
		return m, tea.Quit

	case key.Matches(msg, keys.SelectNS):
		m.nsCursor = m.activeNS
		m.mode = modeNSSelect
		return m, nil

	case key.Matches(msg, keys.SelectSnapshotter):
		if len(m.snapshotters) > 0 {
			m.snCursor = 0
			for i, s := range m.snapshotters {
				if s == m.snapshotter {
					m.snCursor = i
					break
				}
			}
			m.mode = modeSnapshotterSelect
		}
		return m, nil

	case key.Matches(msg, keys.NextResource):
		m.resources[m.activeRes].Table.Blur()
		m.activeRes = (m.activeRes + 1) % len(m.resources)
		m.resources[m.activeRes].Table.Focus()
		return m, nil

	case key.Matches(msg, keys.PrevResource):
		m.resources[m.activeRes].Table.Blur()
		m.activeRes = (m.activeRes - 1 + len(m.resources)) % len(m.resources)
		m.resources[m.activeRes].Table.Focus()
		return m, nil

	case key.Matches(msg, keys.Right):
		m.resources[m.activeRes].Unfold()
		return m, nil

	case key.Matches(msg, keys.Left):
		m.resources[m.activeRes].Fold()
		return m, nil

	case key.Matches(msg, keys.GoTo):
		tab := m.resources[m.activeRes]
		idx := tab.Table.Cursor()
		targetKey := tab.CrossRef(idx)
		var targetTab int
		switch m.activeRes {
		case tabImages, tabContainers:
			targetTab = tabSnapshots
		case tabTasks:
			targetTab = tabContainers
		}
		if targetKey != "" {
			logging.Debug("go to: tab %d -> tab %d, key=%s", m.activeRes, targetTab, resource.ShortDigest(targetKey))
			m.navHistory = append(m.navHistory, navState{
				tabIndex: m.activeRes,
				cursor:   idx,
			})
			m.resources[m.activeRes].Table.Blur()
			m.activeRes = targetTab
			m.resources[m.activeRes].Table.Focus()
			shortKey := resource.ShortDigest(targetKey)
			targetIdx := m.resources[m.activeRes].RevealRow(func(row table.Row) bool {
				return len(row) > 0 && strings.Contains(row[0], shortKey)
			})
			if targetIdx >= 0 {
				m.resources[m.activeRes].Table.SetCursor(targetIdx)
			}
		}
		return m, nil

	case key.Matches(msg, keys.GoBack):
		if len(m.navHistory) > 0 {
			prev := m.navHistory[len(m.navHistory)-1]
			m.navHistory = m.navHistory[:len(m.navHistory)-1]
			m.resources[m.activeRes].Table.Blur()
			m.activeRes = prev.tabIndex
			m.resources[m.activeRes].Table.Focus()
			m.resources[m.activeRes].Table.SetCursor(prev.cursor)
		}
		return m, nil

	case key.Matches(msg, keys.Spec):
		if m.activeRes == tabContainers {
			tab := m.resources[tabContainers]
			title, body := resource.ContainerSpec(tab.RawData, tab.Folded, tab.Table.Cursor())
			if title != "" {
				m.dialog.SetContent(title, body)
				m.mode = modeDialog
			}
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		tab := m.resources[m.activeRes]
		title, body := tab.SelectedDetail()
		if title != "" {
			m.dialog.SetContent(title, body)
			m.mode = modeDialog
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.resources[m.activeRes].Table, cmd = m.resources[m.activeRes].Table.Update(msg)
		return m, cmd
	}
}

// loadSnapshotters fetches available snapshotters from the daemon.
func loadSnapshotters(client *ctr.Client) tea.Cmd {
	return func() tea.Msg {
		sns, err := client.Snapshotters(context.Background())
		if err != nil {
			return errorMsg{err: err}
		}
		return snapshottersLoadedMsg{snapshotters: sns}
	}
}

// loadNamespaces fetches available namespaces from the daemon.
func loadNamespaces(client *ctr.Client) tea.Cmd {
	return func() tea.Msg {
		ns, err := client.Namespaces(context.Background())
		if err != nil {
			return errorMsg{err: err}
		}
		return namespacesLoadedMsg{namespaces: ns}
	}
}

// loadResources fetches all resource data for the active namespace.
func loadResources(client *ctr.Client, ns, snapshotter string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		imgs, err := client.ImageTrees(ctx, ns, snapshotter)
		if err != nil {
			return errorMsg{err: err}
		}
		ctrs, err := client.Containers(ctx, ns)
		if err != nil {
			return errorMsg{err: err}
		}
		tasks, err := client.Tasks(ctx, ns)
		if err != nil {
			return errorMsg{err: err}
		}
		snaps, err := client.Snapshots(ctx, ns, snapshotter)
		if err != nil {
			snaps = nil
		}
		return resourcesLoadedMsg{
			namespace:  ns,
			images:     imgs,
			containers: ctrs,
			tasks:      tasks,
			snapshots:  snaps,
		}
	}
}

// refreshResource reloads a single resource type based on an event topic.
func refreshResource(client *ctr.Client, ns, snapshotter, topic string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		switch {
		case strings.HasPrefix(topic, "/images/"):
			imgs, err := client.ImageTrees(ctx, ns, snapshotter)
			if err != nil {
				return errorMsg{err: err}
			}
			return imagesRefreshedMsg(imgs)
		case strings.HasPrefix(topic, "/containers/"):
			ctrs, err := client.Containers(ctx, ns)
			if err != nil {
				return errorMsg{err: err}
			}
			return containersRefreshedMsg(ctrs)
		case strings.HasPrefix(topic, "/tasks/"):
			tasks, err := client.Tasks(ctx, ns)
			if err != nil {
				return errorMsg{err: err}
			}
			return tasksRefreshedMsg(tasks)
		case strings.HasPrefix(topic, "/snapshot/"):
			snaps, err := client.Snapshots(ctx, ns, snapshotter)
			if err != nil {
				return errorMsg{err: err}
			}
			return snapshotsRefreshedMsg(snaps)
		default:
			return nil
		}
	}
}

// tickCmd triggers a full data refresh every 30 seconds.
func tickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// statsTickCmd triggers a daemon stats refresh every 2 seconds.
func statsTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return statsTickMsg{}
	})
}

// initDaemonStats fetches the initial daemon PID and stats.
func initDaemonStats(client *ctr.Client) tea.Cmd {
	return func() tea.Msg {
		pid, err := client.DaemonPID()
		if err != nil {
			return nil
		}
		stats, err := ctr.ReadDaemonStats(pid)
		if err != nil {
			return nil
		}
		return daemonStatsMsg{stats: stats}
	}
}

// refreshDaemonStats updates daemon resource usage metrics.
func refreshDaemonStats(client *ctr.Client, pid int) tea.Cmd {
	return func() tea.Msg {
		if pid == 0 {
			var err error
			pid, err = client.DaemonPID()
			if err != nil {
				return nil
			}
		}
		stats, err := ctr.ReadDaemonStats(pid)
		if err != nil {
			return nil
		}
		return daemonStatsMsg{stats: stats}
	}
}

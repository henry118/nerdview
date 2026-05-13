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
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/containerd/containerd/v2/core/snapshots"

	"github.com/henry118/nerdtui/ctr"
	"github.com/henry118/nerdtui/logging"
	"github.com/henry118/nerdtui/resource"
	"github.com/henry118/nerdtui/ui"
)

const maxEvents = 200

type viewMode int

const (
	modeNormal viewMode = iota
	modeDialog
	modeNSSelect
	modeSnapshotterSelect
)

type model struct {
	client       *ctr.Client
	namespaces   []string
	activeNS     int
	snapshotter  string
	snapshotters []string
	snCursor     int
	resources    []*resource.Tab
	activeRes    int
	events       []resource.Event
	dialog       ui.DialogModel
	mode         viewMode
	nsCursor     int
	daemonPID    int
	daemonStats  ctr.DaemonStats
	width        int
	height       int
	err          error
}

func newModel(client *ctr.Client, namespace string) model {
	return model{
		client:      client,
		namespaces:  []string{namespace},
		snapshotter: "overlayfs",
		resources: []*resource.Tab{
			ptab(resource.NewTab(resource.ImageKind, 80, 10)),
			ptab(resource.NewTab(resource.ContainerKind, 80, 10)),
			ptab(resource.NewTab(resource.TaskKind, 80, 10)),
			ptab(resource.NewTab(resource.SnapshotKind, 80, 10)),
			ptab(resource.NewTab(resource.EventKind, 80, 10)),
		},
		dialog: ui.NewDialog(80, 24),
	}
}

func ptab(t resource.Tab) *resource.Tab { return &t }

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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// overhead: header (1) + blank (1) + tab bar (1) + help bar (1) = 4
		tableHeight := m.height - 4
		if tableHeight < 3 {
			tableHeight = 3
		}
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
			m.resources[0].UpdateData(msg.images)
			m.resources[1].UpdateData(msg.containers)
			m.resources[2].UpdateData(msg.tasks)
			m.resources[3].UpdateData(msg.snapshots)
		}
		return m, nil

	case ctr.EventMsg:
		var cmds []tea.Cmd
		cmds = append(cmds, ctr.WaitForEvent(m.client))
		if msg.Namespace == m.namespaces[m.activeNS] {
			cmds = append(cmds, refreshResource(m.client, m.namespaces[m.activeNS], m.snapshotter, msg.Topic))
		}
		m.events = append([]resource.Event{{
			Timestamp: msg.Timestamp,
			Namespace: msg.Namespace,
			Topic:     msg.Topic,
		}}, m.events...)
		if len(m.events) > maxEvents {
			m.events = m.events[:maxEvents]
		}
		m.resources[4].UpdateData(m.events)
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
			refreshDaemonStats(m.daemonPID),
			statsTickCmd(),
		)

	case imagesRefreshedMsg:
		m.resources[0].UpdateData([]ctr.ImageTree(msg))
		return m, nil

	case containersRefreshedMsg:
		m.resources[1].UpdateData([]ctr.ContainerInfo(msg))
		return m, nil

	case tasksRefreshedMsg:
		m.resources[2].UpdateData([]ctr.TaskInfo(msg))
		return m, nil

	case snapshotsRefreshedMsg:
		m.resources[3].UpdateData([]snapshots.Info(msg))
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

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeDialog:
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = modeNormal
			return m, nil
		default:
			var cmd tea.Cmd
			m.dialog, cmd = m.dialog.Update(msg)
			return m, cmd
		}

	case modeNSSelect:
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
			logging.Info("switched to namespace: %s", m.namespaces[m.activeNS])
			return m, loadResources(m.client, m.namespaces[m.activeNS], m.snapshotter)
		}
		return m, nil

	case modeSnapshotterSelect:
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

	default:
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

		case key.Matches(msg, keys.ToggleFold):
			tab := m.resources[m.activeRes]
			if tab.CanFold() {
				tab.ToggleFold()
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
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header bar: app name, namespace, daemon stats
	headerContent := styleHeader.Render(" nerdtui ") +
		styleHeaderNS.Render(fmt.Sprintf("[ns:%s]", m.namespaces[m.activeNS])) +
		m.renderDaemonStats()
	headerWidth := lipgloss.Width(headerContent)
	headerPad := ""
	if m.width > headerWidth {
		headerPad = styleHeader.Render(strings.Repeat(" ", m.width-headerWidth))
	}
	header := headerContent + headerPad

	// Tab bar: resource tabs
	var tabs []string
	for i, tab := range m.resources {
		if i == m.activeRes {
			tabs = append(tabs, styleTabActive.Render(tab.Kind.Name))
		} else {
			tabs = append(tabs, styleTabInactive.Render(tab.Kind.Name))
		}
	}
	tabBar := strings.Join(tabs, "")
	tabBarWidth := lipgloss.Width(tabBar)
	if m.width > tabBarWidth {
		tabBar += styleTabInactive.Background(lipgloss.Color("235")).Render(
			strings.Repeat(" ", m.width-tabBarWidth))
	}

	// Table
	tableView := m.resources[m.activeRes].Table.View()

	// Help bar
	helpBar := ui.HelpView(m.width)
	if m.err != nil {
		errText := fmt.Sprintf(" ERROR: %s ", m.err.Error())
		errPad := strings.Repeat(" ", max(0, m.width-len(errText)))
		helpBar = styleError.Render(errText + errPad)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", tabBar, tableView, helpBar)

	// Overlay selectors
	if m.mode == modeNSSelect {
		content = m.overlaySelector("Select Namespace", m.namespaces, m.nsCursor)
	}
	if m.mode == modeSnapshotterSelect {
		content = m.overlaySelector("Select Snapshotter", m.snapshotters, m.snCursor)
	}

	// Overlay detail dialog
	if m.mode == modeDialog {
		dialogView := m.dialog.View()
		content = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialogView,
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return content
}

func (m model) overlaySelector(title string, items []string, cursor int) string {
	titleLine := styleNSListTitle.Render(" " + title + " ")
	var lines []string
	for i, item := range items {
		if i == cursor {
			lines = append(lines, styleNSListSelected.Render("> "+item))
		} else {
			lines = append(lines, styleNSListItem.Render("  "+item))
		}
	}
	list := strings.Join(lines, "\n")
	box := styleNSList.Render(lipgloss.JoinVertical(lipgloss.Left, titleLine, list))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
	)
}

func (m model) renderDaemonStats() string {
	s := m.daemonStats
	if s.PID == 0 {
		return ""
	}
	label := styleHeader.Render(" daemon ")
	pid := styleStatsLabel.Render("pid:") + styleStatsPID.Render(fmt.Sprintf("%d", s.PID))
	cpu := styleStatsLabel.Render(" cpu:") + styleStatsCPU.Render(fmt.Sprintf("%.1f%%", s.CPUPct))
	vms := styleStatsLabel.Render(" vms:") + styleStatsVMS.Render(formatBytes(s.VMS))
	rss := styleStatsLabel.Render(" rss:") + styleStatsRSS.Render(formatBytes(s.RSS))
	threads := styleStatsLabel.Render(" threads:") + styleStatsThreads.Render(fmt.Sprintf("%d", s.Threads))
	up := styleStatsLabel.Render(" up:") + styleStatsUptime.Render(formatDuration(s.Uptime))
	return label + pid + cpu + vms + rss + threads + up
}

func formatBytes(b uint64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func loadSnapshotters(client *ctr.Client) tea.Cmd {
	return func() tea.Msg {
		sns, err := client.Snapshotters(context.Background())
		if err != nil {
			return errorMsg{err: err}
		}
		return snapshottersLoadedMsg{snapshotters: sns}
	}
}

func loadNamespaces(client *ctr.Client) tea.Cmd {
	return func() tea.Msg {
		ns, err := client.Namespaces(context.Background())
		if err != nil {
			return errorMsg{err: err}
		}
		return namespacesLoadedMsg{namespaces: ns}
	}
}

func loadResources(client *ctr.Client, ns, snapshotter string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		imgs, err := client.ImageTrees(ctx, ns)
		if err != nil {
			return errorMsg{err: err}
		}
		ctrs, err := client.Containers(ctx, ns)
		if err != nil {
			return errorMsg{err: err}
		}
		tasks, err := client.TasksWithSpec(ctx, ns)
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

type imagesRefreshedMsg []ctr.ImageTree
type containersRefreshedMsg []ctr.ContainerInfo
type tasksRefreshedMsg []ctr.TaskInfo
type snapshotsRefreshedMsg []snapshots.Info

func refreshResource(client *ctr.Client, ns, snapshotter, topic string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		switch {
		case strings.HasPrefix(topic, "/images/"):
			imgs, err := client.ImageTrees(ctx, ns)
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
			tasks, err := client.TasksWithSpec(ctx, ns)
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

func tickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func statsTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return statsTickMsg{}
	})
}

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

func refreshDaemonStats(pid int) tea.Cmd {
	return func() tea.Msg {
		if pid == 0 {
			return nil
		}
		stats, err := ctr.ReadDaemonStats(pid)
		if err != nil {
			return nil
		}
		return daemonStatsMsg{stats: stats}
	}
}

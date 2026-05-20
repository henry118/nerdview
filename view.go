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
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/henry118/nerdview/resource"
	"github.com/henry118/nerdview/ui"
)

// View renders the full TUI frame.
func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Stats bar: namespace + daemon metrics
	statsBar := m.renderStatsBar()
	statsWidth := lipgloss.Width(statsBar)
	if m.width > statsWidth {
		statsBar += ui.StyleStatsLabel.Render(strings.Repeat(" ", m.width-statsWidth))
	}

	// Tab bar: resource tabs
	var tabs []string
	for i, tab := range m.resources {
		if i == m.activeRes {
			tabs = append(tabs, ui.StyleTabActive.Render(tab.Name()))
		} else {
			tabs = append(tabs, ui.StyleTabInactive.Render(tab.Name()))
		}
	}
	tabBar := strings.Join(tabs, "")
	tabBarWidth := lipgloss.Width(tabBar)
	if m.width > tabBarWidth {
		tabBar += ui.StyleTabBarFill.Render(strings.Repeat(" ", m.width-tabBarWidth))
	}

	// Table
	tableView := m.resources[m.activeRes].Table.View()

	// Help bar with position indicator
	tab := m.resources[m.activeRes]
	rowCount := len(tab.Table.Rows())
	var goToLabel string
	if tab.CrossRef(tab.Table.Cursor()) != "" {
		switch m.activeRes {
		case tabImages, tabContainers:
			goToLabel = "sn"
		case tabTasks:
			goToLabel = "ctr"
		}
	}
	var helpOpts []ui.HelpOption
	if len(m.snapshotters) > 0 {
		helpOpts = append(helpOpts, ui.WithSnapshotter())
	}
	if goToLabel != "" {
		helpOpts = append(helpOpts, ui.WithGoTo(goToLabel))
	}
	if len(m.navHistory) > 0 {
		helpOpts = append(helpOpts, ui.WithBack())
	}
	if tab.HasSpec() {
		helpOpts = append(helpOpts, ui.WithSpec())
	}
	if rowCount > 0 {
		helpOpts = append(helpOpts, ui.WithPosition(fmt.Sprintf("%d/%d", tab.Table.Cursor()+1, rowCount)))
	}
	helpBar := ui.HelpView(m.width, helpOpts...)
	if m.err != nil {
		errText := fmt.Sprintf(" ERROR: %s ", m.err.Error())
		errPad := strings.Repeat(" ", max(0, m.width-len(errText)))
		helpBar = ui.StyleError.Render(errText + errPad)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, statsBar, tabBar, tableView, helpBar)

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

// overlaySelector renders a centered selector popup for namespaces or snapshotters.
func (m model) overlaySelector(title string, items []string, cursor int) string {
	titleLine := ui.StyleNSListTitle.Render(" " + title + " ")
	var lines []string
	for i, item := range items {
		if i == cursor {
			lines = append(lines, ui.StyleNSListSelected.Render("> "+item))
		} else {
			lines = append(lines, ui.StyleNSListItem.Render("  "+item))
		}
	}
	list := strings.Join(lines, "\n")
	box := ui.StyleNSList.Render(lipgloss.JoinVertical(lipgloss.Left, titleLine, list))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
	)
}

const statsNA = "--"

// renderStatsBar renders the namespace and daemon stats line.
func (m model) renderStatsBar() string {
	dot := ui.StyleConnected.Render(" ●")
	if !m.connected {
		dot = ui.StyleDisconnected.Render(" ●")
	}
	ns := dot + ui.StyleStatsLabel.Render(" ns:") + ui.StyleHeaderNS.Render(m.namespaces[m.activeNS])

	s := m.daemonStats
	pidVal, cpuVal, vmsVal, rssVal, threadsVal, upVal := statsNA, statsNA, statsNA, statsNA, statsNA, statsNA
	if s.PID != 0 {
		pidVal = fmt.Sprintf("%d", s.PID)
		cpuVal = fmt.Sprintf("%.1f%%", s.CPUPct)
		vmsVal = resource.FormatBytes(s.VMS)
		rssVal = resource.FormatBytes(s.RSS)
		threadsVal = fmt.Sprintf("%d", s.Threads)
		upVal = formatDuration(s.Uptime)
	}

	pid := ui.StyleStatsLabel.Render("  pid:") + ui.StyleStatsPID.Render(pidVal)
	cpu := ui.StyleStatsLabel.Render(" cpu:") + ui.StyleStatsCPU.Render(cpuVal)
	vms := ui.StyleStatsLabel.Render(" vms:") + ui.StyleStatsVMS.Render(vmsVal)
	rss := ui.StyleStatsLabel.Render(" rss:") + ui.StyleStatsRSS.Render(rssVal)
	threads := ui.StyleStatsLabel.Render(" threads:") + ui.StyleStatsThreads.Render(threadsVal)
	up := ui.StyleStatsLabel.Render(" up:") + ui.StyleStatsUptime.Render(upVal)
	return ns + pid + cpu + vms + rss + threads + up
}

// formatDuration formats a duration into a compact "Xd Xh Xm" string.
func formatDuration(d time.Duration) string {
	totalMins := int(d.Minutes())
	days := totalMins / 1440
	hours := (totalMins % 1440) / 60
	mins := totalMins % 60
	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

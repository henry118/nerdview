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

package ui

import "github.com/charmbracelet/lipgloss"

var (
	// StyleHeaderNS is the namespace label style in the stats bar.
	StyleHeaderNS = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTeal)).
			Bold(true)

	// StyleTabActive is the style for the currently selected tab.
	StyleTabActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBase)).
			Background(lipgloss.Color(ColorMauve)).
			Bold(true).
			Padding(0, 1)

	// StyleTabInactive is the style for unselected tabs.
	StyleTabInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorOverlay0)).
				Padding(0, 1)

	// StyleTabBarFill fills the remaining tab bar width.
	StyleTabBarFill = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorOverlay0)).
			Background(lipgloss.Color(ColorBase))

	// StyleStatsLabel is the label style for daemon stats.
	StyleStatsLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSubtext0)).
			Background(lipgloss.Color(ColorBase))

	// StyleStatsPID is the PID value style.
	StyleStatsPID = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Background(lipgloss.Color(ColorBase))

	// StyleStatsCPU is the CPU value style.
	StyleStatsCPU = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPeach)).
			Background(lipgloss.Color(ColorBase)).
			Bold(true)

	// StyleStatsVMS is the virtual memory value style.
	StyleStatsVMS = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLavender)).
			Background(lipgloss.Color(ColorBase))

	// StyleStatsRSS is the resident memory value style.
	StyleStatsRSS = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGreen)).
			Background(lipgloss.Color(ColorBase)).
			Bold(true)

	// StyleStatsThreads is the thread count value style.
	StyleStatsThreads = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSapphire)).
				Background(lipgloss.Color(ColorBase))

	// StyleStatsUptime is the uptime value style.
	StyleStatsUptime = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorYellow)).
				Background(lipgloss.Color(ColorBase))

	// StyleError is the error message style.
	StyleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorRed)).
			Background(lipgloss.Color(ColorBase)).
			Bold(true)

	// StyleNSList is the namespace/snapshotter selector box style.
	StyleNSList = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorSurface0)).
			Padding(0, 1)

	// StyleNSListTitle is the selector title style.
	StyleNSListTitle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTeal)).
				Bold(true).
				Padding(0, 1)

	// StyleNSListItem is the unselected selector item style.
	StyleNSListItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Padding(0, 1)

	// StyleNSListSelected is the highlighted selector item style.
	StyleNSListSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorBase)).
				Background(lipgloss.Color(ColorMauve)).
				Bold(true).
				Padding(0, 1)

	// StyleTableSelected is the selected row style for all tables.
	StyleTableSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorBase)).
				Background(lipgloss.Color(ColorTeal))
)

// Help bar styles.
var (
	// StyleHelpBar is the background style for help bar text.
	StyleHelpBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSubtext0)).
			Background(lipgloss.Color(ColorBase))

	// StyleHelpKey is the style for key binding labels in the help bar.
	StyleHelpKey = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTeal)).
			Background(lipgloss.Color(ColorBase)).
			Bold(true)

	// StyleHelpPos is the style for the row position indicator.
	StyleHelpPos = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorOverlay0)).
			Background(lipgloss.Color(ColorBase))
)

// Dialog styles.
var (
	// StyleDialogBox is the outer border style for detail popups.
	StyleDialogBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(ColorMauve)).
			Padding(0, 1)

	// StyleDialogTitleBar is the title bar style inside the dialog.
	StyleDialogTitleBar = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorBase)).
				Background(lipgloss.Color(ColorMauve)).
				Padding(0, 1).
				Align(lipgloss.Center)

	// StyleDialogFooter is the footer hint style inside the dialog.
	StyleDialogFooter = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorOverlay0)).
				Align(lipgloss.Right)

	// StyleDialogSeparator is the horizontal separator style inside the dialog.
	StyleDialogSeparator = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorMauve))
)

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
	"github.com/charmbracelet/lipgloss"
	"github.com/henry118/nerdview/ui"
)

var (
	styleHeaderNS = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorTeal)).
			Bold(true)

	styleTabActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorBase)).
			Background(lipgloss.Color(ui.ColorMauve)).
			Bold(true).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorOverlay0)).
				Padding(0, 1)

	styleStatsLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorSubtext0)).
			Background(lipgloss.Color(ui.ColorBase))

	styleStatsPID = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorText)).
			Background(lipgloss.Color(ui.ColorBase))

	styleStatsCPU = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorPeach)).
			Background(lipgloss.Color(ui.ColorBase)).
			Bold(true)

	styleStatsVMS = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorLavender)).
			Background(lipgloss.Color(ui.ColorBase))

	styleStatsRSS = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorGreen)).
			Background(lipgloss.Color(ui.ColorBase)).
			Bold(true)

	styleStatsThreads = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorSapphire)).
				Background(lipgloss.Color(ui.ColorBase))

	styleStatsUptime = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorYellow)).
				Background(lipgloss.Color(ui.ColorBase))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorRed)).
			Background(lipgloss.Color(ui.ColorBase)).
			Bold(true)

	styleNSList = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ui.ColorSurface0)).
			Padding(0, 1)

	styleNSListTitle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorTeal)).
				Bold(true).
				Padding(0, 1)

	styleNSListItem = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorText)).
				Padding(0, 1)

	styleNSListSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorBase)).
				Background(lipgloss.Color(ui.ColorMauve)).
				Bold(true).
				Padding(0, 1)
)

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

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	helpBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236"))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("115")).
			Background(lipgloss.Color("236")).
			Bold(true)
)

// HelpView renders the bottom help bar showing key bindings, padded to width.
// goToLabel is the "go to" target label (e.g. "sn", "ctr"); empty means hidden.
func HelpView(width int, goToLabel string, showBack bool) string {
	parts := []string{
		helpKeyStyle.Render("←/→") + helpBarStyle.Render(":resource  "),
		helpKeyStyle.Render("Tab") + helpBarStyle.Render(":fold/unfold  "),
		helpKeyStyle.Render("n") + helpBarStyle.Render(":namespace  "),
		helpKeyStyle.Render("s") + helpBarStyle.Render(":snapshotter  "),
	}
	if goToLabel != "" {
		parts = append(parts, helpKeyStyle.Render("g")+helpBarStyle.Render(":go to "+goToLabel+"  "))
	}
	if showBack {
		parts = append(parts, helpKeyStyle.Render("b")+helpBarStyle.Render(":back  "))
	}
	parts = append(parts,
		helpKeyStyle.Render("Enter")+helpBarStyle.Render(":detail  "),
		helpKeyStyle.Render("Esc")+helpBarStyle.Render(":quit"),
	)
	text := strings.Join(parts, "")
	textWidth := lipgloss.Width(text)
	if width > textWidth {
		text += helpBarStyle.Render(strings.Repeat(" ", width-textWidth))
	}
	return text
}

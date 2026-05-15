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
			Foreground(lipgloss.Color(ColorSubtext0)).
			Background(lipgloss.Color(ColorBase))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTeal)).
			Background(lipgloss.Color(ColorBase)).
			Bold(true)

	helpPosStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorOverlay0)).
			Background(lipgloss.Color(ColorBase))
)

// HelpOption configures the help bar.
type HelpOption func(*helpConfig)

type helpConfig struct {
	goToLabel string
	showBack  bool
	showSpec  bool
	position  string
}

func WithGoTo(label string) HelpOption {
	return func(c *helpConfig) { c.goToLabel = label }
}

func WithBack() HelpOption {
	return func(c *helpConfig) { c.showBack = true }
}

func WithSpec() HelpOption {
	return func(c *helpConfig) { c.showSpec = true }
}

func WithPosition(pos string) HelpOption {
	return func(c *helpConfig) { c.position = pos }
}

// HelpView renders the bottom help bar showing key bindings, padded to width.
func HelpView(width int, opts ...HelpOption) string {
	var cfg helpConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	parts := []string{
		helpKeyStyle.Render("Tab") + helpBarStyle.Render(":tab  "),
		helpKeyStyle.Render("←/→") + helpBarStyle.Render(":fold  "),
		helpKeyStyle.Render("n") + helpBarStyle.Render(":namespace  "),
		helpKeyStyle.Render("s") + helpBarStyle.Render(":snapshotter  "),
	}
	if cfg.goToLabel != "" {
		parts = append(parts, helpKeyStyle.Render("g")+helpBarStyle.Render(":go to "+cfg.goToLabel+"  "))
	}
	if cfg.showBack {
		parts = append(parts, helpKeyStyle.Render("b")+helpBarStyle.Render(":back  "))
	}
	if cfg.showSpec {
		parts = append(parts, helpKeyStyle.Render("p")+helpBarStyle.Render(":spec  "))
	}
	parts = append(parts,
		helpKeyStyle.Render("Enter")+helpBarStyle.Render(":detail  "),
		helpKeyStyle.Render("q/Esc")+helpBarStyle.Render(":quit"),
	)
	text := strings.Join(parts, "")
	textWidth := lipgloss.Width(text)

	if cfg.position != "" {
		posText := helpPosStyle.Render(cfg.position)
		posWidth := lipgloss.Width(posText)
		gap := width - textWidth - posWidth
		if gap > 0 {
			text += helpBarStyle.Render(strings.Repeat(" ", gap)) + posText
		} else {
			text += " " + posText
		}
	} else if width > textWidth {
		text += helpBarStyle.Render(strings.Repeat(" ", width-textWidth))
	}
	return text
}

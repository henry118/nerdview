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

// HelpOption configures the help bar.
type HelpOption func(*helpConfig)

type helpConfig struct {
	goToLabel       string
	showBack        bool
	showSpec        bool
	showSnapshotter bool
	position        string
}

// WithGoTo shows the "go to" hint with the given target label (e.g. "sn", "ctr").
func WithGoTo(label string) HelpOption {
	return func(c *helpConfig) { c.goToLabel = label }
}

// WithBack shows the "back" navigation hint.
func WithBack() HelpOption {
	return func(c *helpConfig) { c.showBack = true }
}

// WithSpec shows the "spec" hint for viewing runtime specs.
func WithSpec() HelpOption {
	return func(c *helpConfig) { c.showSpec = true }
}

// WithSnapshotter shows the "snapshotter" selector hint.
func WithSnapshotter() HelpOption {
	return func(c *helpConfig) { c.showSnapshotter = true }
}

// WithPosition shows a right-aligned row position indicator (e.g. "3/47").
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
		StyleHelpKey.Render("TAB") + StyleHelpBar.Render(":TAB  "),
		StyleHelpKey.Render("←/→") + StyleHelpBar.Render(":FOLD  "),
		StyleHelpKey.Render("N") + StyleHelpBar.Render(":NAMESPACE  "),
	}
	if cfg.showSnapshotter {
		parts = append(parts, StyleHelpKey.Render("S")+StyleHelpBar.Render(":SNAPSHOTTER  "))
	}
	if cfg.goToLabel != "" {
		parts = append(parts, StyleHelpKey.Render("G")+StyleHelpBar.Render(":GO TO "+strings.ToUpper(cfg.goToLabel)+"  "))
	}
	if cfg.showBack {
		parts = append(parts, StyleHelpKey.Render("B")+StyleHelpBar.Render(":BACK  "))
	}
	if cfg.showSpec {
		parts = append(parts, StyleHelpKey.Render("P")+StyleHelpBar.Render(":SPEC  "))
	}
	parts = append(parts,
		StyleHelpKey.Render("ENTER")+StyleHelpBar.Render(":DETAIL  "),
		StyleHelpKey.Render("Q/ESC")+StyleHelpBar.Render(":QUIT"),
	)
	text := strings.Join(parts, "")
	textWidth := lipgloss.Width(text)

	if cfg.position != "" {
		posText := StyleHelpPos.Render(cfg.position)
		posWidth := lipgloss.Width(posText)
		gap := width - textWidth - posWidth
		if gap > 0 {
			text += StyleHelpBar.Render(strings.Repeat(" ", gap)) + posText
		} else {
			text += " " + posText
		}
	} else if width > textWidth {
		text += StyleHelpBar.Render(strings.Repeat(" ", width-textWidth))
	}
	return text
}

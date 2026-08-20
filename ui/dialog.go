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

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// DialogModel is a scrollable overlay dialog for displaying resource details.
type DialogModel struct {
	Title    string
	body     string
	viewport viewport.Model
	termW    int
	termH    int
	width    int
	height   int
}

// NewDialog creates a dialog with initial terminal dimensions.
func NewDialog(termW, termH int) DialogModel {
	vp := viewport.New(viewport.WithWidth(30), viewport.WithHeight(5))
	return DialogModel{
		viewport: vp,
		termW:    termW,
		termH:    termH,
	}
}

// SetContent sets the dialog title and body, auto-sizing to fit the content.
func (d *DialogModel) SetContent(title, body string) {
	d.Title = title
	d.body = body
	d.resize()
	d.viewport.SetContent(body)
	d.viewport.GotoTop()
}

// SetSize updates the terminal dimensions used for auto-sizing.
func (d *DialogModel) SetSize(termW, termH int) {
	d.termW = termW
	d.termH = termH
	if d.body != "" {
		d.resize()
		d.viewport.SetContent(d.body)
	}
}

func (d *DialogModel) resize() {
	// Measure content dimensions
	lines := strings.Split(d.body, "\n")
	contentHeight := len(lines)
	contentWidth := len(d.Title)
	for _, line := range lines {
		if len(line) > contentWidth {
			contentWidth = len(line)
		}
	}

	// dialog chrome: border (2) + padding (2) horizontally; border (2) + title (1) + sep (1) + sep (1) + footer (1) = 6 vertically
	maxW := d.termW - 6
	maxH := d.termH - 8

	w := max(min(contentWidth+2, maxW), 30)

	h := max(min(contentHeight, maxH), 3)

	d.width = w
	d.height = h
	d.viewport.SetWidth(w)
	d.viewport.SetHeight(h)
}

func (d DialogModel) Update(msg tea.Msg) (DialogModel, tea.Cmd) {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

func (d DialogModel) View() string {
	titleBar := StyleDialogTitleBar.Width(d.width).Render(d.Title)
	separator := StyleDialogSeparator.Render(strings.Repeat("─", d.width))
	content := d.viewport.View()
	footer := StyleDialogFooter.Width(d.width).Render("Esc: close │ j/k: scroll")

	inner := lipgloss.JoinVertical(lipgloss.Left, titleBar, separator, content, separator, footer)
	return StyleDialogBox.Render(inner)
}

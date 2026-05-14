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
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHelpView_ContainsKeys(t *testing.T) {
	view := HelpView(100, "sn", true, "3/47")

	keys := []string{"←/→", "Tab", "n", "s", "Enter", "Esc"}
	for _, key := range keys {
		if !strings.Contains(view, key) {
			t.Errorf("HelpView should contain %q", key)
		}
	}
	if !strings.Contains(view, "go to sn") {
		t.Error("HelpView should contain 'go to sn' when goToLabel is 'sn'")
	}
	if !strings.Contains(view, "back") {
		t.Error("HelpView should contain 'back' when showBack is true")
	}
	if !strings.Contains(view, "3/47") {
		t.Error("HelpView should contain position indicator '3/47'")
	}
}

func TestHelpView_GoToCtr(t *testing.T) {
	view := HelpView(100, "ctr", false, "")
	if !strings.Contains(view, "go to ctr") {
		t.Error("HelpView should contain 'go to ctr' when goToLabel is 'ctr'")
	}
	if strings.Contains(view, "back") {
		t.Error("HelpView should not contain 'back' when showBack is false")
	}
}

func TestHelpView_PadsToWidth(t *testing.T) {
	view := HelpView(120, "", false, "")
	viewWidth := lipgloss.Width(view)

	if viewWidth < 120 {
		t.Errorf("HelpView width = %d, should pad to at least 120", viewWidth)
	}
}

func TestHelpView_NarrowTerminal(t *testing.T) {
	view := HelpView(20, "", false, "")
	if view == "" {
		t.Error("HelpView should produce output even for narrow terminals")
	}
}

func TestHelpView_PositionRightAligned(t *testing.T) {
	view := HelpView(120, "", false, "12/99")
	viewWidth := lipgloss.Width(view)
	if viewWidth < 120 {
		t.Errorf("HelpView with position should fill to width, got %d", viewWidth)
	}
	if !strings.Contains(view, "12/99") {
		t.Error("HelpView should contain position '12/99'")
	}
}

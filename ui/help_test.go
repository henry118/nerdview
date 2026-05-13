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
	view := HelpView(100)

	keys := []string{"←/→", "Tab", "n", "s", "Enter", "Esc"}
	for _, key := range keys {
		if !strings.Contains(view, key) {
			t.Errorf("HelpView should contain %q", key)
		}
	}
}

func TestHelpView_PadsToWidth(t *testing.T) {
	view := HelpView(120)
	viewWidth := lipgloss.Width(view)

	if viewWidth < 120 {
		t.Errorf("HelpView width = %d, should pad to at least 120", viewWidth)
	}
}

func TestHelpView_NarrowTerminal(t *testing.T) {
	view := HelpView(20)
	// Should not panic or produce empty output
	if view == "" {
		t.Error("HelpView should produce output even for narrow terminals")
	}
}

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

	"charm.land/lipgloss/v2"
)

func TestHelpView_ContainsKeys(t *testing.T) {
	view := HelpView(100, WithGoTo("sn"), WithBack(), WithSnapshotter(), WithPosition("3/47"))

	keys := []string{"←/→", "TAB", "N", "S", "ENTER", "ESC"}
	for _, key := range keys {
		if !strings.Contains(view, key) {
			t.Errorf("HelpView should contain %q", key)
		}
	}
	if !strings.Contains(view, "GO TO SN") {
		t.Error("HelpView should contain 'GO TO SN' with WithGoTo('sn')")
	}
	if !strings.Contains(view, "BACK") {
		t.Error("HelpView should contain 'BACK' with WithBack()")
	}
	if !strings.Contains(view, "3/47") {
		t.Error("HelpView should contain position indicator '3/47'")
	}
}

func TestHelpView_GoToCtr(t *testing.T) {
	view := HelpView(100, WithGoTo("ctr"))
	if !strings.Contains(view, "GO TO CTR") {
		t.Error("HelpView should contain 'GO TO CTR' with WithGoTo('ctr')")
	}
	if strings.Contains(view, "BACK") {
		t.Error("HelpView should not contain 'BACK' without WithBack()")
	}
}

func TestHelpView_ShowSpec(t *testing.T) {
	view := HelpView(100, WithSpec())
	if !strings.Contains(view, "SPEC") {
		t.Error("HelpView should contain 'SPEC' with WithSpec()")
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
	if view == "" {
		t.Error("HelpView should produce output even for narrow terminals")
	}
}

func TestHelpView_PositionRightAligned(t *testing.T) {
	view := HelpView(120, WithPosition("12/99"))
	viewWidth := lipgloss.Width(view)
	if viewWidth < 120 {
		t.Errorf("HelpView with position should fill to width, got %d", viewWidth)
	}
	if !strings.Contains(view, "12/99") {
		t.Error("HelpView should contain position '12/99'")
	}
}

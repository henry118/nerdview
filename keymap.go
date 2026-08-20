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

import "charm.land/bubbles/v2/key"

type keyMap struct {
	Up                key.Binding
	Down              key.Binding
	Right             key.Binding
	Left              key.Binding
	NextResource      key.Binding
	PrevResource      key.Binding
	SelectNS          key.Binding
	SelectSnapshotter key.Binding
	GoTo              key.Binding
	GoBack            key.Binding
	Spec              key.Binding
	Enter             key.Binding
	Escape            key.Binding
	Quit              key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "unfold"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "fold"),
	),
	NextResource: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("Tab", "next tab"),
	),
	PrevResource: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("S-Tab", "prev tab"),
	),
	SelectNS: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "namespace"),
	),
	SelectSnapshotter: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "snapshotter"),
	),
	GoTo: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "go to snapshot"),
	),
	GoBack: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "go back"),
	),
	Spec: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "runtime spec"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "detail"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "back/quit"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("q/Ctrl+C", "quit"),
	),
}

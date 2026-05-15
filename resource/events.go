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

package resource

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

type Event struct {
	Timestamp time.Time
	Namespace string
	Topic     string
	Payload   any
}

type eventKind struct{}

var EventKind Kind = eventKind{}

func (eventKind) Name() string { return "Events" }

func (eventKind) Columns() []Column {
	return []Column{
		{Title: "Time", MinWidth: 14},
		{Title: "Namespace", MinWidth: 14},
		{Title: "Topic", MinWidth: 20, Flex: true},
	}
}

func (eventKind) Rows(data any, _ map[string]bool) []table.Row {
	evts, ok := data.([]Event)
	if !ok || len(evts) == 0 {
		return nil
	}
	rows := make([]table.Row, len(evts))
	for i, e := range evts {
		rows[i] = table.Row{
			e.Timestamp.Format("15:04:05.000"),
			e.Namespace,
			e.Topic,
		}
	}
	return rows
}

func (eventKind) FoldKey(_ any, _ map[string]bool, _ int) string {
	return ""
}

func (eventKind) InitFolded(_ any) map[string]bool {
	return nil
}

func (eventKind) Detail(data any, _ map[string]bool, index int) (string, string) {
	evts, ok := data.([]Event)
	if !ok || index < 0 || index >= len(evts) {
		return "", ""
	}
	e := evts[index]
	var b strings.Builder
	fmt.Fprintf(&b, "Timestamp:  %s\n", e.Timestamp.Format(time.RFC3339Nano))
	fmt.Fprintf(&b, "Namespace:  %s\n", e.Namespace)
	fmt.Fprintf(&b, "Topic:      %s\n", e.Topic)
	if e.Payload != nil {
		fmt.Fprintf(&b, "\n--- Payload ---\n")
		data, err := json.MarshalIndent(e.Payload, "", "  ")
		if err == nil {
			fmt.Fprintf(&b, "%s\n", data)
		} else {
			fmt.Fprintf(&b, "%+v\n", e.Payload)
		}
	}
	return e.Topic, b.String()
}

func (eventKind) CrossRefs(_ any, _ map[string]bool) []string {
	return nil
}

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

package ctr

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/containerd/containerd/v2/core/events"
	"github.com/henry118/nerdtui/logging"
)

// EventMsg is a Bubble Tea message carrying a single containerd event.
type EventMsg struct {
	Namespace string
	Topic     string
	Timestamp time.Time
}

// EventErrMsg is sent when the event subscription encounters an error.
type EventErrMsg struct {
	Err error
}

// WaitForEvent returns a Bubble Tea command that blocks until the next
// containerd event arrives, then delivers it as an EventMsg.
func WaitForEvent(c *Client) tea.Cmd {
	return func() tea.Msg {
		select {
		case env, ok := <-c.EventCh():
			if !ok {
				logging.Error("event channel closed")
				return EventErrMsg{Err: fmt.Errorf("event channel closed")}
			}
			logging.Debug("event received: ns=%s topic=%s", env.Namespace, env.Topic)
			return eventFromEnvelope(env)
		case err, ok := <-c.ErrCh():
			if !ok {
				logging.Error("error channel closed")
				return EventErrMsg{Err: fmt.Errorf("error channel closed")}
			}
			if err != nil {
				logging.Error("event stream error: %v", err)
				return EventErrMsg{Err: err}
			}
			return nil
		}
	}
}

func eventFromEnvelope(env *events.Envelope) EventMsg {
	return EventMsg{
		Namespace: env.Namespace,
		Topic:     env.Topic,
		Timestamp: env.Timestamp,
	}
}
